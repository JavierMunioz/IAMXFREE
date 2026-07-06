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

type viewMode int

const (
	viewGrid viewMode = iota
	viewDetail
)

// OpenWizardMsg signals that the user asked to register a new application.
// The dashboard has no notion of "wizard" beyond this — whatever hosts it
// decides what happens next.
type OpenWizardMsg struct{}

type appsLoadedMsg struct {
	apps []*models.Application
}

type appsLoadFailedMsg struct {
	err error
}

type tickMsg time.Time

// Model is the dashboard screen. It depends only on services.ApplicationService
// — never on a repository directly — for every piece of application data it
// shows.
type Model struct {
	service services.ApplicationService

	apps     []*models.Application
	selected int
	mode     viewMode

	width  int
	height int

	status    string
	statusErr error

	loading bool
	loadErr error
}

// New builds a dashboard backed by service.
func New(service services.ApplicationService) Model {
	return Model{service: service, width: 80, height: 24, loading: true}
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
		m.apps = msg.apps
		m.loading = false
		m.loadErr = nil
		if m.selected >= len(m.apps) {
			m.selected = len(m.apps) - 1
		}
		if m.selected < 0 {
			m.selected = 0
		}
		return m, nil

	case appsLoadFailedMsg:
		m.loading = false
		m.loadErr = msg.err
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
	switch {
	case m.mode == viewDetail:
		return m.renderDetail()
	case len(m.apps) == 0:
		return m.renderEmptyState()
	default:
		return m.renderGrid()
	}
}
