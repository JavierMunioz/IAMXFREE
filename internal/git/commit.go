package git

import "time"

// Commit is the repository's current commit (HEAD). SHA is the full hash;
// ShortSHA is the abbreviated form suited for display (e.g. a detail
// panel). A repository with no commits yet reports a zero-value Commit,
// never an error — that's a normal state for a freshly initialized repo.
type Commit struct {
	SHA      string
	ShortSHA string
	Message  string
	Author   string
	Date     time.Time
}
