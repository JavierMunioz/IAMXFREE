package logs

import "github.com/charmbracelet/lipgloss"

// Same validated reference palette used by dashboard/detail (see the
// project's data-viz design skill) — kept as its own copy since those
// packages' styles are unexported. Log event types are distinguished by a
// literal text prefix plus a reserved color, never color alone.
var (
	colorPrimary = lipgloss.AdaptiveColor{Light: "#0b0b0b", Dark: "#ffffff"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#898781", Dark: "#898781"}
	colorAccent  = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#9085e9"}

	colorWarning  = lipgloss.Color("#fab219")
	colorCritical = lipgloss.Color("#d03b3b")
)

var (
	mutedStyle  = lipgloss.NewStyle().Foreground(colorMuted)
	accentStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(colorCritical)
	footerStyle = lipgloss.NewStyle().Foreground(colorMuted)
	stderrStyle = lipgloss.NewStyle().Foreground(colorWarning)
	systemStyle = lipgloss.NewStyle().Foreground(colorAccent)
	logErrStyle = lipgloss.NewStyle().Foreground(colorCritical).Bold(true)
)

// stylePrefix returns the literal text prefix and color for a
// services.LogEvent's Type — stdout/stderr/system/error/eof are always
// distinguishable by the prefix alone, even without color.
func stylePrefix(eventType string) (string, lipgloss.Style) {
	switch eventType {
	case "stdout":
		return "out", mutedStyle
	case "stderr":
		return "err", stderrStyle
	case "system":
		return "sys", systemStyle
	case "error":
		return "ERR", logErrStyle
	case "eof":
		return "end", mutedStyle
	default:
		return "?", mutedStyle
	}
}
