package gitpipe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"gitlab.com/gitlab-org/gitaly/v14/internal/git/catfile"
)

// CatfileObjectResult is a result for the CatfileObject pipeline step.
type CatfileObjectResult struct {
	// Err is an error which occurred during execution of the pipeline.
	Err error

	// ObjectName is the object name as received from the revlistResultChan.
	ObjectName []byte
	// ObjectInfo is the object info of the object.
	ObjectInfo *catfile.ObjectInfo
	// obbjectReader is the reader for the raw object data. The reader must always be consumed
	// by the caller.
	ObjectReader io.Reader
}

// CatfileObject processes catfileInfoResults from the given channel and reads associated objects
// into memory via `git cat-file --batch`. The returned channel will contain all processed objects.
// Any error received via the channel or encountered in this step will cause the pipeline to fail.
// Context cancellation will gracefully halt the pipeline. The returned object readers must always
// be fully consumed by the caller.
func CatfileObject(
	ctx context.Context,
	catfileProcess catfile.Batch,
	catfileInfoResultChan <-chan CatfileInfoResult,
) <-chan CatfileObjectResult {
	resultChan := make(chan CatfileObjectResult)
	go func() {
		defer close(resultChan)

		sendResult := func(result CatfileObjectResult) bool {
			select {
			case resultChan <- result:
				return false
			case <-ctx.Done():
				return true
			}
		}

		var objectReader *signallingReader

		for catfileInfoResult := range catfileInfoResultChan {
			if catfileInfoResult.Err != nil {
				sendResult(CatfileObjectResult{Err: catfileInfoResult.Err})
				return
			}

			// We mustn't try to read another object before reading the previous object
			// has concluded. Given that this is not under our control but under the
			// control of the caller, we thus have to wait until the blocking reader has
			// reached EOF.
			if objectReader != nil {
				select {
				case <-objectReader.doneCh:
				case <-ctx.Done():
					return
				}
			}

			var object *catfile.Object
			var err error

			objectType := catfileInfoResult.ObjectInfo.Type
			switch objectType {
			case "tag":
				object, err = catfileProcess.Tag(ctx, catfileInfoResult.ObjectInfo.Oid.Revision())
			case "commit":
				object, err = catfileProcess.Commit(ctx, catfileInfoResult.ObjectInfo.Oid.Revision())
			case "tree":
				object, err = catfileProcess.Tree(ctx, catfileInfoResult.ObjectInfo.Oid.Revision())
			case "blob":
				object, err = catfileProcess.Blob(ctx, catfileInfoResult.ObjectInfo.Oid.Revision())
			default:
				err = fmt.Errorf("unknown object type %q", objectType)
			}

			if err != nil {
				sendResult(CatfileObjectResult{
					Err: fmt.Errorf("requesting object: %w", err),
				})
				return
			}

			objectReader = &signallingReader{
				reader: object,
				doneCh: make(chan interface{}),
			}

			if isDone := sendResult(CatfileObjectResult{
				ObjectName:   catfileInfoResult.ObjectName,
				ObjectInfo:   catfileInfoResult.ObjectInfo,
				ObjectReader: objectReader,
			}); isDone {
				return
			}
		}
	}()

	return resultChan
}

type signallingReader struct {
	reader    io.Reader
	doneCh    chan interface{}
	closeOnce sync.Once
}

func (r *signallingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if errors.Is(err, io.EOF) {
		r.closeOnce.Do(func() {
			close(r.doneCh)
		})
	}
	return n, err
}
