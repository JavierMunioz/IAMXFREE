package models

// DeploymentStrategy names how a deployment brings the new version up:
// Standard stops the old process before starting the new one (a brief gap
// while it is down); ZeroDowntime is reserved for starting the new version
// alongside the old one, health-checking it, then switching traffic over
// before stopping the old one — not implemented yet, this is the seam a
// future deployment.Engine iteration will build on.
type DeploymentStrategy string

const (
	DeploymentStrategyStandard     DeploymentStrategy = "standard"
	DeploymentStrategyZeroDowntime DeploymentStrategy = "zero_downtime"
)

// IsValid reports whether s is one of the known deployment strategies.
func (s DeploymentStrategy) IsValid() bool {
	switch s {
	case DeploymentStrategyStandard, DeploymentStrategyZeroDowntime:
		return true
	default:
		return false
	}
}
