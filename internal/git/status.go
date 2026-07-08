package git

import (
	"context"
	"strings"
)

// Status reports path's working tree state. Ahead/Behind computation is
// added by a later change; for now they are always 0.
func (m *Manager) Status(ctx context.Context, path string) (RepositoryStatus, error) {
	if strings.TrimSpace(path) == "" {
		return RepositoryStatus{}, ErrEmptyPath
	}

	result, ok, err := m.run(ctx, path, "status", "--porcelain=v1")
	if err != nil {
		return RepositoryStatus{}, err
	}
	if !ok {
		return RepositoryStatus{}, nil
	}

	return RepositoryStatus{WorkingTree: parsePorcelainStatus(result.Stdout)}, nil
}

// parsePorcelainStatus turns `git status --porcelain=v1` output into a
// WorkingTree. Each line is "XY path": untracked files are marked "??";
// everything else (staged or unstaged additions, modifications, deletions,
// renames) is treated as Modified — this iteration does not distinguish
// staged from unstaged.
func parsePorcelainStatus(output string) WorkingTree {
	var tree WorkingTree

	for _, line := range strings.Split(output, "\n") {
		if len(line) < 4 {
			continue
		}

		status := line[:2]
		path := strings.TrimSpace(line[3:])

		if status == "??" {
			tree.Untracked = append(tree.Untracked, path)
		} else {
			tree.Modified = append(tree.Modified, path)
		}
	}

	tree.Clean = len(tree.Modified) == 0 && len(tree.Untracked) == 0
	return tree
}
