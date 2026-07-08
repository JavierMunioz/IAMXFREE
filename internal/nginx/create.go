package nginx

import (
	"context"
	"fmt"
	"path/filepath"
)

// CreateVirtualHost validates vh, renders it, writes it into
// SitesAvailableDir and enables it by symlinking into SitesEnabledDir. It
// returns ErrSiteAlreadyExists if a config file for vh.ServerName already
// exists — use UpdateVirtualHost to change an existing site.
func (m *Manager) CreateVirtualHost(ctx context.Context, vh VirtualHost) error {
	if result := ValidateVirtualHost(vh); !result.Valid {
		return fmt.Errorf("nginx: invalid virtual host: %s", result.Output)
	}

	server, err := m.Detect(ctx)
	if err != nil {
		return err
	}
	if !server.Available || server.SitesAvailableDir == "" || server.SitesEnabledDir == "" {
		return ErrNotAvailable
	}

	availablePath := filepath.Join(server.SitesAvailableDir, vh.FileName())
	enabledPath := filepath.Join(server.SitesEnabledDir, vh.FileName())

	exists, err := m.host.FileExists(availablePath)
	if err != nil {
		return fmt.Errorf("nginx: check existing site: %w", err)
	}
	if exists {
		return ErrSiteAlreadyExists
	}

	content, err := Render(vh)
	if err != nil {
		return fmt.Errorf("nginx: render virtual host: %w", err)
	}

	if err := m.host.WriteFile(availablePath, []byte(content)); err != nil {
		return fmt.Errorf("nginx: write site config: %w", err)
	}

	if err := m.host.Symlink(availablePath, enabledPath); err != nil {
		return fmt.Errorf("nginx: enable site: %w", err)
	}

	return nil
}
