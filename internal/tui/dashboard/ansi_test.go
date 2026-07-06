package dashboard

import (
	"regexp"
	"testing"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes color escape codes so tests can assert on plain text
// regardless of whether lipgloss decided to render color in this
// environment.
func stripANSI(t *testing.T, s string) string {
	t.Helper()
	return ansiEscape.ReplaceAllString(s, "")
}
