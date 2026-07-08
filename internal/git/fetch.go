package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// Fetch updates path's remote-tracking refs via `git fetch`. It never
// touches the working tree or any local branch — the only operation this
// package performs that changes anything on disk at all.
//
// Unlike the query methods (IsRepository, Status, ...), a failed fetch
// (no remote, network unreachable, ...) is a real error here: Fetch is an
// action the caller asked to happen, not a question with a valid negative
// answer.
func (m *Manager) Fetch(ctx context.Context, path string) (FetchResult, error) {
	if strings.TrimSpace(path) == "" {
		return FetchResult{}, ErrEmptyPath
	}

	result, err := m.host.RunCaptured(ctx, runtimehost.Command{Name: "git", Args: []string{"fetch"}, Dir: path})
	output := strings.TrimSpace(result.Stdout + result.Stderr)
	if err != nil {
		return FetchResult{}, fmt.Errorf("git: fetch failed: %s: %w", output, err)
	}

	return FetchResult{Output: output}, nil
}
