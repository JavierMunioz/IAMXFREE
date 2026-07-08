package git

import (
	"context"
	"errors"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// IsRepository reports whether path is inside a Git working tree. false is
// a normal outcome (path simply is not a Git repository), never an error —
// only an infrastructure problem (git missing, path unreadable) is.
func (m *Manager) IsRepository(ctx context.Context, path string) (bool, error) {
	if strings.TrimSpace(path) == "" {
		return false, ErrEmptyPath
	}

	result, ok, err := m.run(ctx, path, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false, err
	}
	if !ok {
		// git ran and reported this is not a repository — expected outcome.
		return false, nil
	}

	return strings.TrimSpace(result.Stdout) == "true", nil
}

// run executes a git subcommand with path as its working directory. The
// returned bool is true only when git ran and exited successfully — false
// with a nil error means git ran but reported a negative result (e.g. "not
// a git repository"), which callers treat as a normal outcome, not a
// failure. A non-nil error means git never actually ran at all (missing
// binary, bad path, ...), a real infrastructure problem.
func (m *Manager) run(ctx context.Context, path string, args ...string) (runtimehost.CommandResult, bool, error) {
	result, err := m.host.RunCaptured(ctx, runtimehost.Command{Name: "git", Args: args, Dir: path})
	if err == nil {
		return result, true, nil
	}

	var execErr *runtimehost.ExecutionError
	if !errors.As(err, &execErr) || execErr.ExitCode < 0 {
		return result, false, err
	}
	return result, false, nil
}
