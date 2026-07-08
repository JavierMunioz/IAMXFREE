package git

import (
	"context"
	"strconv"
	"strings"
)

// Status reports path's working tree state plus, when an upstream is
// configured for the current branch, how many commits it is ahead/behind.
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

	status := RepositoryStatus{WorkingTree: parsePorcelainStatus(result.Stdout)}

	ahead, behind, hasUpstream, err := m.aheadBehind(ctx, path)
	if err != nil {
		return RepositoryStatus{}, err
	}
	if !hasUpstream {
		status.Notes = append(status.Notes, "no upstream configured for the current branch")
		return status, nil
	}
	status.Ahead = ahead
	status.Behind = behind

	return status, nil
}

// aheadBehind reports how many commits the current branch is ahead/behind
// its upstream. hasUpstream is false when no upstream is configured — a
// normal state, not an error.
func (m *Manager) aheadBehind(ctx context.Context, path string) (ahead, behind int, hasUpstream bool, err error) {
	_, ok, err := m.run(ctx, path, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		return 0, 0, false, err
	}
	if !ok {
		return 0, 0, false, nil
	}

	result, ok, err := m.run(ctx, path, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		return 0, 0, false, err
	}
	if !ok {
		return 0, 0, false, nil
	}

	fields := strings.Fields(result.Stdout)
	if len(fields) != 2 {
		return 0, 0, true, nil
	}
	ahead, _ = strconv.Atoi(fields[0])
	behind, _ = strconv.Atoi(fields[1])
	return ahead, behind, true, nil
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
