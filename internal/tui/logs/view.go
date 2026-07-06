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
	footer := footerStyle.Render("↑/↓: scroll  ·  pgup/pgdown: page  ·  home/end: start/end  ·  esc/b: back  ·  q: quit")

	return strings.Join([]string{header, "", body, "", footer}, "\n")
}

func (m Model) renderHeader() string {
	title := fmt.Sprintf("Logs — %d line(s) retained", m.buffer.len())
	if !m.followLive {
		title += " (scrolled — press end to follow live)"
	}
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

// renderVisibleLines renders the lines currently in view: the newest
// retained lines while followLive is true, or the window scrolled to
// topIndex otherwise.
func (m Model) renderVisibleLines() []string {
	all := m.buffer.all()
	h := m.visibleHeight()

	top := m.currentTop(len(all))
	end := top + h
	if end > len(all) {
		end = len(all)
	}
	visible := all[top:end]

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
