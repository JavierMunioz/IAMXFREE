package operations

import "context"

// Operation is one executable unit the Executor runs. It knows which
// component performs the action (Component), what method that component
// exposes (Method), and what to call (Run) — never how Run works
// internally. The caller that builds an Operation (internal/deployment) is
// the only one that understands the component behind it.
type Operation struct {
	// Name is a short, human-readable label (e.g. "Install dependencies").
	Name string

	// Component names who is responsible (e.g. "git", "nginx",
	// "execution"). The Executor never interprets this — it is for
	// display only.
	Component string

	// Method names what's being invoked (e.g. "Fetch", "Install",
	// "Reload"). Display only, same as Component.
	Method string

	// Applicable reports whether this Operation should actually run. When
	// false, the Executor reports it as Skipped with SkipReason and never
	// calls Run.
	Applicable bool
	SkipReason string

	// Run performs the action. A non-nil error means the operation
	// failed; Run never needs to know about OperationState itself — the
	// Executor derives it from whether Run returned an error.
	Run func(ctx context.Context) error

	// Compensate, if set, undoes what Run did — called only when this
	// Operation succeeded and a later Operation in the same run failed.
	// Not every Operation is reversible (e.g. a build has nothing
	// meaningful to undo); leaving Compensate nil is how an Operation
	// says so. Like Run, Compensate never needs to know about
	// OperationState — a non-nil error means the compensation itself
	// failed.
	Compensate func(ctx context.Context) error
}
