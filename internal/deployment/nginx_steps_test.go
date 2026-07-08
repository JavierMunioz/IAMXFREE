package deployment

import (
	"context"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func fakeNginxHost() *runtimehosttest.FakeHost {
	return runtimehosttest.NewFakeHost().
		WithVersion("nginx", runtimehost.ToolInfo{Name: "nginx", Available: true, Version: "nginx/1.24.0"}).
		WithDir("/etc/nginx").
		WithFile("/etc/nginx/nginx.conf").
		WithDir("/etc/nginx/sites-available").
		WithDir("/etc/nginx/sites-enabled")
}

func TestNginxStepsNoDomainConfigured(t *testing.T) {
	engine := &Engine{nginxManager: nginx.NewManager(runtimehosttest.NewFakeHost())}
	app := &models.Application{}

	steps := engine.nginxSteps(context.Background(), app)
	if len(steps) != 1 || steps[0].Status != StepStatusSkipped || steps[0].Required {
		t.Fatalf("steps = %+v, want a single skipped, non-required step", steps)
	}
}

func TestNginxStepsSiteMissing(t *testing.T) {
	host := fakeNginxHost().WithReadDir("/etc/nginx/sites-available", nil, nil)
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	app := &models.Application{Config: models.DeploymentConfig{Domain: "example.com"}}

	steps := engine.nginxSteps(context.Background(), app)
	if len(steps) != 2 {
		t.Fatalf("steps = %+v, want 2 (verify + reload)", steps)
	}
	if steps[0].Status != StepStatusWarning {
		t.Errorf("verify step Status = %q, want %q", steps[0].Status, StepStatusWarning)
	}
	if steps[1].Status != StepStatusReady || !steps[1].Required {
		t.Errorf("reload step = %+v, want ready and required", steps[1])
	}
}

func TestNginxStepsSiteEnabled(t *testing.T) {
	host := fakeNginxHost().
		WithReadDir("/etc/nginx/sites-available", []string{"example.com.conf"}, nil).
		WithFile("/etc/nginx/sites-enabled/example.com.conf")
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	app := &models.Application{Config: models.DeploymentConfig{Domain: "example.com"}}

	steps := engine.nginxSteps(context.Background(), app)
	if steps[0].Status != StepStatusReady {
		t.Errorf("verify step Status = %q, want %q", steps[0].Status, StepStatusReady)
	}
	if steps[1].Status != StepStatusSkipped || steps[1].Required {
		t.Errorf("reload step = %+v, want skipped and not required", steps[1])
	}
}

func TestNginxStepsSiteDisabled(t *testing.T) {
	host := fakeNginxHost().
		WithReadDir("/etc/nginx/sites-available", []string{"example.com.conf"}, nil)
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	app := &models.Application{Config: models.DeploymentConfig{Domain: "example.com"}}

	steps := engine.nginxSteps(context.Background(), app)
	if steps[0].Status != StepStatusWarning {
		t.Errorf("verify step Status = %q, want %q", steps[0].Status, StepStatusWarning)
	}
	if steps[1].Status != StepStatusReady || !steps[1].Required {
		t.Errorf("reload step = %+v, want ready and required", steps[1])
	}
}
