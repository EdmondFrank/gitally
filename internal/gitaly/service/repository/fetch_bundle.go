package repository

import (
	"io"
	"os"
	"path/filepath"

	gitalyerrors "gitlab.com/gitlab-org/gitaly/v14/internal/errors"
	"gitlab.com/gitlab-org/gitaly/v14/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/tempdir"
	"gitlab.com/gitlab-org/gitaly/v14/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/v14/streamio"
)

func (s *server) FetchBundle(stream gitalypb.RepositoryService_FetchBundleServer) error {
	firstRequest, err := stream.Recv()
	if err != nil {
		return helper.ErrInternalf("first request: %v", err)
	}

	if firstRequest.GetRepository() == nil {
		return helper.ErrInvalidArgument(gitalyerrors.ErrEmptyRepository)
	}

	firstRead := true
	reader := streamio.NewReader(func() ([]byte, error) {
		if firstRead {
			firstRead = false
			return firstRequest.GetData(), nil
		}

		request, err := stream.Recv()
		return request.GetData(), err
	})

	ctx := stream.Context()

	tmpDir, err := tempdir.New(ctx, firstRequest.GetRepository().GetStorageName(), s.locator)
	if err != nil {
		return helper.ErrInternal(err)
	}

	bundlePath := filepath.Join(tmpDir.Path(), "repo.bundle")
	file, err := os.Create(bundlePath)
	if err != nil {
		return helper.ErrInternal(err)
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		return helper.ErrInternal(err)
	}

	if _, err := s.FetchRemote(ctx, &gitalypb.FetchRemoteRequest{
		Repository: firstRequest.GetRepository(),
		RemoteParams: &gitalypb.Remote{
			Url: bundlePath,
		},
	}); err != nil {
		return helper.ErrInternal(err)
	}

	return stream.SendAndClose(&gitalypb.FetchBundleResponse{})
}
