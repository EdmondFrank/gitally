package hook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"golang.org/x/sys/unix"
)

// customHooksExecutor executes all custom hooks for a given repository and hook name
type customHooksExecutor func(ctx context.Context, args, env []string, stdin io.Reader, stdout, stderr io.Writer) error

// lookup hook files in this order:
//
// 1. <repository>.git/custom_hooks/<hook_name> - per project hook
// 2. <repository>.git/custom_hooks/<hook_name>.d/* - per project hooks
// 3. <repository>.git/hooks/<hook_name>.d/* - global hooks
func (m *GitLabHookManager) newCustomHooksExecutor(repo *gitalypb.Repository, hookName string) (customHooksExecutor, error) {
	repoPath, err := m.locator.GetRepoPath(repo)
	if err != nil {
		return nil, err
	}

	var hookFiles []string
	projectCustomHookFile := filepath.Join(repoPath, "custom_hooks", hookName)
	s, err := os.Stat(projectCustomHookFile)
	if err == nil && s.Mode()&0100 != 0 {
		hookFiles = append(hookFiles, projectCustomHookFile)
	}

	projectCustomHookDir := filepath.Join(repoPath, "custom_hooks", fmt.Sprintf("%s.d", hookName))
	files, err := matchFiles(projectCustomHookDir)
	if err != nil {
		return nil, err
	}
	hookFiles = append(hookFiles, files...)

	globalCustomHooksDir := filepath.Join(m.hooksConfig.CustomHooksDir, fmt.Sprintf("%s.d", hookName))

	files, err = matchFiles(globalCustomHooksDir)
	if err != nil {
		return nil, err
	}
	hookFiles = append(hookFiles, files...)

	return func(ctx context.Context, args, env []string, stdin io.Reader, stdout, stderr io.Writer) error {
		var stdinBytes []byte
		if stdin != nil {
			stdinBytes, err = ioutil.ReadAll(stdin)
			if err != nil {
				return err
			}
		}

		for _, hookFile := range hookFiles {
			cmd := exec.Command(hookFile, args...)
			cmd.Dir = repoPath
			c, err := command.New(ctx, cmd, bytes.NewReader(stdinBytes), stdout, stderr, env...)
			if err != nil {
				return err
			}
			if err = c.Wait(); err != nil {
				return err
			}
		}

		return nil
	}, nil
}

// match files from path:
// 1. file must be executable
// 2. file must not match backup file
//
// the resulting list is sorted
func matchFiles(dir string) ([]string, error) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	var hookFiles []string

	for _, fi := range fis {
		if fi.IsDir() || strings.HasSuffix(fi.Name(), "~") {
			continue
		}

		filename := filepath.Join(dir, fi.Name())

		if fi.Mode()&os.ModeSymlink != 0 {
			path, err := filepath.EvalSymlinks(filename)
			if err != nil {
				continue
			}

			fi, err = os.Lstat(path)
			if err != nil {
				continue
			}
		}

		if !validHook(fi, filename) {
			continue
		}

		hookFiles = append(hookFiles, filename)
	}

	return hookFiles, nil
}

func validHook(fi os.FileInfo, filename string) bool {
	if fi.IsDir() {
		return false
	}
	if unix.Access(filename, unix.X_OK) != nil {
		return false
	}

	return true
}

func customHooksEnv(payload git.HooksPayload) []string {
	return []string{
		"GL_REPOSITORY=" + payload.Repo.GetGlRepository(),
	}
}
