package deployment

// DeploymentResult describes an outcome a DeploymentStep is expected to
// produce if it runs. This iteration only ever predicts it during
// planning — no step actually executes — but the same shape is meant to
// describe a step's *actual* outcome once execution is implemented, so a
// planned and a real result stay directly comparable.
type DeploymentResult struct {
	Description  string
	WouldSucceed bool
}
