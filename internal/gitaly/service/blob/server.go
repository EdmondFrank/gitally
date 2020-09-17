package blob

import (
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

type server struct {
	ruby *rubyserver.Server
}

// NewServer creates a new instance of a grpc BlobServer
func NewServer(rs *rubyserver.Server) gitalypb.BlobServiceServer {
	return &server{ruby: rs}
}
