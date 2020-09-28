package ref

import (
	"gitlab.com/gitlab-org/gitaly/internal/storage"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

type server struct {
	locator storage.Locator
}

// NewServer creates a new instance of a grpc RefServer
func NewServer(locator storage.Locator) gitalypb.RefServiceServer {
	return &server{locator: locator}
}
