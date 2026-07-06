package execution_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
)

func TestStatusIsValid(t *testing.T) {
	valid := []execution.Status{
		execution.StatusStarting,
		execution.StatusRunning,
		execution.StatusStopping,
		execution.StatusStopped,
		execution.StatusFailed,
	}
	for _, status := range valid {
		if !status.IsValid() {
			t.Errorf("IsValid() = false for %q, want true", status)
		}
	}

	if execution.Status("unknown").IsValid() {
		t.Error("expected an unknown status to be invalid")
	}
}
