package deployment

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestBuildOperationsApplicationNotFound(t *testing.T) {
	engine := &Engine{appService: &fakeAppService{getErr: errBoom}}

	_, err := engine.BuildOperations(context.Background(), DeploymentPlan{ApplicationID: "missing"})
	if !errors.Is(err, errBoom) {
		t.Fatalf("BuildOperations() error = %v, want %v", err, errBoom)
	}
}

func TestBuildOperationsReturnsSixOperationsInOrder(t *testing.T) {
	app := &models.Application{ID: "app-1", Name: "my-api"}
	engine := &Engine{
		appService:       &fakeAppService{app: app},
		executionService: &fakeExecutionService{},
		gitManager:       git.NewManager(runtimehosttest.NewFakeHost()),
		nginxManager:     nginx.NewManager(runtimehosttest.NewFakeHost()),
	}

	ops, err := engine.BuildOperations(context.Background(), DeploymentPlan{ApplicationID: "app-1"})
	if err != nil {
		t.Fatalf("BuildOperations() error = %v", err)
	}

	wantMethods := []string{"Fetch", "Install", "Build", "Stop", "Start", "Reload"}
	if len(ops) != len(wantMethods) {
		t.Fatalf("len(ops) = %d, want %d", len(ops), len(wantMethods))
	}
	for i, want := range wantMethods {
		if ops[i].Method != want {
			t.Errorf("ops[%d].Method = %q, want %q", i, ops[i].Method, want)
		}
	}
}

// TestBuildOperationsAndExecuteEndToEnd wires BuildOperations straight into
// operations.Executor — the same path the TUI will use — for a fully
// applicable deployment (repo ready, commands configured, app running,
// nginx site needing a reload).
func TestBuildOperationsAndExecuteEndToEnd(t *testing.T) {
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

	execSvc := &fakeExecutionService{running: true, session: services.RunSession{PID: 4242}}

	gitHost := fakeCleanRepoHost().
		WithRunResult("git", []string{"fetch"}, runtimehost.CommandResult{ExitCode: 0}, nil)
	nginxHost := fakeNginxHost().
		WithReadDir("/etc/nginx/sites-available", nil, nil).
		WithRunResult("nginx", []string{"-s", "reload"}, runtimehost.CommandResult{ExitCode: 0}, nil)

	engine := &Engine{
		appService:       &fakeAppService{app: app},
		executionService: execSvc,
		gitManager:       git.NewManager(gitHost),
		nginxManager:     nginx.NewManager(nginxHost),
	}

	plan, err := engine.Plan(context.Background(), "app-1")
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	ops, err := engine.BuildOperations(context.Background(), plan)
	if err != nil {
		t.Fatalf("BuildOperations() error = %v", err)
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if summary.Overall != operations.StateSuccess {
		t.Fatalf("Overall = %q, want %q; operations = %+v", summary.Overall, operations.StateSuccess, summary.Operations)
	}
	if summary.Failed != 0 {
		t.Fatalf("Failed = %d, want 0; operations = %+v", summary.Failed, summary.Operations)
	}
	// Fetch, Install, Build, Stop, Start, Reload should all be applicable
	// and succeed given this fixture.
	if summary.Succeeded != 6 {
		t.Fatalf("Succeeded = %d, want 6; operations = %+v", summary.Succeeded, summary.Operations)
	}
}
