package wizard_test

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
)

// fakeStep is a minimal Step used only to exercise wizard.Model's
// navigation logic in isolation from any real step implementation.
type fakeStep struct {
	title      string
	value      string
	invalid    bool
	modal      bool
	focusCalls int
}

func (s *fakeStep) Title() string          { return s.title }
func (s *fakeStep) Focus()                 { s.focusCalls++ }
func (s *fakeStep) Update(tea.Msg) tea.Cmd { return nil }
func (s *fakeStep) View() string           { return s.title }
func (s *fakeStep) Modal() bool            { return s.modal }
func (s *fakeStep) Value() string          { return s.value }

func (s *fakeStep) Validate() error {
	if s.invalid {
		return errors.New("invalid")
	}
	return nil
}

func enter() tea.Msg { return tea.KeyMsg{Type: tea.KeyEnter} }
func esc() tea.Msg   { return tea.KeyMsg{Type: tea.KeyEsc} }
func ctrlC() tea.Msg { return tea.KeyMsg{Type: tea.KeyCtrlC} }

func TestWizardFocusesFirstStepOnNew(t *testing.T) {
	first := &fakeStep{title: "first"}
	second := &fakeStep{title: "second"}
	wizard.New("test", []wizard.StepDef{{Key: "first", Step: first}, {Key: "second", Step: second}})

	if first.focusCalls != 1 {
		t.Fatalf("first.focusCalls = %d, want 1", first.focusCalls)
	}
	if second.focusCalls != 0 {
		t.Fatalf("second.focusCalls = %d, want 0", second.focusCalls)
	}
}

func TestWizardAdvanceStaysOnInvalidStep(t *testing.T) {
	first := &fakeStep{title: "first", invalid: true}
	second := &fakeStep{title: "second"}
	m := wizard.New("test", []wizard.StepDef{{Key: "first", Step: first}, {Key: "second", Step: second}})

	updated, _ := m.Update(enter())
	m = updated.(wizard.Model)

	if m.View() == "" {
		t.Fatal("expected a view")
	}
	if second.focusCalls != 0 {
		t.Fatal("expected second step to not be focused after a failed validation")
	}
}

func TestWizardAdvanceMovesToNextStep(t *testing.T) {
	first := &fakeStep{title: "first"}
	second := &fakeStep{title: "second"}
	m := wizard.New("test", []wizard.StepDef{{Key: "first", Step: first}, {Key: "second", Step: second}})

	updated, cmd := m.Update(enter())
	m = updated.(wizard.Model)

	if cmd != nil {
		t.Fatal("expected no command when moving to a non-final step")
	}
	if second.focusCalls != 1 {
		t.Fatalf("second.focusCalls = %d, want 1", second.focusCalls)
	}
}

func TestWizardBackGoesToPreviousStep(t *testing.T) {
	first := &fakeStep{title: "first"}
	second := &fakeStep{title: "second"}
	m := wizard.New("test", []wizard.StepDef{{Key: "first", Step: first}, {Key: "second", Step: second}})

	updated, _ := m.Update(enter())
	m = updated.(wizard.Model)

	updated, cmd := m.Update(esc())
	m = updated.(wizard.Model)

	if cmd != nil {
		t.Fatal("expected no command when going back")
	}
	if first.focusCalls != 2 {
		t.Fatalf("first.focusCalls = %d, want 2 (initial + re-focus on back)", first.focusCalls)
	}
}

func TestWizardBackOnFirstStepCancels(t *testing.T) {
	first := &fakeStep{title: "first"}
	m := wizard.New("test", []wizard.StepDef{{Key: "first", Step: first}})

	_, cmd := m.Update(esc())
	if cmd == nil {
		t.Fatal("expected a cancel command")
	}
	if _, ok := cmd().(wizard.CancelledMsg); !ok {
		t.Fatal("expected CancelledMsg")
	}
}

func TestWizardCtrlCCancelsFromAnyStep(t *testing.T) {
	first := &fakeStep{title: "first"}
	second := &fakeStep{title: "second", modal: true}
	m := wizard.New("test", []wizard.StepDef{{Key: "first", Step: first}, {Key: "second", Step: second}})

	updated, _ := m.Update(enter())
	m = updated.(wizard.Model)

	_, cmd := m.Update(ctrlC())
	if cmd == nil {
		t.Fatal("expected a cancel command")
	}
	if _, ok := cmd().(wizard.CancelledMsg); !ok {
		t.Fatal("expected CancelledMsg")
	}
}

func TestWizardCompletesOnLastStepConfirmed(t *testing.T) {
	first := &fakeStep{title: "first", value: "alice"}
	second := &fakeStep{title: "second", value: "backend"}
	m := wizard.New("test", []wizard.StepDef{{Key: "name", Step: first}, {Key: "type", Step: second}})

	updated, _ := m.Update(enter())
	m = updated.(wizard.Model)

	_, cmd := m.Update(enter())
	if cmd == nil {
		t.Fatal("expected a completion command")
	}

	msg, ok := cmd().(wizard.CompletedMsg)
	if !ok {
		t.Fatalf("expected CompletedMsg, got %T", cmd())
	}
	if msg.Result.Get("name") != "alice" || msg.Result.Get("type") != "backend" {
		t.Fatalf("unexpected result: %+v", msg.Result)
	}
}

func TestWizardModalStepReceivesEnterAndEscDirectly(t *testing.T) {
	step := &fakeStep{title: "modal", modal: true}
	m := wizard.New("test", []wizard.StepDef{{Key: "k", Step: step}})

	// Neither key should be treated as wizard navigation while Modal() is true.
	updated, cmd := m.Update(enter())
	m = updated.(wizard.Model)
	if cmd != nil {
		t.Fatal("expected enter to be forwarded to the modal step, not trigger completion")
	}

	_, cmd = m.Update(esc())
	if cmd != nil {
		t.Fatal("expected esc to be forwarded to the modal step, not trigger cancel")
	}
}
