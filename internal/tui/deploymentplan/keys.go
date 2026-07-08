package deploymentplan

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes every key while the deployment plan screen is
// active. This screen never executes a step itself — enter only asks
// whatever hosts it to open the execution screen.
func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc", "b":
		return m, func() tea.Msg { return BackMsg{} }

	case "enter":
		if !m.loaded {
			return m, nil
		}
		plan := m.plan
		return m, func() tea.Msg { return ExecutePlanMsg{Plan: plan} }
	}

	return m, nil
}
