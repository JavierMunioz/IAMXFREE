package dashboard

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// Colors below come from the validated reference palette (see the project's
// data-viz design skill): ink/surface tokens as adaptive light/dark pairs,
// and the status scale (good/warning/critical) as fixed values that never
// change with the terminal's theme. Status colors are reserved — they are
// never reused for anything but application state, and always render with
// an icon and a text label so meaning never rides on color alone.
var (
	colorPrimary = lipgloss.AdaptiveColor{Light: "#0b0b0b", Dark: "#ffffff"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#898781", Dark: "#898781"}
	colorBorder  = lipgloss.AdaptiveColor{Light: "#c3c2b7", Dark: "#383835"}
	colorAccent  = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#9085e9"}

	colorGood     = lipgloss.Color("#0ca30c")
	colorWarning  = lipgloss.Color("#fab219")
	colorCritical = lipgloss.Color("#d03b3b")
)

var (
	primaryStyle = lipgloss.NewStyle().Foreground(colorPrimary)
	mutedStyle   = lipgloss.NewStyle().Foreground(colorMuted)
	accentStyle  = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	statusStyle = lipgloss.NewStyle().Foreground(colorGood)
	errorStyle  = lipgloss.NewStyle().Foreground(colorCritical)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	cardSelectedStyle = cardStyle.
				BorderForeground(colorAccent).
				Bold(true)

	topBarStyle = lipgloss.NewStyle().Foreground(colorPrimary)
	footerStyle = lipgloss.NewStyle().Foreground(colorMuted)

	emptyTitleStyle = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
)

// statusPresentation returns the icon and reserved status color for an
// ApplicationStatus. Lifecycle states that are not inherently good or bad
// (installed, configured, stopped) use a neutral muted color with a distinct
// icon rather than borrowing a status color that would misrepresent them.
func statusPresentation(status models.ApplicationStatus) (string, lipgloss.TerminalColor) {
	switch status {
	case models.StatusRunning:
		return "●", colorGood
	case models.StatusUpdating:
		return "↻", colorWarning
	case models.StatusError:
		return "✕", colorCritical
	case models.StatusInstalled:
		return "◆", colorMuted
	case models.StatusConfigured:
		return "◇", colorMuted
	case models.StatusStopped:
		return "○", colorMuted
	default:
		return "?", colorMuted
	}
}
