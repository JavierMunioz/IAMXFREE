package dashboard

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleKey(key tea.KeyMsg) (Model, tea.Cmd) {
	if m.mode == viewDetail {
		return m.handleDetailKey(key)
	}
	return m.handleGridKey(key)
}

func (m Model) handleDetailKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.mode = viewGrid
	}
	return m, nil
}

func (m Model) handleGridKey(key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "a":
		return m, func() tea.Msg { return OpenWizardMsg{} }

	case "r":
		m = m.SetStatus("Refreshing…")
		return m, m.Reload()

	case "enter":
		if len(m.apps) > 0 {
			m.mode = viewDetail
		}

	case "e":
		if len(m.apps) > 0 {
			m = m.SetStatus("Editing is not implemented yet.")
		}

	case "d":
		if len(m.apps) > 0 {
			m = m.SetStatus("Deleting is not implemented yet.")
		}

	case "left":
		m.moveSelection(-1)
	case "right":
		m.moveSelection(1)
	case "up", "k":
		m.moveSelection(-m.columns())
	case "down", "j":
		m.moveSelection(m.columns())
	case "tab":
		m.cycleSelection(1)
	case "shift+tab":
		m.cycleSelection(-1)
	}

	return m, nil
}

// moveSelection shifts the selection by delta, clamped to the valid range.
// Used for spatial movement (arrow keys), which should stop at the edges of
// the grid rather than wrap around.
func (m *Model) moveSelection(delta int) {
	if len(m.apps) == 0 {
		return
	}
	next := m.selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(m.apps) {
		next = len(m.apps) - 1
	}
	m.selected = next
}

// cycleSelection shifts the selection by delta, wrapping around at either
// end. Used for Tab/Shift+Tab, which conventionally cycle focus rather than
// stopping at the edges.
func (m *Model) cycleSelection(delta int) {
	n := len(m.apps)
	if n == 0 {
		return
	}
	m.selected = ((m.selected+delta)%n + n) % n
}

// columns reports how many cards fit per row at the dashboard's current
// width.
func (m Model) columns() int {
	cols := (m.width + cardGap) / (cardWidth + cardGap)
	if cols < 1 {
		cols = 1
	}
	return cols
}
