// Package application composes the generic wizard engine (internal/tui/wizard)
// into the concrete step sequence for registering a new application. It is
// the only place that knows both the wizard engine and the application
// domain; the engine itself stays generic, and models.ApplicationDraft stays
// decoupled from how its fields are captured.
package application
