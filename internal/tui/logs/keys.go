package logs

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes every key while the logs view is the active screen.
// Scroll navigation is added alongside its own capability.
func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc", "b":
		if m.stream != nil {
			m.stream.Close()
		}
		return m, func() tea.Msg { return BackMsg{} }
	}

	return m, nil
}
