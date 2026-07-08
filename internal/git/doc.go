// Package git implements the Git manager: read-only inspection of a
// directory's Git repository — whether it is one, its current branch and
// commit, its configured remotes, its working tree state, and how far it
// has diverged from its upstream. It is technology-agnostic: any kind of
// application (Node, PHP, Go, static, ...) that happens to live in a Git
// repository can be inspected the same way.
//
// This package depends only on runtimehost (every git invocation goes
// through it, never exec.Command directly) and models. It never imports
// the TUI, the Dashboard, the Wizard or the Execution Engine; those consume
// it indirectly, through the services layer.
//
// This iteration is read-only. Fetch is the only operation that touches
// anything (remote-tracking refs) — Pull, Push, Merge, Rebase and Checkout
// are explicitly out of scope until a later iteration.
package git
