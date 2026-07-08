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
}
