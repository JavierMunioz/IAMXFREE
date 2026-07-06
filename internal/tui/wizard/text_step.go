package wizard

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/validation"
)

// TextStep is a reusable Step that captures a single line of free text,
// optionally checked against a validation.Validator.
type TextStep struct {
	title    string
	prompt   string
	validate validation.Validator
	input    textinput.Model
	err      error

	prefill       func() string
	everPrefilled bool
	lastPrefilled string
}

// NewTextStep builds a TextStep. validate may be nil to accept any value.
func NewTextStep(title, prompt, placeholder string, validate validation.Validator) *TextStep {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 256

	return &TextStep{
		title:    title,
		prompt:   prompt,
		validate: validate,
		input:    input,
	}
}

// WithPrefill registers fn as a source of an initial value for this step.
// It is called every time the step is focused, but only applied when the
// current value still matches whatever fn last returned — so a value the
// user typed themselves is never overwritten, while a value nothing has
// touched yet stays in sync with upstream data (e.g. a prior step's
// analysis being re-run after the user goes back and changes it).
func (s *TextStep) WithPrefill(fn func() string) *TextStep {
	s.prefill = fn
	return s
}

func (s *TextStep) Title() string { return s.title }

func (s *TextStep) Focus() {
	s.err = nil
	if s.prefill != nil && (!s.everPrefilled || s.input.Value() == s.lastPrefilled) {
		next := s.prefill()
		s.input.SetValue(next)
		s.lastPrefilled = next
		s.everPrefilled = true
	}
	s.input.Focus()
}

func (s *TextStep) Modal() bool { return false }

func (s *TextStep) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}

func (s *TextStep) Validate() error {
	if s.validate == nil {
		s.err = nil
		return nil
	}
	if err := s.validate(s.input.Value()); err != nil {
		s.err = err
		return err
	}
	s.err = nil
	return nil
}

func (s *TextStep) Value() string {
	return strings.TrimSpace(s.input.Value())
}

func (s *TextStep) View() string {
	view := s.prompt + "\n" + s.input.View()
	if s.err != nil {
		view += "\n" + errorStyle.Render(s.err.Error())
	}
	return view
}
