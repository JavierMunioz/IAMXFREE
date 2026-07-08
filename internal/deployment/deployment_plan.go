package deployment

import "time"

// DeploymentPlan is everything Engine.Plan found out about deploying one
// application: the steps that would be needed, in order, and a summary of
// whether it's currently safe to proceed. Generating one never changes
// anything — it's a snapshot of what the Engine observed at GeneratedAt.
type DeploymentPlan struct {
	ApplicationID   string
	ApplicationName string
	GeneratedAt     time.Time

	Steps []DeploymentStep

	// Warnings flattens every step's Warnings, each prefixed with its
	// step's name, so a caller can show "are there any warnings at all"
	// without walking Steps itself.
	Warnings []string

	Summary DeploymentSummary
}
