package deployment

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestPlanApplicationNotFound(t *testing.T) {
	engine := &Engine{appService: &fakeAppService{getErr: errBoom}}

	_, err := engine.Plan(context.Background(), "missing-app")
	if !errors.Is(err, errBoom) {
		t.Fatalf("Plan() error = %v, want %v", err, errBoom)
	}
}

func TestPlanReadyDeployment(t *testing.T) {
	app := &models.Application{
		ID:   "app-1",
		Name: "my-api",
		Source: models.SourceInfo{
			LocalPath: "/srv/apps/my-api",
		},
		Config: models.DeploymentConfig{
			InstallCommand: "npm install",
			BuildCommand:   "npm run build",
			Domain:         "example.com",
		},
	}

	engine := &Engine{
		appService:       &fakeAppService{app: app},
		executionService: &fakeExecutionService{running: false},
		gitManager:       git.NewManager(fakeCleanRepoHost()),
		nginxManager: nginx.NewManager(fakeNginxHost().
			WithReadDir("/etc/nginx/sites-available", []string{"example.com.conf"}, nil).
			WithFile("/etc/nginx/sites-enabled/example.com.conf")),
	}

	plan, err := engine.Plan(context.Background(), "app-1")
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	if plan.ApplicationID != "app-1" || plan.ApplicationName != "my-api" {
		t.Errorf("plan identity = %+v, want app-1/my-api", plan)
	}
	if plan.GeneratedAt.IsZero() {
		t.Error("expected GeneratedAt to be set")
	}
	if !plan.Summary.Ready {
		t.Errorf("expected the plan to be Ready, got Summary = %+v", plan.Summary)
	}
	if plan.Summary.BlockedSteps != 0 {
		t.Errorf("BlockedSteps = %d, want 0", plan.Summary.BlockedSteps)
	}
	// pre-deploy hook(skipped) + verify repo + local changes + pull(skipped)
	// + install + build + check running + restart + verify nginx +
	// reload(skipped) + post-deploy hook(skipped) = 11
	if plan.Summary.TotalSteps != 11 {
		t.Errorf("TotalSteps = %d, want 11, steps = %+v", plan.Summary.TotalSteps, plan.Steps)
	}
}

func TestPlanBlockedWhenNoRepository(t *testing.T) {
	app := &models.Application{ID: "app-1", Name: "my-api"} // no LocalPath

	engine := &Engine{
		appService:       &fakeAppService{app: app},
		executionService: &fakeExecutionService{},
		gitManager:       git.NewManager(runtimehosttest.NewFakeHost()),
		nginxManager:     nginx.NewManager(runtimehosttest.NewFakeHost()),
	}

	plan, err := engine.Plan(context.Background(), "app-1")
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	if plan.Summary.Ready {
		t.Error("expected the plan to not be Ready when the repository step is blocked")
	}
	if plan.Summary.BlockedSteps == 0 {
		t.Error("expected at least one blocked step")
	}
}

func TestPlanCollectsWarningsAcrossSteps(t *testing.T) {
	app := &models.Application{
		ID:     "app-1",
		Name:   "my-api",
		Source: models.SourceInfo{LocalPath: "/srv/apps/my-api"},
	}

	dirtyHost := fakeCleanRepoHost().
		WithRunResult("git", []string{"status", "--porcelain=v1"},
			runtimehost.CommandResult{ExitCode: 0, Stdout: " M src/main.go\n"}, nil)

	engine := &Engine{
		appService:       &fakeAppService{app: app},
		executionService: &fakeExecutionService{running: true},
		gitManager:       git.NewManager(dirtyHost),
		nginxManager:     nginx.NewManager(runtimehosttest.NewFakeHost()),
	}

	plan, err := engine.Plan(context.Background(), "app-1")
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if len(plan.Warnings) == 0 {
		t.Fatal("expected at least one flattened warning")
	}
	if plan.Summary.WarningCount != len(plan.Warnings) {
		t.Errorf("Summary.WarningCount = %d, want %d (len(plan.Warnings))", plan.Summary.WarningCount, len(plan.Warnings))
	}
}
