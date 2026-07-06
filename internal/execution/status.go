package execution

// Status represents the lifecycle state of a running Session. It is shared
// architecture: every Strategy reports Status this same way, regardless of
// technology.
type Status string

const (
	StatusStarting Status = "starting"
	StatusRunning  Status = "running"
	StatusStopping Status = "stopping"
	StatusStopped  Status = "stopped"
	StatusFailed   Status = "failed"
)

// IsValid reports whether s is one of the known statuses.
func (s Status) IsValid() bool {
	switch s {
	case StatusStarting, StatusRunning, StatusStopping, StatusStopped, StatusFailed:
		return true
	default:
		return false
	}
}
