package deploymentexec

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

func (m Model) View() string {
	if m.buildErr != nil {
		return errorStyle.Render("Could not start deployment: "+m.buildErr.Error()) + "\n\nesc/b: back  ·  q: quit"
	}
	if !m.started {
		return mutedStyle.Render("Preparing deployment…")
	}

	sections := []string{m.renderHeader()}
	for _, result := range m.results {
		sections = append(sections, renderOperationRow(result))
	}
	if m.finished {
		sections = append(sections, m.renderSummary())
	}
	sections = append(sections, footerStyle.Render("esc/b: back  ·  q: quit"))

	return strings.Join(sections, "\n")
}

// renderHeader shows the run's phase — Executing, Failed, Compensating or
// Finished — so the user always understands what's happening at a glance,
// distinct from any single operation's own state.
func (m Model) renderHeader() string {
	phase := currentPhase(m.results, m.finished)

	var status string
	switch phase {
	case phaseFailed:
		status = lipgloss.NewStyle().Foreground(colorCritical).Render(string(phaseFailed))
	case phaseCompensating:
		status = lipgloss.NewStyle().Foreground(colorWarning).Render(string(phaseCompensating) + "…")
	case phaseFinished:
		status = mutedStyle.Render(string(phaseFinished) + ".")
	default:
		status = mutedStyle.Render(string(phaseExecuting) + "…")
	}

	return accentStyle.Render(fmt.Sprintf("Deploying — %s", m.plan.ApplicationName)) + "  " + status + "\n"
}

// renderOperationRow shows one operation's identity, current state, and —
// if a compensation was attempted — its compensation state too. This is
// what lets a user see the current step, every completed step, and the
// step in progress, all in one always-visible list.
func renderOperationRow(result operations.OperationResult) string {
	icon, color := operationStatePresentation(result.State)
	line := lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s %s", icon, result.Name))
	if result.Message != "" {
		line += mutedStyle.Render(" — " + result.Message)
	}

	if result.Compensation != nil {
		compIcon, compColor := operationStatePresentation(result.Compensation.State)
		line += "\n  " + lipgloss.NewStyle().Foreground(compColor).Render(fmt.Sprintf("%s compensation", compIcon))
		if result.Compensation.Message != "" {
			line += mutedStyle.Render(" — " + result.Compensation.Message)
		}
	}

	return line
}

// renderSummary shows the run's final outcome once every operation has
// reached a terminal state.
func (m Model) renderSummary() string {
	summary := operations.Summarize(m.startedAt, m.results)

	verdict := lipgloss.NewStyle().Foreground(colorGood).Render("● Deployment succeeded")
	if summary.Overall == operations.StateFailed {
		verdict = lipgloss.NewStyle().Foreground(colorCritical).Render("✕ Deployment failed")
	} else if summary.Overall == operations.StateCancelled {
		verdict = mutedStyle.Render("◐ Deployment cancelled")
	}

	return "\n" + verdict + "\n" + mutedStyle.Render(fmt.Sprintf(
		"%d succeeded · %d failed · %d skipped · %d cancelled · %d compensated · %d compensation failed",
		summary.Succeeded, summary.Failed, summary.Skipped, summary.Cancelled, summary.Compensated, summary.CompensationFailed,
	))
}
