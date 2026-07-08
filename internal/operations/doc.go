// Package operations implements a small, reusable engine for running a
// sequence of typed operations: the Deployment Executor's runtime. It has
// no notion of Git, Nginx, a build command, or any particular technology —
// an Operation carries its own Run closure, supplied by whichever caller
// (internal/deployment) knows how to actually perform the action. This
// package only sequences them, tracks their state, and stops the run at
// the first failure.
//
// Every operation reports one of six states — Pending, Running, Success,
// Failed, Skipped, Cancelled — never a bare boolean, so a caller streaming
// progress always has enough information to explain what happened.
package operations
