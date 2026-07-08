package deploymentplan

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/JavierMunioz/IAMXFREE/internal/deployment"
)

// Same validated reference palette the rest of the TUI uses — kept as its
// own copy since each screen's styles are unexported to its own package.
// Status colors are reserved and always render with an icon and a text
// label, never color alone.
var (
	colorMuted  = lipgloss.AdaptiveColor{Light: "#898781", Dark: "#898781"}
	colorBorder = lipgloss.AdaptiveColor{Light: "#c3c2b7", Dark: "#383835"}
	colorAccent = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#9085e9"}

	colorGood     = lipgloss.Color("#0ca30c")
	colorWarning  = lipgloss.Color("#fab219")
	colorCritical = lipgloss.Color("#d03b3b")
)

var (
	mutedStyle  = lipgloss.NewStyle().Foreground(colorMuted)
	accentStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	errorStyle  = lipgloss.NewStyle().Foreground(colorCritical)
	footerStyle = lipgloss.NewStyle().Foreground(colorMuted)

	labelStyle = mutedStyle.Width(10)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)
)

// stepStatusPresentation returns the icon and reserved status color for a
// DeploymentStepStatus.
func stepStatusPresentation(status deployment.DeploymentStepStatus) (string, lipgloss.TerminalColor) {
	switch status {
	case deployment.StepStatusReady:
		return "●", colorGood
	case deployment.StepStatusWarning:
		return "▲", colorWarning
	case deployment.StepStatusBlocked:
		return "✕", colorCritical
	case deployment.StepStatusSkipped:
		return "○", colorMuted
	default:
		return "?", colorMuted
	}
}
