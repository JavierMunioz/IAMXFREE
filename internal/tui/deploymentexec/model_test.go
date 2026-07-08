package deploymentexec

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/deployment"
	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
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

func (fakeExecutionService) Install(context.Context, string) error { return nil }
func (fakeExecutionService) Build(context.Context, string) error   { return nil }
func (fakeExecutionService) Start(context.Context, string) (services.RunSession, error) {
	return services.RunSession{}, nil
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
	m := New(testEngine(&models.Application{ID: "app-1"}, nil), deployment.DeploymentPlan{ApplicationID: "app-1"})
	if m.Init() == nil {
		t.Fatal("expected Init() to return a command")
	}
}

func TestBuildFailedSetsBuildErr(t *testing.T) {
	m := New(testEngine(nil, nil), deployment.DeploymentPlan{})
	m, _ = update(t, m, buildFailedMsg{err: errors.New("boom")})

	if m.buildErr == nil {
		t.Fatal("expected buildErr to be set")
	}
}

func TestBuildAndStartCmdReturnsFailedMsgOnBuildError(t *testing.T) {
	m := New(testEngine(nil, errors.New("app not found")), deployment.DeploymentPlan{ApplicationID: "app-1"})

	msg := m.buildAndStartCmd()()
	if _, ok := msg.(buildFailedMsg); !ok {
		t.Fatalf("expected buildFailedMsg, got %T", msg)
	}
}

func TestBuildAndStartCmdReturnsStartedMsgOnSuccess(t *testing.T) {
	app := &models.Application{ID: "app-1", Name: "my-api"}
	m := New(testEngine(app, nil), deployment.DeploymentPlan{ApplicationID: "app-1"})

	msg := m.buildAndStartCmd()()
	started, ok := msg.(executionStartedMsg)
	if !ok {
		t.Fatalf("expected executionStartedMsg, got %T", msg)
	}
	if len(started.ops) != 6 {
		t.Fatalf("len(ops) = %d, want 6", len(started.ops))
	}
}

func TestExecutionStartedInitializesPendingResults(t *testing.T) {
	app := &models.Application{ID: "app-1", Name: "my-api"}
	m := New(testEngine(app, nil), deployment.DeploymentPlan{ApplicationID: "app-1"})

	ch := make(chan operations.OperationProgress)
	close(ch)
	ops := []operations.Operation{{Name: "one", Component: "test", Method: "one"}}

	m, cmd := update(t, m, executionStartedMsg{ops: ops, ch: ch})
	if !m.started {
		t.Fatal("expected started = true")
	}
	if len(m.results) != 1 || m.results[0].State != operations.StatePending {
		t.Fatalf("results = %+v, want a single Pending entry", m.results)
	}
	if cmd == nil {
		t.Fatal("expected a command to start reading progress")
	}
}

func TestOperationProgressUpdatesResultAndReissuesRead(t *testing.T) {
	app := &models.Application{ID: "app-1"}
	m := New(testEngine(app, nil), deployment.DeploymentPlan{ApplicationID: "app-1"})
	m.results = []operations.OperationResult{{State: operations.StatePending}}
	ch := make(chan operations.OperationProgress, 1)
	m.progressCh = ch

	progress := operations.OperationProgress{Index: 0, Total: 1, Result: operations.OperationResult{State: operations.StateSuccess}}
	m, cmd := update(t, m, operationProgressMsg{progress: progress})

	if m.results[0].State != operations.StateSuccess {
		t.Fatalf("results[0].State = %q, want %q", m.results[0].State, operations.StateSuccess)
	}
	if cmd == nil {
		t.Fatal("expected a command to keep reading progress")
	}
}

func TestProgressChannelClosedMarksFinished(t *testing.T) {
	app := &models.Application{ID: "app-1"}
	m := New(testEngine(app, nil), deployment.DeploymentPlan{ApplicationID: "app-1"})

	m, _ = update(t, m, progressChannelClosedMsg{})
	if !m.finished {
		t.Fatal("expected finished = true")
	}
}

func TestBackKeyEmitsBackMsg(t *testing.T) {
	m := New(testEngine(&models.Application{ID: "app-1"}, nil), deployment.DeploymentPlan{ApplicationID: "app-1"})

	_, cmd := m.handleKey(keyMsg("esc"))
	if cmd == nil {
		t.Fatal("expected esc to return a command")
	}
	if _, ok := cmd().(BackMsg); !ok {
		t.Fatalf("expected BackMsg, got %T", cmd())
	}
}

func TestReadProgressCmdReturnsClosedMsgWhenChannelCloses(t *testing.T) {
	ch := make(chan operations.OperationProgress)
	close(ch)

	msg := readProgressCmd(ch)()
	if _, ok := msg.(progressChannelClosedMsg); !ok {
		t.Fatalf("expected progressChannelClosedMsg, got %T", msg)
	}
}

func TestReadProgressCmdReturnsProgressMsg(t *testing.T) {
	ch := make(chan operations.OperationProgress, 1)
	ch <- operations.OperationProgress{Index: 0, Total: 1}

	msg := readProgressCmd(ch)()
	if _, ok := msg.(operationProgressMsg); !ok {
		t.Fatalf("expected operationProgressMsg, got %T", msg)
	}
}

// TestFullExecutionFlowDrainsToFinished exercises Init through a real
// (fast) background run to confirm the read loop actually reaches
// progressChannelClosedMsg without deadlocking.
func TestFullExecutionFlowDrainsToFinished(t *testing.T) {
	app := &models.Application{ID: "app-1", Name: "my-api"}
	m := New(testEngine(app, nil), deployment.DeploymentPlan{ApplicationID: "app-1"})

	cmd := m.Init()
	msg := cmd()
	started, ok := msg.(executionStartedMsg)
	if !ok {
		t.Fatalf("expected executionStartedMsg, got %T", msg)
	}

	m, cmd = update(t, m, started)

	deadline := time.After(2 * time.Second)
	for !m.finished {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for the run to finish")
		default:
		}
		next := cmd()
		m, cmd = update(t, m, next)
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
