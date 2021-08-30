package repository

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v14/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v14/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testserver"
	"gitlab.com/gitlab-org/gitaly/v14/internal/transaction/txinfo"
	"gitlab.com/gitlab-org/gitaly/v14/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/v14/streamio"
	"google.golang.org/grpc"
)

func TestServer_FetchBundle_success(t *testing.T) {
	t.Parallel()
	cfg, _, repoPath, client := setupRepositoryService(t)

	tmp := testhelper.TempDir(t)
	bundlePath := filepath.Join(tmp, "test.bundle")

	gittest.Exec(t, cfg, "-C", repoPath, "bundle", "create", bundlePath, "--all")
	expectedRefs := gittest.Exec(t, cfg, "-C", repoPath, "show-ref", "--head")

	targetRepo, targetRepoPath := gittest.InitRepo(t, cfg, cfg.Storages[0])

	ctx, cancel := testhelper.Context()
	defer cancel()

	stream, err := client.FetchBundle(ctx)
	require.NoError(t, err)

	request := &gitalypb.FetchBundleRequest{Repository: targetRepo}
	writer := streamio.NewWriter(func(p []byte) error {
		request.Data = p

		if err := stream.Send(request); err != nil {
			return err
		}

		request = &gitalypb.FetchBundleRequest{}

		return nil
	})

	bundle, err := os.Open(bundlePath)
	require.NoError(t, err)
	defer testhelper.MustClose(t, bundle)

	_, err = io.Copy(writer, bundle)
	require.NoError(t, err)

	_, err = stream.CloseAndRecv()
	require.NoError(t, err)

	refs := gittest.Exec(t, cfg, "-C", targetRepoPath, "show-ref", "--head")
	require.Equal(t, string(expectedRefs), string(refs))
}

func TestFetchBundle_transaction(t *testing.T) {
	t.Parallel()
	sourceCfg, _, sourceRepoPath := testcfg.BuildWithRepo(t)

	txManager := &mockTxManager{}
	addr := testserver.RunGitalyServer(t, sourceCfg, nil, func(srv *grpc.Server, deps *service.Dependencies) {
		gitalypb.RegisterRepositoryServiceServer(srv, NewServer(
			deps.GetCfg(),
			deps.GetRubyServer(),
			deps.GetLocator(),
			deps.GetTxManager(),
			deps.GetGitCmdFactory(),
			deps.GetCatfileCache(),
		))
	}, testserver.WithTransactionManager(txManager))

	client := newRepositoryClient(t, sourceCfg, addr)

	tmp := testhelper.TempDir(t)
	bundlePath := filepath.Join(tmp, "test.bundle")
	gittest.Exec(t, sourceCfg, "-C", sourceRepoPath, "bundle", "create", bundlePath, "--all")

	targetCfg, targetRepoProto, targetRepoPath := testcfg.BuildWithRepo(t)
	_, stopGitServer := gittest.GitServer(t, targetCfg, targetRepoPath, nil)
	defer func() { require.NoError(t, stopGitServer()) }()

	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, err := txinfo.InjectTransaction(ctx, 1, "node", true)
	require.NoError(t, err)
	ctx = helper.IncomingToOutgoing(ctx)

	require.Equal(t, 0, txManager.votes)

	stream, err := client.FetchBundle(ctx)
	require.NoError(t, err)

	request := &gitalypb.FetchBundleRequest{Repository: targetRepoProto}
	writer := streamio.NewWriter(func(p []byte) error {
		request.Data = p

		if err := stream.Send(request); err != nil {
			return err
		}

		request = &gitalypb.FetchBundleRequest{}

		return nil
	})

	bundle, err := os.Open(bundlePath)
	require.NoError(t, err)
	defer testhelper.MustClose(t, bundle)

	_, err = io.Copy(writer, bundle)
	require.NoError(t, err)

	_, err = stream.CloseAndRecv()
	require.NoError(t, err)

	require.Equal(t, 1, txManager.votes)
}
