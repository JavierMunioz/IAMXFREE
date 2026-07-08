package nginx

import (
	"context"
	"fmt"
	"path/filepath"
)

// UpdateVirtualHost validates vh, re-renders it and overwrites the config
// file already on disk for vh.ServerName. It returns ErrSiteNotFound if no
// such site exists — use CreateVirtualHost first. If the site was
// previously disabled (no symlink in sites-enabled), Update re-enables it;
// an already-enabled site is left as is.
func (m *Manager) UpdateVirtualHost(ctx context.Context, vh VirtualHost) error {
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
	if !exists {
		return ErrSiteNotFound
	}

	content, err := Render(vh)
	if err != nil {
		return fmt.Errorf("nginx: render virtual host: %w", err)
	}

	if err := m.host.WriteFile(availablePath, []byte(content)); err != nil {
		return fmt.Errorf("nginx: write site config: %w", err)
	}

	enabled, err := m.host.FileExists(enabledPath)
	if err != nil {
		return fmt.Errorf("nginx: check enabled site: %w", err)
	}
	if !enabled {
		if err := m.host.Symlink(availablePath, enabledPath); err != nil {
			return fmt.Errorf("nginx: enable site: %w", err)
		}
	}

	return nil
}
