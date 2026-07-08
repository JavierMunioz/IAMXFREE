package deploymentexec

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes every key while the execution screen is active.
// There is no key that cancels or retries a run this iteration.
func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc", "b":
		return m, func() tea.Msg { return BackMsg{} }
	}

	return m, nil
}
