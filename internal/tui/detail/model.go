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

// Model is the application detail screen.
type Model struct {
	appService       services.ApplicationService
	executionService services.ExecutionService

	appID string
	app   *models.Application

	strategyName string
	healthy      bool
	healthKnown  bool

	session    services.RunSession
	hasSession bool

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
	return tea.Batch(m.loadAppCmd(), m.loadHealthCmd())
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

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}
