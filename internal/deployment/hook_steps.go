package deployment

import "github.com/JavierMunioz/IAMXFREE/internal/models"

// preDeployHookStep describes the application's configured pre-deploy hook
// command, taken directly from DeploymentConfig — no execution, just
// reporting what's configured. Skipped when no hook command is set.
func preDeployHookStep(app *models.Application) DeploymentStep {
	return commandStep(ComponentHooks, "Pre-deploy hook", OperationPreDeployHook, app.Config.PreDeployHook, "no pre-deploy hook configured")
}

// postDeployHookStep describes the application's configured post-deploy
// hook command, the same way preDeployHookStep does for the pre-deploy one.
func postDeployHookStep(app *models.Application) DeploymentStep {
	return commandStep(ComponentHooks, "Post-deploy hook", OperationPostDeployHook, app.Config.PostDeployHook, "no post-deploy hook configured")
}
