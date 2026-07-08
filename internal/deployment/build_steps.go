package deployment

import (
	"fmt"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// buildSteps describes the install and build commands a deployment would
// run, taken directly from the application's configured
// DeploymentConfig — no execution, just reporting what's configured.
func buildSteps(app *models.Application) []DeploymentStep {
	return []DeploymentStep{
		commandStep("Install dependencies", OperationInstallDependencies, app.Config.InstallCommand, "no install command configured"),
		commandStep("Build application", OperationBuild, app.Config.BuildCommand, "no build command configured"),
	}
}

// commandStep builds a step for a configured shell command: Skipped when
// command is empty, Ready describing what it would run otherwise.
func commandStep(name string, operation DeploymentOperation, command string, skippedReason string) DeploymentStep {
	step := DeploymentStep{
		Name:      name,
		Component: ComponentExecution,
		Operation: operation,
		Required:  command != "",
	}

	if command == "" {
		step.Status = StepStatusSkipped
		step.Expected = DeploymentResult{Description: skippedReason, WouldSucceed: true}
		return step
	}

	step.Status = StepStatusReady
	step.Expected = DeploymentResult{Description: fmt.Sprintf("run %q", command), WouldSucceed: true}
	return step
}
