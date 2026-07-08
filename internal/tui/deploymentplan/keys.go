package deploymentplan

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes every key while the deployment plan screen is
// active. This screen is read-only — there is no key that executes a step.
func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc", "b":
		return m, func() tea.Msg { return BackMsg{} }
	}

	return m, nil
}
