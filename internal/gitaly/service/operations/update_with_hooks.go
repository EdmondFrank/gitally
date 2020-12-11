package operations

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/updateref"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/hook"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/metadata"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

type preReceiveError struct {
	message string
}

func (e preReceiveError) Error() string {
	return e.message
}

type updateRefError struct {
	reference string
}

func (e updateRefError) Error() string {
	return fmt.Sprintf("Could not update %s. Please refresh and try again.", e.reference)
}

func (s *Server) updateReferenceWithHooks(ctx context.Context, repo *gitalypb.Repository, user *gitalypb.User, reference, newrev, oldrev string) error {
	transaction, praefect, err := metadata.TransactionMetadataFromContext(ctx)
	if err != nil {
		return err
	}

	payload, err := git.NewHooksPayload(s.cfg, repo, transaction, praefect).Env()
	if err != nil {
		return err
	}

	if reference == "" {
		return helper.ErrInternalf("updateReferenceWithHooks: got no reference")
	}
	if err := git.ValidateCommitID(oldrev); err != nil {
		return helper.ErrInternalf("updateReferenceWithHooks: got invalid old value: %w", err)
	}
	if err := git.ValidateCommitID(newrev); err != nil {
		return helper.ErrInternalf("updateReferenceWithHooks: got invalid new value: %w", err)
	}

	env := []string{
		payload,
		"GL_PROTOCOL=web",
		fmt.Sprintf("GL_ID=%s", user.GetGlId()),
		fmt.Sprintf("GL_USERNAME=%s", user.GetGlUsername()),
		fmt.Sprintf("GL_PROJECT_PATH=%s", repo.GetGlProjectPath()),
	}

	changes := fmt.Sprintf("%s %s %s\n", oldrev, newrev, reference)
	var stdout, stderr bytes.Buffer

	if err := s.hookManager.PreReceiveHook(ctx, repo, env, strings.NewReader(changes), &stdout, &stderr); err != nil {
		msg := hookErrorFromStdoutAndStderr(stdout.String(), stderr.String())
		return preReceiveError{message: msg}
	}
	if err := s.hookManager.UpdateHook(ctx, repo, reference, oldrev, newrev, env, &stdout, &stderr); err != nil {
		msg := hookErrorFromStdoutAndStderr(stdout.String(), stderr.String())
		return preReceiveError{message: msg}
	}

	if err := s.hookManager.ReferenceTransactionHook(ctx, hook.ReferenceTransactionPrepared, env, strings.NewReader(changes)); err != nil {
		return preReceiveError{message: err.Error()}
	}

	updater, err := updateref.New(ctx, repo)
	if err != nil {
		return err
	}

	if err := updater.Update(reference, newrev, oldrev); err != nil {
		return err
	}

	if err := updater.Wait(); err != nil {
		return updateRefError{reference: reference}
	}

	if err := s.hookManager.PostReceiveHook(ctx, repo, nil, env, strings.NewReader(changes), &stdout, &stderr); err != nil {
		msg := hookErrorFromStdoutAndStderr(stdout.String(), stderr.String())
		return preReceiveError{message: msg}
	}

	return nil
}
