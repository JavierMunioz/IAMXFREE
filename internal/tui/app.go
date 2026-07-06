// Package tui contains the presentation layer built with Bubble Tea and
// Lipgloss. It renders application state produced by internal/core and
// translates keyboard input into commands, but holds no business logic.
package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/dashboard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizards/application"
)

type screen int

const (
	screenDashboard screen = iota
	screenWizard
)

// applicationRegisteredMsg is emitted once ApplicationService.Register
// succeeds for an application submitted through the create-application
// wizard.
type applicationRegisteredMsg struct {
	app *models.Application
}

// applicationRegistrationFailedMsg is emitted when converting the wizard's
// result into a draft, or persisting it, fails.
type applicationRegistrationFailedMsg struct {
	err error
}

// RootModel is the top-level Bubble Tea model. It owns which screen is
// active — the dashboard or the create-application wizard — and delegates
// the actual persistence decision to services.ApplicationService.
type RootModel struct {
	service services.ApplicationService

	screen    screen
	dashboard dashboard.Model
	wizard    wizard.Model
}

// NewRootModel builds the initial application model.
func NewRootModel(service services.ApplicationService) RootModel {
	return RootModel{
		service:   service,
		screen:    screenDashboard,
		dashboard: dashboard.New(service),
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.dashboard.Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case wizard.CompletedMsg:
		m.screen = screenDashboard

		draft, err := application.DraftFromResult(msg.Result)
		if err != nil {
			m.dashboard = m.dashboard.SetError(err)
			return m, nil
		}
		return m, m.registerCmd(draft.ToApplication())

	case wizard.CancelledMsg:
		m.screen = screenDashboard
		m.dashboard = m.dashboard.SetStatus("Registration cancelled.")
		return m, nil

	case dashboard.OpenWizardMsg:
		m.screen = screenWizard
		m.wizard = wizard.New("New application", application.Steps())
		return m, m.wizard.Init()

	case applicationRegisteredMsg:
		m.dashboard = m.dashboard.SetStatus(fmt.Sprintf("Application %q registered.", msg.app.Name))
		return m, m.dashboard.Reload()

	case applicationRegistrationFailedMsg:
		m.dashboard = m.dashboard.SetError(msg.err)
		return m, nil

	case tea.WindowSizeMsg:
		updated, cmd := m.dashboard.Update(msg)
		m.dashboard = updated.(dashboard.Model)
		return m, cmd
	}

	if m.screen == screenWizard {
		updated, cmd := m.wizard.Update(msg)
		m.wizard = updated.(wizard.Model)
		return m, cmd
	}

	updated, cmd := m.dashboard.Update(msg)
	m.dashboard = updated.(dashboard.Model)
	return m, cmd
}

func (m RootModel) registerCmd(app *models.Application) tea.Cmd {
	service := m.service
	return func() tea.Msg {
		if err := service.Register(context.Background(), app); err != nil {
			return applicationRegistrationFailedMsg{err: err}
		}
		return applicationRegisteredMsg{app: app}
	}
}

func (m RootModel) View() string {
	if m.screen == screenWizard {
		return m.wizard.View()
	}
	return m.dashboard.View()
}

// Run starts the Bubble Tea program using the terminal's real stdin/stdout.
func Run(service services.ApplicationService) error {
	program := tea.NewProgram(NewRootModel(service), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
