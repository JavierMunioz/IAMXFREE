package deployment

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

// BuildOperations turns plan into the ordered, executable operations a
// deployment needs: Fetch, Install, Build, Stop, Start, Reload Nginx —
// each already decided Applicable or Skipped from what the analysis phase
// found. It never runs any of them; that's operations.Executor's job.
//
// An error here means plan.ApplicationID could not be resolved — nothing
// else can fail at this stage, since every builder only reads plan and
// the application, never touches the network or filesystem.
func (e *Engine) BuildOperations(ctx context.Context, plan DeploymentPlan) ([]operations.Operation, error) {
	app, err := e.appService.Get(ctx, plan.ApplicationID)
	if err != nil {
		return nil, err
	}

	return []operations.Operation{
		e.fetchOperation(app, plan),
		e.installOperation(app, plan),
		e.buildOperation(app, plan),
		e.stopOperation(app),
		e.startOperation(app),
		e.reloadNginxOperation(plan),
	}, nil
}
