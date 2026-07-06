package wizard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Model sequences a fixed list of Steps, one at a time, and turns them into a
// Result once the last one is confirmed. It is the only piece of the wizard
// that knows about "the rest of the flow"; individual Steps do not.
type Model struct {
	title string
	steps []StepDef
	index int
}

// New builds a Wizard over steps, in order. title is shown in the header.
func New(title string, steps []StepDef) Model {
	m := Model{title: title, steps: steps}
	if len(m.steps) > 0 {
		m.steps[0].Step.Focus()
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.steps) == 0 {
		return m, nil
	}

	step := m.steps[m.index].Step

	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "ctrl+c" {
			return m, cancelCmd
		}

		if !step.Modal() {
			switch key.String() {
			case "esc":
				return m.back()
			case "enter":
				return m.advance()
			}
		}
	}

	cmd := step.Update(msg)
	return m, cmd
}

func (m Model) back() (Model, tea.Cmd) {
	if m.index == 0 {
		return m, cancelCmd
	}
	m.index--
	m.steps[m.index].Step.Focus()
	return m, nil
}

func (m Model) advance() (Model, tea.Cmd) {
	step := m.steps[m.index].Step

	if err := step.Validate(); err != nil {
		return m, nil
	}

	if m.index == len(m.steps)-1 {
		return m, m.completeCmd()
	}

	m.index++
	m.steps[m.index].Step.Focus()
	return m, nil
}

func (m Model) completeCmd() tea.Cmd {
	values := make(map[string]string, len(m.steps))
	for _, def := range m.steps {
		values[def.Key] = def.Step.Value()
	}
	result := Result{Values: values}
	return func() tea.Msg { return CompletedMsg{Result: result} }
}

func cancelCmd() tea.Msg { return CancelledMsg{} }

func (m Model) View() string {
	if len(m.steps) == 0 {
		return ""
	}

	def := m.steps[m.index]
	header := headerStyle.Render(fmt.Sprintf("%s — %d/%d · %s", m.title, m.index+1, len(m.steps), def.Step.Title()))
	hint := hintStyle.Render("enter: confirm  ·  esc: back  ·  ctrl+c: cancel")

	return header + "\n\n" + def.Step.View() + "\n\n" + hint
}
