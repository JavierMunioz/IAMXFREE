package deployment

import (
	"context"
	"fmt"
	"time"
)

// Plan builds a DeploymentPlan for appID: what a deployment would need to
// do, and whether it's currently safe to do it. It never changes
// anything — every fact comes from read-only calls to ApplicationService,
// ExecutionService, the Git Manager and the Nginx Manager.
//
// An error is only returned when appID itself can't be resolved; every
// other collaborator's failure is captured as a Risk on its own step
// instead of failing plan generation outright, so a problem with e.g.
// Nginx never hides what was still learned about Git.
func (e *Engine) Plan(ctx context.Context, appID string) (DeploymentPlan, error) {
	app, err := e.appService.Get(ctx, appID)
	if err != nil {
		return DeploymentPlan{}, err
	}

	var steps []DeploymentStep

	steps = append(steps, preDeployHookStep(app))

	gitVerifySteps, repo := e.gitSteps(ctx, app)
	steps = append(steps, gitVerifySteps...)
	if repo.IsRepo {
		steps = append(steps, pullStep(repo))
	}

	steps = append(steps, buildSteps(app)...)
	steps = append(steps, e.executionSteps(app)...)
	steps = append(steps, e.nginxSteps(ctx, app)...)
	steps = append(steps, postDeployHookStep(app))

	plan := DeploymentPlan{
		ApplicationID:   app.ID,
		ApplicationName: app.Name,
		GeneratedAt:     time.Now().UTC(),
		Steps:           steps,
		Summary:         summarize(steps),
	}
	for _, step := range steps {
		for _, warning := range step.Warnings {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("%s: %s", step.Name, warning))
		}
	}

	return plan, nil
}

// summarize reduces steps to the numbers DeploymentSummary reports.
func summarize(steps []DeploymentStep) DeploymentSummary {
	summary := DeploymentSummary{TotalSteps: len(steps)}

	for _, step := range steps {
		if step.Required {
			summary.RequiredSteps++
		}
		if step.Status == StepStatusBlocked {
			summary.BlockedSteps++
		}
		summary.WarningCount += len(step.Warnings)
	}

	summary.Ready = summary.BlockedSteps == 0
	return summary
}
