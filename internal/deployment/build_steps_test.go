package deployment

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestBuildStepsNoCommandsConfigured(t *testing.T) {
	app := &models.Application{}
	steps := buildSteps(app)

	if len(steps) != 2 {
		t.Fatalf("steps = %+v, want 2 (install + build)", steps)
	}
	for _, step := range steps {
		if step.Status != StepStatusSkipped || step.Required {
			t.Errorf("step %q = %+v, want skipped and not required", step.Name, step)
		}
	}
}

func TestBuildStepsCommandsConfigured(t *testing.T) {
	app := &models.Application{Config: models.DeploymentConfig{
		InstallCommand: "npm install",
		BuildCommand:   "npm run build",
	}}
	steps := buildSteps(app)

	if steps[0].Status != StepStatusReady || !steps[0].Required {
		t.Errorf("install step = %+v, want ready and required", steps[0])
	}
	if steps[1].Status != StepStatusReady || !steps[1].Required {
		t.Errorf("build step = %+v, want ready and required", steps[1])
	}
}
