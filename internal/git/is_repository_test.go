package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

const repoPath = "/srv/apps/my-api"

func TestIsRepositoryTrue(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"rev-parse", "--is-inside-work-tree"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: "true\n"}, nil)
	manager := git.NewManager(host)

	ok, err := manager.IsRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("IsRepository() error = %v", err)
	}
	if !ok {
		t.Fatal("expected IsRepository = true")
	}
}

func TestIsRepositoryFalse(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"rev-parse", "--is-inside-work-tree"},
			runtimehost.CommandResult{ExitCode: 128, Stderr: "fatal: not a git repository"},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	manager := git.NewManager(host)

	ok, err := manager.IsRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("IsRepository() error = %v, want nil (not a repo is not an error)", err)
	}
	if ok {
		t.Fatal("expected IsRepository = false")
	}
}

func TestIsRepositoryEmptyPath(t *testing.T) {
	manager := git.NewManager(runtimehosttest.NewFakeHost())

	_, err := manager.IsRepository(context.Background(), "")
	if !errors.Is(err, git.ErrEmptyPath) {
		t.Fatalf("IsRepository() error = %v, want ErrEmptyPath", err)
	}
}

func TestIsRepositoryGitNotAvailable(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"rev-parse", "--is-inside-work-tree"},
			runtimehost.CommandResult{ExitCode: -1},
			&runtimehost.ExecutionError{Command: "git", ExitCode: -1, Err: errors.New("exec: \"git\": executable file not found in $PATH")},
		)
	manager := git.NewManager(host)

	_, err := manager.IsRepository(context.Background(), repoPath)
	if err == nil {
		t.Fatal("expected an error when git never actually ran")
	}
}
