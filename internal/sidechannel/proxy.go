package sidechannel

import (
	"context"
	"io"
	"net"

	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/v14/internal/helper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewUnaryProxy creates a gRPC client middleware that proxies sidechannels.
func (s *Registry) NewUnaryProxy() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !hasSidechannelMetadata(ctx) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		ctx, waiter := RegisterSidechannel(ctx, s, proxy(ctx))
		defer logWaiterClose(s.logger, waiter)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// NewStreamProxy creates a gRPC client middleware that proxies sidechannels.
func (s *Registry) NewStreamProxy() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if !hasSidechannelMetadata(ctx) {
			return streamer(ctx, desc, cc, method, opts...)
		}

		ctx, waiter := RegisterSidechannel(ctx, s, proxy(ctx))
		go func() {
			<-ctx.Done()
			logWaiterClose(s.logger, waiter)
		}()

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func hasSidechannelMetadata(ctx context.Context) bool {
	md, ok := metadata.FromOutgoingContext(ctx)
	return ok && len(md.Get(sidechannelMetadataKey)) > 0
}

func proxy(ctx context.Context) func(net.Conn) error {
	return func(upstream net.Conn) error {
		downstream, err := OpenSidechannel(helper.OutgoingToIncoming(ctx))
		if err != nil {
			return err
		}
		defer downstream.Close()

		errC := make(chan error, 2)
		go func() {
			// TODO use CloseWrite()
			// https://gitlab.com/gitlab-com/gl-infra/scalability/-/issues/1278
			_, err := io.Copy(upstream, downstream)
			errC <- err
		}()
		go func() {
			_, err := io.Copy(downstream, upstream)
			errC <- err
		}()

		return <-errC
	}
}

func logWaiterClose(logger *logrus.Entry, waiter *Waiter) {
	if err := waiter.Close(); err != nil {
		logger.WithError(err).Error("proxy sidechannel")
	}
}
