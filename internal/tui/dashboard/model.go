package dashboard

import (
	"context"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// OpenWizardMsg signals that the user asked to register a new application.
// The dashboard has no notion of "wizard" beyond this — whatever hosts it
// decides what happens next.
type OpenWizardMsg struct{}

// OpenDetailMsg signals that the user asked to open the detail view for one
// application. The dashboard has no notion of "detail view" beyond this —
// whatever hosts it decides what happens next.
type OpenDetailMsg struct {
	AppID string
}

type appsLoadedMsg struct {
	apps []*models.Application
}

type appsLoadFailedMsg struct {
	err error
}

// healthLoadedMsg carries execution health, keyed by application ID, for
// whichever applications have a resolvable execution.Strategy. An
// application missing from the map simply has no strategy yet — that's not
// an error, just nothing to show.
type healthLoadedMsg struct {
	healthByID map[string]services.ExecutionHealth
}

// sessionsLoadedMsg carries active sessions, keyed by application ID, for
// whichever applications ExecutionService is currently tracking a session
// for. An application missing from the map simply has none — that's not an
// error, just nothing to show.
type sessionsLoadedMsg struct {
	sessionByID map[string]services.RunSession
}

// gitLoadedMsg carries Git status, keyed by application ID, for every
// currently loaded application. An application missing from the map (or
// present with IsRepo == false) simply has no repository to show — that's
// not an error.
type gitLoadedMsg struct {
	gitByID map[string]services.GitStatus
}

type tickMsg time.Time

// Model is the dashboard screen. It depends only on services.ApplicationService
// and services.ExecutionService — never on a repository or internal/monitor
// directly — for every piece of application data it shows.
type Model struct {
	service          services.ApplicationService
	executionService services.ExecutionService

	apps        []*models.Application
	healthByID  map[string]services.ExecutionHealth
	sessionByID map[string]services.RunSession
	gitByID     map[string]services.GitStatus
	selected    int

	width  int
	height int

	status    string
	statusErr error

	loading bool
	loadErr error
}

// New builds a dashboard backed by service and executionService.
func New(service services.ApplicationService, executionService services.ExecutionService) Model {
	return Model{service: service, executionService: executionService, width: 80, height: 24, loading: true}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.Reload(), tickCmd())
}

// Reload re-fetches the application list from the service. Exported so a
// host (e.g. after a successful registration) can trigger a refresh without
// reaching into the dashboard's internals.
func (m Model) Reload() tea.Cmd {
	service := m.service
	return func() tea.Msg {
		apps, err := service.List(context.Background())
		if err != nil {
			return appsLoadFailedMsg{err: err}
		}
		sort.Slice(apps, func(i, j int) bool {
			return apps[i].CreatedAt.Before(apps[j].CreatedAt)
		})
		return appsLoadedMsg{apps: apps}
	}
}

// loadHealthCmd fetches execution health for every currently loaded
// application. It is enrichment, not core data: an application whose health
// could not be determined (no strategy, or a lookup error) is simply left
// out of the result rather than failing the whole batch.
func (m Model) loadHealthCmd() tea.Cmd {
	service := m.service
	apps := m.apps
	return func() tea.Msg {
		healthByID := make(map[string]services.ExecutionHealth, len(apps))
		for _, app := range apps {
			health, err := service.CheckExecutionHealth(context.Background(), app.ID)
			if err == nil {
				healthByID[app.ID] = health
			}
		}
		return healthLoadedMsg{healthByID: healthByID}
	}
}

// loadSessionsCmd fetches the active session for every currently loaded
// application, keyed by application ID. It performs no I/O — ActiveSession
// is a pure in-memory lookup — so this never fails; an application with no
// tracked session is simply left out of the result.
func (m Model) loadSessionsCmd() tea.Cmd {
	execService := m.executionService
	apps := m.apps
	return func() tea.Msg {
		sessionByID := make(map[string]services.RunSession, len(apps))
		for _, app := range apps {
			if session, ok := execService.ActiveSession(app.ID); ok {
				sessionByID[app.ID] = session
			}
		}
		return sessionsLoadedMsg{sessionByID: sessionByID}
	}
}

// loadGitCmd fetches Git status for every currently loaded application. It
// is enrichment, not core data: an application whose status could not be
// determined (no strategy — err from CheckGitStatus — or simply no repo)
// is left out of the result rather than failing the whole batch.
func (m Model) loadGitCmd() tea.Cmd {
	service := m.service
	apps := m.apps
	return func() tea.Msg {
		gitByID := make(map[string]services.GitStatus, len(apps))
		for _, app := range apps {
			status, err := service.CheckGitStatus(context.Background(), app.ID)
			if err == nil {
				gitByID[app.ID] = status
			}
		}
		return gitLoadedMsg{gitByID: gitByID}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// SetStatus records a transient, non-error status message to show in the
// dashboard's status line.
func (m Model) SetStatus(text string) Model {
	m.status = text
	m.statusErr = nil
	return m
}

// SetError records an error to show in the dashboard's status line.
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

	case tickMsg:
		return m, tickCmd()

	case appsLoadedMsg:
		// wasLoading distinguishes the very first load (Init) from an
		// explicit refresh (the "r" key) — only the latter gets a
		// "Refreshed." status, so startup doesn't show a confusing message
		// the user never asked for.
		wasLoading := m.loading
		m.apps = msg.apps
		m.loading = false
		m.loadErr = nil
		if m.selected >= len(m.apps) {
			m.selected = len(m.apps) - 1
		}
		if m.selected < 0 {
			m.selected = 0
		}
		if !wasLoading {
			m = m.SetStatus("Refreshed.")
		}
		return m, tea.Batch(m.loadHealthCmd(), m.loadSessionsCmd(), m.loadGitCmd())

	case appsLoadFailedMsg:
		m.loading = false
		m.loadErr = msg.err
		m = m.SetError(msg.err)
		return m, nil

	case healthLoadedMsg:
		m.healthByID = msg.healthByID
		return m, nil

	case sessionsLoadedMsg:
		m.sessionByID = msg.sessionByID
		return m, nil

	case gitLoadedMsg:
		m.gitByID = msg.gitByID
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) View() string {
	sections := []string{m.renderTopBar(), "", m.renderBody()}

	if status := m.renderStatusLine(); status != "" {
		sections = append(sections, status)
	}

	sections = append(sections, "", m.renderFooter())

	return strings.Join(sections, "\n")
}

// renderStatusLine shows the dashboard's transient status/error message, if
// any.
func (m Model) renderStatusLine() string {
	switch {
	case m.statusErr != nil:
		return errorStyle.Render(m.statusErr.Error())
	case m.status != "":
		return statusStyle.Render(m.status)
	default:
		return ""
	}
}

func (m Model) renderBody() string {
	if len(m.apps) == 0 {
		return m.renderEmptyState()
	}
	return m.renderGrid()
}
