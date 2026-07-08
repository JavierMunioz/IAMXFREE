package git

// Repository is everything Manager.Inspect gathers about a directory in
// one call. IsRepo == false is a normal, expected outcome (the directory
// simply is not a Git repository) — everything else is left zero-valued in
// that case, never an error.
type Repository struct {
	Path   string
	IsRepo bool

	Branch  Branch
	Commit  Commit
	Remotes []Remote
	Status  RepositoryStatus
}
