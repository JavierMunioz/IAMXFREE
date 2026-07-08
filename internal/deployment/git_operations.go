package deployment

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

// fetchOperation updates the repository's remote-tracking refs before the
// rest of the deployment runs, so the plan's Ahead/Behind reflects the
// current state of the remote. It only applies when the plan found a
// valid Git repository to begin with.
func (e *Engine) fetchOperation(app *models.Application, plan DeploymentPlan) operations.Operation {
	op := operations.Operation{Name: "Fetch latest changes", Component: string(ComponentGit), Method: "Fetch"}

	verify, found := findStep(plan, OperationVerifyRepository)
	if !found || verify.Status != StepStatusReady {
		op.SkipReason = "no valid Git repository to fetch"
		return op
	}

	op.Applicable = true
	op.Run = func(ctx context.Context) error {
		_, err := e.gitManager.Fetch(ctx, app.Source.LocalPath)
		return err
	}
	return op
}
