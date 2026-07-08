package operations_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

func applicableOp(name string, run func(ctx context.Context) error) operations.Operation {
	return operations.Operation{Name: name, Component: "test", Method: name, Applicable: true, Run: run}
}

func compensableOp(name string, run, compensate func(ctx context.Context) error) operations.Operation {
	return operations.Operation{Name: name, Component: "test", Method: name, Applicable: true, Run: run, Compensate: compensate}
}

func succeed(context.Context) error { return nil }

func TestExecuteAllSucceed(t *testing.T) {
	ops := []operations.Operation{
		applicableOp("one", succeed),
		applicableOp("two", succeed),
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if summary.Overall != operations.StateSuccess {
		t.Fatalf("Overall = %q, want %q", summary.Overall, operations.StateSuccess)
	}
	if summary.Succeeded != 2 || summary.Failed != 0 || summary.Skipped != 0 {
		t.Fatalf("summary = %+v, want 2 succeeded, 0 failed, 0 skipped", summary)
	}
}

func TestExecuteStopsOnFailureAndSkipsRest(t *testing.T) {
	wantErr := errors.New("boom")
	var thirdRan bool

	ops := []operations.Operation{
		applicableOp("one", succeed),
		applicableOp("two", func(context.Context) error { return wantErr }),
		applicableOp("three", func(context.Context) error { thirdRan = true; return nil }),
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if summary.Overall != operations.StateFailed {
		t.Fatalf("Overall = %q, want %q", summary.Overall, operations.StateFailed)
	}
	if summary.Succeeded != 1 || summary.Failed != 1 || summary.Skipped != 1 {
		t.Fatalf("summary = %+v, want 1/1/1", summary)
	}
	if thirdRan {
		t.Fatal("expected the third operation to never run")
	}
	if summary.Operations[2].State != operations.StateSkipped {
		t.Fatalf("Operations[2].State = %q, want %q", summary.Operations[2].State, operations.StateSkipped)
	}
	if summary.Operations[1].Err == nil || summary.Operations[1].Err.Cause != wantErr {
		t.Fatalf("Operations[1].Err = %+v, want it to wrap %v", summary.Operations[1].Err, wantErr)
	}
}

func TestExecuteNotApplicableIsSkippedWithoutRunning(t *testing.T) {
	var ran bool
	ops := []operations.Operation{
		{
			Name: "conditional", Component: "test", Method: "conditional",
			Applicable: false, SkipReason: "not needed",
			Run: func(context.Context) error { ran = true; return nil },
		},
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if ran {
		t.Fatal("expected Run to never be called for a non-applicable operation")
	}
	if summary.Operations[0].State != operations.StateSkipped {
		t.Fatalf("State = %q, want %q", summary.Operations[0].State, operations.StateSkipped)
	}
	if summary.Operations[0].Message != "not needed" {
		t.Fatalf("Message = %q, want %q", summary.Operations[0].Message, "not needed")
	}
}

func TestExecuteCancelledContextSkipsWithCancelledState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var ran bool
	ops := []operations.Operation{applicableOp("one", func(context.Context) error { ran = true; return nil })}

	summary := operations.NewExecutor().Execute(ctx, ops, nil)

	if ran {
		t.Fatal("expected Run to never be called once ctx is cancelled")
	}
	if summary.Operations[0].State != operations.StateCancelled {
		t.Fatalf("State = %q, want %q", summary.Operations[0].State, operations.StateCancelled)
	}
	if summary.Overall != operations.StateCancelled {
		t.Fatalf("Overall = %q, want %q", summary.Overall, operations.StateCancelled)
	}
}

func TestExecuteEmitsRunningThenTerminalProgress(t *testing.T) {
	ops := []operations.Operation{applicableOp("one", succeed)}

	var states []operations.OperationState
	operations.NewExecutor().Execute(context.Background(), ops, func(p operations.OperationProgress) {
		states = append(states, p.Result.State)
	})

	if len(states) != 2 || states[0] != operations.StateRunning || states[1] != operations.StateSuccess {
		t.Fatalf("states = %v, want [running success]", states)
	}
}

func TestExecuteSkippedEmitsOnlyOneProgressUpdate(t *testing.T) {
	ops := []operations.Operation{{Name: "skip", Applicable: false, SkipReason: "n/a"}}

	var updates int
	operations.NewExecutor().Execute(context.Background(), ops, func(operations.OperationProgress) { updates++ })

	if updates != 1 {
		t.Fatalf("updates = %d, want 1 (no Running phase for a skipped operation)", updates)
	}
}

func TestExecuteCompensatesSuccessfulOperationsOnFailure(t *testing.T) {
	var compensated bool
	ops := []operations.Operation{
		compensableOp("stop", succeed, func(context.Context) error { compensated = true; return nil }),
		applicableOp("fail", func(context.Context) error { return errors.New("boom") }),
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if !compensated {
		t.Fatal("expected the compensation to run")
	}
	if summary.Operations[0].State != operations.StateSuccess {
		t.Fatalf("Operations[0].State = %q, want %q (compensation never changes the original State)", summary.Operations[0].State, operations.StateSuccess)
	}
	if summary.Operations[0].Compensation == nil || summary.Operations[0].Compensation.State != operations.StateCompensated {
		t.Fatalf("Operations[0].Compensation = %+v, want StateCompensated", summary.Operations[0].Compensation)
	}
	if summary.Compensated != 1 {
		t.Fatalf("Compensated = %d, want 1", summary.Compensated)
	}
	if summary.Overall != operations.StateFailed {
		t.Fatalf("Overall = %q, want %q (a compensated failure is still a failure)", summary.Overall, operations.StateFailed)
	}
}

func TestExecuteSkipsCompensationForOperationsWithoutOne(t *testing.T) {
	ops := []operations.Operation{
		applicableOp("build", succeed), // no Compensate — Build may not need one
		applicableOp("fail", func(context.Context) error { return errors.New("boom") }),
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if summary.Operations[0].Compensation != nil {
		t.Fatalf("Operations[0].Compensation = %+v, want nil (no Compensate defined)", summary.Operations[0].Compensation)
	}
	if summary.Compensated != 0 {
		t.Fatalf("Compensated = %d, want 0", summary.Compensated)
	}
}

func TestExecuteCompensatesInReverseOrder(t *testing.T) {
	var order []string
	ops := []operations.Operation{
		compensableOp("first", succeed, func(context.Context) error { order = append(order, "first"); return nil }),
		compensableOp("second", succeed, func(context.Context) error { order = append(order, "second"); return nil }),
		applicableOp("fail", func(context.Context) error { return errors.New("boom") }),
	}

	operations.NewExecutor().Execute(context.Background(), ops, nil)

	if len(order) != 2 || order[0] != "second" || order[1] != "first" {
		t.Fatalf("compensation order = %v, want [second first]", order)
	}
}

func TestExecuteRecordsCompensationFailureWithoutStoppingEarlierCompensations(t *testing.T) {
	var firstCompensated bool
	ops := []operations.Operation{
		compensableOp("first", succeed, func(context.Context) error { firstCompensated = true; return nil }),
		compensableOp("second", succeed, func(context.Context) error { return errors.New("cannot undo") }),
		applicableOp("fail", func(context.Context) error { return errors.New("boom") }),
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if !firstCompensated {
		t.Fatal("expected the earlier operation to still be compensated despite a later compensation failing")
	}
	if summary.Operations[1].Compensation == nil || summary.Operations[1].Compensation.State != operations.StateCompensationFailed {
		t.Fatalf("Operations[1].Compensation = %+v, want StateCompensationFailed", summary.Operations[1].Compensation)
	}
	if summary.Operations[1].Compensation.Err == nil {
		t.Fatal("expected a CompensationResult.Err explaining the failure")
	}
	if summary.CompensationFailed != 1 || summary.Compensated != 1 {
		t.Fatalf("CompensationFailed = %d, Compensated = %d, want 1 and 1", summary.CompensationFailed, summary.Compensated)
	}
}

func TestExecuteEmitsCompensatingThenTerminalProgress(t *testing.T) {
	ops := []operations.Operation{
		compensableOp("stop", succeed, succeed),
		applicableOp("fail", func(context.Context) error { return errors.New("boom") }),
	}

	var compensationStates []operations.OperationState
	operations.NewExecutor().Execute(context.Background(), ops, func(p operations.OperationProgress) {
		if p.Index == 0 && p.Result.Compensation != nil {
			compensationStates = append(compensationStates, p.Result.Compensation.State)
		}
	})

	if len(compensationStates) != 2 || compensationStates[0] != operations.StateCompensating || compensationStates[1] != operations.StateCompensated {
		t.Fatalf("compensation progress states = %v, want [compensating compensated]", compensationStates)
	}
}
