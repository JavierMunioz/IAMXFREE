package operations

import "time"

// CompensationResult is what one Operation's Compensate produced. It is
// nil on an OperationResult until the Executor actually attempts a
// compensation — an operation that succeeded but was never compensated
// (because nothing failed after it, or it has no Compensate at all) has
// no CompensationResult, not a zero-valued one.
type CompensationResult struct {
	State   OperationState // StateCompensating while running, then StateCompensated or StateCompensationFailed
	Message string
	Err     *OperationError

	StartedAt  time.Time
	FinishedAt time.Time
}
