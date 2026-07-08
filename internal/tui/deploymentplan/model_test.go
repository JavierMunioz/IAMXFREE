package deploymentplan

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/deployment"
	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// fakeAppService is a minimal services.ApplicationService test double.
type fakeAppService struct {
	app    *models.Application
	getErr error
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
func (f *fakeAppService) CheckExecutionHealth(context.Context, string) (services.ExecutionHealth, error) {
	return services.ExecutionHealth{}, execution.ErrNoStrategyFound
}
func (f *fakeAppService) CheckGitStatus(context.Context, string) (services.GitStatus, error) {
	return services.GitStatus{}, nil
}

// fakeExecutionService is a minimal services.ExecutionService test double.
type fakeExecutionService struct{}

func (fakeExecutionService) Install(context.Context, string) error {
	return execution.ErrNotImplemented
}
func (fakeExecutionService) Build(context.Context, string) error {
	return execution.ErrNotImplemented
}
func (fakeExecutionService) Start(context.Context, string) (services.RunSession, error) {
	return services.RunSession{}, execution.ErrNotImplemented
}
func (fakeExecutionService) Stop(context.Context, string, services.RunSession) error { return nil }
func (fakeExecutionService) RefreshSession(context.Context, string, services.RunSession) (services.RunSession, error) {
	return services.RunSession{}, nil
}
func (fakeExecutionService) OpenLogs(context.Context, string, services.RunSession) (services.LogStream, error) {
	return nil, nil
}
func (fakeExecutionService) Snapshot(context.Context, services.RunSession) (services.RuntimeSnapshot, error) {
	return services.RuntimeSnapshot{}, nil
}
func (fakeExecutionService) ActiveSession(string) (services.RunSession, bool) {
	return services.RunSession{}, false
}

func testEngine(app *models.Application, getErr error) *deployment.Engine {
	return deployment.NewEngine(
		&fakeAppService{app: app, getErr: getErr},
		fakeExecutionService{},
		git.NewManager(runtimehosttest.NewFakeHost()),
		nginx.NewManager(runtimehosttest.NewFakeHost()),
	)
}

func TestInitReturnsACommand(t *testing.T) {
	m := New(testEngine(&models.Application{ID: "app-1", Name: "my-api"}, nil), "app-1")
	if m.Init() == nil {
		t.Fatal("expected Init() to return a command")
	}
}

func TestPlanLoadedUpdatesModel(t *testing.T) {
	m := New(testEngine(&models.Application{ID: "app-1", Name: "my-api"}, nil), "app-1")

	plan := deployment.DeploymentPlan{ApplicationID: "app-1", ApplicationName: "my-api"}
	m, _ = update(t, m, planLoadedMsg{plan: plan})

	if !m.loaded || m.plan.ApplicationName != "my-api" {
		t.Fatalf("expected the plan to be loaded, got loaded=%v plan=%+v", m.loaded, m.plan)
	}
}

func TestPlanLoadFailedSetsLoadErr(t *testing.T) {
	m := New(testEngine(nil, nil), "app-1")

	m, _ = update(t, m, planLoadFailedMsg{err: errors.New("boom")})
	if m.loadErr == nil {
		t.Fatal("expected loadErr to be set")
	}
}

func TestLoadPlanCmdReturnsLoadedMsgOnSuccess(t *testing.T) {
	app := &models.Application{ID: "app-1", Name: "my-api"}
	m := New(testEngine(app, nil), "app-1")

	msg := m.loadPlanCmd()()
	loaded, ok := msg.(planLoadedMsg)
	if !ok {
		t.Fatalf("expected planLoadedMsg, got %T", msg)
	}
	if loaded.plan.ApplicationName != "my-api" {
		t.Fatalf("plan.ApplicationName = %q, want %q", loaded.plan.ApplicationName, "my-api")
	}
}

func TestLoadPlanCmdReturnsFailedMsgOnError(t *testing.T) {
	m := New(testEngine(nil, errors.New("app not found")), "app-1")

	msg := m.loadPlanCmd()()
	if _, ok := msg.(planLoadFailedMsg); !ok {
		t.Fatalf("expected planLoadFailedMsg, got %T", msg)
	}
}

func TestBackKeyEmitsBackMsg(t *testing.T) {
	m := New(testEngine(&models.Application{ID: "app-1"}, nil), "app-1")

	_, cmd := m.handleKey(keyMsg("esc"))
	if cmd == nil {
		t.Fatal("expected esc to return a command")
	}
	if _, ok := cmd().(BackMsg); !ok {
		t.Fatalf("expected BackMsg, got %T", cmd())
	}
}

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
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}
