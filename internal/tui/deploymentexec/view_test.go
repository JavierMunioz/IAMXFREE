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
