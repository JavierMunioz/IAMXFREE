// Package monitor is the Runtime Monitor: it observes a Session's real
// operating-system state — process status, CPU/memory usage, and the
// context it was launched with — and reports it as a RuntimeSnapshot. It
// never starts, stops or otherwise modifies the process it observes, and
// every snapshot is built on demand; there is no background polling.
//
// It depends only on internal/execution (for the Session it observes) and
// internal/runtimehost (the only thing allowed to actually query the
// operating system) — never on internal/tui or internal/tui/dashboard, and
// never on internal/services. This is shared architecture: every current
// and future execution.Strategy is observed through this same Monitor,
// without any technology-specific code here.
package monitor
