// Package detail is the application detail screen: the place from which a
// single registered application is administered (diagnosed, started,
// stopped, and — later — configured). It depends only on
// services.ApplicationService and services.ExecutionService, never on
// internal/execution, internal/runtimehost or a repository directly.
package detail

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// BackMsg signals that the user asked to return to whatever hosts this
// screen (the dashboard). The detail view has no notion of "dashboard"
// beyond this — whatever hosts it decides what happens next.
type BackMsg struct{}

// OpenLogsMsg signals that the user asked to open the real-time logs view
// for the currently tracked session. The detail view has no notion of
// "logs view" beyond this — whatever hosts it decides what happens next.
type OpenLogsMsg struct {
	AppID   string
	Session services.RunSession
}

type appLoadedMsg struct {
	app *models.Application
}

type appLoadFailedMsg struct {
	err error
}

// healthLoadedMsg carries the resolved strategy's name and health. It is
// enrichment, not core data — see healthUnavailableMsg.
type healthLoadedMsg struct {
	strategyName string
	healthy      bool
}

// healthUnavailableMsg means no execution.Strategy resolved for this
// application (or checking its health failed) — not an error to show, just
// nothing to display in the Strategy/Health fields.
type healthUnavailableMsg struct{}

// gitStatusLoadedMsg carries the application's Git status, obtained on
// Init and on every explicit refresh (f5) — never automatically.
type gitStatusLoadedMsg struct {
	status services.GitStatus
}

// gitStatusUnavailableMsg means CheckGitStatus failed outright (as opposed
// to a normal "not a repository" result, which still arrives as
// gitStatusLoadedMsg with IsRepo == false) — not an error to show, just
// nothing to display in the Source Control panel.
type gitStatusUnavailableMsg struct{}

// activeSessionLoadedMsg carries whatever session ExecutionService is
// already tracking for this application, if any. This is how the screen
// learns about a session it did not itself start in this Model instance —
// e.g. one started the last time this screen was open, before the user
// went back to the dashboard and reopened it, which otherwise would look
// exactly like nothing was ever started.
type activeSessionLoadedMsg struct {
	session services.RunSession
	found   bool
}

// snapshotLoadedMsg carries a fresh RuntimeSnapshot, obtained only through
// an explicit refresh (see refresh.go) — never automatically.
type snapshotLoadedMsg struct {
	snapshot services.RuntimeSnapshot
}

// snapshotFailedMsg means observing the session's runtime state failed.
// It never clears an already-displayed snapshot — the metrics panel keeps
// showing the last known values alongside the error.
type snapshotFailedMsg struct {
	err error
}

// Model is the application detail screen.
type Model struct {
	appService       services.ApplicationService
	executionService services.ExecutionService

	appID string
	app   *models.Application

	strategyName string
	healthy      bool
	healthKnown  bool

	gitStatus    services.GitStatus
	hasGitStatus bool

	session    services.RunSession
	hasSession bool

	snapshot    services.RuntimeSnapshot
	hasSnapshot bool

	status    string
	statusErr error

	loading bool
	loadErr error

	width  int
	height int
}

// New builds the detail view for the application identified by appID.
func New(appService services.ApplicationService, executionService services.ExecutionService, appID string) Model {
	return Model{
		appService:       appService,
		executionService: executionService,
		appID:            appID,
		width:            80,
		height:           24,
		loading:          true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadAppCmd(), m.loadHealthCmd(), m.loadActiveSessionCmd(), m.loadGitStatusCmd())
}

func (m Model) loadAppCmd() tea.Cmd {
	service := m.appService
	id := m.appID
	return func() tea.Msg {
		app, err := service.Get(context.Background(), id)
		if err != nil {
			return appLoadFailedMsg{err: err}
		}
		return appLoadedMsg{app: app}
	}
}

// loadHealthCmd re-resolves the application's execution strategy and its
// health. Used both on Init and on an explicit refresh.
func (m Model) loadHealthCmd() tea.Cmd {
	service := m.appService
	id := m.appID
	return func() tea.Msg {
		health, err := service.CheckExecutionHealth(context.Background(), id)
		if err != nil {
			return healthUnavailableMsg{}
		}
		return healthLoadedMsg{strategyName: health.StrategyName, healthy: health.Healthy}
	}
}

// loadGitStatusCmd re-checks the application's Git repository status. Used
// both on Init and on an explicit refresh.
func (m Model) loadGitStatusCmd() tea.Cmd {
	service := m.appService
	id := m.appID
	return func() tea.Msg {
		status, err := service.CheckGitStatus(context.Background(), id)
		if err != nil {
			return gitStatusUnavailableMsg{}
		}
		return gitStatusLoadedMsg{status: status}
	}
}

// loadActiveSessionCmd asks ExecutionService whether it's already tracking
// a session for this application — populated in memory by a previous
// Start/RefreshSession call, possibly from an earlier time this screen was
// open. It performs no I/O (ActiveSession is a pure in-memory lookup), so
// this never fails.
func (m Model) loadActiveSessionCmd() tea.Cmd {
	service := m.executionService
	id := m.appID
	return func() tea.Msg {
		session, ok := service.ActiveSession(id)
		return activeSessionLoadedMsg{session: session, found: ok}
	}
}

// SetStatus records a transient, non-error status message to show in the
// detail view's status line.
func (m Model) SetStatus(text string) Model {
	m.status = text
	m.statusErr = nil
	return m
}

// SetError records an error to show in the detail view's status line.
func (m Model) SetError(err error) Model {
	m.statusErr = err
	m.status = ""
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case appLoadedMsg:
		m.app = msg.app
		m.loading = false
		m.loadErr = nil
		return m, nil

	case appLoadFailedMsg:
		m.loading = false
		m.loadErr = msg.err
		return m, nil

	case healthLoadedMsg:
		m.strategyName = msg.strategyName
		m.healthy = msg.healthy
		m.healthKnown = true
		return m, nil

	case healthUnavailableMsg:
		m.healthKnown = false
		m.strategyName = ""
		return m, nil

	case gitStatusLoadedMsg:
		m.gitStatus = msg.status
		m.hasGitStatus = true
		return m, nil

	case gitStatusUnavailableMsg:
		m.hasGitStatus = false
		m.gitStatus = services.GitStatus{}
		return m, nil

	case activeSessionLoadedMsg:
		if msg.found {
			m.session = msg.session
			m.hasSession = true
		}
		return m, nil

	case startedMsg:
		m.session = msg.session
		m.hasSession = true
		m.hasSnapshot = false
		m.snapshot = services.RuntimeSnapshot{}
		m = m.SetStatus("Application started.")
		return m, nil

	case startFailedMsg:
		m = m.SetError(msg.err)
		return m, nil

	case stoppedMsg:
		m.hasSession = false
		m.session = services.RunSession{}
		m.hasSnapshot = false
		m.snapshot = services.RuntimeSnapshot{}
		m = m.SetStatus("Application stopped.")
		return m, nil

	case stopFailedMsg:
		m = m.SetError(msg.err)
		return m, nil

	case sessionRefreshedMsg:
		m.session = msg.session
		m = m.SetStatus("Refreshed.")
		return m, nil

	case sessionRefreshFailedMsg:
		m = m.SetError(msg.err)
		return m, nil

	case snapshotLoadedMsg:
		m.snapshot = msg.snapshot
		m.hasSnapshot = true
		return m, nil

	case snapshotFailedMsg:
		m = m.SetError(msg.err)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}
