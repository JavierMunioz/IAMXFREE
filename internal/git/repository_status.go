package git

// RepositoryStatus is a repository's working tree state plus how far its
// current branch has diverged from its upstream. Ahead is commits present
// locally but not yet pushed; Behind is commits present on the upstream
// but not yet pulled. Both are 0 when no upstream is configured — Notes
// explains why rather than that being indistinguishable from "up to date".
type RepositoryStatus struct {
	WorkingTree WorkingTree
	Ahead       int
	Behind      int
	Notes       []string
}
