package operations

import (
	"context"

	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

func (s *Server) UserDeleteTag(ctx context.Context, req *gitalypb.UserDeleteTagRequest) (*gitalypb.UserDeleteTagResponse, error) {
	client, err := s.ruby.OperationServiceClient(ctx)
	if err != nil {
		return nil, err
	}

	clientCtx, err := rubyserver.SetHeaders(ctx, s.locator, req.GetRepository())
	if err != nil {
		return nil, err
	}

	return client.UserDeleteTag(clientCtx, req)
}

func (s *Server) UserCreateTag(ctx context.Context, req *gitalypb.UserCreateTagRequest) (*gitalypb.UserCreateTagResponse, error) {
	client, err := s.ruby.OperationServiceClient(ctx)
	if err != nil {
		return nil, err
	}

	clientCtx, err := rubyserver.SetHeaders(ctx, s.locator, req.GetRepository())
	if err != nil {
		return nil, err
	}

	return client.UserCreateTag(clientCtx, req)
}
