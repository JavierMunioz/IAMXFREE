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

func (s *TextStep) Title() string { return s.title }

func (s *TextStep) Focus() {
	s.err = nil
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
