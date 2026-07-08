package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestInspectNotARepository(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"rev-parse", "--is-inside-work-tree"},
			runtimehost.CommandResult{ExitCode: 128, Stderr: "fatal: not a git repository"},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	manager := git.NewManager(host)

	repo, err := manager.Inspect(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if repo.IsRepo {
		t.Fatal("expected IsRepo = false")
	}
	if repo.Path != repoPath {
		t.Errorf("Path = %q, want %q", repo.Path, repoPath)
	}
	// No other command was configured on the FakeHost; if Inspect had
	// attempted one, it would have failed with "no result configured".
}

func TestInspectRepository(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"rev-parse", "--is-inside-work-tree"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: "true\n"}, nil).
		WithRunResult("git", []string{"rev-parse", "--abbrev-ref", "HEAD"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: "main\n"}, nil).
		WithRunResult("git", []string{"log", "-1", "--format=%H\x1f%h\x1f%s\x1f%an\x1f%aI"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: "abc123\x1fabc\x1fInitial\x1fJavier\x1f2026-07-07T10:00:00-05:00\n"}, nil).
		WithRunResult("git", []string{"remote", "-v"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: "origin\thttps://github.com/user/repo.git (fetch)\n"}, nil).
		WithRunResult("git", []string{"status", "--porcelain=v1"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: ""}, nil).
		WithRunResult("git", []string{"rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"},
			runtimehost.CommandResult{ExitCode: 128},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	manager := git.NewManager(host)

	repo, err := manager.Inspect(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if !repo.IsRepo {
		t.Fatal("expected IsRepo = true")
	}
	if repo.Branch.Name != "main" {
		t.Errorf("Branch.Name = %q, want %q", repo.Branch.Name, "main")
	}
	if repo.Commit.ShortSHA != "abc" {
		t.Errorf("Commit.ShortSHA = %q, want %q", repo.Commit.ShortSHA, "abc")
	}
	if len(repo.Remotes) != 1 || repo.Remotes[0].Name != "origin" {
		t.Errorf("Remotes = %v, want a single origin entry", repo.Remotes)
	}
	if !repo.Status.WorkingTree.Clean {
		t.Error("expected a clean working tree")
	}
}

func TestInspectEmptyPath(t *testing.T) {
	manager := git.NewManager(runtimehosttest.NewFakeHost())

	_, err := manager.Inspect(context.Background(), "")
	if !errors.Is(err, git.ErrEmptyPath) {
		t.Fatalf("Inspect() error = %v, want ErrEmptyPath", err)
	}
}
