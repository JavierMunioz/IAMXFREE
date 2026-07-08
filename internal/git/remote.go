package git

// Remote is one remote a repository is configured to fetch/push against
// (e.g. "origin" -> "https://github.com/user/repo.git").
type Remote struct {
	Name string
	URL  string
}
