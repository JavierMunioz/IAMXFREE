package wizard_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
)

func newTestChoices() []wizard.Choice {
	return []wizard.Choice{
		{Label: "Frontend", Value: "frontend"},
		{Label: "Backend", Value: "backend"},
	}
}

func TestChoiceStepDefaultsToFirstOption(t *testing.T) {
	step := wizard.NewChoiceStep("Type", "Application type:", newTestChoices(), false)
	step.Focus()

	if got := step.Value(); got != "frontend" {
		t.Fatalf("Value() = %q, want %q", got, "frontend")
	}
	if err := step.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestChoiceStepCursorMovesSelection(t *testing.T) {
	step := wizard.NewChoiceStep("Type", "Application type:", newTestChoices(), false)
	step.Focus()

	step.Update(tea.KeyMsg{Type: tea.KeyDown})
	if got := step.Value(); got != "backend" {
		t.Fatalf("Value() = %q, want %q", got, "backend")
	}

	step.Update(tea.KeyMsg{Type: tea.KeyUp})
	if got := step.Value(); got != "frontend" {
		t.Fatalf("Value() = %q, want %q", got, "frontend")
	}
}

func TestChoiceStepNotModalByDefault(t *testing.T) {
	step := wizard.NewChoiceStep("Type", "Application type:", newTestChoices(), true)
	step.Focus()
	if step.Modal() {
		t.Fatal("expected ChoiceStep to start out non-modal")
	}
}

func TestChoiceStepCustomOptionEntersModalOnValidate(t *testing.T) {
	step := wizard.NewChoiceStep("Framework", "Framework:", newTestChoices(), true)
	step.Focus()

	// Move cursor to the appended "Otro..." option (index 2).
	step.Update(tea.KeyMsg{Type: tea.KeyDown})
	step.Update(tea.KeyMsg{Type: tea.KeyDown})

	if err := step.Validate(); err == nil {
		t.Fatal("expected Validate() to fail and enter custom mode when no custom value was typed")
	}
	if !step.Modal() {
		t.Fatal("expected the step to become modal after selecting the custom option")
	}
}

func TestChoiceStepCustomValueConfirmAndValidate(t *testing.T) {
	step := wizard.NewChoiceStep("Framework", "Framework:", newTestChoices(), true)
	step.Focus()
	step.Update(tea.KeyMsg{Type: tea.KeyDown})
	step.Update(tea.KeyMsg{Type: tea.KeyDown})
	_ = step.Validate() // enters modal mode

	for _, r := range "Remix" {
		step.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	step.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if step.Modal() {
		t.Fatal("expected confirming the custom value to exit modal mode")
	}
	if err := step.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	if got := step.Value(); got != "Remix" {
		t.Fatalf("Value() = %q, want %q", got, "Remix")
	}
}

func TestChoiceStepEscInCustomModeReturnsToList(t *testing.T) {
	step := wizard.NewChoiceStep("Framework", "Framework:", newTestChoices(), true)
	step.Focus()
	step.Update(tea.KeyMsg{Type: tea.KeyDown})
	step.Update(tea.KeyMsg{Type: tea.KeyDown})
	_ = step.Validate() // enters modal mode

	step.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if step.Modal() {
		t.Fatal("expected esc to exit modal mode")
	}
	if got := step.Value(); got != "" {
		t.Fatalf("Value() = %q, want empty after discarding custom entry", got)
	}
}

func TestChoiceStepPrefillSelectsMatchingOption(t *testing.T) {
	step := wizard.NewChoiceStep("Type", "Application type:", newTestChoices(), false).
		WithPrefill(func() string { return "backend" })

	step.Focus()

	if got := step.Value(); got != "backend" {
		t.Fatalf("Value() = %q, want %q", got, "backend")
	}
}

func TestChoiceStepPrefillIgnoresUnknownValue(t *testing.T) {
	step := wizard.NewChoiceStep("Type", "Application type:", newTestChoices(), false).
		WithPrefill(func() string { return "not-a-known-choice" })

	step.Focus()

	if got := step.Value(); got != "frontend" {
		t.Fatalf("Value() = %q, want %q (default, unknown prefill ignored)", got, "frontend")
	}
}

func TestChoiceStepPrefillNeverOverwritesUserSelection(t *testing.T) {
	value := "backend"
	step := wizard.NewChoiceStep("Type", "Application type:", newTestChoices(), false).
		WithPrefill(func() string { return value })

	step.Focus()                                    // selects backend
	step.Update(tea.KeyMsg{Type: tea.KeyUp})         // user picks frontend instead
	value = "backend"                                // upstream unchanged, still "backend"
	step.Focus()                                     // revisited

	if got := step.Value(); got != "frontend" {
		t.Fatalf("Value() = %q, want %q (user selection preserved)", got, "frontend")
	}
}
