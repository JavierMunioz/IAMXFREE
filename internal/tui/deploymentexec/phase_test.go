package deploymentexec

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

func TestCurrentPhaseExecuting(t *testing.T) {
	results := []operations.OperationResult{{State: operations.StateSuccess}, {State: operations.StatePending}}
	if got := currentPhase(results, false); got != phaseExecuting {
		t.Fatalf("currentPhase() = %q, want %q", got, phaseExecuting)
	}
}

func TestCurrentPhaseFailed(t *testing.T) {
	results := []operations.OperationResult{{State: operations.StateSuccess}, {State: operations.StateFailed}}
	if got := currentPhase(results, false); got != phaseFailed {
		t.Fatalf("currentPhase() = %q, want %q", got, phaseFailed)
	}
}

func TestCurrentPhaseCompensating(t *testing.T) {
	results := []operations.OperationResult{
		{State: operations.StateSuccess, Compensation: &operations.CompensationResult{State: operations.StateCompensating}},
		{State: operations.StateFailed},
	}
	if got := currentPhase(results, false); got != phaseCompensating {
		t.Fatalf("currentPhase() = %q, want %q", got, phaseCompensating)
	}
}

func TestCurrentPhaseFinished(t *testing.T) {
	results := []operations.OperationResult{{State: operations.StateFailed}}
	if got := currentPhase(results, true); got != phaseFinished {
		t.Fatalf("currentPhase() = %q, want %q", got, phaseFinished)
	}
}
