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

func (m Model) renderHeader() string {
	status := mutedStyle.Render("Running…")
	if m.finished {
		status = mutedStyle.Render("Finished.")
	}
	return accentStyle.Render(fmt.Sprintf("Deploying — %s", m.plan.ApplicationName)) + "  " + status + "\n"
}

// renderOperationRow shows one operation's identity and current state —
// this is what lets a user see the current step, every completed step,
// and the step in progress, all in one always-visible list.
func renderOperationRow(result operations.OperationResult) string {
	icon, color := operationStatePresentation(result.State)
	line := lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s %s", icon, result.Name))
	if result.Message != "" {
		line += mutedStyle.Render(" — " + result.Message)
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
		"%d succeeded · %d failed · %d skipped · %d cancelled",
		summary.Succeeded, summary.Failed, summary.Skipped, summary.Cancelled,
	))
}
