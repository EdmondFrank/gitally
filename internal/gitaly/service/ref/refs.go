package ref

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"gitlab.com/gitlab-org/gitaly/v14/internal/git"
	"gitlab.com/gitlab-org/gitaly/v14/internal/git/catfile"
	"gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v14/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/helper/lines"
	"gitlab.com/gitlab-org/gitaly/v14/proto/go/gitalypb"
)

const (
	tagFormat = "%(objectname) %(objecttype) %(refname:lstrip=2)"
)

var (
	// We declare the following functions in variables so that we can override them in our tests
	headReference = _headReference
	// FindBranchNames is exported to be used in other packages
	FindBranchNames = _findBranchNames
)

type findRefsOpts struct {
	cmdArgs []git.Option
	delim   byte
	lines.SenderOpts
}

func (s *server) findRefs(ctx context.Context, writer lines.Sender, repo git.RepositoryExecutor, patterns []string, opts *findRefsOpts) error {
	var options []git.Option

	if len(opts.cmdArgs) == 0 {
		options = append(options, git.Flag{Name: "--format=%(refname)"}) // Default format
	} else {
		options = append(options, opts.cmdArgs...)
	}

	cmd, err := repo.Exec(ctx, git.SubCmd{
		Name:  "for-each-ref",
		Flags: options,
		Args:  patterns,
	})
	if err != nil {
		return err
	}

	if err := lines.Send(cmd, writer, lines.SenderOpts{
		IsPageToken: opts.IsPageToken,
		Delimiter:   opts.delim,
		Limit:       opts.Limit,
	}); err != nil {
		return err
	}

	return cmd.Wait()
}

func _findBranchNames(ctx context.Context, repo git.RepositoryExecutor) ([][]byte, error) {
	var names [][]byte

	cmd, err := repo.Exec(ctx, git.SubCmd{
		Name:  "for-each-ref",
		Flags: []git.Option{git.Flag{Name: "--format=%(refname)"}},
		Args:  []string{"refs/heads"},
	},
	)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(cmd)
	for scanner.Scan() {
		names = lines.CopyAndAppend(names, scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading standard input: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return names, nil
}

func _headReference(ctx context.Context, repo git.RepositoryExecutor) ([]byte, error) {
	var headRef []byte

	cmd, err := repo.Exec(ctx, git.SubCmd{
		Name:  "rev-parse",
		Flags: []git.Option{git.Flag{Name: "--symbolic-full-name"}},
		Args:  []string{"HEAD"},
	})
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(cmd)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	headRef = scanner.Bytes()

	if err := cmd.Wait(); err != nil {
		// If the ref pointed at by HEAD doesn't exist, the rev-parse fails
		// returning the string `"HEAD"`, so we return `nil` without error.
		if bytes.Equal(headRef, []byte("HEAD")) {
			return nil, nil
		}

		return nil, err
	}

	return headRef, nil
}

// SetDefaultBranchRef overwrites the default branch ref for the repository
func SetDefaultBranchRef(ctx context.Context, repo git.RepositoryExecutor, ref string, cfg config.Cfg) error {
	if err := repo.ExecAndWait(ctx, git.SubCmd{
		Name: "symbolic-ref",
		Args: []string{"HEAD", ref},
	}, git.WithRefTxHook(ctx, repo, cfg)); err != nil {
		return err
	}
	return nil
}

// DefaultBranchName looks up the name of the default branch given a repoPath
func DefaultBranchName(ctx context.Context, repo git.RepositoryExecutor) ([]byte, error) {
	branches, err := FindBranchNames(ctx, repo)
	if err != nil {
		return nil, err
	}

	// Return empty ref name if there are no branches
	if len(branches) == 0 {
		return nil, nil
	}

	// Return first branch name if there's only one
	if len(branches) == 1 {
		return branches[0], nil
	}

	hasDefaultRef, hasLegacyDefaultRef := false, false
	headRef, err := headReference(ctx, repo)
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		// Return HEAD if it exists and corresponds to a branch
		if headRef != nil && bytes.Equal(headRef, branch) {
			return headRef, nil
		}

		if bytes.Equal(branch, git.DefaultRef) {
			hasDefaultRef = true
		}

		hasLegacyDefaultRef = hasLegacyDefaultRef || bytes.Equal(branch, git.LegacyDefaultRef)
	}

	// Return the default ref if it exists
	if hasDefaultRef {
		return git.DefaultRef, nil
	}

	if hasLegacyDefaultRef {
		return git.LegacyDefaultRef, nil
	}

	// If all else fails, return the first branch name
	return branches[0], nil
}

// FindDefaultBranchName returns the default branch name for the given repository
func (s *server) FindDefaultBranchName(ctx context.Context, in *gitalypb.FindDefaultBranchNameRequest) (*gitalypb.FindDefaultBranchNameResponse, error) {
	repo := s.localrepo(in.GetRepository())

	defaultBranchName, err := DefaultBranchName(ctx, repo)
	if err != nil {
		return nil, helper.ErrInternal(err)
	}

	return &gitalypb.FindDefaultBranchNameResponse{Name: defaultBranchName}, nil
}

func parseSortKey(sortKey gitalypb.FindLocalBranchesRequest_SortBy) string {
	switch sortKey {
	case gitalypb.FindLocalBranchesRequest_NAME:
		return "refname"
	case gitalypb.FindLocalBranchesRequest_UPDATED_ASC:
		return "committerdate"
	case gitalypb.FindLocalBranchesRequest_UPDATED_DESC:
		return "-committerdate"
	}

	panic("never reached") // famous last words
}

// FindLocalBranches creates a stream of branches for all local branches in the given repository
func (s *server) FindLocalBranches(in *gitalypb.FindLocalBranchesRequest, stream gitalypb.RefService_FindLocalBranchesServer) error {
	if err := s.findLocalBranches(in, stream); err != nil {
		return helper.ErrInternal(err)
	}

	return nil
}

func (s *server) findLocalBranches(in *gitalypb.FindLocalBranchesRequest, stream gitalypb.RefService_FindLocalBranchesServer) error {
	ctx := stream.Context()
	repo := s.localrepo(in.GetRepository())

	c, err := s.catfileCache.BatchProcess(ctx, repo)
	if err != nil {
		return err
	}

	writer := newFindLocalBranchesWriter(stream, c)
	opts := paginationParamsToOpts(in.GetPaginationParams())
	opts.cmdArgs = []git.Option{
		// %00 inserts the null character into the output (see for-each-ref docs)
		git.Flag{Name: "--format=" + strings.Join(localBranchFormatFields, "%00")},
		git.Flag{Name: "--sort=" + parseSortKey(in.GetSortBy())},
	}

	return s.findRefs(ctx, writer, repo, []string{"refs/heads"}, opts)
}

func (s *server) FindAllBranches(in *gitalypb.FindAllBranchesRequest, stream gitalypb.RefService_FindAllBranchesServer) error {
	if err := s.findAllBranches(in, stream); err != nil {
		return helper.ErrInternal(err)
	}

	return nil
}

func (s *server) findAllBranches(in *gitalypb.FindAllBranchesRequest, stream gitalypb.RefService_FindAllBranchesServer) error {
	repo := s.localrepo(in.GetRepository())

	args := []git.Option{
		// %00 inserts the null character into the output (see for-each-ref docs)
		git.Flag{Name: "--format=" + strings.Join(localBranchFormatFields, "%00")},
	}

	patterns := []string{"refs/heads", "refs/remotes"}

	if in.MergedOnly {
		defaultBranchName, err := DefaultBranchName(stream.Context(), repo)
		if err != nil {
			return err
		}

		args = append(args, git.Flag{Name: fmt.Sprintf("--merged=%s", string(defaultBranchName))})

		if len(in.MergedBranches) > 0 {
			patterns = nil

			for _, mergedBranch := range in.MergedBranches {
				patterns = append(patterns, string(mergedBranch))
			}
		}
	}

	ctx := stream.Context()
	c, err := s.catfileCache.BatchProcess(ctx, repo)
	if err != nil {
		return err
	}

	opts := paginationParamsToOpts(nil)
	opts.cmdArgs = args

	writer := newFindAllBranchesWriter(stream, c)

	return s.findRefs(ctx, writer, repo, patterns, opts)
}

func (s *server) FindTag(ctx context.Context, in *gitalypb.FindTagRequest) (*gitalypb.FindTagResponse, error) {
	if err := s.validateFindTagRequest(in); err != nil {
		return nil, helper.ErrInvalidArgument(err)
	}

	repo := s.localrepo(in.GetRepository())

	tag, err := s.findTag(ctx, repo, in.GetTagName())
	if err != nil {
		return nil, helper.ErrInternal(err)
	}

	return &gitalypb.FindTagResponse{Tag: tag}, nil
}

// parseTagLine parses a line of text with the output format %(objectname) %(objecttype) %(refname:lstrip=2)
func parseTagLine(ctx context.Context, c catfile.Batch, tagLine string) (*gitalypb.Tag, error) {
	fields := strings.SplitN(tagLine, " ", 3)
	if len(fields) != 3 {
		return nil, fmt.Errorf("invalid output from for-each-ref command: %v", tagLine)
	}

	tagID, refType, refName := fields[0], fields[1], fields[2]

	tag := &gitalypb.Tag{
		Id:   tagID,
		Name: []byte(refName),
	}

	switch refType {
	// annotated tag
	case "tag":
		tag, err := catfile.GetTag(ctx, c, git.Revision(tagID), refName, true, true)
		if err != nil {
			return nil, fmt.Errorf("getting annotated tag: %v", err)
		}
		return tag, nil
	case "commit":
		commit, err := catfile.GetCommit(ctx, c, git.Revision(tagID))
		if err != nil {
			return nil, fmt.Errorf("getting commit catfile: %v", err)
		}
		tag.TargetCommit = commit
		return tag, nil
	default:
		return tag, nil
	}
}

func (s *server) findTag(ctx context.Context, repo git.RepositoryExecutor, tagName []byte) (*gitalypb.Tag, error) {
	tagCmd, err := repo.Exec(ctx,
		git.SubCmd{
			Name: "tag",
			Flags: []git.Option{
				git.Flag{Name: "-l"}, git.ValueFlag{Name: "--format", Value: tagFormat},
			},
			Args: []string{string(tagName)},
		},
		git.WithRefTxHook(ctx, repo, s.cfg),
	)
	if err != nil {
		return nil, fmt.Errorf("for-each-ref error: %v", err)
	}

	c, err := s.catfileCache.BatchProcess(ctx, repo)
	if err != nil {
		return nil, err
	}

	var tag *gitalypb.Tag

	scanner := bufio.NewScanner(tagCmd)
	if scanner.Scan() {
		tag, err = parseTagLine(ctx, c, scanner.Text())
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("no tag found")
	}

	if err = tagCmd.Wait(); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *server) validateFindTagRequest(in *gitalypb.FindTagRequest) error {
	if in.GetRepository() == nil {
		return errors.New("repository is empty")
	}

	if _, err := s.locator.GetRepoPath(in.GetRepository()); err != nil {
		return fmt.Errorf("invalid git directory: %v", err)
	}

	if in.GetTagName() == nil {
		return errors.New("tag name is empty")
	}
	return nil
}

func paginationParamsToOpts(p *gitalypb.PaginationParameter) *findRefsOpts {
	opts := &findRefsOpts{delim: '\n'}
	opts.IsPageToken = func(_ []byte) bool { return true }
	opts.Limit = math.MaxInt32

	if p == nil {
		return opts
	}

	if p.GetLimit() >= 0 {
		opts.Limit = int(p.GetLimit())
	}

	if p.GetPageToken() != "" {
		opts.IsPageToken = func(l []byte) bool { return bytes.Compare(l, []byte(p.GetPageToken())) >= 0 }
	}

	return opts
}

// getTagSortField returns a field that needs to be used to sort the tags.
// If sorting is not provided the default sorting is used: by refname.
func getTagSortField(sortBy *gitalypb.FindAllTagsRequest_SortBy) (string, error) {
	if sortBy == nil {
		return "", nil
	}

	var dir string
	switch sortBy.Direction {
	case gitalypb.SortDirection_ASCENDING:
		dir = ""
	case gitalypb.SortDirection_DESCENDING:
		dir = "-"
	default:
		return "", fmt.Errorf("unsupported sorting direction: %s", sortBy.Direction)
	}

	var key string
	switch sortBy.Key {
	case gitalypb.FindAllTagsRequest_SortBy_REFNAME:
		key = "refname"
	case gitalypb.FindAllTagsRequest_SortBy_CREATORDATE:
		key = "creatordate"
	default:
		return "", fmt.Errorf("unsupported sorting key: %s", sortBy.Key)
	}

	return dir + key, nil
}
