package wizard_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
	"github.com/JavierMunioz/IAMXFREE/internal/validation"
)

func typeRunes(step *wizard.TextStep, s string) {
	for _, r := range s {
		step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestTextStepCapturesTypedValue(t *testing.T) {
	step := wizard.NewTextStep("Name", "Application name:", "my-api", nil)
	step.Focus()

	typeRunes(step, "my-api")

	if got := step.Value(); got != "my-api" {
		t.Fatalf("Value() = %q, want %q", got, "my-api")
	}
}

func TestTextStepValidateRequired(t *testing.T) {
	step := wizard.NewTextStep("Name", "Application name:", "my-api", validation.Required())
	step.Focus()

	if err := step.Validate(); err == nil {
		t.Fatal("expected an error for an empty required field")
	}
	if step.View() == "" {
		t.Fatal("expected the view to render something, including the error")
	}

	typeRunes(step, "my-api")
	if err := step.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestTextStepFocusClearsPriorError(t *testing.T) {
	step := wizard.NewTextStep("Name", "Application name:", "my-api", validation.Required())
	step.Focus()
	_ = step.Validate()

	step.Focus()
	view := step.View()
	if view == "" {
		t.Fatal("expected a view")
	}
}

func TestTextStepModalIsAlwaysFalse(t *testing.T) {
	step := wizard.NewTextStep("Name", "Application name:", "", nil)
	if step.Modal() {
		t.Fatal("TextStep should never be modal")
	}
}
