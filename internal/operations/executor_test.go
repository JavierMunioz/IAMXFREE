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
