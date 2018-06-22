package log

import (
	"context"
	"io/ioutil"
	"strings"

	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	log "github.com/sirupsen/logrus"
)

var commitLogFormatFields = []string{
	"%H",  // commit hash
	"%an", // author name
	"%ae", // author email
	"%aI", // author date, strict ISO 8601 format
	"%cn", // committer name
	"%ce", // committer email
	"%cI", // committer date, strict ISO 8601 format
	"%P",  // parent hashes
}

const fieldDelimiterGitFormatString = "%x1f"

// LastCommitForPath returns the last commit which modified path.
func LastCommitForPath(ctx context.Context, repo *pb.Repository, revision string, path string) (*pb.GitCommit, error) {
	cmd, err := git.Command(ctx, repo, "log", "--format=%H", "--max-count=1", revision, "--", path)
	if err != nil {
		return nil, err
	}

	commitId, err := ioutil.ReadAll(cmd)
	if err != nil {
		return nil, err
	}

	return GetCommit(ctx, repo, strings.TrimSpace(string(commitId)))
}

// GitLogCommand returns a Command that executes git log with the given the arguments
func GitLogCommand(ctx context.Context, repo *pb.Repository, revisions []string, paths []string, extraArgs ...string) (*command.Command, error) {
	grpc_logrus.Extract(ctx).WithFields(log.Fields{
		"Revisions": revisions,
	}).Debug("GitLog")

	formatFlag := "--pretty=format:" + strings.Join(commitLogFormatFields, fieldDelimiterGitFormatString)

	args := []string{
		"log",
		"-z", // use 0x00 as the entry terminator (instead of \n)
		formatFlag,
	}
	args = append(args, extraArgs...)
	args = append(args, revisions...)
	args = append(args, "--")
	args = append(args, paths...)

	cmd, err := git.Command(ctx, repo, args...)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
