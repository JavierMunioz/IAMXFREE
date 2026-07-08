package deployment

import (
	"context"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestInstallOperationNotApplicableWhenSkipped(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{}}
	app := &models.Application{}
	plan := planWithStep(DeploymentStep{Operation: OperationInstallDependencies, Status: StepStatusSkipped})

	op := engine.installOperation(app, plan)
	if op.Applicable {
		t.Fatal("expected Install to not be applicable")
	}
}

func TestInstallOperationApplicableAndRuns(t *testing.T) {
	var calledWith string
	engine := &Engine{executionService: &fakeExecutionService{
		installFn: func(_ context.Context, appID string) error { calledWith = appID; return nil },
	}}
	app := &models.Application{ID: "app-1"}
	plan := planWithStep(DeploymentStep{Operation: OperationInstallDependencies, Status: StepStatusReady, Required: true})

	op := engine.installOperation(app, plan)
	if !op.Applicable {
		t.Fatal("expected Install to be applicable")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if calledWith != "app-1" {
		t.Errorf("ExecutionService.Install called with %q, want %q", calledWith, "app-1")
	}
}

func TestBuildOperationNotApplicableWhenSkipped(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{}}
	app := &models.Application{}
	plan := planWithStep(DeploymentStep{Operation: OperationBuild, Status: StepStatusSkipped})

	op := engine.buildOperation(app, plan)
	if op.Applicable {
		t.Fatal("expected Build to not be applicable")
	}
}

func TestBuildOperationApplicableAndRuns(t *testing.T) {
	var ran bool
	engine := &Engine{executionService: &fakeExecutionService{
		buildFn: func(context.Context, string) error { ran = true; return nil },
	}}
	app := &models.Application{ID: "app-1"}
	plan := planWithStep(DeploymentStep{Operation: OperationBuild, Status: StepStatusReady, Required: true})

	op := engine.buildOperation(app, plan)
	if !op.Applicable {
		t.Fatal("expected Build to be applicable")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !ran {
		t.Error("expected ExecutionService.Build to be called")
	}
}

func TestBuildOperationRunPropagatesError(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{buildErr: errBoom}}
	app := &models.Application{ID: "app-1"}
	plan := planWithStep(DeploymentStep{Operation: OperationBuild, Status: StepStatusReady, Required: true})

	op := engine.buildOperation(app, plan)
	if err := op.Run(context.Background()); err == nil {
		t.Fatal("expected Run to propagate the build error")
	}
}
