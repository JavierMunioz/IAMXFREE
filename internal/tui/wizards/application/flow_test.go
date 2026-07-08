package application_test

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
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

func pressEsc(t *testing.T, m wizard.Model) (wizard.Model, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
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

// TestFullApplicationWizardFlowPreFillsFromAnalysis drives the whole wizard
// pressing Enter through every pre-filled step without typing anything for
// them, proving the analysis proposal actually reaches and fills Name,
// Type, Framework, Runtime, PackageManager and the three commands — while
// Port (required, never pre-filled) still has to be typed, and Domain/
// Repository (optional, never pre-filled) are left blank.
func TestFullApplicationWizardFlowPreFillsFromAnalysis(t *testing.T) {
	setup := &fakeSetupService{proposal: services.ApplicationSetupProposal{
		ProjectType:    "node",
		SuggestedName:  "my-api",
		Type:           "backend",
		Framework:      "express",
		Runtime:        "node",
		PackageManager: "npm",
		InstallCommand: "npm install",
		BuildCommand:   "",
		StartCommand:   "npm start",
	}}

	m := wizard.New("New application", application.Steps(setup))

	m = typeText(m, "/srv/apps/my-api")
	m, cmd := pressEnter(t, m) // path -> analysis (runs inspection)
	if cmd != nil {
		t.Fatal("did not expect completion yet (path)")
	}

	m, cmd = pressEnter(t, m) // analysis -> name (pre-filled: "my-api")
	if cmd != nil {
		t.Fatal("did not expect completion yet (analysis)")
	}

	m, cmd = pressEnter(t, m) // name -> type (pre-selected: backend)
	if cmd != nil {
		t.Fatal("did not expect completion yet (name)")
	}

	m, cmd = pressEnter(t, m) // type -> framework (pre-selected: express)
	if cmd != nil {
		t.Fatal("did not expect completion yet (type)")
	}

	m, cmd = pressEnter(t, m) // framework -> runtime (pre-selected: node)
	if cmd != nil {
		t.Fatal("did not expect completion yet (framework)")
	}

	m, cmd = pressEnter(t, m) // runtime -> package manager (pre-filled: npm)
	if cmd != nil {
		t.Fatal("did not expect completion yet (runtime)")
	}

	m, cmd = pressEnter(t, m) // package manager -> install command
	if cmd != nil {
		t.Fatal("did not expect completion yet (package manager)")
	}

	m, cmd = pressEnter(t, m) // install command -> build command (blank, optional)
	if cmd != nil {
		t.Fatal("did not expect completion yet (install command)")
	}

	m, cmd = pressEnter(t, m) // build command -> start command
	if cmd != nil {
		t.Fatal("did not expect completion yet (build command)")
	}

	m, cmd = pressEnter(t, m) // start command -> port
	if cmd != nil {
		t.Fatal("did not expect completion yet (start command)")
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

	m, cmd = pressEnter(t, m) // repo_url (blank, optional) -> pre_deploy_hook
	if cmd != nil {
		t.Fatal("did not expect completion yet (repo_url)")
	}

	m, cmd = pressEnter(t, m) // pre_deploy_hook (blank, optional) -> post_deploy_hook
	if cmd != nil {
		t.Fatal("did not expect completion yet (pre_deploy_hook)")
	}

	m, cmd = pressEnter(t, m) // post_deploy_hook (blank, optional) -> confirm
	if cmd != nil {
		t.Fatal("did not expect completion yet (post_deploy_hook)")
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
		t.Errorf("Name = %q, want %q (pre-filled)", draft.Name, "my-api")
	}
	if draft.Type != "backend" {
		t.Errorf("Type = %q, want %q (pre-filled)", draft.Type, "backend")
	}
	if draft.Framework != "express" {
		t.Errorf("Framework = %q, want %q (pre-filled)", draft.Framework, "express")
	}
	if draft.Runtime != "node" {
		t.Errorf("Runtime = %q, want %q (pre-filled)", draft.Runtime, "node")
	}
	if draft.PackageManager != "npm" {
		t.Errorf("PackageManager = %q, want %q (pre-filled)", draft.PackageManager, "npm")
	}
	if draft.InstallCommand != "npm install" {
		t.Errorf("InstallCommand = %q, want %q (pre-filled)", draft.InstallCommand, "npm install")
	}
	if draft.BuildCommand != "" {
		t.Errorf("BuildCommand = %q, want empty (never invented)", draft.BuildCommand)
	}
	if draft.StartCommand != "npm start" {
		t.Errorf("StartCommand = %q, want %q (pre-filled)", draft.StartCommand, "npm start")
	}
	if draft.Path != "/srv/apps/my-api" {
		t.Errorf("Path = %q, want %q", draft.Path, "/srv/apps/my-api")
	}
	if draft.Port != 3000 {
		t.Errorf("Port = %d, want 3000 (never pre-filled, always typed)", draft.Port)
	}
	if draft.Domain != "" || draft.RepoURL != "" {
		t.Errorf("expected optional, non-pre-filled fields to stay empty, got Domain=%q RepoURL=%q", draft.Domain, draft.RepoURL)
	}
}

// TestUserEditOverridesPreFilledValue proves a pre-filled field can still
// be freely changed: the user always has the final say.
func TestUserEditOverridesPreFilledValue(t *testing.T) {
	setup := &fakeSetupService{proposal: services.ApplicationSetupProposal{
		SuggestedName: "detected-name",
	}}

	m := wizard.New("New application", application.Steps(setup))

	m = typeText(m, "/srv/apps/my-api")
	m, _ = pressEnter(t, m) // path -> analysis
	m, _ = pressEnter(t, m) // analysis -> name (pre-filled "detected-name")

	// Clear the pre-filled value and type something else.
	for range "detected-name" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = updated.(wizard.Model)
	}
	m = typeText(m, "my-own-name")

	m, cmd := pressEnter(t, m) // name -> type
	if cmd != nil {
		t.Fatal("did not expect completion yet")
	}

	for _, step := range []string{"type", "framework", "runtime", "package_manager", "install", "build", "start"} {
		m, cmd = pressEnter(t, m)
		if cmd != nil {
			t.Fatalf("did not expect completion yet (%s)", step)
		}
	}

	m = typeText(m, "3000")
	m, cmd = pressEnter(t, m) // port -> domain
	if cmd != nil {
		t.Fatal("did not expect completion yet (port)")
	}
	m, cmd = pressEnter(t, m) // domain -> repo_url
	if cmd != nil {
		t.Fatal("did not expect completion yet (domain)")
	}
	m, cmd = pressEnter(t, m) // repo_url -> pre_deploy_hook
	if cmd != nil {
		t.Fatal("did not expect completion yet (repo_url)")
	}
	m, cmd = pressEnter(t, m) // pre_deploy_hook -> post_deploy_hook
	if cmd != nil {
		t.Fatal("did not expect completion yet (pre_deploy_hook)")
	}
	m, cmd = pressEnter(t, m) // post_deploy_hook -> confirm
	if cmd != nil {
		t.Fatal("did not expect completion yet (post_deploy_hook)")
	}
	_, cmd = pressEnter(t, m) // confirm -> completes
	if cmd == nil {
		t.Fatal("expected the wizard to complete")
	}

	msg, ok := cmd().(wizard.CompletedMsg)
	if !ok {
		t.Fatalf("expected CompletedMsg, got %T", cmd())
	}
	if got := msg.Result.Get(application.KeyName); got != "my-own-name" {
		t.Fatalf("Name = %q, want %q (user edit preserved)", got, "my-own-name")
	}
}

// TestAnalysisFailureBlocksContinueAndAllowsGoingBack proves a hard
// inspection failure (e.g. an unreadable path) keeps the wizard on the
// analysis screen until the user goes back and fixes the path.
func TestAnalysisFailureBlocksContinueAndAllowsGoingBack(t *testing.T) {
	setup := &fakeSetupService{err: errors.New("permission denied")}
	m := wizard.New("New application", application.Steps(setup))

	m = typeText(m, "/no/access")
	m, _ = pressEnter(t, m) // path -> analysis (inspection fails)

	if !strings.Contains(m.View(), "permission denied") {
		t.Fatalf("expected the analysis view to show the error, got:\n%s", m.View())
	}

	m, cmd := pressEnter(t, m) // enter must NOT advance past a failed analysis
	if cmd != nil {
		t.Fatal("did not expect completion")
	}
	if !strings.Contains(m.View(), "permission denied") {
		t.Fatal("expected to still be on the analysis screen after a blocked enter")
	}

	m, cmd = pressEsc(t, m) // esc goes back to path
	if cmd != nil {
		t.Fatal("did not expect a command when going back")
	}
	if !strings.Contains(m.View(), "Local project path") {
		t.Fatalf("expected esc to return to the path step, got:\n%s", m.View())
	}
}
