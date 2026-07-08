package deployment

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestExecutionStepsNotRunning(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{running: false}}
	app := &models.Application{ID: "app-1"}

	steps := engine.executionSteps(app)
	if len(steps) != 2 {
		t.Fatalf("steps = %+v, want 2 (check + restart)", steps)
	}
	if len(steps[0].Warnings) != 0 {
		t.Errorf("expected no warnings when not running, got %v", steps[0].Warnings)
	}
	if len(steps[1].Risks) != 0 {
		t.Errorf("expected no restart risk when not running, got %v", steps[1].Risks)
	}
}

func TestExecutionStepsRunning(t *testing.T) {
	engine := &Engine{executionService: &fakeExecutionService{
		running: true,
		session: services.RunSession{PID: 4242},
	}}
	app := &models.Application{ID: "app-1"}

	steps := engine.executionSteps(app)
	if len(steps[0].Warnings) == 0 {
		t.Error("expected a warning that the application is currently running")
	}
	if len(steps[1].Risks) == 0 {
		t.Error("expected a restart risk about interrupting active connections")
	}
}
