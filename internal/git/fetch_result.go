package git

// FetchResult is the outcome of running `git fetch`: it updates the
// repository's remote-tracking refs (needed to compute Ahead/Behind
// accurately) without touching the working tree or any local branch.
type FetchResult struct {
	Output string
}
