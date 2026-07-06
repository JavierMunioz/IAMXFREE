package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
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

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
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

func TestRootModelCancelledWizardReturnsToSplash(t *testing.T) {
	m := NewRootModel(&fakeApplicationService{})
	m.screen = screenWizard

	updated, _ := m.Update(wizard.CancelledMsg{})
	m = updated.(RootModel)

	if m.screen != screenSplash {
		t.Fatalf("screen = %v, want screenSplash", m.screen)
	}
	if m.status == "" {
		t.Fatal("expected a status message after cancelling")
	}
}

func TestRootModelCompletedWizardRegistersApplication(t *testing.T) {
	m := NewRootModel(&fakeApplicationService{})
	m.screen = screenWizard

	updated, cmd := m.Update(wizard.CompletedMsg{Result: validResult()})
	m = updated.(RootModel)

	if m.screen != screenSplash {
		t.Fatalf("screen = %v, want screenSplash", m.screen)
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
	if m.status == "" {
		t.Fatal("expected a success status message")
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
	if m.statusErr == nil {
		t.Fatal("expected statusErr to be set")
	}
}
