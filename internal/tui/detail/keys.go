package detail

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes every key while the detail view is the active screen.
// More actions (start, stop, refresh, restart/logs/edit placeholders) are
// added alongside their own capability.
func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc", "b":
		return m, func() tea.Msg { return BackMsg{} }

	case "s":
		if m.hasSession {
			return m.SetStatus("Already tracking a session — stop it first."), nil
		}
		m = m.SetStatus("Starting…")
		return m, m.startCmd()
	}

	return m, nil
}
