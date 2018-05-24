package repository

import (
	"fmt"
	"strings"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"
	"gitlab.com/gitlab-org/gitaly/internal/helper"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	squashWorktreePrefix = "squash"
)

func (s *server) IsSquashInProgress(ctx context.Context, req *pb.IsSquashInProgressRequest) (*pb.IsSquashInProgressResponse, error) {
	if err := validateIsSquashInProgressRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "IsSquashInProgress: %v", err)
	}

	repoPath, err := helper.GetRepoPath(req.GetRepository())
	if err != nil {
		return nil, err
	}

	inProg, err := freshWorktree(repoPath, squashWorktreePrefix, req.GetSquashId())
	if err != nil {
		return nil, err
	}
	return &pb.IsSquashInProgressResponse{InProgress: inProg}, nil
}

func validateIsSquashInProgressRequest(req *pb.IsSquashInProgressRequest) error {
	if req.GetRepository() == nil {
		return fmt.Errorf("empty Repository")
	}

	if req.GetSquashId() == "" {
		return fmt.Errorf("empty SquashId")
	}

	if strings.Contains(req.GetSquashId(), "/") {
		return fmt.Errorf("SquashId contains '/'")
	}

	return nil
}
