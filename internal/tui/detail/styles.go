package detail

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// Same validated reference palette the dashboard uses (see the project's
// data-viz design skill) — kept as its own copy since dashboard's styles are
// unexported to that package. Status colors are reserved and always render
// with an icon and a text label, never color alone.
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
	footerStyle = lipgloss.NewStyle().Foreground(colorMuted)

	labelStyle = mutedStyle.Width(12)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)
)

// statusPresentation returns the icon and reserved status color for an
// ApplicationStatus — same convention as the dashboard's cards.
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

// valueOrDash returns value, or an em-dash placeholder when it is empty —
// so an unset field is visibly "not set" rather than a blank space.
func valueOrDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}
