package git

import (
	"context"
	"strings"
)

// CurrentBranch reports the branch path is currently checked out on. A
// detached HEAD (not pointing at any branch) reports Branch{Detached:
// true}, never an error.
func (m *Manager) CurrentBranch(ctx context.Context, path string) (Branch, error) {
	if strings.TrimSpace(path) == "" {
		return Branch{}, ErrEmptyPath
	}

	result, ok, err := m.run(ctx, path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return Branch{}, err
	}
	if !ok {
		return Branch{}, nil
	}

	name := strings.TrimSpace(result.Stdout)
	if name == "HEAD" {
		return Branch{Detached: true}, nil
	}
	return Branch{Name: name}, nil
}
