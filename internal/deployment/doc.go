// Package deployment implements the Deployment Engine: the orchestrator
// that coordinates IAMXFREE's existing managers to plan a deployment. It
// represents a workflow, not a technology — it has no logic of its own for
// talking to Git, Nginx or a running process; every question it answers is
// delegated to the component that actually knows the answer (the Git
// Manager, the Nginx Manager, ExecutionService).
//
// This package depends only on ApplicationService, ExecutionService, the
// Git Manager, the Nginx Manager and runtimehost. It never imports the
// TUI, the Dashboard or the Wizard; those consume it indirectly, through
// the services layer.
//
// This iteration is entirely read-only and deterministic: Plan gathers
// facts and describes the steps a deployment would need, but never
// executes any of them. Pull, Build, Restart, Reload, SSL, rollback and
// hooks are all future work — this iteration only explains what would
// happen.
package deployment
