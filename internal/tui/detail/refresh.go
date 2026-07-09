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

// snapshotCmd observes the currently tracked session's real runtime state
// (process status, CPU/memory usage) through the Runtime Monitor. It is
// only ever invoked from refreshCmd — there is no automatic polling.
func (m Model) snapshotCmd() tea.Cmd {
	service := m.executionService
	session := m.session
	return func() tea.Msg {
		snapshot, err := service.Snapshot(context.Background(), session)
		if err != nil {
			return snapshotFailedMsg{err: err}
		}
		return snapshotLoadedMsg{snapshot: snapshot}
	}
}

// refreshCmd re-queries the strategy's health, the application's Git
// status and, if a session is being tracked, its current status and
// runtime snapshot. This is the only refresh IAMXFREE performs — there is
// no automatic/periodic refresh yet.
func (m Model) refreshCmd() tea.Cmd {
	cmds := []tea.Cmd{m.loadHealthCmd(), m.loadGitStatusCmd(), m.loadCandidateSessionCmd()}
	if m.hasSession {
		cmds = append(cmds, m.refreshSessionCmd(), m.snapshotCmd())
	}
	return tea.Batch(cmds...)
}
