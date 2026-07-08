package deployment

// DeploymentStep is one step of a DeploymentPlan: what it's called, which
// component would carry it out, whether it's required or optional for a
// correct deployment, and what the Engine found out about it — without
// running it.
type DeploymentStep struct {
	Name      string
	Component DeploymentComponent
	Operation DeploymentOperation
	Status    DeploymentStepStatus
	Required  bool

	// Risks are reasons this step might fail or cause harm if run.
	Risks []string
	// Warnings are things worth the user's attention that don't block the
	// step from running.
	Warnings []string

	Expected DeploymentResult
}
