package operations

import (
	"context"
	"time"
)

// Executor runs a sequence of Operations, one at a time, in order.
type Executor struct{}

// NewExecutor returns an Executor. It holds no state of its own — every
// Execute call is independent.
func NewExecutor() *Executor {
	return &Executor{}
}

// Execute runs ops sequentially. If onProgress is non-nil, it is called
// synchronously after every state transition — once with StateRunning
// right before an applicable Operation's Run is called, then once more
// with its terminal state — so a caller can render a live view without
// polling. onProgress may be nil.
//
// The first Failed operation stops the run: every Operation after it is
// reported Skipped without ever calling Run. If ctx is already cancelled
// when Execute reaches an Operation, that Operation and everything after
// it is reported Cancelled instead.
//
// After a failure, Execute walks every Success operation that ran before
// it in reverse order and, for each one with a Compensate, runs it —
// recording the outcome on that operation's Compensation rather than
// touching its original State. An operation with no Compensate is simply
// left alone; not every operation is reversible.
func (e *Executor) Execute(ctx context.Context, ops []Operation, onProgress func(OperationProgress)) ExecutionSummary {
	started := time.Now().UTC()
	results := make([]OperationResult, len(ops))
	halted := false
	failedAt := -1

	for i, op := range ops {
		var result OperationResult

		switch {
		case ctx.Err() != nil:
			result = newResult(op, StateCancelled, ctx.Err().Error(), nil)
			halted = true

		case halted:
			result = newResult(op, StateSkipped, "skipped because a previous operation failed", nil)

		case !op.Applicable:
			result = newResult(op, StateSkipped, op.SkipReason, nil)

		default:
			if onProgress != nil {
				onProgress(OperationProgress{Index: i, Total: len(ops), Result: newResult(op, StateRunning, "", nil)})
			}
			result = e.run(ctx, op)
			if result.State == StateFailed {
				halted = true
				failedAt = i
			}
		}

		results[i] = result
		if onProgress != nil {
			onProgress(OperationProgress{Index: i, Total: len(ops), Result: result})
		}
	}

	if failedAt >= 0 {
		e.compensate(ctx, ops, results, failedAt, onProgress)
	}

	return Summarize(started, results)
}

// compensate walks results[:failedAt] in reverse order — undoing the most
// recent success first — running Compensate for every Success operation
// that has one. It never stops early on a CompensationFailed: recovering
// as much as possible matters more than an all-or-nothing rollback, and
// every failure is still recorded, never hidden.
func (e *Executor) compensate(ctx context.Context, ops []Operation, results []OperationResult, failedAt int, onProgress func(OperationProgress)) {
	for i := failedAt - 1; i >= 0; i-- {
		if results[i].State != StateSuccess || ops[i].Compensate == nil {
			continue
		}
		op := ops[i]

		if onProgress != nil {
			compensating := results[i]
			compensating.Compensation = &CompensationResult{State: StateCompensating}
			onProgress(OperationProgress{Index: i, Total: len(ops), Result: compensating})
		}

		start := time.Now().UTC()
		err := op.Compensate(ctx)
		finish := time.Now().UTC()

		comp := &CompensationResult{StartedAt: start, FinishedAt: finish}
		if err != nil {
			comp.State = StateCompensationFailed
			comp.Message = err.Error()
			comp.Err = &OperationError{Message: op.Name + " compensation failed", Cause: err}
		} else {
			comp.State = StateCompensated
			comp.Message = "compensated successfully"
		}

		results[i].Compensation = comp
		if onProgress != nil {
			onProgress(OperationProgress{Index: i, Total: len(ops), Result: results[i]})
		}
	}
}

// run calls op.Run and turns its outcome into a terminal OperationResult.
func (e *Executor) run(ctx context.Context, op Operation) OperationResult {
	start := time.Now().UTC()
	err := op.Run(ctx)
	finish := time.Now().UTC()

	if err != nil {
		return OperationResult{
			Name: op.Name, Component: op.Component, Method: op.Method,
			State:      StateFailed,
			Message:    err.Error(),
			Err:        &OperationError{Message: op.Name + " failed", Cause: err},
			StartedAt:  start,
			FinishedAt: finish,
		}
	}

	return OperationResult{
		Name: op.Name, Component: op.Component, Method: op.Method,
		State:      StateSuccess,
		Message:    "completed successfully",
		StartedAt:  start,
		FinishedAt: finish,
	}
}

func newResult(op Operation, state OperationState, message string, err *OperationError) OperationResult {
	return OperationResult{
		Name: op.Name, Component: op.Component, Method: op.Method,
		State:   state,
		Message: message,
		Err:     err,
	}
}

// Summarize reduces a run's results into an ExecutionSummary. Exported so
// a caller streaming OperationProgress (which never sees the whole slice
// at once) can compute the same summary from what it accumulated.
func Summarize(started time.Time, results []OperationResult) ExecutionSummary {
	summary := ExecutionSummary{StartedAt: started, FinishedAt: time.Now().UTC(), Operations: results}

	for _, r := range results {
		switch r.State {
		case StateSuccess:
			summary.Succeeded++
		case StateFailed:
			summary.Failed++
		case StateSkipped:
			summary.Skipped++
		case StateCancelled:
			summary.Cancelled++
		}

		if r.Compensation != nil {
			switch r.Compensation.State {
			case StateCompensated:
				summary.Compensated++
			case StateCompensationFailed:
				summary.CompensationFailed++
			}
		}
	}

	switch {
	case summary.Failed > 0:
		summary.Overall = StateFailed
	case summary.Cancelled > 0:
		summary.Overall = StateCancelled
	default:
		summary.Overall = StateSuccess
	}

	return summary
}
