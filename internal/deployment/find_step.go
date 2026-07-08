package deployment

// findStep returns the first step in plan.Steps with the given Operation,
// so an operations.Operation builder can decide Applicable/SkipReason
// from what the analysis phase already found — without re-deriving it.
func findStep(plan DeploymentPlan, op DeploymentOperation) (DeploymentStep, bool) {
	for _, step := range plan.Steps {
		if step.Operation == op {
			return step, true
		}
	}
	return DeploymentStep{}, false
}
