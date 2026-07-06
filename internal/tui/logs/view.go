package logs

import (
	"fmt"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func (m Model) View() string {
	if m.openErr != nil {
		return errorStyle.Render("Could not open logs: "+m.openErr.Error()) + "\n\nesc/b: back  ·  q: quit"
	}

	header := m.renderHeader()
	body := strings.Join(m.renderVisibleLines(), "\n")
	footer := footerStyle.Render("esc/b: back  ·  q: quit")

	return strings.Join([]string{header, "", body, "", footer}, "\n")
}

func (m Model) renderHeader() string {
	title := fmt.Sprintf("Logs — %d line(s) retained", m.buffer.len())
	if m.streamEnded {
		title += " (stream ended)"
	}
	return accentStyle.Render(title)
}

// visibleHeight reports how many log lines fit given the screen's current
// height, reserving room for the header and footer chrome.
func (m Model) visibleHeight() int {
	h := m.height - 4
	if h < 1 {
		h = 1
	}
	return h
}

// renderVisibleLines renders the lines currently in view, pinned to the
// bottom (the most recent lines) — scrolling arrives in a later commit.
func (m Model) renderVisibleLines() []string {
	all := m.buffer.all()
	h := m.visibleHeight()

	start := len(all) - h
	if start < 0 {
		start = 0
	}
	visible := all[start:]

	lines := make([]string, 0, len(visible))
	for _, e := range visible {
		lines = append(lines, renderLine(e))
	}
	return lines
}

func renderLine(e services.LogEvent) string {
	prefix, style := stylePrefix(e.Type)
	return style.Render(prefix) + " " + e.Content
}
