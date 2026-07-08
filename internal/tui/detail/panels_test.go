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

func TestRenderMetricsPanelShowsNoMetricsYet(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	panel := m.renderMetricsPanel(60)
	if !strings.Contains(panel, "No metrics yet") {
		t.Errorf("expected a clear 'no metrics yet' message, got:\n%s", panel)
	}
}

func TestRenderMetricsPanelShowsAvailableValues(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSnapshot = true
	m.snapshot = services.RuntimeSnapshot{
		CPUPercent:     services.Metric{Value: 12.5, Available: true},
		MemoryRSSBytes: services.Metric{Value: 100 * 1024 * 1024, Available: true},
		MemoryVSZBytes: services.Metric{Available: false},
	}

	panel := m.renderMetricsPanel(60)
	for _, want := range []string{"12.5%", "100.0 MB", "n/a"} {
		if !strings.Contains(panel, want) {
			t.Errorf("metrics panel missing %q:\n%s", want, panel)
		}
	}
}

func TestRenderGitPanelShowsLoadingBeforeFirstLoad(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	panel := m.renderGitPanel(60)
	if !strings.Contains(panel, "Loading") {
		t.Errorf("expected a loading message before the first check, got:\n%s", panel)
	}
}

func TestRenderGitPanelShowsNoRepository(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasGitStatus = true
	m.gitStatus = services.GitStatus{IsRepo: false}

	panel := m.renderGitPanel(60)
	if !strings.Contains(panel, "No Git repository") {
		t.Errorf("expected a clear 'no repository' message, got:\n%s", panel)
	}
}

func TestRenderGitPanelShowsRepositoryDetails(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasGitStatus = true
	m.gitStatus = services.GitStatus{
		IsRepo:     true,
		Branch:     "main",
		CommitSHA:  "abc123d",
		Clean:      true,
		RemoteName: "origin",
	}

	panel := m.renderGitPanel(60)
	for _, want := range []string{"main", "abc123d", "clean", "origin"} {
		if !strings.Contains(panel, want) {
			t.Errorf("git panel missing %q:\n%s", want, panel)
		}
	}
}

func TestRenderGitPanelShowsDirtyWorkingTree(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasGitStatus = true
	m.gitStatus = services.GitStatus{IsRepo: true, Branch: "main", Clean: false, Modified: 2, Untracked: 1}

	panel := m.renderGitPanel(60)
	if !strings.Contains(panel, "dirty (2 modified, 1 untracked)") {
		t.Errorf("expected a dirty working tree summary, got:\n%s", panel)
	}
}

func TestRenderGitPanelShowsDetachedHead(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasGitStatus = true
	m.gitStatus = services.GitStatus{IsRepo: true, Detached: true}

	panel := m.renderGitPanel(60)
	if !strings.Contains(panel, "detached HEAD") {
		t.Errorf("expected a detached HEAD indicator, got:\n%s", panel)
	}
}

func TestViewAlwaysShowsSourceControlPanel(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m, _ = update(t, m, appLoadedMsg{app: app})

	if !strings.Contains(m.View(), "Source Control") {
		t.Fatalf("expected the Source Control panel regardless of session state, got:\n%s", m.View())
	}
}

func TestViewShowsMetricsPanelOnlyWithASession(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m, _ = update(t, m, appLoadedMsg{app: app})

	if strings.Contains(m.View(), "Metrics") {
		t.Fatalf("expected no metrics panel without a session, got:\n%s", m.View())
	}

	m.hasSession = true
	m.session = services.RunSession{PID: 4242, Status: "running"}
	if !strings.Contains(m.View(), "Metrics") {
		t.Fatalf("expected a metrics panel once a session exists, got:\n%s", m.View())
	}
}
