package operations

// OperationState is the lifecycle state of one Operation. An Operation
// always reports one of these — never a bare boolean — so a caller can
// always explain what happened, not just whether it "worked".
type OperationState string

const (
	// StatePending is an Operation's state before the Executor reaches
	// it.
	StatePending OperationState = "pending"

	// StateRunning means the Executor is currently calling this
	// Operation's Run.
	StateRunning OperationState = "running"

	// StateSuccess means Run completed without error.
	StateSuccess OperationState = "success"

	// StateFailed means Run returned an error. The Executor stops the
	// run immediately after a Failed operation.
	StateFailed OperationState = "failed"

	// StateSkipped means this Operation was never run: either it did not
	// apply (Operation.Applicable == false), or it comes after a Failed
	// operation in the same run.
	StateSkipped OperationState = "skipped"

	// StateCancelled means the run's context was cancelled before this
	// Operation could start.
	StateCancelled OperationState = "cancelled"

	// StateCompensating means the Executor is currently calling this
	// Operation's Compensate, after a later operation failed.
	StateCompensating OperationState = "compensating"

	// StateCompensated means a Success operation's Compensate ran without
	// error after a later operation failed. It never reuses Success —
	// the operation did succeed, and was then deliberately undone.
	StateCompensated OperationState = "compensated"

	// StateCompensationFailed means a Success operation's Compensate
	// itself returned an error while recovering from a later failure. It
	// never reuses Failed — that state is reserved for the operation
	// that actually broke the run.
	StateCompensationFailed OperationState = "compensation_failed"
)
