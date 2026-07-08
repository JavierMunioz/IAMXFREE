package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestFetchSuccess(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"fetch"},
			runtimehost.CommandResult{ExitCode: 0, Stderr: "From github.com:user/repo\n   abc123..def456  main -> origin/main"}, nil)
	manager := git.NewManager(host)

	result, err := manager.Fetch(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if result.Output == "" {
		t.Error("expected non-empty Output")
	}
}

func TestFetchFailure(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"fetch"},
			runtimehost.CommandResult{ExitCode: 128, Stderr: "fatal: unable to access: Could not resolve host"},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	manager := git.NewManager(host)

	_, err := manager.Fetch(context.Background(), repoPath)
	if err == nil {
		t.Fatal("expected an error for a failed fetch")
	}
}

func TestFetchEmptyPath(t *testing.T) {
	manager := git.NewManager(runtimehosttest.NewFakeHost())

	_, err := manager.Fetch(context.Background(), "")
	if !errors.Is(err, git.ErrEmptyPath) {
		t.Fatalf("Fetch() error = %v, want ErrEmptyPath", err)
	}
}
