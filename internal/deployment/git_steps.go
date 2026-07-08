package deployment

import (
	"context"
	"fmt"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// gitSteps answers "does a Git repository exist" and "are there local
// changes" for app, delegating entirely to the Git Manager. It returns the
// Repository it inspected alongside the steps so pullStep can reason about
// it without a second inspection.
func (e *Engine) gitSteps(ctx context.Context, app *models.Application) ([]DeploymentStep, git.Repository) {
	verify := DeploymentStep{
		Name:      "Verify Git repository",
		Component: ComponentGit,
		Operation: OperationVerifyRepository,
		Required:  true,
	}

	if strings.TrimSpace(app.Source.LocalPath) == "" {
		verify.Status = StepStatusBlocked
		verify.Risks = []string{"application has no source path configured"}
		verify.Expected = DeploymentResult{Description: "confirm the application's source is a Git repository", WouldSucceed: false}
		return []DeploymentStep{verify}, git.Repository{}
	}

	repo, err := e.gitManager.Inspect(ctx, app.Source.LocalPath)
	if err != nil {
		verify.Status = StepStatusBlocked
		verify.Risks = []string{fmt.Sprintf("could not inspect repository: %v", err)}
		verify.Expected = DeploymentResult{Description: "confirm the application's source is a Git repository", WouldSucceed: false}
		return []DeploymentStep{verify}, git.Repository{}
	}

	if !repo.IsRepo {
		verify.Status = StepStatusBlocked
		verify.Risks = []string{fmt.Sprintf("%q is not a Git repository", app.Source.LocalPath)}
		verify.Expected = DeploymentResult{Description: "confirm the application's source is a Git repository", WouldSucceed: false}
		return []DeploymentStep{verify}, repo
	}

	verify.Status = StepStatusReady
	verify.Expected = DeploymentResult{Description: fmt.Sprintf("repository on branch %q", repo.Branch.Name), WouldSucceed: true}

	changes := DeploymentStep{
		Name:      "Check for local changes",
		Component: ComponentGit,
		Operation: OperationCheckLocalChanges,
		Required:  true,
		Status:    StepStatusReady,
		Expected:  DeploymentResult{Description: "confirm the working tree matches what would be deployed", WouldSucceed: true},
	}

	if !repo.Status.WorkingTree.Clean {
		changes.Status = StepStatusWarning
		changes.Warnings = append(changes.Warnings, fmt.Sprintf(
			"%d modified and %d untracked file(s) are not part of any commit",
			len(repo.Status.WorkingTree.Modified), len(repo.Status.WorkingTree.Untracked),
		))
	}

	return []DeploymentStep{verify, changes}, repo
}

// pullStep describes whether the branch repo is on needs to be fast-forwarded
// from its upstream. It never calls git.Manager.Fetch — Behind/Ahead reflect
// remote-tracking refs as of the last fetch, exactly like the rest of this
// package's read-only guarantee requires.
func pullStep(repo git.Repository) DeploymentStep {
	step := DeploymentStep{
		Name:      "Pull latest changes",
		Component: ComponentGit,
		Operation: OperationPull,
		Required:  repo.Status.Behind > 0,
	}

	switch {
	case !repo.IsRepo:
		step.Status = StepStatusSkipped
		step.Expected = DeploymentResult{Description: "not applicable — no repository", WouldSucceed: false}

	case repo.Status.Behind == 0:
		step.Status = StepStatusSkipped
		step.Expected = DeploymentResult{Description: "already up to date with upstream", WouldSucceed: true}

	default:
		step.Status = StepStatusReady
		step.Expected = DeploymentResult{
			Description:  fmt.Sprintf("fast-forward local branch by %d commit(s)", repo.Status.Behind),
			WouldSucceed: true,
		}
		if repo.Status.Ahead > 0 {
			step.Risks = append(step.Risks, fmt.Sprintf(
				"local branch has %d unpushed commit(s); pulling may require a merge", repo.Status.Ahead,
			))
		}
	}

	return step
}
