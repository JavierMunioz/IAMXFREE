// Package tui contains the presentation layer built with Bubble Tea and
// Lipgloss. It renders application state produced by internal/core and
// translates keyboard input into commands, but holds no business logic.
package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizards/application"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")).
			Padding(1, 1, 0)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4CAF50")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E45858")).
			Padding(0, 1)
)

type screen int

const (
	screenSplash screen = iota
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
// active (today: the splash screen or the create-application wizard) and
// delegates the actual persistence decision to services.ApplicationService.
type RootModel struct {
	service services.ApplicationService

	screen screen
	wizard wizard.Model

	status    string
	statusErr error

	quitting bool
}

// NewRootModel builds the initial application model.
func NewRootModel(service services.ApplicationService) RootModel {
	return RootModel{service: service, screen: screenSplash}
}

func (m RootModel) Init() tea.Cmd {
	return nil
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case wizard.CompletedMsg:
		m.screen = screenSplash
		m.status = ""
		m.statusErr = nil

		draft, err := application.DraftFromResult(msg.Result)
		if err != nil {
			m.statusErr = err
			return m, nil
		}
		return m, m.registerCmd(draft.ToApplication())

	case wizard.CancelledMsg:
		m.screen = screenSplash
		m.status = "Registration cancelled."
		m.statusErr = nil
		return m, nil

	case applicationRegisteredMsg:
		m.status = fmt.Sprintf("Application %q registered.", msg.app.Name)
		m.statusErr = nil
		return m, nil

	case applicationRegistrationFailedMsg:
		m.statusErr = msg.err
		return m, nil

	case tea.KeyMsg:
		if m.screen == screenSplash {
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "a":
				m.screen = screenWizard
				m.wizard = wizard.New("New application", application.Steps())
				m.status = ""
				m.statusErr = nil
				return m, m.wizard.Init()
			}
			return m, nil
		}
	}

	if m.screen == screenWizard {
		updated, cmd := m.wizard.Update(msg)
		m.wizard = updated.(wizard.Model)
		return m, cmd
	}

	return m, nil
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
	if m.quitting {
		return ""
	}

	if m.screen == screenWizard {
		return m.wizard.View()
	}

	body := fmt.Sprintf(
		"%s\n%s\n%s",
		titleStyle.Render("IAMXFREE"),
		subtitleStyle.Render("VPS application manager"),
		hintStyle.Render("a: new application  ·  q: quit"),
	)

	switch {
	case m.statusErr != nil:
		body += "\n" + errorStyle.Render(m.statusErr.Error())
	case m.status != "":
		body += "\n" + statusStyle.Render(m.status)
	}

	return body + "\n"
}

// Run starts the Bubble Tea program using the terminal's real stdin/stdout.
func Run(service services.ApplicationService) error {
	program := tea.NewProgram(NewRootModel(service), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
