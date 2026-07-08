package deployment

// DeploymentStepStatus states what a DeploymentStep found, not what it did
// — this iteration never executes a step.
type DeploymentStepStatus string

const (
	// StepStatusReady means the step could run right now with no known
	// problem.
	StepStatusReady DeploymentStepStatus = "ready"

	// StepStatusWarning means the step could run, but something about it
	// is worth the user's attention first (see the step's Warnings).
	StepStatusWarning DeploymentStepStatus = "warning"

	// StepStatusBlocked means the step cannot run as things stand (see
	// the step's Risks).
	StepStatusBlocked DeploymentStepStatus = "blocked"

	// StepStatusSkipped means the step does not apply to this application
	// (e.g. no build command configured) — not a problem, just nothing to
	// do.
	StepStatusSkipped DeploymentStepStatus = "skipped"
)
