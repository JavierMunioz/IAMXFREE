package logs

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes every key while the logs view is the active screen.
func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc", "b":
		if m.stream != nil {
			m.stream.Close()
		}
		return m, func() tea.Msg { return BackMsg{} }

	case "up", "k":
		return m.scrollBy(-1), nil
	case "down", "j":
		return m.scrollBy(1), nil
	case "pgup":
		return m.scrollBy(-m.visibleHeight()), nil
	case "pgdown":
		return m.scrollBy(m.visibleHeight()), nil
	case "home", "g":
		return m.scrollToStart(), nil
	case "end", "G":
		return m.scrollToEnd(), nil
	}

	return m, nil
}
