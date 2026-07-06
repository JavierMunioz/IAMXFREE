// Package tui contains the presentation layer built with Bubble Tea and
// Lipgloss. It renders application state produced by internal/core and
// translates keyboard input into commands, but holds no business logic.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
)

// RootModel is the top-level Bubble Tea model. Later iterations will replace
// its body with a real dashboard; for now it only proves the render loop and
// key handling work end to end.
type RootModel struct {
	quitting bool
}

// NewRootModel builds the initial application model.
func NewRootModel() RootModel {
	return RootModel{}
}

func (m RootModel) Init() tea.Cmd {
	return nil
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m RootModel) View() string {
	if m.quitting {
		return ""
	}

	body := fmt.Sprintf(
		"%s\n%s\n%s",
		titleStyle.Render("IAMXFREE"),
		subtitleStyle.Render("VPS application manager"),
		hintStyle.Render("q: quit"),
	)

	return body + "\n"
}

// Run starts the Bubble Tea program using the terminal's real stdin/stdout.
func Run() error {
	program := tea.NewProgram(NewRootModel(), tea.WithAltScreen())
	_, err := program.Run()
	return err
}
