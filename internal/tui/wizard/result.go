package wizard

// Result is the set of values a completed Wizard collected, keyed by each
// StepDef's Key.
type Result struct {
	Values map[string]string
}

// Get returns the captured value for key, or "" if it was never collected.
func (r Result) Get(key string) string {
	return r.Values[key]
}

// CompletedMsg is emitted once the last step has been confirmed. Whatever
// hosts the Wizard is responsible for turning Result into a domain object
// and persisting it — the Wizard itself never does either.
type CompletedMsg struct {
	Result Result
}

// CancelledMsg is emitted when the user aborts the Wizard before completing
// it (Ctrl+C anywhere, or Esc on the first step).
type CancelledMsg struct{}
