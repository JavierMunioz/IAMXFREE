package wizard

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E45858"))
)
