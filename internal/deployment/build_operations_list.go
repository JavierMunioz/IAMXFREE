package deployment

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

// BuildOperations turns plan into the ordered, executable operations a
// deployment needs, dispatching on the application's configured
// DeploymentStrategy: DeploymentStrategyZeroDowntime gets
// zeroDowntimeOperations; anything else (including blank) gets the
// standard pre-deploy hook/Fetch/Install/Build/Stop/Start/Reload
// Nginx/post-deploy hook sequence — each already decided Applicable or
// Skipped from what the analysis phase found. It never runs any of them;
// that's operations.Executor's job.
//
// An error here means plan.ApplicationID could not be resolved — nothing
// else can fail at this stage, since every builder only reads plan and
// the application, never touches the network or filesystem.
func (e *Engine) BuildOperations(ctx context.Context, plan DeploymentPlan) ([]operations.Operation, error) {
	app, err := e.appService.Get(ctx, plan.ApplicationID)
	if err != nil {
		return nil, err
	}

	if app.Config.Strategy == models.DeploymentStrategyZeroDowntime {
		return e.zeroDowntimeOperations(app, plan), nil
	}
	return e.standardOperations(app, plan), nil
}

func (e *Engine) standardOperations(app *models.Application, plan DeploymentPlan) []operations.Operation {
	return []operations.Operation{
		e.preDeployHookOperation(app, plan),
		e.fetchOperation(app, plan),
		e.installOperation(app, plan),
		e.buildOperation(app, plan),
		e.stopOperation(app),
		e.startOperation(app),
		e.reloadNginxOperation(plan),
		e.postDeployHookOperation(app, plan),
	}
}
