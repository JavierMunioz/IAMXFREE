package git

import (
	"context"
	"strconv"
	"strings"
)

// DiffSummary summarizes path's uncommitted changes (staged and unstaged,
// combined) as per-file line counts via `git diff HEAD --numstat` — not
// the line-by-line diff itself. A repository with no commits yet (no HEAD
// to diff against) reports an empty DiffStat, never an error.
func (m *Manager) DiffSummary(ctx context.Context, path string) (DiffStat, error) {
	if strings.TrimSpace(path) == "" {
		return DiffStat{}, ErrEmptyPath
	}

	result, ok, err := m.run(ctx, path, "diff", "HEAD", "--numstat")
	if err != nil {
		return DiffStat{}, err
	}
	if !ok {
		return DiffStat{}, nil
	}

	var stat DiffStat
	for _, line := range strings.Split(result.Stdout, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.SplitN(line, "\t", 3)
		if len(fields) != 3 {
			continue
		}

		entry := FileDiffStat{Path: fields[2]}
		if fields[0] == "-" || fields[1] == "-" {
			entry.Binary = true
		} else {
			entry.Insertions, _ = strconv.Atoi(fields[0])
			entry.Deletions, _ = strconv.Atoi(fields[1])
		}
		stat.Files = append(stat.Files, entry)
	}

	return stat, nil
}
