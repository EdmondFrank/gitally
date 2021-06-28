package gitpipe

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"gitlab.com/gitlab-org/gitaly/v14/internal/git"
	"gitlab.com/gitlab-org/gitaly/v14/internal/git/catfile"
	"gitlab.com/gitlab-org/gitaly/v14/internal/git/localrepo"
)

// CatfileInfoResult is a result for the CatfileInfo pipeline step.
type CatfileInfoResult struct {
	// Err is an error which occurred during execution of the pipeline.
	Err error

	// ObjectName is the object name as received from the revlistResultChan.
	ObjectName []byte
	// ObjectInfo is the object info of the object.
	ObjectInfo *catfile.ObjectInfo
}

// CatfileInfo processes revlistResults from the given channel and extracts object information via
// `git cat-file --batch-check`. The returned channel will contain all processed catfile info
// results. Any error received via the channel or encountered in this step will cause the pipeline
// to fail. Context cancellation will gracefully halt the pipeline.
func CatfileInfo(ctx context.Context, catfile catfile.Batch, revlistResultChan <-chan RevlistResult) <-chan CatfileInfoResult {
	resultChan := make(chan CatfileInfoResult)

	go func() {
		defer close(resultChan)

		sendResult := func(result CatfileInfoResult) bool {
			select {
			case resultChan <- result:
				return false
			case <-ctx.Done():
				return true
			}
		}

		for revlistResult := range revlistResultChan {
			if revlistResult.Err != nil {
				sendResult(CatfileInfoResult{Err: revlistResult.Err})
				return
			}

			objectInfo, err := catfile.Info(ctx, revlistResult.OID.Revision())
			if err != nil {
				sendResult(CatfileInfoResult{
					Err: fmt.Errorf("retrieving object info for %q: %w", revlistResult.OID, err),
				})
				return
			}

			if isDone := sendResult(CatfileInfoResult{
				ObjectName: revlistResult.ObjectName,
				ObjectInfo: objectInfo,
			}); isDone {
				return
			}
		}
	}()

	return resultChan
}

// CatfileInfoAllObjects enumerates all Git objects part of the repository's object directory and
// extracts their object info via `git cat-file --batch-check`. The returned channel will contain
// all processed results. Any error encountered during execution of this pipeline step will cause
// the pipeline to fail. Context cancellation will gracefully halt the pipeline. Note that with this
// pipeline step, the resulting catfileInfoResults will never have an object name.
func CatfileInfoAllObjects(ctx context.Context, repo *localrepo.Repo) <-chan CatfileInfoResult {
	resultChan := make(chan CatfileInfoResult)

	go func() {
		defer close(resultChan)

		sendResult := func(result CatfileInfoResult) bool {
			select {
			case resultChan <- result:
				return false
			case <-ctx.Done():
				return true
			}
		}

		cmd, err := repo.Exec(ctx, git.SubCmd{
			Name: "cat-file",
			Flags: []git.Option{
				git.Flag{Name: "--batch-all-objects"},
				git.Flag{Name: "--batch-check"},
				git.Flag{Name: "--buffer"},
				git.Flag{Name: "--unordered"},
			},
		})
		if err != nil {
			sendResult(CatfileInfoResult{
				Err: fmt.Errorf("spawning cat-file failed: %w", err),
			})
			return
		}

		reader := bufio.NewReader(cmd)
		for {
			objectInfo, err := catfile.ParseObjectInfo(reader)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				sendResult(CatfileInfoResult{
					Err: fmt.Errorf("parsing object info: %w", err),
				})
				return
			}

			if isDone := sendResult(CatfileInfoResult{
				ObjectInfo: objectInfo,
			}); isDone {
				return
			}
		}

		if err := cmd.Wait(); err != nil {
			sendResult(CatfileInfoResult{
				Err: fmt.Errorf("cat-file failed: %w", err),
			})
			return
		}
	}()

	return resultChan
}

// CatfileInfoFilter filters the catfileInfoResults from the provided channel with the filter
// function: if the filter returns `false` for a given item, then it will be dropped from the
// pipeline. Errors cannot be filtered and will always be passed through.
func CatfileInfoFilter(ctx context.Context, c <-chan CatfileInfoResult, filter func(CatfileInfoResult) bool) <-chan CatfileInfoResult {
	resultChan := make(chan CatfileInfoResult)
	go func() {
		defer close(resultChan)

		for result := range c {
			if result.Err != nil || filter(result) {
				select {
				case resultChan <- result:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return resultChan
}
