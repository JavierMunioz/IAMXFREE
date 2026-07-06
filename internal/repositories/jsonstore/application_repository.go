package jsonstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories"
)

// ApplicationRepository stores each Application as its own "<id>.json" file
// inside a directory. One file per application keeps a write to one
// application from risking the rest of the data set, and keeps the files
// inspectable/editable by hand during development.
type ApplicationRepository struct {
	dir string
	mu  sync.RWMutex
}

var _ repositories.ApplicationRepository = (*ApplicationRepository)(nil)

// NewApplicationRepository returns a repository that stores its files under
// dir, creating dir if it does not exist yet.
func NewApplicationRepository(dir string) (*ApplicationRepository, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("jsonstore: create applications directory: %w", err)
	}
	return &ApplicationRepository{dir: dir}, nil
}

func (r *ApplicationRepository) path(id string) string {
	return filepath.Join(r.dir, id+".json")
}

func (r *ApplicationRepository) Create(_ context.Context, app *models.Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := os.Stat(r.path(app.ID)); err == nil {
		return repositories.ErrApplicationAlreadyExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("jsonstore: stat application file: %w", err)
	}

	return writeJSON(r.path(app.ID), app)
}

func (r *ApplicationRepository) Update(_ context.Context, app *models.Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := os.Stat(r.path(app.ID)); errors.Is(err, os.ErrNotExist) {
		return repositories.ErrApplicationNotFound
	} else if err != nil {
		return fmt.Errorf("jsonstore: stat application file: %w", err)
	}

	return writeJSON(r.path(app.ID), app)
}

func (r *ApplicationRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := os.Remove(r.path(id)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repositories.ErrApplicationNotFound
		}
		return fmt.Errorf("jsonstore: remove application file: %w", err)
	}
	return nil
}

func (r *ApplicationRepository) FindByID(_ context.Context, id string) (*models.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return readJSON(r.path(id))
}

func (r *ApplicationRepository) FindByName(ctx context.Context, name string) (*models.Application, error) {
	apps, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, app := range apps {
		if app.Name == name {
			return app, nil
		}
	}
	return nil, repositories.ErrApplicationNotFound
}

func (r *ApplicationRepository) List(_ context.Context) ([]*models.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, fmt.Errorf("jsonstore: read applications directory: %w", err)
	}

	apps := make([]*models.Application, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		app, err := readJSON(filepath.Join(r.dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func readJSON(path string) (*models.Application, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, repositories.ErrApplicationNotFound
		}
		return nil, fmt.Errorf("jsonstore: read application file: %w", err)
	}

	var app models.Application
	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("jsonstore: decode application file %s: %w", path, err)
	}
	return &app, nil
}

// writeJSON writes v as indented JSON, via a temp file plus rename, so a
// process interrupted mid-write can never leave a corrupt application file
// behind.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("jsonstore: encode application: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("jsonstore: write temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("jsonstore: rename temp file: %w", err)
	}
	return nil
}
