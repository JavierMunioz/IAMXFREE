package deployment

import (
	"fmt"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// executionSteps answers "is the application running" and describes the
// restart every deployment needs, delegating entirely to ExecutionService.
// ActiveSession performs no I/O — it's an in-memory lookup — so this never
// touches the operating system.
func (e *Engine) executionSteps(app *models.Application) []DeploymentStep {
	session, running := e.executionService.ActiveSession(app.ID)

	check := DeploymentStep{
		Name:      "Check running status",
		Component: ComponentExecution,
		Operation: OperationCheckRunningStatus,
		Required:  true,
		Status:    StepStatusReady,
	}
	if running {
		check.Expected = DeploymentResult{Description: "application is currently running", WouldSucceed: true}
		check.Warnings = append(check.Warnings, fmt.Sprintf("application is currently running (PID %d)", session.PID))
	} else {
		check.Expected = DeploymentResult{Description: "application is not currently running", WouldSucceed: true}
	}

	restart := DeploymentStep{
		Name:      "Restart application",
		Component: ComponentExecution,
		Operation: OperationRestartApplication,
		Required:  true,
		Status:    StepStatusReady,
		Expected:  DeploymentResult{Description: "application running with the latest deployed code", WouldSucceed: true},
	}
	if running {
		restart.Risks = append(restart.Risks, "restarting will interrupt active connections")
	}

	return []DeploymentStep{check, restart}
}
