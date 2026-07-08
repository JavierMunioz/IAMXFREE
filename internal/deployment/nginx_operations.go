package deployment

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

// reloadNginxOperation reloads Nginx, applicable only when the analysis
// phase's "Reload Nginx" step found it necessary (a missing or disabled
// site) — an already-enabled site needs no reload.
func (e *Engine) reloadNginxOperation(plan DeploymentPlan) operations.Operation {
	op := operations.Operation{Name: "Reload Nginx", Component: string(ComponentNginx), Method: "Reload"}

	step, found := findStep(plan, OperationReloadNginx)
	if !found || step.Status == StepStatusSkipped || !step.Required {
		op.SkipReason = "Nginx reload is not necessary"
		return op
	}

	op.Applicable = true
	op.Run = func(ctx context.Context) error {
		return e.nginxManager.Reload(ctx)
	}
	return op
}
