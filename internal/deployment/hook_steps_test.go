package deployment

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestPreDeployHookStepNoCommandConfigured(t *testing.T) {
	step := preDeployHookStep(&models.Application{})
	if step.Status != StepStatusSkipped || step.Required {
		t.Errorf("step = %+v, want skipped and not required", step)
	}
}

func TestPreDeployHookStepCommandConfigured(t *testing.T) {
	app := &models.Application{Config: models.DeploymentConfig{PreDeployHook: "echo hi"}}
	step := preDeployHookStep(app)
	if step.Status != StepStatusReady || !step.Required {
		t.Errorf("step = %+v, want ready and required", step)
	}
	if step.Component != ComponentHooks {
		t.Errorf("Component = %q, want %q", step.Component, ComponentHooks)
	}
}

func TestPostDeployHookStepNoCommandConfigured(t *testing.T) {
	step := postDeployHookStep(&models.Application{})
	if step.Status != StepStatusSkipped || step.Required {
		t.Errorf("step = %+v, want skipped and not required", step)
	}
}

func TestPostDeployHookStepCommandConfigured(t *testing.T) {
	app := &models.Application{Config: models.DeploymentConfig{PostDeployHook: "echo bye"}}
	step := postDeployHookStep(app)
	if step.Status != StepStatusReady || !step.Required {
		t.Errorf("step = %+v, want ready and required", step)
	}
	if step.Component != ComponentHooks {
		t.Errorf("Component = %q, want %q", step.Component, ComponentHooks)
	}
}
