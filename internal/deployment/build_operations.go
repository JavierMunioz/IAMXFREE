package deployment

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

// installOperation runs the application's configured install command
// through ExecutionService, applicable only when the analysis phase found
// one configured.
func (e *Engine) installOperation(app *models.Application, plan DeploymentPlan) operations.Operation {
	return e.commandOperation(app, plan, OperationInstallDependencies, "Install dependencies", "Install", e.executionService.Install)
}

// buildOperation runs the application's configured build command through
// ExecutionService, applicable only when the analysis phase found one
// configured.
func (e *Engine) buildOperation(app *models.Application, plan DeploymentPlan) operations.Operation {
	return e.commandOperation(app, plan, OperationBuild, "Build application", "Build", e.executionService.Build)
}

// commandOperation is shared by installOperation and buildOperation: both
// are gated by a step already Skipped when no command was configured, and
// both invoke a same-shaped ExecutionService method.
func (e *Engine) commandOperation(app *models.Application, plan DeploymentPlan, stepOperation DeploymentOperation, name, method string, invoke func(ctx context.Context, appID string) error) operations.Operation {
	op := operations.Operation{Name: name, Component: string(ComponentExecution), Method: method}

	step, found := findStep(plan, stepOperation)
	if !found || step.Status == StepStatusSkipped {
		op.SkipReason = "no " + method + " command configured"
		return op
	}

	op.Applicable = true
	op.Run = func(ctx context.Context) error {
		return invoke(ctx, app.ID)
	}
	return op
}
