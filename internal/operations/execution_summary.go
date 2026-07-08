package operations

import "time"

// ExecutionSummary is a completed run reduced to what a caller needs to
// answer "did the deployment work" without walking every OperationResult
// itself.
type ExecutionSummary struct {
	StartedAt  time.Time
	FinishedAt time.Time

	Operations []OperationResult

	Succeeded int
	Failed    int
	Skipped   int
	Cancelled int

	// Compensated and CompensationFailed count operations whose
	// Compensation reached that state — always a subset of Succeeded,
	// since only a successful operation is ever compensated.
	Compensated        int
	CompensationFailed int

	// Overall is Success only when every Operation ended Success or
	// Skipped; Failed if any Operation Failed; Cancelled if the run was
	// aborted before every Operation reached a terminal state.
	// Compensation outcome never changes Overall — a failed deployment
	// that was cleanly compensated is still a failed deployment.
	Overall OperationState
}
