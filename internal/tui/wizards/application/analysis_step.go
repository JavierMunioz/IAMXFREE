package application

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
)

var _ wizard.Step = (*AnalysisStep)(nil)

// AnalysisStep runs the project inspection/planning pipeline — via
// services.ApplicationSetupService, never directly against
// internal/inspection or internal/planner — the first time it is focused
// after the user confirms a project path, and shows the resulting proposal
// so the user can see exactly what was detected, inferred, or left
// undetermined before any field is pre-filled. The user can continue,
// go back to change the path, or cancel the wizard entirely (the same
// Enter/Esc/Ctrl+C the engine already gives every step).
//
// Inspection re-runs only when the path actually changes — navigating back
// and forth without editing the path reuses the cached proposal, so the
// filesystem is analyzed at most once per distinct path.
type AnalysisStep struct {
	setup     services.ApplicationSetupService
	pathValue func() string

	inspectedPath string
	hasInspected  bool
	proposal      services.ApplicationSetupProposal
	err           error
}

// NewAnalysisStep builds an AnalysisStep. pathValue returns the
// currently-confirmed project path — typically a preceding TextStep's
// Value method.
func NewAnalysisStep(setup services.ApplicationSetupService, pathValue func() string) *AnalysisStep {
	return &AnalysisStep{setup: setup, pathValue: pathValue}
}

func (s *AnalysisStep) Title() string { return "Analysis" }

func (s *AnalysisStep) Modal() bool { return false }

func (s *AnalysisStep) Focus() {
	path := s.pathValue()
	if s.hasInspected && path == s.inspectedPath {
		return
	}
	s.hasInspected = true
	s.inspectedPath = path
	s.proposal, s.err = s.setup.Inspect(context.Background(), path)
}

func (s *AnalysisStep) Update(tea.Msg) tea.Cmd { return nil }

// Validate blocks continuing past this step only when inspection itself
// failed (e.g. the path does not exist or cannot be read) — never because
// nothing interesting was detected, since manual entry is always a valid
// fallback.
func (s *AnalysisStep) Validate() error { return s.err }

// Value carries nothing forward on its own; later steps read Proposal
// directly via a closure to pre-fill themselves.
func (s *AnalysisStep) Value() string { return "" }

// Proposal returns whatever this step's last inspection found.
func (s *AnalysisStep) Proposal() services.ApplicationSetupProposal { return s.proposal }

func (s *AnalysisStep) View() string {
	if s.err != nil {
		return fmt.Sprintf(
			"Could not analyze this path:\n\n  %s\n\nesc: go back and fix the path",
			s.err.Error(),
		)
	}

	p := s.proposal
	var b strings.Builder
	b.WriteString("Here's what was detected at this path:\n\n")
	fmt.Fprintf(&b, "Project:         %s\n", valueOrDash(p.ProjectType))
	fmt.Fprintf(&b, "Framework:       %s\n", valueOrDash(string(p.Framework)))
	fmt.Fprintf(&b, "Runtime:         %s\n", valueOrDash(string(p.Runtime)))
	fmt.Fprintf(&b, "Package manager: %s\n", valueOrDash(p.PackageManager))
	fmt.Fprintf(&b, "Files found:     %s\n", valueOrDash(strings.Join(p.MatchedFiles, ", ")))
	fmt.Fprintf(&b, "Confidence:      %s\n", valueOrDash(p.Confidence))

	if len(p.Warnings) > 0 {
		b.WriteString("\nWarnings:\n")
		for _, w := range p.Warnings {
			fmt.Fprintf(&b, "  - %s\n", w)
		}
	}
	if len(p.Notes) > 0 {
		b.WriteString("\nNotes:\n")
		for _, n := range p.Notes {
			fmt.Fprintf(&b, "  - %s\n", n)
		}
	}

	b.WriteString("\nenter: continue  ·  esc: back  ·  ctrl+c: cancel")
	return b.String()
}
