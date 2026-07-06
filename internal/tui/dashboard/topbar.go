package dashboard

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderTopBar shows server identity on the left and live counters on the
// right. Data IAMXFREE cannot provide yet (server uptime) renders as a
// placeholder rather than a guess.
func (m Model) renderTopBar() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown-host"
	}
	osInfo := runtime.GOOS + "/" + runtime.GOARCH

	left := fmt.Sprintf("%s  %s", accentStyle.Render(host), mutedStyle.Render(osInfo))
	right := fmt.Sprintf(
		"uptime %s  ·  apps %s  ·  %s",
		mutedStyle.Render("—"),
		primaryStyle.Render(fmt.Sprintf("%d", len(m.apps))),
		primaryStyle.Render(time.Now().Format("15:04:05")),
	)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	line := left + strings.Repeat(" ", gap) + right
	rule := mutedStyle.Render(strings.Repeat("─", max(m.width, 1)))

	return topBarStyle.Render(line) + "\n" + rule
}
