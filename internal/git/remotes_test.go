package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestRemotes(t *testing.T) {
	stdout := "origin\thttps://github.com/user/repo.git (fetch)\n" +
		"origin\thttps://github.com/user/repo.git (push)\n" +
		"upstream\thttps://github.com/other/repo.git (fetch)\n" +
		"upstream\thttps://github.com/other/repo.git (push)\n"
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"remote", "-v"}, runtimehost.CommandResult{ExitCode: 0, Stdout: stdout}, nil)
	manager := git.NewManager(host)

	remotes, err := manager.Remotes(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("Remotes() error = %v", err)
	}
	if len(remotes) != 2 {
		t.Fatalf("Remotes() = %v, want 2 entries", remotes)
	}
	if remotes[0].Name != "origin" || remotes[0].URL != "https://github.com/user/repo.git" {
		t.Errorf("remotes[0] = %+v, want origin", remotes[0])
	}
	if remotes[1].Name != "upstream" || remotes[1].URL != "https://github.com/other/repo.git" {
		t.Errorf("remotes[1] = %+v, want upstream", remotes[1])
	}
}

func TestRemotesNone(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"remote", "-v"}, runtimehost.CommandResult{ExitCode: 0, Stdout: ""}, nil)
	manager := git.NewManager(host)

	remotes, err := manager.Remotes(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("Remotes() error = %v", err)
	}
	if len(remotes) != 0 {
		t.Fatalf("Remotes() = %v, want empty", remotes)
	}
}

func TestRemotesEmptyPath(t *testing.T) {
	manager := git.NewManager(runtimehosttest.NewFakeHost())

	_, err := manager.Remotes(context.Background(), "")
	if !errors.Is(err, git.ErrEmptyPath) {
		t.Fatalf("Remotes() error = %v, want ErrEmptyPath", err)
	}
}
