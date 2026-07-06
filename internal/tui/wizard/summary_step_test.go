package wizard_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
)

func TestSummaryStepRendersAndAlwaysValidates(t *testing.T) {
	step := wizard.NewSummaryStep("Summary", func() string {
		return "Name: my-api"
	})

	if err := step.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	if step.Modal() {
		t.Fatal("SummaryStep should never be modal")
	}
	if step.Value() != "" {
		t.Fatalf("Value() = %q, want empty", step.Value())
	}
	if got := step.View(); got == "" {
		t.Fatal("expected a non-empty view")
	}
}

func TestSummaryStepReflectsLiveState(t *testing.T) {
	name := "my-api"
	step := wizard.NewSummaryStep("Summary", func() string {
		return "Name: " + name
	})

	first := step.View()
	name = "renamed-api"
	second := step.View()

	if first == second {
		t.Fatal("expected the summary view to reflect updated closed-over state")
	}
}
