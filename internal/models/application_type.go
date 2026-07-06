package models

// ApplicationType classifies the architectural role an application plays.
type ApplicationType string

const (
	ApplicationTypeFrontend     ApplicationType = "frontend"
	ApplicationTypeBackend      ApplicationType = "backend"
	ApplicationTypeMonolith     ApplicationType = "monolith"
	ApplicationTypeWorker       ApplicationType = "worker"
	ApplicationTypeService      ApplicationType = "service"
	ApplicationTypeAPI          ApplicationType = "api"
	ApplicationTypeMicroservice ApplicationType = "microservice"
)

// IsValid reports whether t is one of the known application types.
func (t ApplicationType) IsValid() bool {
	switch t {
	case ApplicationTypeFrontend, ApplicationTypeBackend, ApplicationTypeMonolith,
		ApplicationTypeWorker, ApplicationTypeService, ApplicationTypeAPI, ApplicationTypeMicroservice:
		return true
	default:
		return false
	}
}
