package jsonstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories/jsonstore"
)

func newSessionRepo(t *testing.T) *jsonstore.SessionRepository {
	t.Helper()
	repo, err := jsonstore.NewSessionRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewSessionRepository() error = %v", err)
	}
	return repo
}

func TestSessionRepositorySaveAndList(t *testing.T) {
	ctx := context.Background()
	repo := newSessionRepo(t)

	session := execution.Session{
		PID:        4242,
		StartedAt:  time.Now().UTC().Truncate(time.Second),
		Command:    "npm",
		Args:       []string{"start"},
		WorkingDir: "/srv/apps/my-api",
		Status:     execution.StatusRunning,
		Runtime:    models.RuntimeNode,
	}

	if err := repo.Save(ctx, "app-1", session); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	sessions, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	got, ok := sessions["app-1"]
	if !ok {
		t.Fatal("expected sessions to contain app-1")
	}
	if got.PID != 4242 || got.Command != "npm" || got.Runtime != models.RuntimeNode {
		t.Errorf("List()[app-1] = %+v, unexpected", got)
	}
	if !got.StartedAt.Equal(session.StartedAt) {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, session.StartedAt)
	}
}

func TestSessionRepositorySaveOverwrites(t *testing.T) {
	ctx := context.Background()
	repo := newSessionRepo(t)

	if err := repo.Save(ctx, "app-1", execution.Session{PID: 1, Status: execution.StatusRunning}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := repo.Save(ctx, "app-1", execution.Session{PID: 2, Status: execution.StatusStopped}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	sessions, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if sessions["app-1"].PID != 2 {
		t.Fatalf("PID = %d, want 2 (overwritten)", sessions["app-1"].PID)
	}
}

func TestSessionRepositoryDeleteRemovesEntry(t *testing.T) {
	ctx := context.Background()
	repo := newSessionRepo(t)

	if err := repo.Save(ctx, "app-1", execution.Session{PID: 4242}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := repo.Delete(ctx, "app-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	sessions, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if _, ok := sessions["app-1"]; ok {
		t.Fatal("expected app-1 to be removed after Delete")
	}
}

func TestSessionRepositoryDeleteUnknownAppIDIsNotAnError(t *testing.T) {
	ctx := context.Background()
	repo := newSessionRepo(t)

	if err := repo.Delete(ctx, "never-existed"); err != nil {
		t.Fatalf("Delete() error = %v, want nil for a never-persisted appID", err)
	}
}

func TestSessionRepositoryListEmptyDirectory(t *testing.T) {
	ctx := context.Background()
	repo := newSessionRepo(t)

	sessions, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("len(sessions) = %d, want 0", len(sessions))
	}
}
