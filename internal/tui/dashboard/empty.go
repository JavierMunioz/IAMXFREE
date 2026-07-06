package dashboard

import "github.com/charmbracelet/lipgloss"

// renderEmptyState replaces the card grid when there are zero registered
// applications, so a first-time user sees guidance instead of blank space.
func (m Model) renderEmptyState() string {
	keyHint := accentStyle.Render("a")

	message := lipgloss.JoinVertical(
		lipgloss.Center,
		emptyTitleStyle.Render("No applications registered yet"),
		"",
		"Press "+keyHint+mutedStyle.Render(" to register your first application."),
	)

	height := m.height - 6
	if height < 3 {
		height = 3
	}

	return lipgloss.Place(m.width, height, lipgloss.Center, lipgloss.Center, message)
}
