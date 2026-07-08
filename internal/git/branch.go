package git

// Branch identifies a repository's current branch. Detached is true when
// HEAD does not point at a branch (a detached checkout) — in that case
// Name is empty rather than the literal string "HEAD" git itself reports.
type Branch struct {
	Name     string
	Detached bool
}
