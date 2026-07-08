package deploymentplan

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/deployment"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestViewShowsLoadingBeforePlanArrives(t *testing.T) {
	m := New(testEngine(&models.Application{}, nil), "app-1")
	if !strings.Contains(m.View(), "Building deployment plan") {
		t.Fatalf("expected a loading message, got:\n%s", m.View())
	}
}

func TestViewShowsErrorOnLoadFailure(t *testing.T) {
	m := New(testEngine(nil, nil), "app-1")
	m, _ = update(t, m, planLoadFailedMsg{err: errors.New("boom")})

	if !strings.Contains(m.View(), "Could not build deployment plan") {
		t.Fatalf("expected an error message, got:\n%s", m.View())
	}
}

func TestViewRendersPlanSummaryAndSteps(t *testing.T) {
	m := New(testEngine(&models.Application{}, nil), "app-1")
	m, _ = update(t, m, planLoadedMsg{plan: deployment.DeploymentPlan{
		ApplicationName: "my-api",
		GeneratedAt:     time.Now(),
		Steps: []deployment.DeploymentStep{
			{
				Name:      "Verify Git repository",
				Component: deployment.ComponentGit,
				Operation: deployment.OperationVerifyRepository,
				Status:    deployment.StepStatusReady,
				Required:  true,
				Expected:  deployment.DeploymentResult{Description: "repository on branch main"},
			},
			{
				Name:      "Restart application",
				Component: deployment.ComponentExecution,
				Operation: deployment.OperationRestartApplication,
				Status:    deployment.StepStatusWarning,
				Required:  true,
				Warnings:  []string{"restarting will interrupt active connections"},
				Expected:  deployment.DeploymentResult{Description: "application running with the latest deployed code"},
			},
		},
		Summary: deployment.DeploymentSummary{TotalSteps: 2, RequiredSteps: 2, Ready: true},
	}})

	view := m.View()
	for _, want := range []string{
		"my-api", "Ready to deploy",
		"Verify Git repository", "repository on branch main",
		"Restart application", "restarting will interrupt active connections",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestViewShowsNotReadyWhenBlocked(t *testing.T) {
	m := New(testEngine(&models.Application{}, nil), "app-1")
	m, _ = update(t, m, planLoadedMsg{plan: deployment.DeploymentPlan{
		ApplicationName: "my-api",
		Steps: []deployment.DeploymentStep{
			{Name: "Verify Git repository", Status: deployment.StepStatusBlocked, Required: true, Risks: []string{"no source path configured"}},
		},
		Summary: deployment.DeploymentSummary{TotalSteps: 1, RequiredSteps: 1, BlockedSteps: 1, Ready: false},
	}})

	view := m.View()
	if !strings.Contains(view, "Not ready") {
		t.Errorf("expected a not-ready indicator, got:\n%s", view)
	}
	if !strings.Contains(view, "no source path configured") {
		t.Errorf("expected the risk to be shown, got:\n%s", view)
	}
}
