package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Choice is one selectable option in a ChoiceStep.
type Choice struct {
	Label string
	Value string
}

const customChoiceValue = "__custom__"

// ChoiceStep is a reusable Step that lets the user pick one of a fixed list
// of options. If allowCustom is true, an extra "Otro..." option is appended;
// selecting it opens an inline text field (a self-contained sub-interaction
// signaled via Modal) so a value outside the predefined list can be entered.
type ChoiceStep struct {
	title       string
	prompt      string
	choices     []Choice
	allowCustom bool

	cursor      int
	customMode  bool
	customInput textinput.Model
	err         error
}

// NewChoiceStep builds a ChoiceStep over choices. When allowCustom is true,
// the user may type a value not present in choices.
func NewChoiceStep(title, prompt string, choices []Choice, allowCustom bool) *ChoiceStep {
	input := textinput.New()
	input.Placeholder = "type a custom value"
	input.CharLimit = 128

	return &ChoiceStep{
		title:       title,
		prompt:      prompt,
		choices:     choices,
		allowCustom: allowCustom,
		customInput: input,
	}
}

func (s *ChoiceStep) Title() string { return s.title }

func (s *ChoiceStep) Focus() {
	s.err = nil
	s.customMode = false
}

func (s *ChoiceStep) Modal() bool { return s.customMode }

func (s *ChoiceStep) options() []Choice {
	if !s.allowCustom {
		return s.choices
	}
	return append(append([]Choice{}, s.choices...), Choice{Label: "Otro...", Value: customChoiceValue})
}

func (s *ChoiceStep) Update(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	if s.customMode {
		switch key.String() {
		case "enter":
			if strings.TrimSpace(s.customInput.Value()) == "" {
				s.err = fmt.Errorf("this field is required")
				return nil
			}
			s.err = nil
			s.customMode = false
			return nil
		case "esc":
			s.customMode = false
			s.err = nil
			return nil
		}
		var cmd tea.Cmd
		s.customInput, cmd = s.customInput.Update(msg)
		return cmd
	}

	options := s.options()
	switch key.String() {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(options)-1 {
			s.cursor++
		}
	}
	return nil
}

func (s *ChoiceStep) Validate() error {
	options := s.options()
	if len(options) == 0 {
		s.err = nil
		return nil
	}

	selected := options[s.cursor]
	if selected.Value == customChoiceValue && strings.TrimSpace(s.customInput.Value()) == "" {
		s.customMode = true
		s.customInput.Focus()
		s.err = fmt.Errorf("type a custom value")
		return s.err
	}

	s.err = nil
	return nil
}

func (s *ChoiceStep) Value() string {
	options := s.options()
	if len(options) == 0 {
		return ""
	}
	selected := options[s.cursor]
	if selected.Value == customChoiceValue {
		return strings.TrimSpace(s.customInput.Value())
	}
	return selected.Value
}

func (s *ChoiceStep) View() string {
	var b strings.Builder
	b.WriteString(s.prompt + "\n")

	for i, choice := range s.options() {
		marker := "  "
		if i == s.cursor {
			marker = "> "
		}
		b.WriteString(marker + choice.Label + "\n")
	}

	if s.customMode {
		b.WriteString("\n" + s.customInput.View())
	}
	if s.err != nil {
		b.WriteString("\n" + errorStyle.Render(s.err.Error()))
	}

	return strings.TrimRight(b.String(), "\n")
}
