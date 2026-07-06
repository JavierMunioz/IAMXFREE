package detail

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

type startedMsg struct {
	session services.RunSession
}

type startFailedMsg struct {
	err error
}

// startCmd asks the ExecutionService to start the application's configured
// start command.
func (m Model) startCmd() tea.Cmd {
	service := m.executionService
	id := m.appID
	return func() tea.Msg {
		session, err := service.Start(context.Background(), id)
		if err != nil {
			return startFailedMsg{err: err}
		}
		return startedMsg{session: session}
	}
}
