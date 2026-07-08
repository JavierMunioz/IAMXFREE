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
		WithRunResult("git", []string{"status", "--porcelain=v1"}, runtimehost.CommandResult{ExitCode: 0, Stdout: ""}, nil)
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
		WithRunResult("git", []string{"status", "--porcelain=v1"}, runtimehost.CommandResult{ExitCode: 0, Stdout: stdout}, nil)
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

func TestStatusEmptyPath(t *testing.T) {
	manager := git.NewManager(runtimehosttest.NewFakeHost())

	_, err := manager.Status(context.Background(), "")
	if !errors.Is(err, git.ErrEmptyPath) {
		t.Fatalf("Status() error = %v, want ErrEmptyPath", err)
	}
}
