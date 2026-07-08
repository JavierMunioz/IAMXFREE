package deployment

import (
	"context"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestStopOperationNotApplicableWhenNotRunning(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{running: false}}
	app := &models.Application{ID: "app-1"}

	op := engine.stopOperation(app)
	if op.Applicable {
		t.Fatal("expected Stop to not be applicable when nothing is running")
	}
}

func TestStopOperationApplicableAndRuns(t *testing.T) {
	exec := &fakeExecutionService{running: true, session: services.RunSession{PID: 4242}}
	engine := &Engine{executionService: exec}
	app := &models.Application{ID: "app-1"}

	op := engine.stopOperation(app)
	if !op.Applicable {
		t.Fatal("expected Stop to be applicable")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if exec.stoppedWith.PID != 4242 {
		t.Errorf("Stop called with session PID = %d, want 4242", exec.stoppedWith.PID)
	}
}

func TestStopOperationRunPropagatesError(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{running: true, stopErr: errBoom}}
	app := &models.Application{ID: "app-1"}

	op := engine.stopOperation(app)
	if err := op.Run(context.Background()); err == nil {
		t.Fatal("expected Run to propagate the stop error")
	}
}

func TestStartOperationAlwaysApplicable(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{}}
	app := &models.Application{ID: "app-1"}

	op := engine.startOperation(app)
	if !op.Applicable {
		t.Fatal("expected Start to always be applicable")
	}
}

func TestStartOperationRuns(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{startSession: services.RunSession{PID: 99}}}
	app := &models.Application{ID: "app-1"}

	op := engine.startOperation(app)
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestStartOperationRunPropagatesError(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{startErr: errBoom}}
	app := &models.Application{ID: "app-1"}

	op := engine.startOperation(app)
	if err := op.Run(context.Background()); err == nil {
		t.Fatal("expected Run to propagate the start error")
	}
}
