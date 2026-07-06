package wizard

import tea "github.com/charmbracelet/bubbletea"

// Step is a single, self-contained screen in a Wizard. A Step only knows how
// to render itself, how to validate the value it currently holds, and how to
// hand that value back — it has no awareness of sibling steps, its position
// in the sequence, or how to move between steps. The Wizard alone is
// responsible for coordinating navigation.
type Step interface {
	// Title is a short label shown in the Wizard's progress indicator.
	Title() string

	// Focus (re)initializes the step right before it becomes the active
	// one, e.g. focusing a text input or clearing a stale error.
	Focus()

	// Update handles one input message not claimed by the Wizard itself
	// (see Modal). It never receives "enter"/"esc" unless Modal() is true.
	Update(msg tea.Msg) tea.Cmd

	// View renders the step's prompt and current input state, including
	// its own validation error if it has one.
	View() string

	// Validate checks the step's current value, recording an error that
	// View can render if it is invalid.
	Validate() error

	// Value returns the step's captured value once it is valid.
	Value() string

	// Modal reports whether the step is in a self-contained sub-interaction
	// (e.g. typing a custom option) that should receive "enter"/"esc"
	// directly instead of having the Wizard treat them as
	// confirm-and-advance / go-back.
	Modal() bool
}

// StepDef pairs a Step with the key its captured value is stored under in
// the Wizard's final Result.
type StepDef struct {
	Key  string
	Step Step
}
