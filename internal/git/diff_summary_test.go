package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestDiffSummary(t *testing.T) {
	stdout := "5\t2\tsrc/main.go\n" +
		"0\t10\told_file.go\n" +
		"-\t-\tlogo.png\n"
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"diff", "HEAD", "--numstat"}, runtimehost.CommandResult{ExitCode: 0, Stdout: stdout}, nil)
	manager := git.NewManager(host)

	stat, err := manager.DiffSummary(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("DiffSummary() error = %v", err)
	}
	if len(stat.Files) != 3 {
		t.Fatalf("Files = %v, want 3 entries", stat.Files)
	}

	if stat.Files[0].Path != "src/main.go" || stat.Files[0].Insertions != 5 || stat.Files[0].Deletions != 2 {
		t.Errorf("Files[0] = %+v, want src/main.go +5/-2", stat.Files[0])
	}
	if stat.Files[2].Path != "logo.png" || !stat.Files[2].Binary {
		t.Errorf("Files[2] = %+v, want logo.png marked Binary", stat.Files[2])
	}
}

func TestDiffSummaryNoChanges(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"diff", "HEAD", "--numstat"}, runtimehost.CommandResult{ExitCode: 0, Stdout: ""}, nil)
	manager := git.NewManager(host)

	stat, err := manager.DiffSummary(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("DiffSummary() error = %v", err)
	}
	if len(stat.Files) != 0 {
		t.Fatalf("Files = %v, want empty", stat.Files)
	}
}

func TestDiffSummaryNoCommitsYet(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"diff", "HEAD", "--numstat"},
			runtimehost.CommandResult{ExitCode: 128, Stderr: "fatal: ambiguous argument 'HEAD'"},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	manager := git.NewManager(host)

	stat, err := manager.DiffSummary(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("DiffSummary() error = %v, want nil", err)
	}
	if len(stat.Files) != 0 {
		t.Fatalf("Files = %v, want empty", stat.Files)
	}
}

func TestDiffSummaryEmptyPath(t *testing.T) {
	manager := git.NewManager(runtimehosttest.NewFakeHost())

	_, err := manager.DiffSummary(context.Background(), "")
	if !errors.Is(err, git.ErrEmptyPath) {
		t.Fatalf("DiffSummary() error = %v, want ErrEmptyPath", err)
	}
}
