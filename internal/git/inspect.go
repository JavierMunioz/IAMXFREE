package git

import (
	"context"
	"strings"
)

// Inspect gathers everything this package knows about path in one call:
// whether it is a Git repository, and if so its current branch, commit,
// remotes and status. IsRepo == false is a normal outcome — path just
// isn't a Git repository — and every other field is left zero-valued
// without attempting any further git calls.
func (m *Manager) Inspect(ctx context.Context, path string) (Repository, error) {
	if strings.TrimSpace(path) == "" {
		return Repository{}, ErrEmptyPath
	}

	isRepo, err := m.IsRepository(ctx, path)
	if err != nil {
		return Repository{}, err
	}
	if !isRepo {
		return Repository{Path: path}, nil
	}

	repo := Repository{Path: path, IsRepo: true}

	if repo.Branch, err = m.CurrentBranch(ctx, path); err != nil {
		return Repository{}, err
	}
	if repo.Commit, err = m.CurrentCommit(ctx, path); err != nil {
		return Repository{}, err
	}
	if repo.Remotes, err = m.Remotes(ctx, path); err != nil {
		return Repository{}, err
	}
	if repo.Status, err = m.Status(ctx, path); err != nil {
		return Repository{}, err
	}

	return repo, nil
}
