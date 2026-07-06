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

func TestTextStepPrefillAppliesOnFirstFocus(t *testing.T) {
	step := wizard.NewTextStep("Name", "Application name:", "", nil).
		WithPrefill(func() string { return "my-api" })

	step.Focus()

	if got := step.Value(); got != "my-api" {
		t.Fatalf("Value() = %q, want %q", got, "my-api")
	}
}

func TestTextStepPrefillNeverOverwritesUserEdit(t *testing.T) {
	value := "my-api"
	step := wizard.NewTextStep("Name", "Application name:", "", nil).
		WithPrefill(func() string { return value })

	step.Focus()
	typeRunes(step, "-custom") // user edits the prefilled value
	value = "a-different-name" // upstream data changes (e.g. re-inspection)
	step.Focus()               // step is revisited

	if got := step.Value(); got != "my-api-custom" {
		t.Fatalf("Value() = %q, want %q (user edit preserved)", got, "my-api-custom")
	}
}

func TestTextStepPrefillRefreshesUntouchedValue(t *testing.T) {
	value := "my-api"
	step := wizard.NewTextStep("Name", "Application name:", "", nil).
		WithPrefill(func() string { return value })

	step.Focus() // applies "my-api", untouched by the user
	value = "a-different-name"
	step.Focus() // should refresh since the user never edited it

	if got := step.Value(); got != "a-different-name" {
		t.Fatalf("Value() = %q, want %q (refreshed since untouched)", got, "a-different-name")
	}
}
