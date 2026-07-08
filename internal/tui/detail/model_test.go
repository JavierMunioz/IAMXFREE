package detail

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// fakeAppService is a minimal services.ApplicationService test double.
type fakeAppService struct {
	app       *models.Application
	getErr    error
	health    services.ExecutionHealth
	healthErr error
	git       services.GitStatus
	gitErr    error
}

func (f *fakeAppService) Register(context.Context, *models.Application) error { return nil }
func (f *fakeAppService) Get(context.Context, string) (*models.Application, error) {
	return f.app, f.getErr
}
func (f *fakeAppService) List(context.Context) ([]*models.Application, error) { return nil, nil }
func (f *fakeAppService) UpdateConfig(context.Context, string, models.DeploymentConfig) (*models.Application, error) {
	return nil, nil
}
func (f *fakeAppService) ChangeStatus(context.Context, string, models.ApplicationStatus) (*models.Application, error) {
	return nil, nil
}
func (f *fakeAppService) Remove(context.Context, string) error { return nil }
func (f *fakeAppService) ResolveExecutionStrategy(context.Context, string) (execution.Metadata, error) {
	return execution.Metadata{}, execution.ErrNoStrategyFound
}
func (f *fakeAppService) CheckGitStatus(context.Context, string) (services.GitStatus, error) {
	return f.git, f.gitErr
}
func (f *fakeAppService) CheckExecutionHealth(context.Context, string) (services.ExecutionHealth, error) {
	return f.health, f.healthErr
}

// fakeExecutionService is a minimal services.ExecutionService test double.
type fakeExecutionService struct {
	startSession services.RunSession
	startErr     error
	stopErr      error
	refreshed    services.RunSession
	refreshErr   error
	logStream    services.LogStream
	logsErr      error
	snapshot     services.RuntimeSnapshot
	snapshotErr  error

	activeSession    services.RunSession
	hasActiveSession bool
}

func (f *fakeExecutionService) Install(context.Context, string) error { return nil }
func (f *fakeExecutionService) Build(context.Context, string) error   { return nil }
func (f *fakeExecutionService) Start(context.Context, string) (services.RunSession, error) {
	return f.startSession, f.startErr
}
func (f *fakeExecutionService) Stop(context.Context, string, services.RunSession) error {
	return f.stopErr
}
func (f *fakeExecutionService) OpenLogs(context.Context, string, services.RunSession) (services.LogStream, error) {
	return f.logStream, f.logsErr
}
func (f *fakeExecutionService) RefreshSession(context.Context, string, services.RunSession) (services.RunSession, error) {
	return f.refreshed, f.refreshErr
}
func (f *fakeExecutionService) Snapshot(context.Context, services.RunSession) (services.RuntimeSnapshot, error) {
	return f.snapshot, f.snapshotErr
}
func (f *fakeExecutionService) ActiveSession(string) (services.RunSession, bool) {
	return f.activeSession, f.hasActiveSession
}

func newTestApp() *models.Application {
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	app.Framework = models.FrameworkExpress
	app.Config.InternalPort = 3000
	app.Config.Domain = "my-api.example.com"
	return app
}

// update sends msg through m.Update and asserts the result is still a
// detail.Model, returning it unwrapped for convenient chaining in tests.
func update(t *testing.T, m Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(msg)
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update() returned %T, want Model", updated)
	}
	return next, cmd
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "f5":
		return tea.KeyMsg{Type: tea.KeyF5}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestInitReturnsACommand(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	if cmd := m.Init(); cmd == nil {
		t.Fatal("expected Init to return a command")
	}
}

func TestAppLoadedUpdatesModel(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m, _ = update(t, m, appLoadedMsg{app: app})
	if m.app == nil || m.app.Name != "my-api" {
		t.Fatalf("expected app to be loaded, got %+v", m.app)
	}
	if m.loading {
		t.Fatal("expected loading to be false after appLoadedMsg")
	}
}

func TestAppLoadFailedSetsLoadErr(t *testing.T) {
	wantErr := errors.New("not found")
	m := New(&fakeAppService{getErr: wantErr}, &fakeExecutionService{}, "missing-id")

	m, _ = update(t, m, appLoadFailedMsg{err: wantErr})
	if !errors.Is(m.loadErr, wantErr) {
		t.Fatalf("loadErr = %v, want %v", m.loadErr, wantErr)
	}
	if !strings.Contains(m.View(), "not found") {
		t.Fatalf("expected the view to show the load error, got:\n%s", m.View())
	}
}

func TestHealthLoadedUpdatesModel(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m, _ = update(t, m, healthLoadedMsg{strategyName: "Node.js (npm)", healthy: true})
	if !m.healthKnown || m.strategyName != "Node.js (npm)" || !m.healthy {
		t.Fatalf("expected health to be loaded, got known=%v name=%q healthy=%v", m.healthKnown, m.strategyName, m.healthy)
	}
}

func TestHealthUnavailableClearsStrategy(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m, _ = update(t, m, healthLoadedMsg{strategyName: "Node.js (npm)", healthy: true})
	m, _ = update(t, m, healthUnavailableMsg{})

	if m.healthKnown || m.strategyName != "" {
		t.Fatalf("expected health to be cleared, got known=%v name=%q", m.healthKnown, m.strategyName)
	}
}

func TestGitStatusLoadedUpdatesModel(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m, _ = update(t, m, gitStatusLoadedMsg{status: services.GitStatus{IsRepo: true, Branch: "main"}})
	if !m.hasGitStatus || !m.gitStatus.IsRepo || m.gitStatus.Branch != "main" {
		t.Fatalf("expected git status to be loaded, got hasGitStatus=%v status=%+v", m.hasGitStatus, m.gitStatus)
	}
}

func TestGitStatusUnavailableClearsStatus(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m, _ = update(t, m, gitStatusLoadedMsg{status: services.GitStatus{IsRepo: true, Branch: "main"}})
	m, _ = update(t, m, gitStatusUnavailableMsg{})

	if m.hasGitStatus {
		t.Fatalf("expected git status to be cleared, got hasGitStatus=%v", m.hasGitStatus)
	}
}

func TestLoadGitStatusCmdReturnsLoadedMsg(t *testing.T) {
	app := newTestApp()
	appSvc := &fakeAppService{app: app, git: services.GitStatus{IsRepo: true, Branch: "main"}}
	m := New(appSvc, &fakeExecutionService{}, app.ID)

	msg := m.loadGitStatusCmd()()
	loaded, ok := msg.(gitStatusLoadedMsg)
	if !ok {
		t.Fatalf("expected gitStatusLoadedMsg, got %T", msg)
	}
	if !loaded.status.IsRepo || loaded.status.Branch != "main" {
		t.Fatalf("status = %+v, want IsRepo=true Branch=main", loaded.status)
	}
}

func TestLoadGitStatusCmdReturnsUnavailableOnError(t *testing.T) {
	app := newTestApp()
	appSvc := &fakeAppService{app: app, gitErr: errors.New("boom")}
	m := New(appSvc, &fakeExecutionService{}, app.ID)

	msg := m.loadGitStatusCmd()()
	if _, ok := msg.(gitStatusUnavailableMsg); !ok {
		t.Fatalf("expected gitStatusUnavailableMsg, got %T", msg)
	}
}

func TestViewRendersTopPanelFields(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m, _ = update(t, m, appLoadedMsg{app: app})
	m, _ = update(t, m, healthLoadedMsg{strategyName: "Node.js (npm)", healthy: true})

	view := m.View()
	for _, want := range []string{
		"my-api", string(models.ApplicationTypeAPI), string(models.FrameworkExpress),
		string(models.RuntimeNode), "Node.js (npm)", "3000", "my-api.example.com",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestViewShowsLoadingBeforeAppArrives(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	if !strings.Contains(m.View(), "Loading") {
		t.Fatalf("expected a loading placeholder, got:\n%s", m.View())
	}
}

func TestQuitReturnsCommand(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	_, cmd := m.handleKey(keyMsg("q"))
	if cmd == nil {
		t.Fatal("expected quit to return a command")
	}
}

func TestInitHydratesAnAlreadyTrackedSession(t *testing.T) {
	app := newTestApp()
	execSvc := &fakeExecutionService{
		activeSession:    services.RunSession{PID: 4242, Status: "running"},
		hasActiveSession: true,
	}
	m := New(&fakeAppService{app: app}, execSvc, app.ID)

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a command")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected a tea.BatchMsg, got %T", cmd())
	}

	for _, batched := range batch {
		if msg, ok := batched().(activeSessionLoadedMsg); ok {
			m, _ = update(t, m, msg)
		}
	}

	if !m.hasSession || m.session.PID != 4242 {
		t.Fatalf("expected the already-tracked session to be hydrated, got hasSession=%v session=%+v", m.hasSession, m.session)
	}
}

func TestInitLeavesNoSessionWhenNoneTracked(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m, _ = update(t, m, activeSessionLoadedMsg{found: false})
	if m.hasSession {
		t.Fatal("expected hasSession to stay false when ExecutionService tracks nothing")
	}
}

func TestBackEmitsBackMsg(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	for _, key := range []string{"esc", "b"} {
		_, cmd := m.handleKey(keyMsg(key))
		if cmd == nil {
			t.Fatalf("expected %q to return a command", key)
		}
		if _, ok := cmd().(BackMsg); !ok {
			t.Fatalf("expected %q to emit BackMsg, got %T", key, cmd())
		}
	}
}
