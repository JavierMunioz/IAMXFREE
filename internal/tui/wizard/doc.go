// Package wizard is a generic, reusable multi-step form engine for the TUI.
// It knows nothing about applications, Nginx, SSL or any other feature —
// only how to sequence a list of Step values, one screen at a time, and hand
// back whatever they collected. Every concrete wizard (create an
// application, configure Nginx, add a VPS, ...) is built by composing Step
// implementations and handing them to wizard.New; the engine itself never
// changes.
package wizard
