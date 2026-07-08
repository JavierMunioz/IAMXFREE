package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories/jsonstore"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func newService(t *testing.T) services.ApplicationService {
	t.Helper()
	return newServiceWithResolver(t, execution.NewResolver(execution.NewRegistry()))
}

func newServiceWithResolver(t *testing.T, resolver *execution.Resolver) services.ApplicationService {
	t.Helper()
	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
	gitManager := git.NewManager(runtimehosttest.NewFakeHost())
	return services.NewApplicationService(repo, resolver, gitManager)
}

// fakeStrategy is a minimal execution.Strategy used to test
// ResolveExecutionStrategy/CheckExecutionHealth/ExecutionService's wiring.
type fakeStrategy struct {
	name      string
	runtime   models.Runtime
	health    execution.HealthCheck
	healthErr error

	installErr error
	buildErr   error

	startSession  execution.Session
	startErr      error
	stopErr       error
	statusSession execution.Session
	statusErr     error
	logStream     execution.LogStream
	logsErr       error
}

func (s *fakeStrategy) Metadata() execution.Metadata {
	return execution.Metadata{Name: s.name, SupportedRuntimes: []models.Runtime{s.runtime}}
}
func (s *fakeStrategy) CanHandle(app *models.Application) bool { return app.Runtime == s.runtime }
func (s *fakeStrategy) HealthCheck(context.Context, *models.Application) (execution.HealthCheck, error) {
	return s.health, s.healthErr
}
func (s *fakeStrategy) Readiness(context.Context, *models.Application) (execution.Readiness, error) {
	return execution.Readiness{}, execution.ErrNotImplemented
}
func (s *fakeStrategy) Install(context.Context, *models.Application) error {
	return s.installErr
}
func (s *fakeStrategy) Build(context.Context, *models.Application) error {
	return s.buildErr
}
func (s *fakeStrategy) Start(context.Context, *models.Application, int) (execution.Session, error) {
	return s.startSession, s.startErr
}
func (s *fakeStrategy) Stop(context.Context, *models.Application, execution.Session) error {
	return s.stopErr
}
func (s *fakeStrategy) Status(context.Context, *models.Application, execution.Session) (execution.Session, error) {
	return s.statusSession, s.statusErr
}
func (s *fakeStrategy) Logs(context.Context, *models.Application, execution.Session) (execution.LogStream, error) {
	return s.logStream, s.logsErr
}
func (s *fakeStrategy) Restart(context.Context, *models.Application) error {
	return execution.ErrNotImplemented
}
func (s *fakeStrategy) Update(context.Context, *models.Application) error {
	return execution.ErrNotImplemented
}

func TestApplicationServiceRegister(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)

	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	got, err := svc.Get(ctx, app.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name != "my-api" {
		t.Errorf("Name = %q, want %q", got.Name, "my-api")
	}
}

func TestApplicationServiceRegisterRejectsInvalidApplication(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("", models.ApplicationTypeAPI)

	if err := svc.Register(ctx, app); !errors.Is(err, models.ErrApplicationNameRequired) {
		t.Fatalf("Register() error = %v, want %v", err, models.ErrApplicationNameRequired)
	}
}

func TestApplicationServiceRegisterRejectsDuplicateName(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	first := models.NewApplication("my-api", models.ApplicationTypeAPI)
	if err := svc.Register(ctx, first); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	second := models.NewApplication("my-api", models.ApplicationTypeBackend)
	if err := svc.Register(ctx, second); !errors.Is(err, services.ErrApplicationNameTaken) {
		t.Fatalf("Register() error = %v, want %v", err, services.ErrApplicationNameTaken)
	}
}

func TestApplicationServiceList(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	for _, name := range []string{"app-one", "app-two"} {
		app := models.NewApplication(name, models.ApplicationTypeAPI)
		if err := svc.Register(ctx, app); err != nil {
			t.Fatalf("Register() error = %v", err)
		}
	}

	apps, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("len(apps) = %d, want 2", len(apps))
	}
}

func TestApplicationServiceUpdateConfig(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	updated, err := svc.UpdateConfig(ctx, app.ID, models.DeploymentConfig{InternalPort: 4000})
	if err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}
	if updated.Config.InternalPort != 4000 {
		t.Errorf("Config.InternalPort = %d, want 4000", updated.Config.InternalPort)
	}
}

func TestApplicationServiceChangeStatus(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	updated, err := svc.ChangeStatus(ctx, app.ID, models.StatusRunning)
	if err != nil {
		t.Fatalf("ChangeStatus() error = %v", err)
	}
	if updated.Status != models.StatusRunning {
		t.Errorf("Status = %q, want %q", updated.Status, models.StatusRunning)
	}
}

func TestApplicationServiceRemove(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := svc.Remove(ctx, app.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := svc.Get(ctx, app.ID); err == nil {
		t.Fatal("expected Get() to fail after Remove()")
	}
}

func TestApplicationServiceResolveExecutionStrategyNoneRegistered(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if _, err := svc.ResolveExecutionStrategy(ctx, app.ID); !errors.Is(err, execution.ErrNoStrategyFound) {
		t.Fatalf("ResolveExecutionStrategy() error = %v, want %v", err, execution.ErrNoStrategyFound)
	}
}

func TestApplicationServiceResolveExecutionStrategyReturnsMetadata(t *testing.T) {
	ctx := context.Background()
	registry := execution.NewRegistry()
	registry.Register(&fakeStrategy{name: "node-npm", runtime: models.RuntimeNode})
	svc := newServiceWithResolver(t, execution.NewResolver(registry))

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	meta, err := svc.ResolveExecutionStrategy(ctx, app.ID)
	if err != nil {
		t.Fatalf("ResolveExecutionStrategy() error = %v", err)
	}
	if meta.Name != "node-npm" {
		t.Fatalf("ResolveExecutionStrategy() name = %q, want %q", meta.Name, "node-npm")
	}
}

func TestApplicationServiceResolveExecutionStrategyUnknownApplication(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	if _, err := svc.ResolveExecutionStrategy(ctx, "does-not-exist"); err == nil {
		t.Fatal("expected an error for an unknown application")
	}
}

func TestApplicationServiceCheckExecutionHealthHealthy(t *testing.T) {
	ctx := context.Background()
	registry := execution.NewRegistry()
	registry.Register(&fakeStrategy{
		name:    "node-npm",
		runtime: models.RuntimeNode,
		health: execution.HealthCheck{Items: []execution.HealthCheckItem{
			{Name: execution.HealthCheckRuntimeInstalled, Status: execution.HealthStatusPass},
		}},
	})
	svc := newServiceWithResolver(t, execution.NewResolver(registry))

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	health, err := svc.CheckExecutionHealth(ctx, app.ID)
	if err != nil {
		t.Fatalf("CheckExecutionHealth() error = %v", err)
	}
	if health.StrategyName != "node-npm" {
		t.Errorf("StrategyName = %q, want %q", health.StrategyName, "node-npm")
	}
	if !health.Healthy {
		t.Error("expected Healthy to be true")
	}
}

func TestApplicationServiceCheckExecutionHealthUnhealthy(t *testing.T) {
	ctx := context.Background()
	registry := execution.NewRegistry()
	registry.Register(&fakeStrategy{
		name:    "node-npm",
		runtime: models.RuntimeNode,
		health: execution.HealthCheck{Items: []execution.HealthCheckItem{
			{Name: execution.HealthCheckRuntimeInstalled, Status: execution.HealthStatusFail, Detail: "node not found"},
		}},
	})
	svc := newServiceWithResolver(t, execution.NewResolver(registry))

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	health, err := svc.CheckExecutionHealth(ctx, app.ID)
	if err != nil {
		t.Fatalf("CheckExecutionHealth() error = %v", err)
	}
	if health.Healthy {
		t.Error("expected Healthy to be false")
	}
}

func TestApplicationServiceCheckExecutionHealthNoStrategy(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if _, err := svc.CheckExecutionHealth(ctx, app.ID); !errors.Is(err, execution.ErrNoStrategyFound) {
		t.Fatalf("CheckExecutionHealth() error = %v, want %v", err, execution.ErrNoStrategyFound)
	}
}

func TestApplicationServiceCheckGitStatusNoLocalPath(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	status, err := svc.CheckGitStatus(ctx, app.ID)
	if err != nil {
		t.Fatalf("CheckGitStatus() error = %v", err)
	}
	if status.IsRepo {
		t.Error("expected IsRepo = false when the application has no LocalPath configured")
	}
}

func TestApplicationServiceCheckGitStatusRepository(t *testing.T) {
	ctx := context.Background()

	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
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
			runtimehost.CommandResult{ExitCode: 0, Stdout: " M src/main.go\n"}, nil).
		WithRunResult("git", []string{"rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"},
			runtimehost.CommandResult{ExitCode: 128},
			&runtimehost.ExecutionError{Command: "git", ExitCode: 128, Err: errors.New("exit status 128")},
		)
	svc := services.NewApplicationService(repo, execution.NewResolver(execution.NewRegistry()), git.NewManager(host))

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Source.LocalPath = "/srv/apps/my-api"
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	status, err := svc.CheckGitStatus(ctx, app.ID)
	if err != nil {
		t.Fatalf("CheckGitStatus() error = %v", err)
	}
	if !status.IsRepo {
		t.Fatal("expected IsRepo = true")
	}
	if status.Branch != "main" {
		t.Errorf("Branch = %q, want %q", status.Branch, "main")
	}
	if status.CommitSHA != "abc" {
		t.Errorf("CommitSHA = %q, want %q", status.CommitSHA, "abc")
	}
	if status.Clean {
		t.Error("expected Clean = false (one modified file)")
	}
	if status.Modified != 1 {
		t.Errorf("Modified = %d, want 1", status.Modified)
	}
	if status.RemoteName != "origin" {
		t.Errorf("RemoteName = %q, want %q", status.RemoteName, "origin")
	}
}
