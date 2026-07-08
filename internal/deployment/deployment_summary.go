package deployment

// DeploymentSummary is a DeploymentPlan's steps reduced to the numbers a
// caller needs to answer "is this safe to deploy" without walking every
// step itself.
type DeploymentSummary struct {
	TotalSteps    int
	RequiredSteps int
	BlockedSteps  int
	WarningCount  int

	// Ready is true only when no required step is Blocked. It says
	// nothing about Warnings — a plan can be Ready and still have things
	// worth reviewing first.
	Ready bool
}
