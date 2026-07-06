package jsonstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories"
)

// SessionRepository stores each application's tracked session as its own
// "<appID>.json" file inside a directory — same one-file-per-entity shape
// as ApplicationRepository, for the same reasons (an isolated write per
// application, files inspectable by hand).
type SessionRepository struct {
	dir string
	mu  sync.RWMutex
}

var _ repositories.SessionRepository = (*SessionRepository)(nil)

// NewSessionRepository returns a repository that stores its files under
// dir, creating dir if it does not exist yet.
func NewSessionRepository(dir string) (*SessionRepository, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("jsonstore: create sessions directory: %w", err)
	}
	return &SessionRepository{dir: dir}, nil
}

func (r *SessionRepository) path(appID string) string {
	return filepath.Join(r.dir, appID+".json")
}

func (r *SessionRepository) Save(_ context.Context, appID string, session execution.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return writeJSON(r.path(appID), session)
}

func (r *SessionRepository) Delete(_ context.Context, appID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := os.Remove(r.path(appID)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("jsonstore: remove session file: %w", err)
	}
	return nil
}

func (r *SessionRepository) List(_ context.Context) (map[string]execution.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, fmt.Errorf("jsonstore: read sessions directory: %w", err)
	}

	sessions := make(map[string]execution.Session, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(r.dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("jsonstore: read session file: %w", err)
		}

		var session execution.Session
		if err := json.Unmarshal(data, &session); err != nil {
			return nil, fmt.Errorf("jsonstore: decode session file %s: %w", entry.Name(), err)
		}

		appID := strings.TrimSuffix(entry.Name(), ".json")
		sessions[appID] = session
	}
	return sessions, nil
}
