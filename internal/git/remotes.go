package git

import (
	"context"
	"strings"
)

// Remotes lists path's configured remotes. A repository with none
// configured reports an empty slice, never an error.
func (m *Manager) Remotes(ctx context.Context, path string) ([]Remote, error) {
	if strings.TrimSpace(path) == "" {
		return nil, ErrEmptyPath
	}

	result, ok, err := m.run(ctx, path, "remote", "-v")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	var remotes []Remote
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasSuffix(line, "(fetch)") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		remotes = append(remotes, Remote{Name: fields[0], URL: fields[1]})
	}

	return remotes, nil
}
