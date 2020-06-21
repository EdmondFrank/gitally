package log

import (
	"context"
	"io/ioutil"

	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/catfile"
	"gitlab.com/gitlab-org/gitaly/internal/helper/text"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

// LastCommitForPath returns the last commit which modified path.
func LastCommitForPath(ctx context.Context, batch *catfile.Batch, repo *gitalypb.Repository, revision string, path string, literalPathspec bool) (*gitalypb.GitCommit, error) {
	var globals []git.Option
	if literalPathspec {
		globals = []git.Option{git.Flag{"--literal-pathspecs"}}
	}

	cmd, err := git.SafeCmd(ctx, repo, globals, git.SubCmd{
		Name:        "log",
		Flags:       []git.Option{git.Flag{"--format=%H"}, git.Flag{"--max-count=1"}},
		Args:        []string{revision},
		PostSepArgs: []string{path}})
	if err != nil {
		return nil, err
	}

	commitID, err := ioutil.ReadAll(cmd)
	if err != nil {
		return nil, err
	}

	return GetCommitCatfile(batch, text.ChompBytes(commitID))
}

// GitLogCommand returns a Command that executes git log with the given the arguments
func GitLogCommand(ctx context.Context, repo *gitalypb.Repository, revisions []string, paths []string, literalPathspecs bool, extraArgs ...git.Option) (*command.Command, error) {
	var globals []git.Option
	if literalPathspecs {
		globals = []git.Option{git.Flag{"--literal-pathspecs"}}
	}

	return git.SafeCmd(ctx, repo, globals, git.SubCmd{
		Name:        "log",
		Flags:       append([]git.Option{git.Flag{"--pretty=%H"}}, extraArgs...),
		Args:        revisions,
		PostSepArgs: paths,
	})
}
