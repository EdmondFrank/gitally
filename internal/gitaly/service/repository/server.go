package repository

import (
	"context"

	"gitlab.com/gitlab-org/gitaly/client"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/storage"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

type server struct {
	ruby                 *rubyserver.Server
	conns                *client.Pool
	internalGitalySocket string
	locator              storage.Locator
}

// NewServer creates a new instance of a gRPC repo server
func NewServer(rs *rubyserver.Server, locator storage.Locator, internalGitalySocket string) gitalypb.RepositoryServiceServer {
	return &server{
		ruby:                 rs,
		locator:              locator,
		conns:                client.NewPool(),
		internalGitalySocket: internalGitalySocket,
	}
}

func (*server) FetchHTTPRemote(context.Context, *gitalypb.FetchHTTPRemoteRequest) (*gitalypb.FetchHTTPRemoteResponse, error) {
	return nil, helper.Unimplemented
}