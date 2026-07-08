package deploymentplan

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JavierMunioz/IAMXFREE/internal/deployment"
)

func (m Model) View() string {
	if m.loadErr != nil {
		return errorStyle.Render("Could not build deployment plan: "+m.loadErr.Error()) + "\n\nesc/b: back  ·  q: quit"
	}
	if !m.loaded {
		return mutedStyle.Render("Building deployment plan…")
	}

	sections := []string{m.renderHeader()}
	for _, step := range m.plan.Steps {
		sections = append(sections, m.renderStep(step))
	}
	sections = append(sections, footerStyle.Render("enter: execute plan  ·  esc/b: back  ·  q: quit"))

	return strings.Join(sections, "\n\n")
}

// renderHeader shows the application, overall readiness, and the numbers
// DeploymentSummary reports — everything a caller needs before reading a
// single step.
func (m Model) renderHeader() string {
	readiness := lipgloss.NewStyle().Foreground(colorGood).Render("● Ready to deploy")
	if !m.plan.Summary.Ready {
		readiness = lipgloss.NewStyle().Foreground(colorCritical).Render("✕ Not ready — blocked step(s) present")
	}

	lines := []string{
		accentStyle.Render(fmt.Sprintf("Deployment Plan — %s", m.plan.ApplicationName)),
		"",
		readiness,
		mutedStyle.Render(fmt.Sprintf(
			"%d step(s) · %d required · %d blocked · %d warning(s)",
			m.plan.Summary.TotalSteps, m.plan.Summary.RequiredSteps, m.plan.Summary.BlockedSteps, m.plan.Summary.WarningCount,
		)),
		mutedStyle.Render("Generated " + m.plan.GeneratedAt.Format("2006-01-02 15:04:05")),
	}
	return strings.Join(lines, "\n")
}

// renderStep shows one DeploymentStep in full: what it's called, who's
// responsible, whether it's required, and everything the Engine found out
// about it — never whether it ran, since it never does this iteration.
func (m Model) renderStep(step deployment.DeploymentStep) string {
	icon, color := stepStatusPresentation(step.Status)
	requiredText := "optional"
	if step.Required {
		requiredText = "required"
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("%s %s", icon, step.Name)))
	lines = append(lines, mutedStyle.Render(fmt.Sprintf("%s · %s · %s", step.Component, step.Operation, requiredText)))
	lines = append(lines, labelStyle.Render("Expected")+step.Expected.Description)

	for _, risk := range step.Risks {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorCritical).Render("risk: "+risk))
	}
	for _, warning := range step.Warnings {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorWarning).Render("warn: "+warning))
	}

	width := m.width - 4
	if width < 20 {
		width = 20
	}
	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}
