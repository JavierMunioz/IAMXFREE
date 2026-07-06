package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/dashboard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizards/application"
)

type fakeApplicationService struct {
	registerErr error
}

func (f *fakeApplicationService) Register(context.Context, *models.Application) error {
	return f.registerErr
}
func (f *fakeApplicationService) Get(context.Context, string) (*models.Application, error) {
	return nil, nil
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
	m := NewRootModel(&fakeApplicationService{})

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
	m := NewRootModel(&fakeApplicationService{})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected a quit command")
	}
}

func TestRootModelCancelledWizardReturnsToDashboard(t *testing.T) {
	m := NewRootModel(&fakeApplicationService{})
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
	m := NewRootModel(&fakeApplicationService{})
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

func TestRootModelRegistrationFailureShowsError(t *testing.T) {
	wantErr := errors.New("boom")
	m := NewRootModel(&fakeApplicationService{registerErr: wantErr})
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
