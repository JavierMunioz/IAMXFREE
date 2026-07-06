package models

// ApplicationStatus represents the lifecycle state of a managed application.
type ApplicationStatus string

const (
	StatusInstalled  ApplicationStatus = "installed"
	StatusConfigured ApplicationStatus = "configured"
	StatusRunning    ApplicationStatus = "running"
	StatusStopped    ApplicationStatus = "stopped"
	StatusError      ApplicationStatus = "error"
	StatusUpdating   ApplicationStatus = "updating"
)

// IsValid reports whether s is one of the known application statuses.
func (s ApplicationStatus) IsValid() bool {
	switch s {
	case StatusInstalled, StatusConfigured, StatusRunning, StatusStopped, StatusError, StatusUpdating:
		return true
	default:
		return false
	}
}
