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

	// Overall is Success only when every Operation ended Success or
	// Skipped; Failed if any Operation Failed; Cancelled if the run was
	// aborted before every Operation reached a terminal state.
	Overall OperationState
}
