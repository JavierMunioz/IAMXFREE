package detail

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

type sessionRefreshedMsg struct {
	session services.RunSession
}

type sessionRefreshFailedMsg struct {
	err error
}

// refreshSessionCmd re-checks whether the currently tracked session's
// process is still alive.
func (m Model) refreshSessionCmd() tea.Cmd {
	service := m.executionService
	id := m.appID
	session := m.session
	return func() tea.Msg {
		updated, err := service.RefreshSession(context.Background(), id, session)
		if err != nil {
			return sessionRefreshFailedMsg{err: err}
		}
		return sessionRefreshedMsg{session: updated}
	}
}

// refreshCmd re-queries the strategy's health and, if a session is being
// tracked, its current status. This is the only refresh IAMXFREE performs —
// there is no automatic/periodic refresh yet.
func (m Model) refreshCmd() tea.Cmd {
	cmds := []tea.Cmd{m.loadHealthCmd()}
	if m.hasSession {
		cmds = append(cmds, m.refreshSessionCmd())
	}
	return tea.Batch(cmds...)
}
