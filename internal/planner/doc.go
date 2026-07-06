// Package planner turns an inspection.Result into a DeploymentPlan: a
// proposed IAMXFREE configuration (suggested name, type, framework,
// runtime, package manager, install/build/start commands, ...) for the
// project that was inspected.
//
// The Inspector detects. The Planner interprets. Neither runs a command,
// writes a file, or persists anything — this package only produces a
// proposal for something else (the Wizard, in a later iteration) to show
// the user for confirmation or editing.
//
// This package depends only on internal/models and internal/inspection. It
// never depends on internal/tui, internal/execution, or any wizard/dashboard
// code, so it stays usable from anywhere.
package planner
