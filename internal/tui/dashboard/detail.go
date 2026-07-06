package dashboard

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var detailLabelStyle = mutedStyle.Width(14)

// renderDetail shows every stored field of the selected application. It is
// read-only: no process control, no editing — just a fuller view of data
// the card doesn't have room for.
func (m Model) renderDetail() string {
	app := m.apps[m.selected]
	icon, color := statusPresentation(app.Status)
	status := lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s %s", icon, app.Status))

	rows := [][2]string{
		{"Name", app.Name},
		{"Type", string(app.Type)},
		{"Framework", orDash(string(app.Framework))},
		{"Runtime", orDash(string(app.Runtime))},
		{"Status", status},
		{"Local path", orDash(app.Source.LocalPath)},
		{"Repository", orDash(app.Source.RepositoryURL)},
		{"Branch", orDash(app.Source.MainBranch)},
		{"Internal port", fmt.Sprintf("%d", app.Config.InternalPort)},
		{"Domain", orDash(app.Config.Domain)},
		{"Created", app.CreatedAt.Format(time.RFC3339)},
		{"Updated", app.UpdatedAt.Format(time.RFC3339)},
	}

	var lines []string
	lines = append(lines, accentStyle.Render(app.Name))
	lines = append(lines, "")
	for _, row := range rows {
		lines = append(lines, detailLabelStyle.Render(row[0])+row[1])
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))

	return box
}
