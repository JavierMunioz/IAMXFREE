package deploymentexec

import (
	"errors"
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/deployment"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

func TestViewShowsPreparingBeforeStarted(t *testing.T) {
	m := New(testEngine(&models.Application{}, nil), deployment.DeploymentPlan{})
	if !strings.Contains(m.View(), "Preparing deployment") {
		t.Fatalf("expected a preparing message, got:\n%s", m.View())
	}
}

func TestViewShowsBuildError(t *testing.T) {
	m := New(testEngine(nil, nil), deployment.DeploymentPlan{})
	m, _ = update(t, m, buildFailedMsg{err: errors.New("app not found")})

	if !strings.Contains(m.View(), "Could not start deployment") {
		t.Fatalf("expected an error message, got:\n%s", m.View())
	}
}

func TestViewShowsOperationRows(t *testing.T) {
	m := New(testEngine(&models.Application{}, nil), deployment.DeploymentPlan{ApplicationName: "my-api"})
	ch := make(chan operations.OperationProgress)
	close(ch)
	ops := []operations.Operation{
		{Name: "Fetch latest changes", Component: "git", Method: "Fetch"},
		{Name: "Install dependencies", Component: "execution", Method: "Install"},
	}
	m, _ = update(t, m, executionStartedMsg{ops: ops, ch: ch})

	view := m.View()
	for _, want := range []string{"my-api", "Fetch latest changes", "Install dependencies"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestViewShowsSummaryOnlyWhenFinished(t *testing.T) {
	m := New(testEngine(&models.Application{}, nil), deployment.DeploymentPlan{ApplicationName: "my-api"})
	ch := make(chan operations.OperationProgress)
	close(ch)
	ops := []operations.Operation{{Name: "one", Component: "test", Method: "one"}}
	m, _ = update(t, m, executionStartedMsg{ops: ops, ch: ch})

	if strings.Contains(m.View(), "Deployment succeeded") || strings.Contains(m.View(), "Deployment failed") {
		t.Fatalf("expected no summary before finished, got:\n%s", m.View())
	}

	m, _ = update(t, m, progressChannelClosedMsg{})
	if !strings.Contains(m.View(), "succeeded") {
		t.Fatalf("expected a summary once finished, got:\n%s", m.View())
	}
}

func TestViewShowsFailedSummary(t *testing.T) {
	m := New(testEngine(&models.Application{}, nil), deployment.DeploymentPlan{ApplicationName: "my-api"})
	m.started = true
	m.results = []operations.OperationResult{{Name: "one", State: operations.StateFailed}}

	m, _ = update(t, m, progressChannelClosedMsg{})
	if !strings.Contains(m.View(), "Deployment failed") {
		t.Fatalf("expected a failed verdict, got:\n%s", m.View())
	}
}

func TestViewShowsHeaderPhase(t *testing.T) {
	m := New(testEngine(&models.Application{}, nil), deployment.DeploymentPlan{ApplicationName: "my-api"})
	m.started = true
	m.results = []operations.OperationResult{{Name: "one", State: operations.StateSuccess}}

	if !strings.Contains(m.View(), "Executing") {
		t.Fatalf("expected an Executing phase indicator, got:\n%s", m.View())
	}

	m.results = []operations.OperationResult{{Name: "one", State: operations.StateFailed}}
	if !strings.Contains(m.View(), "Failed") {
		t.Fatalf("expected a Failed phase indicator, got:\n%s", m.View())
	}

	m.results = []operations.OperationResult{
		{Name: "stop", State: operations.StateSuccess, Compensation: &operations.CompensationResult{State: operations.StateCompensating}},
		{Name: "fail", State: operations.StateFailed},
	}
	if !strings.Contains(m.View(), "Compensating") {
		t.Fatalf("expected a Compensating phase indicator, got:\n%s", m.View())
	}
}

func TestViewShowsCompensationRow(t *testing.T) {
	m := New(testEngine(&models.Application{}, nil), deployment.DeploymentPlan{ApplicationName: "my-api"})
	m.started = true
	m.finished = true
	m.results = []operations.OperationResult{
		{Name: "Stop application", State: operations.StateSuccess, Compensation: &operations.CompensationResult{
			State: operations.StateCompensated, Message: "compensated successfully",
		}},
		{Name: "Reload Nginx", State: operations.StateFailed},
	}

	view := m.View()
	if !strings.Contains(view, "compensation") || !strings.Contains(view, "compensated successfully") {
		t.Fatalf("expected the compensation row to be shown, got:\n%s", view)
	}
}
