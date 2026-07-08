package git

// WorkingTree is the state of a repository's uncommitted changes. Clean is
// true only when both Modified and Untracked are empty. Modified covers
// both staged and unstaged changes to tracked files (added, changed,
// deleted) — this iteration does not distinguish the two.
type WorkingTree struct {
	Clean     bool
	Modified  []string
	Untracked []string
}
