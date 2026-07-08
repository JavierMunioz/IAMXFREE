package deployment

import (
	"context"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

func planWithStep(step DeploymentStep) DeploymentPlan {
	return DeploymentPlan{Steps: []DeploymentStep{step}}
}

func TestFetchOperationNotApplicableWhenNoRepository(t *testing.T) {
	engine := &Engine{gitManager: git.NewManager(fakeCleanRepoHost())}
	app := &models.Application{Source: models.SourceInfo{LocalPath: "/srv/apps/my-api"}}
	plan := planWithStep(DeploymentStep{Operation: OperationVerifyRepository, Status: StepStatusBlocked})

	op := engine.fetchOperation(app, plan)
	if op.Applicable {
		t.Fatal("expected Fetch to not be applicable when the repository step is blocked")
	}
	if op.SkipReason == "" {
		t.Error("expected a SkipReason")
	}
}

func TestFetchOperationApplicableAndRuns(t *testing.T) {
	host := fakeCleanRepoHost().
		WithRunResult("git", []string{"fetch"}, runtimehost.CommandResult{ExitCode: 0}, nil)
	engine := &Engine{gitManager: git.NewManager(host)}
	app := &models.Application{Source: models.SourceInfo{LocalPath: "/srv/apps/my-api"}}
	plan := planWithStep(DeploymentStep{Operation: OperationVerifyRepository, Status: StepStatusReady})

	op := engine.fetchOperation(app, plan)
	if !op.Applicable {
		t.Fatal("expected Fetch to be applicable")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestFetchOperationRunPropagatesError(t *testing.T) {
	host := fakeCleanRepoHost().
		WithRunResult("git", []string{"fetch"},
			runtimehost.CommandResult{ExitCode: 1, Stderr: "fatal: unable to access"},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 1, Err: errBoom},
		)
	engine := &Engine{gitManager: git.NewManager(host)}
	app := &models.Application{Source: models.SourceInfo{LocalPath: "/srv/apps/my-api"}}
	plan := planWithStep(DeploymentStep{Operation: OperationVerifyRepository, Status: StepStatusReady})

	op := engine.fetchOperation(app, plan)
	if err := op.Run(context.Background()); err == nil {
		t.Fatal("expected Run to propagate the fetch error")
	}
}

// ensure the operations package's Operation shape is what we expect (name
// non-empty, method set) regardless of applicability.
func TestFetchOperationAlwaysHasIdentity(t *testing.T) {
	engine := &Engine{gitManager: git.NewManager(fakeCleanRepoHost())}
	app := &models.Application{}
	op := engine.fetchOperation(app, DeploymentPlan{})

	if op.Name == "" || op.Method == "" || op.Component != string(ComponentGit) {
		t.Fatalf("op = %+v, want identity fields set", op)
	}
}
