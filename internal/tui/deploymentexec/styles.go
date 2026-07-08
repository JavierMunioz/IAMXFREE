package deploymentexec

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

// Same validated reference palette the rest of the TUI uses — kept as its
// own copy since each screen's styles are unexported to its own package.
// Status colors are reserved and always render with an icon and a text
// label, never color alone.
var (
	colorMuted  = lipgloss.AdaptiveColor{Light: "#898781", Dark: "#898781"}
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
)

// operationStatePresentation returns the icon and reserved status color
// for an OperationState.
func operationStatePresentation(state operations.OperationState) (string, lipgloss.TerminalColor) {
	switch state {
	case operations.StatePending:
		return "○", colorMuted
	case operations.StateRunning:
		return "◌", colorWarning
	case operations.StateSuccess:
		return "●", colorGood
	case operations.StateFailed:
		return "✕", colorCritical
	case operations.StateSkipped:
		return "○", colorMuted
	case operations.StateCancelled:
		return "◐", colorMuted
	default:
		return "?", colorMuted
	}
}
