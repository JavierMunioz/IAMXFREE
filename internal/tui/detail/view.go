package detail

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.loadErr != nil {
		return errorStyle.Render("Could not load application: "+m.loadErr.Error()) + "\n\nesc/b: back  ·  q: quit"
	}
	if m.loading || m.app == nil {
		return mutedStyle.Render("Loading…")
	}

	panelWidth := m.width - 4
	if panelWidth < 20 {
		panelWidth = 20
	}

	sections := []string{m.renderTopPanel(), m.renderMiddleRow(), m.renderGitPanel(panelWidth)}

	if m.hasSession {
		sections = append(sections, m.renderMetricsPanel(panelWidth))
	}

	if status := m.renderStatusLine(); status != "" {
		sections = append(sections, status)
	}

	sections = append(sections, m.renderActionBar())

	return strings.Join(sections, "\n\n")
}

// renderActionBar lists every keybinding this screen responds to. Actions
// without real logic yet are still listed — pressing them reports "not
// implemented yet" via the status line instead of doing nothing silently.
func (m Model) renderActionBar() string {
	return footerStyle.Render(
		"s: start  ·  x: stop  ·  l: logs  ·  p: deployment plan  ·  r: restart*  ·  e: edit*  ·  f5: refresh  ·  b/esc: back  ·  q: quit    (* not implemented yet)",
	)
}

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

// renderTopPanel shows the application's general information: identity,
// how it's managed, and its current status/health.
func (m Model) renderTopPanel() string {
	app := m.app

	icon, color := statusPresentation(app.Status)
	statusText := lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s %s", icon, app.Status))

	healthText := "—"
	if m.healthKnown {
		if m.healthy {
			healthText = lipgloss.NewStyle().Foreground(colorGood).Render("● healthy")
		} else {
			healthText = lipgloss.NewStyle().Foreground(colorCritical).Render("✕ unhealthy")
		}
	}

	rows := [][2]string{
		{"Name", app.Name},
		{"Type", string(app.Type)},
		{"Framework", valueOrDash(string(app.Framework))},
		{"Runtime", valueOrDash(string(app.Runtime))},
		{"Strategy", valueOrDash(m.strategyName)},
		{"Status", statusText},
		{"Health", healthText},
		{"Port", fmt.Sprintf("%d", app.Config.InternalPort)},
		{"Domain", valueOrDash(app.Config.Domain)},
	}

	var lines []string
	lines = append(lines, accentStyle.Render(app.Name))
	lines = append(lines, "")
	for _, row := range rows {
		lines = append(lines, labelStyle.Render(row[0])+row[1])
	}

	width := m.width - 4
	if width < 20 {
		width = 20
	}

	return panelStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
