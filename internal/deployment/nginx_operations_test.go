package deployment

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

func TestReloadNginxOperationNotApplicableWhenSkipped(t *testing.T) {
	engine := &Engine{nginxManager: nginx.NewManager(fakeNginxHost())}
	plan := planWithStep(DeploymentStep{Operation: OperationReloadNginx, Status: StepStatusSkipped, Required: false})

	op := engine.reloadNginxOperation(plan)
	if op.Applicable {
		t.Fatal("expected Reload to not be applicable")
	}
}

func TestReloadNginxOperationApplicableAndRuns(t *testing.T) {
	host := fakeNginxHost().WithRunResult("nginx", []string{"-s", "reload"}, runtimehost.CommandResult{ExitCode: 0}, nil)
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	plan := planWithStep(DeploymentStep{Operation: OperationReloadNginx, Status: StepStatusReady, Required: true})

	op := engine.reloadNginxOperation(plan)
	if !op.Applicable {
		t.Fatal("expected Reload to be applicable")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestReloadNginxOperationRunPropagatesError(t *testing.T) {
	host := fakeNginxHost().WithRunResult("nginx", []string{"-s", "reload"},
		runtimehost.CommandResult{ExitCode: 1, Stderr: "nginx: [error] invalid PID"},
		&runtimehost.ExecutionError{Command: "nginx", ExitCode: 1, Err: errors.New("exit status 1")},
	)
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	plan := planWithStep(DeploymentStep{Operation: OperationReloadNginx, Status: StepStatusReady, Required: true})

	op := engine.reloadNginxOperation(plan)
	if err := op.Run(context.Background()); err == nil {
		t.Fatal("expected Run to propagate the reload error")
	}
}
