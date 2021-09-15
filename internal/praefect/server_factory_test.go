package praefect

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v14/client"
	"gitlab.com/gitlab-org/gitaly/v14/internal/backchannel"
	"gitlab.com/gitlab-org/gitaly/v14/internal/bootstrap/starter"
	"gitlab.com/gitlab-org/gitaly/v14/internal/git/gittest"
	gconfig "gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/service/setup"
	"gitlab.com/gitlab-org/gitaly/v14/internal/helper/text"
	"gitlab.com/gitlab-org/gitaly/v14/internal/listenmux"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/datastore/glsql"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/nodes"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/protoregistry"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/service/transaction"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/transactions"
	"gitlab.com/gitlab-org/gitaly/v14/internal/sidechannel"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/promtest"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testserver"
	"gitlab.com/gitlab-org/gitaly/v14/proto/go/gitalypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestServerFactory(t *testing.T) {
	t.Parallel()
	cfg, repo, repoPath := testcfg.BuildWithRepo(t)
	register := func(srv *grpc.Server, deps *service.Dependencies) {
		setup.RegisterAll(srv, deps)
		gitalypb.RegisterTestServiceServer(srv, &testSidechannelService{})
	}
	gitalyAddr := testserver.RunGitalyServer(t, cfg, nil, register, testserver.WithDisablePraefect())

	certFile, keyFile := testhelper.GenerateCerts(t)

	conf := config.Config{
		TLS: gconfig.TLS{
			CertPath: certFile,
			KeyPath:  keyFile,
		},
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "praefect",
				Nodes: []*config.Node{
					{
						Storage: cfg.Storages[0].Name,
						Address: gitalyAddr,
						Token:   cfg.Auth.Token,
					},
				},
			},
		},
		Failover: config.Failover{Enabled: true},
	}

	repo.StorageName = conf.VirtualStorages[0].Name // storage must be re-written to virtual to be properly redirected by praefect
	revision := text.ChompBytes(gittest.Exec(t, cfg, "-C", repoPath, "rev-parse", "HEAD"))

	logger := testhelper.DiscardTestEntry(t)
	queue := datastore.NewPostgresReplicationEventQueue(glsql.NewDB(t))

	rs := datastore.MockRepositoryStore{}
	txMgr := transactions.NewManager(conf)
	sidechannelRegistry := sidechannel.NewRegistry(logger)
	clientHandshaker := backchannel.NewClientHandshaker(logger, NewBackchannelServerFactory(logger, transaction.NewServer(txMgr), sidechannelRegistry))
	nodeMgr, err := nodes.NewManager(logger, conf, nil, rs, &promtest.MockHistogramVec{}, protoregistry.GitalyProtoPreregistered, nil, clientHandshaker, sidechannelRegistry)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Second)
	registry := protoregistry.GitalyProtoPreregistered

	coordinator := NewCoordinator(
		queue,
		rs,
		NewNodeManagerRouter(nodeMgr, rs),
		txMgr,
		conf,
		registry,
	)

	checkOwnRegisteredServices := func(ctx context.Context, t *testing.T, cc *grpc.ClientConn) healthpb.HealthClient {
		t.Helper()

		healthClient := healthpb.NewHealthClient(cc)
		resp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{})
		require.NoError(t, err)
		require.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.Status)
		return healthClient
	}

	checkProxyingOntoGitaly := func(ctx context.Context, t *testing.T, cc *grpc.ClientConn) {
		t.Helper()

		commitClient := gitalypb.NewCommitServiceClient(cc)
		resp, err := commitClient.FindCommit(ctx, &gitalypb.FindCommitRequest{
			Repository: repo,
			Revision:   []byte(revision),
		})
		require.NoError(t, err)
		require.Equal(t, revision, resp.Commit.Id)
	}

	checkSidechannelGitaly := func(ctx context.Context, t *testing.T, addr string, creds credentials.TransportCredentials) {
		t.Helper()
		factory := func() backchannel.Server {
			lm := listenmux.New(insecure.NewCredentials())
			lm.Register(sidechannel.NewServerHandshaker(sidechannelRegistry))
			return grpc.NewServer(grpc.Creds(lm))
		}

		clientHandshaker := backchannel.NewClientHandshaker(logger, factory)
		dialOpt := grpc.WithTransportCredentials(clientHandshaker.ClientHandshake(creds))

		conn, err := grpc.Dial(addr, dialOpt)
		require.NoError(t, err)
		t.Cleanup(func() { conn.Close() })

		ctx, waiter := sidechannel.RegisterSidechannel(ctx, sidechannelRegistry, func(conn net.Conn) error {
			const message = "hello world01234"
			if len(message) != sidechannelTestMessageSize {
				return errors.New("test setup error: incorrect message size")
			}
			if _, err := io.WriteString(conn, message); err != nil {
				return err
			}
			buf := make([]byte, len(message))
			if _, err := io.ReadFull(conn, buf); err != nil {
				return err
			}
			if string(buf) != message {
				return fmt.Errorf("unexpected response: %q", buf)
			}
			return nil
		})
		defer waiter.Close()

		_, err = gitalypb.NewTestServiceClient(conn).TestSidechannel(ctx,
			&gitalypb.TestSidechannelRequest{Repository: repo},
		)
		require.NoError(t, err)
		require.NoError(t, waiter.Wait())
	}

	t.Run("insecure", func(t *testing.T) {
		praefectServerFactory := NewServerFactory(conf, logger, coordinator.StreamDirector, nodeMgr, txMgr, queue, rs, datastore.AssignmentStore{}, registry, nil, nil)
		defer praefectServerFactory.Stop()

		listener, err := net.Listen(starter.TCP, "localhost:0")
		require.NoError(t, err)
		defer func() { require.NoError(t, listener.Close()) }()

		go praefectServerFactory.Serve(listener, false)

		praefectAddr, err := starter.ComposeEndpoint(listener.Addr().Network(), listener.Addr().String())
		require.NoError(t, err)

		creds := insecure.NewCredentials()

		cc, err := client.Dial(praefectAddr, nil)
		require.NoError(t, err)
		defer func() { require.NoError(t, cc.Close()) }()

		ctx, cancel := testhelper.Context()
		defer cancel()

		t.Run("handles registered RPCs", func(t *testing.T) {
			checkOwnRegisteredServices(ctx, t, cc)
		})

		t.Run("proxies RPCs onto gitaly server", func(t *testing.T) {
			checkProxyingOntoGitaly(ctx, t, cc)
		})

		t.Run("proxies sidechannel RPCs onto gitaly server", func(t *testing.T) {
			checkSidechannelGitaly(ctx, t, listener.Addr().String(), creds)
		})
	})

	t.Run("secure", func(t *testing.T) {
		praefectServerFactory := NewServerFactory(conf, logger, coordinator.StreamDirector, nodeMgr, txMgr, queue, rs, datastore.AssignmentStore{}, registry, nil, nil)
		defer praefectServerFactory.Stop()

		listener, err := net.Listen(starter.TCP, "localhost:0")
		require.NoError(t, err)
		defer func() { require.NoError(t, listener.Close()) }()

		go praefectServerFactory.Serve(listener, true)

		ctx, cancel := testhelper.Context()
		defer cancel()

		certPool, err := x509.SystemCertPool()
		require.NoError(t, err)

		pem := testhelper.MustReadFile(t, conf.TLS.CertPath)

		require.True(t, certPool.AppendCertsFromPEM(pem))

		creds := credentials.NewTLS(&tls.Config{
			RootCAs:    certPool,
			MinVersion: tls.VersionTLS12,
		})

		cc, err := grpc.DialContext(ctx, listener.Addr().String(), grpc.WithTransportCredentials(creds))
		require.NoError(t, err)
		defer func() { require.NoError(t, cc.Close()) }()

		t.Run("handles registered RPCs", func(t *testing.T) {
			checkOwnRegisteredServices(ctx, t, cc)
		})

		t.Run("proxies RPCs onto gitaly server", func(t *testing.T) {
			checkProxyingOntoGitaly(ctx, t, cc)
		})

		t.Run("proxies sidechannel RPCs onto gitaly server", func(t *testing.T) {
			checkSidechannelGitaly(ctx, t, listener.Addr().String(), creds)
		})
	})

	t.Run("stops all listening servers", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		praefectServerFactory := NewServerFactory(conf, logger, coordinator.StreamDirector, nodeMgr, txMgr, queue, rs, datastore.AssignmentStore{}, registry, nil, nil)
		defer praefectServerFactory.Stop()

		// start with tcp address
		tcpListener, err := net.Listen(starter.TCP, "localhost:0")
		require.NoError(t, err)
		defer tcpListener.Close()

		go praefectServerFactory.Serve(tcpListener, false)

		praefectTCPAddr, err := starter.ComposeEndpoint(tcpListener.Addr().Network(), tcpListener.Addr().String())
		require.NoError(t, err)

		tcpCC, err := client.Dial(praefectTCPAddr, nil)
		require.NoError(t, err)
		defer func() { require.NoError(t, tcpCC.Close()) }()

		tcpHealthClient := checkOwnRegisteredServices(ctx, t, tcpCC)

		// start with tls address
		tlsListener, err := net.Listen(starter.TCP, "localhost:0")
		require.NoError(t, err)
		defer tlsListener.Close()

		go praefectServerFactory.Serve(tlsListener, true)

		praefectTLSAddr, err := starter.ComposeEndpoint(tcpListener.Addr().Network(), tcpListener.Addr().String())
		require.NoError(t, err)

		tlsCC, err := client.Dial(praefectTLSAddr, nil)
		require.NoError(t, err)
		defer func() { require.NoError(t, tlsCC.Close()) }()

		tlsHealthClient := checkOwnRegisteredServices(ctx, t, tlsCC)

		// start with socket address
		socketPath := testhelper.GetTemporaryGitalySocketFileName(t)
		defer func() { require.NoError(t, os.RemoveAll(socketPath)) }()
		socketListener, err := net.Listen(starter.Unix, socketPath)
		require.NoError(t, err)
		defer socketListener.Close()

		go praefectServerFactory.Serve(socketListener, false)

		praefectSocketAddr, err := starter.ComposeEndpoint(socketListener.Addr().Network(), socketListener.Addr().String())
		require.NoError(t, err)

		socketCC, err := client.Dial(praefectSocketAddr, nil)
		require.NoError(t, err)
		defer func() { require.NoError(t, socketCC.Close()) }()

		unixHealthClient := checkOwnRegisteredServices(ctx, t, socketCC)

		praefectServerFactory.GracefulStop()

		_, err = tcpHealthClient.Check(ctx, nil)
		require.Error(t, err)

		_, err = tlsHealthClient.Check(ctx, nil)
		require.Error(t, err)

		_, err = unixHealthClient.Check(ctx, nil)
		require.Error(t, err)
	})

	t.Run("tls key path invalid", func(t *testing.T) {
		badTLSKeyPath := conf
		badTLSKeyPath.TLS.KeyPath = "invalid"
		praefectServerFactory := NewServerFactory(badTLSKeyPath, logger, coordinator.StreamDirector, nodeMgr, txMgr, queue, rs, datastore.AssignmentStore{}, registry, nil, nil)

		err := praefectServerFactory.Serve(nil, true)
		require.EqualError(t, err, "load certificate key pair: open invalid: no such file or directory")
	})

	t.Run("tls cert path invalid", func(t *testing.T) {
		badTLSKeyPath := conf
		badTLSKeyPath.TLS.CertPath = "invalid"
		praefectServerFactory := NewServerFactory(badTLSKeyPath, logger, coordinator.StreamDirector, nodeMgr, txMgr, queue, rs, datastore.AssignmentStore{}, registry, nil, nil)

		err := praefectServerFactory.Serve(nil, true)
		require.EqualError(t, err, "load certificate key pair: open invalid: no such file or directory")
	})
}

const sidechannelTestMessageSize = 16

type testSidechannelService struct {
	gitalypb.UnimplementedTestServiceServer
}

func (testSidechannelService) TestSidechannel(ctx context.Context, _ *gitalypb.TestSidechannelRequest) (*gitalypb.TestSidechannelResponse, error) {
	conn, err := sidechannel.OpenSidechannel(ctx)
	if err != nil {
		return nil, fmt.Errorf("upstream handler: %w", err)
	}
	defer conn.Close()

	if _, err := io.CopyN(conn, conn, sidechannelTestMessageSize); err != nil {
		return nil, fmt.Errorf("upstream handler: %w", err)
	}

	return &gitalypb.TestSidechannelResponse{}, nil
}
