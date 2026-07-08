package deployment

import (
	"context"
	"fmt"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// preDeployHookOperation runs the application's configured pre-deploy hook,
// applicable only when the analysis phase found one configured. It runs
// first in the sequence, before anything else touches the application.
func (e *Engine) preDeployHookOperation(app *models.Application, plan DeploymentPlan) operations.Operation {
	return e.hookOperation(app, plan, OperationPreDeployHook, "Pre-deploy hook", app.Config.PreDeployHook)
}

// postDeployHookOperation runs the application's configured post-deploy
// hook, applicable only when the analysis phase found one configured. It
// runs last, once the deployment otherwise succeeded.
func (e *Engine) postDeployHookOperation(app *models.Application, plan DeploymentPlan) operations.Operation {
	return e.hookOperation(app, plan, OperationPostDeployHook, "Post-deploy hook", app.Config.PostDeployHook)
}

// hookOperation is shared by preDeployHookOperation/postDeployHookOperation:
// both run an arbitrary shell command directly through the host, since a
// hook is technology-agnostic and not tied to any execution.Strategy. It
// has no Compensate — a hook can have any side effect at all (notify a
// channel, warm a cache), so nothing here can safely be assumed reversible.
func (e *Engine) hookOperation(app *models.Application, plan DeploymentPlan, stepOperation DeploymentOperation, name, command string) operations.Operation {
	op := operations.Operation{Name: name, Component: string(ComponentHooks), Method: "Run"}

	step, found := findStep(plan, stepOperation)
	if !found || step.Status == StepStatusSkipped {
		op.SkipReason = "no " + strings.ToLower(name) + " command configured"
		return op
	}

	op.Applicable = true
	op.Run = func(ctx context.Context) error {
		result, err := e.host.RunCaptured(ctx, runtimehost.Command{
			Name: "sh",
			Args: []string{"-c", command},
			Dir:  app.Source.LocalPath,
		})
		if err != nil {
			output := strings.TrimSpace(result.Stdout + result.Stderr)
			return fmt.Errorf("%s failed: %s: %w", name, output, err)
		}
		return nil
	}
	return op
}
