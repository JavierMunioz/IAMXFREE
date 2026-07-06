package repositories

import (
	"context"
	"errors"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

var (
	// ErrApplicationNotFound is returned when no application matches the
	// requested ID or name.
	ErrApplicationNotFound = errors.New("application not found")
	// ErrApplicationAlreadyExists is returned by Create when an application
	// with the same ID is already persisted.
	ErrApplicationAlreadyExists = errors.New("application already exists")
)

// ApplicationRepository persists and retrieves Application entities. It is
// the only contract the rest of the system depends on; concrete storage
// backends (JSON files today, SQLite/Postgres/etc. later) implement it in
// their own package so swapping one for another never touches callers.
type ApplicationRepository interface {
	Create(ctx context.Context, app *models.Application) error
	Update(ctx context.Context, app *models.Application) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.Application, error)
	FindByName(ctx context.Context, name string) (*models.Application, error)
	List(ctx context.Context) ([]*models.Application, error)
}
