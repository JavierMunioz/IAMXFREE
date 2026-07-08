package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestStatusClean(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"status", "--porcelain=v1"}, runtimehost.CommandResult{ExitCode: 0, Stdout: ""}, nil).
		WithRunResult("git", []string{"rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"},
			runtimehost.CommandResult{ExitCode: 128},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	manager := git.NewManager(host)

	status, err := manager.Status(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.WorkingTree.Clean {
		t.Error("expected a clean working tree")
	}
}

func TestStatusModifiedAndUntracked(t *testing.T) {
	stdout := " M src/main.go\n" +
		"M  src/util.go\n" +
		"?? notes.txt\n" +
		"?? build/\n"
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"status", "--porcelain=v1"}, runtimehost.CommandResult{ExitCode: 0, Stdout: stdout}, nil).
		WithRunResult("git", []string{"rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"},
			runtimehost.CommandResult{ExitCode: 128},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	manager := git.NewManager(host)

	status, err := manager.Status(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.WorkingTree.Clean {
		t.Fatal("expected a dirty working tree")
	}
	if len(status.WorkingTree.Modified) != 2 {
		t.Errorf("Modified = %v, want 2 entries", status.WorkingTree.Modified)
	}
	if len(status.WorkingTree.Untracked) != 2 {
		t.Errorf("Untracked = %v, want 2 entries", status.WorkingTree.Untracked)
	}
}

func TestStatusAheadBehind(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"status", "--porcelain=v1"}, runtimehost.CommandResult{ExitCode: 0, Stdout: ""}, nil).
		WithRunResult("git", []string{"rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: "origin/main\n"}, nil).
		WithRunResult("git", []string{"rev-list", "--left-right", "--count", "HEAD...@{u}"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: "2\t5\n"}, nil)
	manager := git.NewManager(host)

	status, err := manager.Status(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Ahead != 2 {
		t.Errorf("Ahead = %d, want 2", status.Ahead)
	}
	if status.Behind != 5 {
		t.Errorf("Behind = %d, want 5", status.Behind)
	}
	if len(status.Notes) != 0 {
		t.Errorf("Notes = %v, want empty", status.Notes)
	}
}

func TestStatusNoUpstream(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"status", "--porcelain=v1"}, runtimehost.CommandResult{ExitCode: 0, Stdout: ""}, nil).
		WithRunResult("git", []string{"rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"},
			runtimehost.CommandResult{ExitCode: 128, Stderr: "fatal: no upstream configured for branch 'main'"},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	manager := git.NewManager(host)

	status, err := manager.Status(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Ahead != 0 || status.Behind != 0 {
		t.Errorf("Ahead/Behind = %d/%d, want 0/0", status.Ahead, status.Behind)
	}
	if len(status.Notes) == 0 {
		t.Error("expected a note explaining no upstream is configured")
	}
}

func TestStatusEmptyPath(t *testing.T) {
	manager := git.NewManager(runtimehosttest.NewFakeHost())

	_, err := manager.Status(context.Background(), "")
	if !errors.Is(err, git.ErrEmptyPath) {
		t.Fatalf("Status() error = %v, want ErrEmptyPath", err)
	}
}
