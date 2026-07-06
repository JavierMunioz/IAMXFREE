package dashboard

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderTopBarShowsApplicationCount(t *testing.T) {
	m := modelWithApps(2)

	bar := m.renderTopBar()
	if !strings.Contains(bar, "apps 2") {
		t.Fatalf("expected top bar to mention apps 2, got:\n%s", bar)
	}
}

func TestRenderTopBarFitsConfiguredWidth(t *testing.T) {
	m := modelWithApps(1)
	m.width = 100

	line := strings.Split(m.renderTopBar(), "\n")[0]
	if got := lipgloss.Width(line); got != m.width {
		t.Fatalf("top bar line width = %d, want %d", got, m.width)
	}
}

func TestViewIncludesStatusLine(t *testing.T) {
	m := modelWithApps(1).SetStatus("hello")
	view := m.View()
	if !strings.Contains(view, "hello") {
		t.Fatalf("expected view to include the status message, got:\n%s", view)
	}
}
