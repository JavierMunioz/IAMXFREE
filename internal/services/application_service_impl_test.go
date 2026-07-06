package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories/jsonstore"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func newService(t *testing.T) services.ApplicationService {
	t.Helper()
	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
	return services.NewApplicationService(repo)
}

func TestApplicationServiceRegister(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)

	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	got, err := svc.Get(ctx, app.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name != "my-api" {
		t.Errorf("Name = %q, want %q", got.Name, "my-api")
	}
}

func TestApplicationServiceRegisterRejectsInvalidApplication(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("", models.ApplicationTypeAPI)

	if err := svc.Register(ctx, app); !errors.Is(err, models.ErrApplicationNameRequired) {
		t.Fatalf("Register() error = %v, want %v", err, models.ErrApplicationNameRequired)
	}
}

func TestApplicationServiceRegisterRejectsDuplicateName(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	first := models.NewApplication("my-api", models.ApplicationTypeAPI)
	if err := svc.Register(ctx, first); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	second := models.NewApplication("my-api", models.ApplicationTypeBackend)
	if err := svc.Register(ctx, second); !errors.Is(err, services.ErrApplicationNameTaken) {
		t.Fatalf("Register() error = %v, want %v", err, services.ErrApplicationNameTaken)
	}
}

func TestApplicationServiceList(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	for _, name := range []string{"app-one", "app-two"} {
		app := models.NewApplication(name, models.ApplicationTypeAPI)
		if err := svc.Register(ctx, app); err != nil {
			t.Fatalf("Register() error = %v", err)
		}
	}

	apps, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("len(apps) = %d, want 2", len(apps))
	}
}

func TestApplicationServiceUpdateConfig(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	updated, err := svc.UpdateConfig(ctx, app.ID, models.DeploymentConfig{InternalPort: 4000})
	if err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}
	if updated.Config.InternalPort != 4000 {
		t.Errorf("Config.InternalPort = %d, want 4000", updated.Config.InternalPort)
	}
}

func TestApplicationServiceChangeStatus(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	updated, err := svc.ChangeStatus(ctx, app.ID, models.StatusRunning)
	if err != nil {
		t.Fatalf("ChangeStatus() error = %v", err)
	}
	if updated.Status != models.StatusRunning {
		t.Errorf("Status = %q, want %q", updated.Status, models.StatusRunning)
	}
}

func TestApplicationServiceRemove(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	if err := svc.Register(ctx, app); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := svc.Remove(ctx, app.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := svc.Get(ctx, app.ID); err == nil {
		t.Fatal("expected Get() to fail after Remove()")
	}
}
