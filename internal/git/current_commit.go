package git

import (
	"context"
	"strings"
	"time"
)

// commitFieldSeparator is git's %x1f format placeholder (ASCII unit
// separator) — a byte that will never appear in a commit's own fields, so
// splitting on it is unambiguous.
const commitFieldSeparator = "\x1f"

const commitLogFormat = "%H" + commitFieldSeparator + "%h" + commitFieldSeparator + "%s" + commitFieldSeparator + "%an" + commitFieldSeparator + "%aI"

// CurrentCommit reports path's current commit (HEAD). A repository with no
// commits yet reports a zero-value Commit, never an error — that is a
// normal state for a freshly initialized repo.
func (m *Manager) CurrentCommit(ctx context.Context, path string) (Commit, error) {
	if strings.TrimSpace(path) == "" {
		return Commit{}, ErrEmptyPath
	}

	result, ok, err := m.run(ctx, path, "log", "-1", "--format="+commitLogFormat)
	if err != nil {
		return Commit{}, err
	}
	if !ok {
		return Commit{}, nil
	}

	fields := strings.Split(strings.TrimSpace(result.Stdout), commitFieldSeparator)
	if len(fields) != 5 {
		return Commit{}, nil
	}

	commit := Commit{
		SHA:      fields[0],
		ShortSHA: fields[1],
		Message:  fields[2],
		Author:   fields[3],
	}
	if date, err := time.Parse(time.RFC3339, fields[4]); err == nil {
		commit.Date = date
	}

	return commit, nil
}
