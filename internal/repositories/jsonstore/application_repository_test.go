package jsonstore_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories/jsonstore"
)

func newRepo(t *testing.T) *jsonstore.ApplicationRepository {
	t.Helper()
	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
	return repo
}

func TestApplicationRepositoryCreateAndFindByID(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)

	if err := repo.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, app.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Name != app.Name {
		t.Fatalf("FindByID() name = %q, want %q", got.Name, app.Name)
	}
}

func TestApplicationRepositoryFindByIDMissing(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	if _, err := repo.FindByID(ctx, "does-not-exist"); !errors.Is(err, repositories.ErrApplicationNotFound) {
		t.Fatalf("FindByID() error = %v, want %v", err, repositories.ErrApplicationNotFound)
	}
}

func TestApplicationRepositoryCreateDuplicate(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)

	if err := repo.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.Create(ctx, app); !errors.Is(err, repositories.ErrApplicationAlreadyExists) {
		t.Fatalf("Create() duplicate error = %v, want %v", err, repositories.ErrApplicationAlreadyExists)
	}
}

func TestApplicationRepositoryUpdate(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)

	if err := repo.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	app.Status = models.StatusRunning
	if err := repo.Update(ctx, app); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.FindByID(ctx, app.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != models.StatusRunning {
		t.Fatalf("FindByID() status = %q, want %q", got.Status, models.StatusRunning)
	}
}

func TestApplicationRepositoryUpdateMissing(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)

	if err := repo.Update(ctx, app); !errors.Is(err, repositories.ErrApplicationNotFound) {
		t.Fatalf("Update() error = %v, want %v", err, repositories.ErrApplicationNotFound)
	}
}

func TestApplicationRepositoryDelete(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)

	if err := repo.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.Delete(ctx, app.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.FindByID(ctx, app.ID); !errors.Is(err, repositories.ErrApplicationNotFound) {
		t.Fatalf("FindByID() after delete error = %v, want %v", err, repositories.ErrApplicationNotFound)
	}
}

func TestApplicationRepositoryDeleteMissing(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	if err := repo.Delete(ctx, "does-not-exist"); !errors.Is(err, repositories.ErrApplicationNotFound) {
		t.Fatalf("Delete() error = %v, want %v", err, repositories.ErrApplicationNotFound)
	}
}

func TestApplicationRepositoryFindByName(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)

	if err := repo.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByName(ctx, "my-api")
	if err != nil {
		t.Fatalf("FindByName() error = %v", err)
	}
	if got.ID != app.ID {
		t.Fatalf("FindByName() id = %q, want %q", got.ID, app.ID)
	}
}

func TestApplicationRepositoryFindByNameMissing(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	if _, err := repo.FindByName(ctx, "does-not-exist"); !errors.Is(err, repositories.ErrApplicationNotFound) {
		t.Fatalf("FindByName() error = %v, want %v", err, repositories.ErrApplicationNotFound)
	}
}

func TestApplicationRepositoryList(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	for _, name := range []string{"app-one", "app-two"} {
		app := models.NewApplication(name, models.ApplicationTypeAPI)
		if err := repo.Create(ctx, app); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	apps, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("List() len = %d, want 2", len(apps))
	}
}
