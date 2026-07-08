package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestCurrentCommit(t *testing.T) {
	stdout := "abc123def456\x1fabc123d\x1fFix the bug\x1fJavier Munioz\x1f2026-07-07T10:30:00-05:00\n"
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"log", "-1", "--format=%H\x1f%h\x1f%s\x1f%an\x1f%aI"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: stdout}, nil)
	manager := git.NewManager(host)

	commit, err := manager.CurrentCommit(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("CurrentCommit() error = %v", err)
	}
	if commit.SHA != "abc123def456" {
		t.Errorf("SHA = %q, want %q", commit.SHA, "abc123def456")
	}
	if commit.ShortSHA != "abc123d" {
		t.Errorf("ShortSHA = %q, want %q", commit.ShortSHA, "abc123d")
	}
	if commit.Message != "Fix the bug" {
		t.Errorf("Message = %q, want %q", commit.Message, "Fix the bug")
	}
	if commit.Author != "Javier Munioz" {
		t.Errorf("Author = %q, want %q", commit.Author, "Javier Munioz")
	}
	if commit.Date.IsZero() {
		t.Error("expected Date to be parsed, got zero value")
	}
}

func TestCurrentCommitNoCommitsYet(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"log", "-1", "--format=%H\x1f%h\x1f%s\x1f%an\x1f%aI"},
			runtimehost.CommandResult{ExitCode: 128, Stderr: "fatal: your current branch 'main' does not have any commits yet"},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	manager := git.NewManager(host)

	commit, err := manager.CurrentCommit(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("CurrentCommit() error = %v, want nil", err)
	}
	if commit != (git.Commit{}) {
		t.Errorf("CurrentCommit() = %+v, want zero value", commit)
	}
}

func TestCurrentCommitEmptyPath(t *testing.T) {
	manager := git.NewManager(runtimehosttest.NewFakeHost())

	_, err := manager.CurrentCommit(context.Background(), "")
	if !errors.Is(err, git.ErrEmptyPath) {
		t.Fatalf("CurrentCommit() error = %v, want ErrEmptyPath", err)
	}
}
