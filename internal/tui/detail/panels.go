package detail

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// renderMiddleRow lays out the technical panel (left) and the execution
// status panel (right) side by side.
func (m Model) renderMiddleRow() string {
	half := (m.width - 4) / 2
	if half < 20 {
		half = 20
	}

	left := m.renderTechnicalPanel(half)
	right := m.renderExecutionPanel(half)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// renderTechnicalPanel shows where the application lives and how it is
// installed/built/started — the configuration behind the general info shown
// on the top panel.
func (m Model) renderTechnicalPanel(width int) string {
	app := m.app

	rows := [][2]string{
		{"Path", valueOrDash(app.Source.LocalPath)},
		{"Repository", valueOrDash(app.Source.RepositoryURL)},
		{"Pkg manager", valueOrDash(app.Config.PackageManager)},
		{"Install", valueOrDash(app.Config.InstallCommand)},
		{"Build", valueOrDash(app.Config.BuildCommand)},
		{"Start", valueOrDash(app.Config.StartCommand)},
		{"Registered", app.CreatedAt.Format("2006-01-02 15:04")},
	}

	var lines []string
	lines = append(lines, primaryStyle.Bold(true).Render("Technical"))
	lines = append(lines, "")
	for _, row := range rows {
		lines = append(lines, labelStyle.Render(row[0])+row[1])
	}

	return panelStyle.Width(width).Height(9).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// renderMetricsPanel shows the currently tracked session's resource usage
// as last observed by the Runtime Monitor. It only ever reflects the last
// explicit refresh (f5) — there is no automatic polling — so a metric this
// environment cannot report, or that was never fetched, is shown as "n/a"
// rather than a fabricated value.
func (m Model) renderMetricsPanel(width int) string {
	var lines []string
	lines = append(lines, primaryStyle.Bold(true).Render("Metrics"))
	lines = append(lines, "")

	if !m.hasSnapshot {
		lines = append(lines, mutedStyle.Render("No metrics yet."))
		lines = append(lines, mutedStyle.Render("Press f5 to refresh."))
	} else {
		rows := [][2]string{
			{"CPU", formatPercentMetric(m.snapshot.CPUPercent)},
			{"Memory (RSS)", formatBytesMetric(m.snapshot.MemoryRSSBytes)},
			{"Memory (VSZ)", formatBytesMetric(m.snapshot.MemoryVSZBytes)},
		}
		for _, row := range rows {
			lines = append(lines, labelStyle.Render(row[0])+row[1])
		}
	}

	return panelStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func formatPercentMetric(metric services.Metric) string {
	if !metric.Available {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", metric.Value)
}

func formatBytesMetric(metric services.Metric) string {
	if !metric.Available {
		return "n/a"
	}
	return fmt.Sprintf("%.1f MB", metric.Value/(1024*1024))
}

// renderGitPanel shows the application's Git repository state: current
// branch, last commit (short SHA), working tree cleanliness and configured
// remote. Read-only, and only ever reflects the last load (Init or f5) —
// there is no automatic polling.
func (m Model) renderGitPanel(width int) string {
	var lines []string
	lines = append(lines, primaryStyle.Bold(true).Render("Source Control"))
	lines = append(lines, "")

	switch {
	case !m.hasGitStatus:
		lines = append(lines, mutedStyle.Render("Loading…"))
	case !m.gitStatus.IsRepo:
		lines = append(lines, mutedStyle.Render("No Git repository."))
	default:
		branch := m.gitStatus.Branch
		if m.gitStatus.Detached {
			branch = "detached HEAD"
		}

		workingTree := lipgloss.NewStyle().Foreground(colorGood).Render("● clean")
		if !m.gitStatus.Clean {
			workingTree = lipgloss.NewStyle().Foreground(colorCritical).Render(
				fmt.Sprintf("✕ dirty (%d modified, %d untracked)", m.gitStatus.Modified, m.gitStatus.Untracked))
		}

		rows := [][2]string{
			{"Branch", valueOrDash(branch)},
			{"Commit", valueOrDash(m.gitStatus.CommitSHA)},
			{"Working tree", workingTree},
			{"Remote", valueOrDash(m.gitStatus.RemoteName)},
		}
		for _, row := range rows {
			lines = append(lines, labelStyle.Render(row[0])+row[1])
		}
	}

	return panelStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// renderExecutionPanel shows the live session(s) IAMXFREE is tracking for
// this application. When no session is active, that is stated clearly
// rather than leaving fields blank. While a zero-downtime deployment has a
// candidate session started alongside the active one, both are shown,
// labeled Active/Candidate, with a note on which one currently receives
// traffic — the panel simply grows taller for as long as that lasts.
func (m Model) renderExecutionPanel(width int) string {
	var lines []string
	lines = append(lines, primaryStyle.Bold(true).Render("Execution"))
	lines = append(lines, "")

	if !m.hasSession {
		lines = append(lines, mutedStyle.Render("No active session."))
		lines = append(lines, mutedStyle.Render("Press s to start."))
	} else {
		label := "Active"
		if m.hasCandidateSession {
			label = "Active — receiving traffic"
		}
		lines = append(lines, mutedStyle.Render(label))
		lines = append(lines, sessionRows(m.session)...)
	}

	if m.hasCandidateSession {
		lines = append(lines, "")
		lines = append(lines, mutedStyle.Render("Candidate — not yet receiving traffic"))
		lines = append(lines, sessionRows(m.candidateSession)...)
	}

	return panelStyle.Width(width).Height(9).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// sessionRows renders the fields shared by the active and candidate
// sessions in renderExecutionPanel — a single source of layout so the two
// never drift apart in shape.
func sessionRows(session services.RunSession) []string {
	running := session.Status == "running"
	runningText := lipgloss.NewStyle().Foreground(colorCritical).Render("✕ stopped")
	if running {
		runningText = lipgloss.NewStyle().Foreground(colorGood).Render("● running")
	}

	port := ""
	if session.Port != 0 {
		port = fmt.Sprintf("%d", session.Port)
	}

	rows := [][2]string{
		{"Running", runningText},
		{"PID", fmt.Sprintf("%d", session.PID)},
		{"Port", valueOrDash(port)},
		{"Started", session.StartedAt.Format("2006-01-02 15:04:05")},
		{"Work dir", valueOrDash(session.WorkingDir)},
		{"Runtime", valueOrDash(string(session.Runtime))},
	}
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, labelStyle.Render(row[0])+row[1])
	}
	return lines
}
