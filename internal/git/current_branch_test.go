package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestCurrentBranch(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"rev-parse", "--abbrev-ref", "HEAD"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: "main\n"}, nil)
	manager := git.NewManager(host)

	branch, err := manager.CurrentBranch(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if branch.Name != "main" || branch.Detached {
		t.Errorf("CurrentBranch() = %+v, want {Name: main, Detached: false}", branch)
	}
}

func TestCurrentBranchDetachedHead(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"rev-parse", "--abbrev-ref", "HEAD"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: "HEAD\n"}, nil)
	manager := git.NewManager(host)

	branch, err := manager.CurrentBranch(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if !branch.Detached || branch.Name != "" {
		t.Errorf("CurrentBranch() = %+v, want {Name: \"\", Detached: true}", branch)
	}
}

func TestCurrentBranchEmptyPath(t *testing.T) {
	manager := git.NewManager(runtimehosttest.NewFakeHost())

	_, err := manager.CurrentBranch(context.Background(), "")
	if !errors.Is(err, git.ErrEmptyPath) {
		t.Fatalf("CurrentBranch() error = %v, want ErrEmptyPath", err)
	}
}
