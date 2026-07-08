package detail

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes every key while the detail view is the active screen.
// More actions (start, stop, refresh, restart/logs/edit placeholders) are
// added alongside their own capability.
func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	// A pending delete confirmation intercepts every key: "d" again
	// confirms it, anything else cancels. This runs before the normal
	// switch so no other binding can fire mid-confirmation.
	if m.confirmingDelete {
		m.confirmingDelete = false
		if key.String() == "d" {
			m = m.SetStatus("Deleting…")
			return m, m.deleteCmd()
		}
		return m.SetStatus("Delete cancelled."), nil
	}

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

	case "x":
		if !m.hasSession {
			return m.SetStatus("No active session to stop."), nil
		}
		m = m.SetStatus("Stopping…")
		return m, m.stopCmd()

	case "f5":
		m = m.SetStatus("Refreshing…")
		return m, m.refreshCmd()

	case "r":
		return m.SetStatus("Restart is not implemented yet."), nil

	case "l":
		if !m.hasSession {
			return m.SetStatus("No active session to show logs for."), nil
		}
		appID := m.appID
		session := m.session
		return m, func() tea.Msg { return OpenLogsMsg{AppID: appID, Session: session} }

	case "e":
		return m.SetStatus("Editing configuration is not implemented yet."), nil

	case "p":
		appID := m.appID
		return m, func() tea.Msg { return OpenDeploymentPlanMsg{AppID: appID} }

	case "d":
		m.confirmingDelete = true
		return m.SetStatus("Press d again to permanently delete this application, any other key cancels."), nil
	}

	return m, nil
}
