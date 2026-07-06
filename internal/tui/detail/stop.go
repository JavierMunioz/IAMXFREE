package detail

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

type stoppedMsg struct{}

type stopFailedMsg struct {
	err error
}

// stopCmd asks the ExecutionService to terminate the currently tracked
// session.
func (m Model) stopCmd() tea.Cmd {
	service := m.executionService
	id := m.appID
	session := m.session
	return func() tea.Msg {
		if err := service.Stop(context.Background(), id, session); err != nil {
			return stopFailedMsg{err: err}
		}
		return stoppedMsg{}
	}
}
