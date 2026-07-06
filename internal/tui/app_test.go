package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/dashboard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/detail"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/logs"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizards/application"
)

type fakeApplicationService struct {
	registerErr error
	app         *models.Application
}

func (f *fakeApplicationService) Register(context.Context, *models.Application) error {
	return f.registerErr
}
func (f *fakeApplicationService) Get(context.Context, string) (*models.Application, error) {
	return f.app, nil
}
func (f *fakeApplicationService) List(context.Context) ([]*models.Application, error) {
	return nil, nil
}
func (f *fakeApplicationService) UpdateConfig(context.Context, string, models.DeploymentConfig) (*models.Application, error) {
	return nil, nil
}
func (f *fakeApplicationService) ChangeStatus(context.Context, string, models.ApplicationStatus) (*models.Application, error) {
	return nil, nil
}
func (f *fakeApplicationService) Remove(context.Context, string) error {
	return nil
}
func (f *fakeApplicationService) ResolveExecutionStrategy(context.Context, string) (execution.Metadata, error) {
	return execution.Metadata{}, execution.ErrNoStrategyFound
}
func (f *fakeApplicationService) CheckExecutionHealth(context.Context, string) (services.ExecutionHealth, error) {
	return services.ExecutionHealth{}, execution.ErrNoStrategyFound
}

type fakeExecutionService struct{}

func (fakeExecutionService) Start(context.Context, string) (services.RunSession, error) {
	return services.RunSession{}, execution.ErrNotImplemented
}
func (fakeExecutionService) Stop(context.Context, string, services.RunSession) error {
	return execution.ErrNotImplemented
}
func (fakeExecutionService) RefreshSession(context.Context, string, services.RunSession) (services.RunSession, error) {
	return services.RunSession{}, execution.ErrNotImplemented
}
func (fakeExecutionService) OpenLogs(context.Context, string, services.RunSession) (services.LogStream, error) {
	return nil, execution.ErrNotImplemented
}
func (fakeExecutionService) Snapshot(context.Context, services.RunSession) (services.RuntimeSnapshot, error) {
	return services.RuntimeSnapshot{}, execution.ErrNotImplemented
}

type fakeApplicationSetupService struct{}

func (fakeApplicationSetupService) Inspect(context.Context, string) (services.ApplicationSetupProposal, error) {
	return services.ApplicationSetupProposal{}, nil
}

func validResult() wizard.Result {
	return wizard.Result{Values: map[string]string{
		application.KeyName:      "my-api",
		application.KeyType:      "api",
		application.KeyFramework: "fastapi",
		application.KeyRuntime:   "python",
		application.KeyPath:      "/srv/apps/my-api",
		application.KeyPort:      "8000",
	}}
}

func TestRootModelPressingAOpensWizard(t *testing.T) {
	m := NewRootModel(&fakeApplicationService{}, fakeExecutionService{}, fakeApplicationSetupService{})

	// "a" is handled by the dashboard, which asks the root model to open
	// the wizard via a command rather than switching screens itself.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(RootModel)
	if cmd == nil {
		t.Fatal("expected a command requesting the wizard to open")
	}

	msg, ok := cmd().(dashboard.OpenWizardMsg)
	if !ok {
		t.Fatalf("expected dashboard.OpenWizardMsg, got %T", cmd())
	}

	updated, _ = m.Update(msg)
	m = updated.(RootModel)
	if m.screen != screenWizard {
		t.Fatalf("screen = %v, want screenWizard", m.screen)
	}
}

func TestRootModelQuits(t *testing.T) {
	m := NewRootModel(&fakeApplicationService{}, fakeExecutionService{}, fakeApplicationSetupService{})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected a quit command")
	}
}

func TestRootModelCancelledWizardReturnsToDashboard(t *testing.T) {
	m := NewRootModel(&fakeApplicationService{}, fakeExecutionService{}, fakeApplicationSetupService{})
	m.screen = screenWizard

	updated, _ := m.Update(wizard.CancelledMsg{})
	m = updated.(RootModel)

	if m.screen != screenDashboard {
		t.Fatalf("screen = %v, want screenDashboard", m.screen)
	}
	if !strings.Contains(m.dashboard.View(), "cancelled") {
		t.Fatal("expected the dashboard to show a cancellation status message")
	}
}

func TestRootModelCompletedWizardRegistersApplication(t *testing.T) {
	m := NewRootModel(&fakeApplicationService{}, fakeExecutionService{}, fakeApplicationSetupService{})
	m.screen = screenWizard

	updated, cmd := m.Update(wizard.CompletedMsg{Result: validResult()})
	m = updated.(RootModel)

	if m.screen != screenDashboard {
		t.Fatalf("screen = %v, want screenDashboard", m.screen)
	}
	if cmd == nil {
		t.Fatal("expected a registration command")
	}

	msg := cmd()
	registered, ok := msg.(applicationRegisteredMsg)
	if !ok {
		t.Fatalf("expected applicationRegisteredMsg, got %T", msg)
	}
	if registered.app.Name != "my-api" {
		t.Errorf("app.Name = %q, want %q", registered.app.Name, "my-api")
	}

	updated, _ = m.Update(registered)
	m = updated.(RootModel)
	if !strings.Contains(m.dashboard.View(), `"my-api" registered`) {
		t.Fatal("expected the dashboard to show a success status message")
	}
}

func TestRootModelOpenDetailMsgSwitchesToDetailScreen(t *testing.T) {
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	m := NewRootModel(&fakeApplicationService{app: app}, fakeExecutionService{}, fakeApplicationSetupService{})

	updated, cmd := m.Update(dashboard.OpenDetailMsg{AppID: app.ID})
	m = updated.(RootModel)
	if m.screen != screenDetail {
		t.Fatalf("screen = %v, want screenDetail", m.screen)
	}
	if cmd == nil {
		t.Fatal("expected opening the detail view to return its Init command")
	}
}

func TestRootModelBackMsgReturnsToDashboard(t *testing.T) {
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	m := NewRootModel(&fakeApplicationService{app: app}, fakeExecutionService{}, fakeApplicationSetupService{})
	m.screen = screenDetail

	updated, cmd := m.Update(detail.BackMsg{})
	m = updated.(RootModel)
	if m.screen != screenDashboard {
		t.Fatalf("screen = %v, want screenDashboard", m.screen)
	}
	if cmd == nil {
		t.Fatal("expected returning to the dashboard to trigger a reload")
	}
}

func TestRootModelOpenLogsMsgSwitchesToLogsScreen(t *testing.T) {
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	m := NewRootModel(&fakeApplicationService{app: app}, fakeExecutionService{}, fakeApplicationSetupService{})
	m.screen = screenDetail

	updated, cmd := m.Update(detail.OpenLogsMsg{AppID: app.ID, Session: services.RunSession{PID: 4242}})
	m = updated.(RootModel)
	if m.screen != screenLogs {
		t.Fatalf("screen = %v, want screenLogs", m.screen)
	}
	if cmd == nil {
		t.Fatal("expected opening the logs view to return its Init command")
	}
}

func TestRootModelLogsBackMsgReturnsToDetail(t *testing.T) {
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	m := NewRootModel(&fakeApplicationService{app: app}, fakeExecutionService{}, fakeApplicationSetupService{})
	m.screen = screenLogs

	updated, _ := m.Update(logs.BackMsg{})
	m = updated.(RootModel)
	if m.screen != screenDetail {
		t.Fatalf("screen = %v, want screenDetail", m.screen)
	}
}

func TestRootModelRegistrationFailureShowsError(t *testing.T) {
	wantErr := errors.New("boom")
	m := NewRootModel(&fakeApplicationService{registerErr: wantErr}, fakeExecutionService{}, fakeApplicationSetupService{})
	m.screen = screenWizard

	updated, cmd := m.Update(wizard.CompletedMsg{Result: validResult()})
	m = updated.(RootModel)

	msg := cmd()
	failed, ok := msg.(applicationRegistrationFailedMsg)
	if !ok {
		t.Fatalf("expected applicationRegistrationFailedMsg, got %T", msg)
	}

	updated, _ = m.Update(failed)
	m = updated.(RootModel)
	if !strings.Contains(m.dashboard.View(), wantErr.Error()) {
		t.Fatal("expected the dashboard to show the registration error")
	}
}
