package deployment

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

var errBoom = errors.New("boom")

func TestGitStepsNoLocalPath(t *testing.T) {
	engine := &Engine{gitManager: git.NewManager(runtimehosttest.NewFakeHost())}
	app := &models.Application{ID: "app-1"}

	steps, repo := engine.gitSteps(context.Background(), app)
	if len(steps) != 1 {
		t.Fatalf("steps = %v, want 1 (verify only)", steps)
	}
	if steps[0].Status != StepStatusBlocked {
		t.Errorf("Status = %q, want %q", steps[0].Status, StepStatusBlocked)
	}
	if repo.IsRepo {
		t.Error("expected a zero-value repository")
	}
}

func TestGitStepsNotARepository(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("git", []string{"rev-parse", "--is-inside-work-tree"},
			runtimehost.CommandResult{ExitCode: 128},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errBoom},
		)
	engine := &Engine{gitManager: git.NewManager(host)}
	app := &models.Application{ID: "app-1", Source: models.SourceInfo{LocalPath: "/srv/apps/my-api"}}

	steps, repo := engine.gitSteps(context.Background(), app)
	if len(steps) != 1 || steps[0].Status != StepStatusBlocked {
		t.Fatalf("steps = %+v, want a single blocked step", steps)
	}
	if repo.IsRepo {
		t.Error("expected IsRepo = false")
	}
}

func TestGitStepsCleanRepository(t *testing.T) {
	host := fakeCleanRepoHost()
	engine := &Engine{gitManager: git.NewManager(host)}
	app := &models.Application{ID: "app-1", Source: models.SourceInfo{LocalPath: "/srv/apps/my-api"}}

	steps, repo := engine.gitSteps(context.Background(), app)
	if len(steps) != 2 {
		t.Fatalf("steps = %+v, want 2 (verify + local changes)", steps)
	}
	if steps[0].Status != StepStatusReady {
		t.Errorf("verify step Status = %q, want %q", steps[0].Status, StepStatusReady)
	}
	if steps[1].Status != StepStatusReady {
		t.Errorf("changes step Status = %q, want %q", steps[1].Status, StepStatusReady)
	}
	if len(steps[1].Warnings) != 0 {
		t.Errorf("expected no warnings for a clean tree, got %v", steps[1].Warnings)
	}
	if !repo.IsRepo {
		t.Error("expected IsRepo = true")
	}
}

func TestGitStepsDirtyWorkingTree(t *testing.T) {
	host := fakeCleanRepoHost().
		WithRunResult("git", []string{"status", "--porcelain=v1"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: " M src/main.go\n?? new.txt\n"}, nil)
	engine := &Engine{gitManager: git.NewManager(host)}
	app := &models.Application{ID: "app-1", Source: models.SourceInfo{LocalPath: "/srv/apps/my-api"}}

	steps, _ := engine.gitSteps(context.Background(), app)
	if steps[1].Status != StepStatusWarning {
		t.Fatalf("changes step Status = %q, want %q", steps[1].Status, StepStatusWarning)
	}
	if len(steps[1].Warnings) == 0 {
		t.Error("expected a warning describing the dirty working tree")
	}
}

func TestPullStepUpToDate(t *testing.T) {
	step := pullStep(git.Repository{IsRepo: true, Status: git.RepositoryStatus{Behind: 0}})
	if step.Status != StepStatusSkipped || step.Required {
		t.Fatalf("step = %+v, want skipped and not required", step)
	}
}

func TestPullStepBehind(t *testing.T) {
	step := pullStep(git.Repository{IsRepo: true, Status: git.RepositoryStatus{Behind: 3}})
	if step.Status != StepStatusReady || !step.Required {
		t.Fatalf("step = %+v, want ready and required", step)
	}
}

func TestPullStepBehindAndAheadWarnsOfPossibleMerge(t *testing.T) {
	step := pullStep(git.Repository{IsRepo: true, Status: git.RepositoryStatus{Behind: 3, Ahead: 1}})
	if len(step.Risks) == 0 {
		t.Error("expected a risk noting unpushed local commits")
	}
}

func TestPullStepNotARepository(t *testing.T) {
	step := pullStep(git.Repository{IsRepo: false})
	if step.Status != StepStatusSkipped {
		t.Fatalf("step.Status = %q, want %q", step.Status, StepStatusSkipped)
	}
}

// fakeCleanRepoHost configures a FakeHost that reports a clean repository
// on branch "main" with no upstream configured.
func fakeCleanRepoHost() *runtimehosttest.FakeHost {
	return runtimehosttest.NewFakeHost().
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
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errBoom},
		)
}
