package repositories

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
)

// SessionRepository persists the session ExecutionService is tracking for
// each application, so that knowledge survives IAMXFREE being closed and
// reopened. The OS process itself already survives on its own (it's just a
// detached background process) — without this, only IAMXFREE's own record
// of it was lost on every restart, making ActiveSession report "nothing
// running" for a session that was, in fact, still alive.
type SessionRepository interface {
	// Save persists session as the one currently tracked for appID,
	// overwriting whatever was there before.
	Save(ctx context.Context, appID string, session execution.Session) error

	// Delete removes any persisted session for appID. Deleting an appID
	// with nothing persisted is not an error.
	Delete(ctx context.Context, appID string) error

	// List returns every persisted session, keyed by application ID. Used
	// once at startup to rehydrate ExecutionService's in-memory registry —
	// never called on a recurring basis.
	List(ctx context.Context) (map[string]execution.Session, error)
}
