package repository

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"google.golang.org/grpc/codes"
)

func TestCreateRepositorySuccess(t *testing.T) {
	server, serverSocketPath := runRepoServer(t)
	defer server.Stop()

	client, conn := newRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	ctx, cancel := testhelper.Context()
	defer cancel()

	storageDir, err := helper.GetStorageByName("default")
	require.NoError(t, err)
	relativePath := "create-repository-test.git"
	repoDir := path.Join(storageDir, relativePath)
	require.NoError(t, os.RemoveAll(repoDir))

	repo := &gitalypb.Repository{StorageName: "default", RelativePath: relativePath}
	req := &gitalypb.CreateRepositoryRequest{Repository: repo}
	_, err = client.CreateRepository(ctx, req)
	require.NoError(t, err)

	for _, dir := range []string{repoDir, path.Join(repoDir, "refs")} {
		fi, err := os.Stat(dir)
		require.NoError(t, err)
		require.True(t, fi.IsDir(), "%q must be a directory", fi.Name())
	}
}

func TestCreateRepositoryFailure(t *testing.T) {
	server, serverSocketPath := runRepoServer(t)
	defer server.Stop()

	client, conn := newRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	ctx, cancel := testhelper.Context()
	defer cancel()

	storagePath, err := helper.GetStorageByName("default")
	require.NoError(t, err)
	fullPath := path.Join(storagePath, "foo.git")

	_, err = os.Create(fullPath)
	require.NoError(t, err)
	defer os.RemoveAll(fullPath)

	_, err = client.CreateRepository(ctx, &gitalypb.CreateRepositoryRequest{
		Repository: &gitalypb.Repository{StorageName: "default", RelativePath: "foo.git"},
	})

	testhelper.RequireGrpcError(t, err, codes.Internal)
}

func TestCreateRepositoryFailureInvalidArgs(t *testing.T) {
	server, serverSocketPath := runRepoServer(t)
	defer server.Stop()

	client, conn := newRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	ctx, cancel := testhelper.Context()
	defer cancel()

	testCases := []struct {
		repo *gitalypb.Repository
		code codes.Code
	}{
		{
			repo: &gitalypb.Repository{StorageName: "does not exist", RelativePath: "foobar.git"},
			code: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%+v", tc.repo), func(t *testing.T) {
			_, err := client.CreateRepository(ctx, &gitalypb.CreateRepositoryRequest{Repository: tc.repo})

			require.Error(t, err)
			testhelper.RequireGrpcError(t, err, tc.code)
		})
	}
}
