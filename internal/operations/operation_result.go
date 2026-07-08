package operations

import "time"

// OperationResult is what one Operation produced: enough to display on its
// own, without needing the original Operation alongside it.
type OperationResult struct {
	Name      string
	Component string
	Method    string

	State   OperationState
	Message string
	Err     *OperationError

	StartedAt  time.Time
	FinishedAt time.Time

	// Compensation is nil unless the Executor attempted to compensate
	// this operation (only possible when State == StateSuccess and a
	// later operation failed). It is the operation's execution outcome
	// and its compensation outcome, kept separate — an operation can have
	// State == StateSuccess and a Compensation that says it was later
	// undone.
	Compensation *CompensationResult
}
