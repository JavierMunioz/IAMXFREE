package wizard

import tea "github.com/charmbracelet/bubbletea"

// SummaryStep is a reusable, read-only review screen: it renders whatever
// render returns, always validates successfully, and captures no value of
// its own. It is typically the last StepDef in a wizard, letting the caller
// review previously captured values (read from the other steps by closure,
// which only the composing feature package needs to know about) before the
// Wizard emits CompletedMsg.
type SummaryStep struct {
	title  string
	render func() string
}

// NewSummaryStep builds a SummaryStep. render is called on every View() so
// it always reflects the current state of whatever it closes over.
func NewSummaryStep(title string, render func() string) *SummaryStep {
	return &SummaryStep{title: title, render: render}
}

func (s *SummaryStep) Title() string          { return s.title }
func (s *SummaryStep) Focus()                 {}
func (s *SummaryStep) Modal() bool            { return false }
func (s *SummaryStep) Update(tea.Msg) tea.Cmd { return nil }
func (s *SummaryStep) Validate() error        { return nil }
func (s *SummaryStep) Value() string          { return "" }

func (s *SummaryStep) View() string {
	return s.render() + "\n\n" + hintStyle.Render("enter: confirm and save")
}
