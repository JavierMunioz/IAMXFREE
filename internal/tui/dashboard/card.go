package dashboard

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// renderGrid lays out every application as a card, wrapping into rows of
// m.columns() cards each.
func (m Model) renderGrid() string {
	cols := m.columns()

	var rows []string
	for start := 0; start < len(m.apps); start += cols {
		end := start + cols
		if end > len(m.apps) {
			end = len(m.apps)
		}

		var cards []string
		for i := start; i < end; i++ {
			cards = append(cards, m.renderCard(m.apps[i], i == m.selected))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderCard(app *models.Application, selected bool) string {
	style := cardStyle
	nameStyle := primaryStyle.Bold(true)
	if selected {
		style = cardSelectedStyle
		nameStyle = accentStyle
	}

	icon, color := statusPresentation(app.Status)
	statusText := lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s %s", icon, app.Status))
	statusColumn := lipgloss.NewStyle().Width(cardWidth - 12).Render(statusText)
	portText := mutedStyle.Render(fmt.Sprintf("port %d", app.Config.InternalPort))

	lines := []string{
		nameStyle.Render(truncate(app.Name, cardWidth)),
		mutedStyle.Render(truncate(fmt.Sprintf("%s · %s · %s", orDash(string(app.Type)), orDash(string(app.Framework)), orDash(string(app.Runtime))), cardWidth)),
		lipgloss.JoinHorizontal(lipgloss.Top, statusColumn, portText),
		mutedStyle.Render(truncate(orDash(app.Config.Domain), cardWidth)),
	}

	if health, ok := m.healthByID[app.ID]; ok {
		lines = append(lines,
			mutedStyle.Render(truncate("Strategy: "+health.StrategyName, cardWidth)),
			renderHealthLine(health.Healthy),
		)
	}

	return style.Width(cardWidth).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// renderHealthLine shows execution health with both an icon and a label —
// never color alone — following the same status-color convention as the
// application's own run status.
func renderHealthLine(healthy bool) string {
	if healthy {
		return lipgloss.NewStyle().Foreground(colorGood).Render("● Health: healthy")
	}
	return lipgloss.NewStyle().Foreground(colorCritical).Render("✕ Health: unhealthy")
}

func orDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

func truncate(value string, width int) string {
	if lipgloss.Width(value) <= width {
		return value
	}
	runes := []rune(value)
	if width <= 1 || len(runes) <= width {
		return value
	}
	return string(runes[:width-1]) + "…"
}
