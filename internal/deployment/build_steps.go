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
		commandStep(ComponentExecution, "Install dependencies", OperationInstallDependencies, app.Config.InstallCommand, "no install command configured"),
		commandStep(ComponentExecution, "Build application", OperationBuild, app.Config.BuildCommand, "no build command configured"),
	}
}

// commandStep builds a step for a configured shell command: Skipped when
// command is empty, Ready describing what it would run otherwise.
func commandStep(component DeploymentComponent, name string, operation DeploymentOperation, command string, skippedReason string) DeploymentStep {
	step := DeploymentStep{
		Name:      name,
		Component: component,
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
