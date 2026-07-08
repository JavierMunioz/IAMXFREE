package detail

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// DeletedMsg signals that the user confirmed removing this application and
// it succeeded. The detail view has no notion of "dashboard" beyond this —
// whatever hosts it decides what happens next (return to the dashboard and
// refresh its list).
type DeletedMsg struct {
	AppID string
	Name  string
}

type deleteFailedMsg struct {
	err error
}

// deleteCmd asks ApplicationService to permanently remove the application
// from the registry. It only forgets the application — it does not stop a
// running session or touch its files; that's out of scope here.
func (m Model) deleteCmd() tea.Cmd {
	service := m.appService
	id := m.appID
	// m.app may still be nil if the delete confirmation somehow lands
	// before Init's load completes; fall back to an empty name rather
	// than dereferencing a nil pointer.
	name := ""
	if m.app != nil {
		name = m.app.Name
	}
	return func() tea.Msg {
		if err := service.Remove(context.Background(), id); err != nil {
			return deleteFailedMsg{err: err}
		}
		return DeletedMsg{AppID: id, Name: name}
	}
}
