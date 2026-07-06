package execution

import (
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// Session represents one running instance of an application — never just a
// bare PID. It is shared architecture: every Strategy that starts a process
// returns a Session shaped this way, so the rest of IAMXFREE never needs to
// know which technology is behind it. Only basic information is captured
// today; the shape is deliberately room enough to grow (e.g. exit code,
// restart count) without breaking callers.
type Session struct {
	PID        int
	StartedAt  time.Time
	Command    string
	Args       []string
	WorkingDir string
	Status     Status
	Runtime    models.Runtime
}
