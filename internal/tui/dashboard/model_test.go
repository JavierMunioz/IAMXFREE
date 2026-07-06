package dashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

type fakeService struct {
	apps    []*models.Application
	listErr error
	health  map[string]services.ExecutionHealth
}

func (f *fakeService) Register(context.Context, *models.Application) error { return nil }
func (f *fakeService) Get(context.Context, string) (*models.Application, error) {
	return nil, nil
}
func (f *fakeService) List(context.Context) ([]*models.Application, error) {
	return f.apps, f.listErr
}
func (f *fakeService) UpdateConfig(context.Context, string, models.DeploymentConfig) (*models.Application, error) {
	return nil, nil
}
func (f *fakeService) ChangeStatus(context.Context, string, models.ApplicationStatus) (*models.Application, error) {
	return nil, nil
}
func (f *fakeService) Remove(context.Context, string) error { return nil }
func (f *fakeService) ResolveExecutionStrategy(context.Context, string) (execution.Metadata, error) {
	return execution.Metadata{}, execution.ErrNoStrategyFound
}
func (f *fakeService) CheckExecutionHealth(_ context.Context, id string) (services.ExecutionHealth, error) {
	health, ok := f.health[id]
	if !ok {
		return services.ExecutionHealth{}, execution.ErrNoStrategyFound
	}
	return health, nil
}

func newApp(name string, createdAt time.Time) *models.Application {
	app := models.NewApplication(name, models.ApplicationTypeAPI)
	app.CreatedAt = createdAt
	return app
}

func TestNewDefaults(t *testing.T) {
	m := New(&fakeService{})
	if m.width != 80 || m.height != 24 {
		t.Fatalf("width/height = %d/%d, want 80/24", m.width, m.height)
	}
	if !m.loading {
		t.Fatal("expected a new dashboard to start in a loading state")
	}
}

func TestReloadSortsByCreatedAt(t *testing.T) {
	now := time.Now()
	svc := &fakeService{apps: []*models.Application{
		newApp("second", now.Add(time.Minute)),
		newApp("first", now),
	}}
	m := New(svc)

	msg := m.Reload()()
	loaded, ok := msg.(appsLoadedMsg)
	if !ok {
		t.Fatalf("expected appsLoadedMsg, got %T", msg)
	}
	if len(loaded.apps) != 2 || loaded.apps[0].Name != "first" || loaded.apps[1].Name != "second" {
		t.Fatalf("unexpected order: %+v", loaded.apps)
	}
}

func TestReloadFailure(t *testing.T) {
	wantErr := errors.New("boom")
	svc := &fakeService{listErr: wantErr}
	m := New(svc)

	msg := m.Reload()()
	failed, ok := msg.(appsLoadFailedMsg)
	if !ok {
		t.Fatalf("expected appsLoadFailedMsg, got %T", msg)
	}
	if !errors.Is(failed.err, wantErr) {
		t.Fatalf("err = %v, want %v", failed.err, wantErr)
	}
}

func TestUpdateAppsLoadedClampsSelection(t *testing.T) {
	m := New(&fakeService{})
	m.selected = 5

	updated, _ := m.Update(appsLoadedMsg{apps: []*models.Application{
		newApp("only-one", time.Now()),
	}})
	m = updated.(Model)

	if m.selected != 0 {
		t.Fatalf("selected = %d, want 0", m.selected)
	}
	if m.loading {
		t.Fatal("expected loading to be false after apps loaded")
	}
}

func TestUpdateAppsLoadFailedSetsError(t *testing.T) {
	m := New(&fakeService{})
	wantErr := errors.New("boom")

	updated, _ := m.Update(appsLoadFailedMsg{err: wantErr})
	m = updated.(Model)

	if !errors.Is(m.loadErr, wantErr) {
		t.Fatalf("loadErr = %v, want %v", m.loadErr, wantErr)
	}
}

func TestUpdateWindowSize(t *testing.T) {
	m := New(&fakeService{})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)

	if m.width != 120 || m.height != 40 {
		t.Fatalf("width/height = %d/%d, want 120/40", m.width, m.height)
	}
}

func TestSetStatusAndSetErrorAreMutuallyExclusive(t *testing.T) {
	m := New(&fakeService{})

	m = m.SetStatus("done")
	if m.status != "done" || m.statusErr != nil {
		t.Fatalf("after SetStatus: status=%q statusErr=%v", m.status, m.statusErr)
	}

	m = m.SetError(errors.New("failed"))
	if m.statusErr == nil || m.status != "" {
		t.Fatalf("after SetError: status=%q statusErr=%v", m.status, m.statusErr)
	}
}

func TestAppsLoadedTriggersHealthLoad(t *testing.T) {
	app := newApp("my-api", time.Now())
	svc := &fakeService{health: map[string]services.ExecutionHealth{
		app.ID: {StrategyName: "Node.js (npm)", Healthy: true},
	}}
	m := New(svc)

	updated, cmd := m.Update(appsLoadedMsg{apps: []*models.Application{app}})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("expected appsLoadedMsg to also trigger a health-loading command")
	}

	msg := cmd()
	loaded, ok := msg.(healthLoadedMsg)
	if !ok {
		t.Fatalf("expected healthLoadedMsg, got %T", msg)
	}
	if loaded.healthByID[app.ID].StrategyName != "Node.js (npm)" {
		t.Fatalf("healthByID[%s] = %+v, want strategy Node.js (npm)", app.ID, loaded.healthByID[app.ID])
	}
}

func TestHealthLoadedUpdatesModel(t *testing.T) {
	m := New(&fakeService{})

	updated, _ := m.Update(healthLoadedMsg{healthByID: map[string]services.ExecutionHealth{
		"app-1": {StrategyName: "Node.js (npm)", Healthy: true},
	}})
	m = updated.(Model)

	if m.healthByID["app-1"].StrategyName != "Node.js (npm)" {
		t.Fatalf("healthByID = %+v, want an entry for app-1", m.healthByID)
	}
}

func TestLoadHealthCmdOmitsAppsWithoutAResolvableStrategy(t *testing.T) {
	app := newApp("my-api", time.Now())
	m := New(&fakeService{}) // no health configured -> ErrNoStrategyFound for every app
	m.apps = []*models.Application{app}

	msg := m.loadHealthCmd()()
	loaded, ok := msg.(healthLoadedMsg)
	if !ok {
		t.Fatalf("expected healthLoadedMsg, got %T", msg)
	}
	if _, present := loaded.healthByID[app.ID]; present {
		t.Fatal("expected an app with no resolvable strategy to be omitted from healthByID")
	}
}

func TestTickReschedulesItself(t *testing.T) {
	m := New(&fakeService{})
	_, cmd := m.Update(tickMsg(time.Now()))
	if cmd == nil {
		t.Fatal("expected tick to reschedule itself")
	}
	if _, ok := cmd().(tickMsg); !ok {
		t.Fatal("expected the rescheduled command to produce another tickMsg")
	}
}
