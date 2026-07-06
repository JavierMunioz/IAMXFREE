package execution_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// fakeStrategy is a minimal Strategy used across this package's tests to
// verify the contract is implementable and behaves as documented. Real
// strategies (Node+npm, Docker, systemd, ...) arrive in later iterations.
type fakeStrategy struct {
	name    string
	runtime models.Runtime
}

func (s *fakeStrategy) Metadata() execution.Metadata {
	return execution.Metadata{
		Name:              s.name,
		SupportedRuntimes: []models.Runtime{s.runtime},
		Capabilities:      []execution.Capability{execution.CapabilityStart},
	}
}

func (s *fakeStrategy) CanHandle(app *models.Application) bool {
	return app.Runtime == s.runtime
}

func (s *fakeStrategy) HealthCheck(context.Context, *models.Application) (execution.HealthCheck, error) {
	return execution.HealthCheck{}, execution.ErrNotImplemented
}
func (s *fakeStrategy) Readiness(context.Context, *models.Application) (execution.Readiness, error) {
	return execution.Readiness{}, execution.ErrNotImplemented
}
func (s *fakeStrategy) Install(context.Context, *models.Application) error {
	return execution.ErrNotImplemented
}
func (s *fakeStrategy) Build(context.Context, *models.Application) error {
	return execution.ErrNotImplemented
}
func (s *fakeStrategy) Start(context.Context, *models.Application) (execution.Session, error) {
	return execution.Session{}, execution.ErrNotImplemented
}
func (s *fakeStrategy) Stop(context.Context, *models.Application, execution.Session) error {
	return execution.ErrNotImplemented
}
func (s *fakeStrategy) Restart(context.Context, *models.Application) error {
	return execution.ErrNotImplemented
}
func (s *fakeStrategy) Update(context.Context, *models.Application) error {
	return execution.ErrNotImplemented
}

var _ execution.Strategy = (*fakeStrategy)(nil)

func TestStrategyLifecycleMethodsReportNotImplemented(t *testing.T) {
	s := &fakeStrategy{name: "fake", runtime: models.RuntimeNode}
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	ctx := context.Background()

	_, healthErr := s.HealthCheck(ctx, app)
	_, readinessErr := s.Readiness(ctx, app)
	_, startErr := s.Start(ctx, app)

	for _, err := range []error{
		healthErr,
		readinessErr,
		s.Install(ctx, app),
		s.Build(ctx, app),
		startErr,
		s.Stop(ctx, app, execution.Session{}),
		s.Restart(ctx, app),
		s.Update(ctx, app),
	} {
		if !errors.Is(err, execution.ErrNotImplemented) {
			t.Errorf("got %v, want ErrNotImplemented", err)
		}
	}
}

func TestStrategyMetadata(t *testing.T) {
	s := &fakeStrategy{name: "fake", runtime: models.RuntimeNode}

	meta := s.Metadata()
	if meta.Name != "fake" {
		t.Errorf("Name = %q, want %q", meta.Name, "fake")
	}
	if len(meta.Capabilities) != 1 || meta.Capabilities[0] != execution.CapabilityStart {
		t.Errorf("Capabilities = %v", meta.Capabilities)
	}
}

func TestStrategyCanHandle(t *testing.T) {
	s := &fakeStrategy{name: "fake", runtime: models.RuntimeNode}

	nodeApp := models.NewApplication("node-app", models.ApplicationTypeBackend)
	nodeApp.Runtime = models.RuntimeNode
	if !s.CanHandle(nodeApp) {
		t.Error("expected CanHandle to be true for a matching runtime")
	}

	pythonApp := models.NewApplication("python-app", models.ApplicationTypeBackend)
	pythonApp.Runtime = models.RuntimePython
	if s.CanHandle(pythonApp) {
		t.Error("expected CanHandle to be false for a non-matching runtime")
	}
}
