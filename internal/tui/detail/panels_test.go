package detail

import (
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestRenderTechnicalPanelShowsConfiguredFields(t *testing.T) {
	app := newTestApp()
	app.Source.LocalPath = "/srv/apps/my-api"
	app.Source.RepositoryURL = "https://github.com/user/my-api.git"
	app.Config.PackageManager = "npm"
	app.Config.InstallCommand = "npm install"
	app.Config.BuildCommand = "npm run build"
	app.Config.StartCommand = "npm start"

	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m, _ = update(t, m, appLoadedMsg{app: app})

	panel := m.renderTechnicalPanel(60)
	for _, want := range []string{
		"/srv/apps/my-api", "https://github.com/user/my-api.git",
		"npm", "npm install", "npm run build", "npm start",
	} {
		if !strings.Contains(panel, want) {
			t.Errorf("technical panel missing %q:\n%s", want, panel)
		}
	}
}

func TestRenderExecutionPanelShowsNoActiveSession(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m, _ = update(t, m, appLoadedMsg{app: app})

	panel := m.renderExecutionPanel(60)
	if !strings.Contains(panel, "No active session") {
		t.Errorf("expected a clear 'no active session' message, got:\n%s", panel)
	}
}

func TestRenderExecutionPanelShowsRunningSession(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m, _ = update(t, m, appLoadedMsg{app: app})
	m.hasSession = true
	m.session = services.RunSession{
		PID:        4242,
		Status:     "running",
		WorkingDir: "/srv/apps/my-api",
		Runtime:    app.Runtime,
	}

	panel := m.renderExecutionPanel(60)
	for _, want := range []string{"running", "4242", "/srv/apps/my-api", string(app.Runtime)} {
		if !strings.Contains(panel, want) {
			t.Errorf("execution panel missing %q:\n%s", want, panel)
		}
	}
}
