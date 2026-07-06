package application_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizards/application"
)

func pressEnter(t *testing.T, m wizard.Model) (wizard.Model, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, ok := updated.(wizard.Model)
	if !ok {
		t.Fatalf("Update() returned %T, want wizard.Model", updated)
	}
	return next, cmd
}

func typeText(m wizard.Model, s string) wizard.Model {
	for _, r := range s {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(wizard.Model)
	}
	return m
}

func TestFullApplicationWizardFlow(t *testing.T) {
	m := wizard.New("New application", application.Steps())

	m = typeText(m, "my-api")
	m, cmd := pressEnter(t, m) // name -> type
	if cmd != nil {
		t.Fatal("did not expect completion yet (name)")
	}

	m, cmd = pressEnter(t, m) // type (default: frontend) -> framework
	if cmd != nil {
		t.Fatal("did not expect completion yet (type)")
	}

	m, cmd = pressEnter(t, m) // framework (default: react) -> runtime
	if cmd != nil {
		t.Fatal("did not expect completion yet (framework)")
	}

	m, cmd = pressEnter(t, m) // runtime (default: node) -> path
	if cmd != nil {
		t.Fatal("did not expect completion yet (runtime)")
	}

	m = typeText(m, "/srv/apps/my-api")
	m, cmd = pressEnter(t, m) // path -> port
	if cmd != nil {
		t.Fatal("did not expect completion yet (path)")
	}

	m = typeText(m, "3000")
	m, cmd = pressEnter(t, m) // port -> domain
	if cmd != nil {
		t.Fatal("did not expect completion yet (port)")
	}

	m, cmd = pressEnter(t, m) // domain (blank, optional) -> repo_url
	if cmd != nil {
		t.Fatal("did not expect completion yet (domain)")
	}

	m, cmd = pressEnter(t, m) // repo_url (blank, optional) -> confirm
	if cmd != nil {
		t.Fatal("did not expect completion yet (repo_url)")
	}

	_, cmd = pressEnter(t, m) // confirm -> completes
	if cmd == nil {
		t.Fatal("expected a completion command after confirming the summary")
	}

	msg, ok := cmd().(wizard.CompletedMsg)
	if !ok {
		t.Fatalf("expected CompletedMsg, got %T", cmd())
	}

	draft, err := application.DraftFromResult(msg.Result)
	if err != nil {
		t.Fatalf("DraftFromResult() error = %v", err)
	}
	if draft.Name != "my-api" {
		t.Errorf("Name = %q, want %q", draft.Name, "my-api")
	}
	if draft.Path != "/srv/apps/my-api" {
		t.Errorf("Path = %q, want %q", draft.Path, "/srv/apps/my-api")
	}
	if draft.Port != 3000 {
		t.Errorf("Port = %d, want 3000", draft.Port)
	}
	if draft.Domain != "" || draft.RepoURL != "" {
		t.Errorf("expected optional fields to stay empty, got Domain=%q RepoURL=%q", draft.Domain, draft.RepoURL)
	}
}
