package git

// DiffStat is a summarized diff between HEAD and the working tree (staged
// and unstaged changes combined) — file names and line counts, not the
// actual line-by-line diff.
type DiffStat struct {
	Files []FileDiffStat
}

// FileDiffStat is one file's contribution to a DiffStat. Binary is true
// when git could not report line counts for the file (it reports "-" for
// both), in which case Insertions and Deletions are 0 rather than a
// fabricated value.
type FileDiffStat struct {
	Path       string
	Insertions int
	Deletions  int
	Binary     bool
}
