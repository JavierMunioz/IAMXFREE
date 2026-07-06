package detail

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// renderMiddleRow lays out the technical panel (left) and the execution
// status panel (right) side by side.
func (m Model) renderMiddleRow() string {
	half := (m.width - 4) / 2
	if half < 20 {
		half = 20
	}

	left := m.renderTechnicalPanel(half)
	right := m.renderExecutionPanel(half)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// renderTechnicalPanel shows where the application lives and how it is
// installed/built/started — the configuration behind the general info shown
// on the top panel.
func (m Model) renderTechnicalPanel(width int) string {
	app := m.app

	rows := [][2]string{
		{"Path", valueOrDash(app.Source.LocalPath)},
		{"Repository", valueOrDash(app.Source.RepositoryURL)},
		{"Pkg manager", valueOrDash(app.Config.PackageManager)},
		{"Install", valueOrDash(app.Config.InstallCommand)},
		{"Build", valueOrDash(app.Config.BuildCommand)},
		{"Start", valueOrDash(app.Config.StartCommand)},
		{"Registered", app.CreatedAt.Format("2006-01-02 15:04")},
	}

	var lines []string
	lines = append(lines, primaryStyle.Bold(true).Render("Technical"))
	lines = append(lines, "")
	for _, row := range rows {
		lines = append(lines, labelStyle.Render(row[0])+row[1])
	}

	return panelStyle.Width(width).Height(9).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// renderExecutionPanel shows the live session IAMXFREE is tracking for this
// application, if any. When no session is active, that is stated clearly
// rather than leaving fields blank.
func (m Model) renderExecutionPanel(width int) string {
	var lines []string
	lines = append(lines, primaryStyle.Bold(true).Render("Execution"))
	lines = append(lines, "")

	if !m.hasSession {
		lines = append(lines, mutedStyle.Render("No active session."))
		lines = append(lines, mutedStyle.Render("Press s to start."))
	} else {
		running := m.session.Status == "running"
		runningText := lipgloss.NewStyle().Foreground(colorCritical).Render("✕ stopped")
		if running {
			runningText = lipgloss.NewStyle().Foreground(colorGood).Render("● running")
		}

		rows := [][2]string{
			{"Running", runningText},
			{"PID", fmt.Sprintf("%d", m.session.PID)},
			{"Started", m.session.StartedAt.Format("2006-01-02 15:04:05")},
			{"Work dir", valueOrDash(m.session.WorkingDir)},
			{"Runtime", valueOrDash(string(m.session.Runtime))},
		}
		for _, row := range rows {
			lines = append(lines, labelStyle.Render(row[0])+row[1])
		}
	}

	return panelStyle.Width(width).Height(9).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
