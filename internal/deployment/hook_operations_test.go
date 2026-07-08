package deployment

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestPreDeployHookOperationNotApplicableWhenSkipped(t *testing.T) {
	engine := &Engine{host: runtimehosttest.NewFakeHost()}
	app := &models.Application{}
	plan := planWithStep(DeploymentStep{Operation: OperationPreDeployHook, Status: StepStatusSkipped})

	op := engine.preDeployHookOperation(app, plan)
	if op.Applicable {
		t.Fatal("expected pre-deploy hook to not be applicable")
	}
}

func TestPreDeployHookOperationApplicableAndRuns(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("sh", []string{"-c", "echo hi"}, runtimehost.CommandResult{ExitCode: 0}, nil)
	engine := &Engine{host: host}
	app := &models.Application{Config: models.DeploymentConfig{PreDeployHook: "echo hi"}}
	plan := planWithStep(DeploymentStep{Operation: OperationPreDeployHook, Status: StepStatusReady, Required: true})

	op := engine.preDeployHookOperation(app, plan)
	if !op.Applicable {
		t.Fatal("expected pre-deploy hook to be applicable")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestPreDeployHookOperationRunPropagatesError(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("sh", []string{"-c", "exit 1"},
			runtimehost.CommandResult{ExitCode: 1, Stderr: "boom"},
			&runtimehost.ExecutionError{Command: "sh", ExitCode: 1, Err: errors.New("exit status 1")},
		)
	engine := &Engine{host: host}
	app := &models.Application{Config: models.DeploymentConfig{PreDeployHook: "exit 1"}}
	plan := planWithStep(DeploymentStep{Operation: OperationPreDeployHook, Status: StepStatusReady, Required: true})

	op := engine.preDeployHookOperation(app, plan)
	if err := op.Run(context.Background()); err == nil {
		t.Fatal("expected Run to propagate the hook error")
	}
}

func TestPostDeployHookOperationNotApplicableWhenSkipped(t *testing.T) {
	engine := &Engine{host: runtimehosttest.NewFakeHost()}
	app := &models.Application{}
	plan := planWithStep(DeploymentStep{Operation: OperationPostDeployHook, Status: StepStatusSkipped})

	op := engine.postDeployHookOperation(app, plan)
	if op.Applicable {
		t.Fatal("expected post-deploy hook to not be applicable")
	}
}

func TestPostDeployHookOperationApplicableAndRuns(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("sh", []string{"-c", "echo bye"}, runtimehost.CommandResult{ExitCode: 0}, nil)
	engine := &Engine{host: host}
	app := &models.Application{Config: models.DeploymentConfig{PostDeployHook: "echo bye"}}
	plan := planWithStep(DeploymentStep{Operation: OperationPostDeployHook, Status: StepStatusReady, Required: true})

	op := engine.postDeployHookOperation(app, plan)
	if !op.Applicable {
		t.Fatal("expected post-deploy hook to be applicable")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestPostDeployHookOperationHasNoCompensation(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("sh", []string{"-c", "echo bye"}, runtimehost.CommandResult{ExitCode: 0}, nil)
	engine := &Engine{host: host}
	app := &models.Application{Config: models.DeploymentConfig{PostDeployHook: "echo bye"}}
	plan := planWithStep(DeploymentStep{Operation: OperationPostDeployHook, Status: StepStatusReady, Required: true})

	op := engine.postDeployHookOperation(app, plan)
	if op.Compensate != nil {
		t.Fatal("expected a hook operation to have no Compensate: side effects can't be safely assumed reversible")
	}
}
