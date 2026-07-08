package nginx

import (
	"context"
	"fmt"
	"path/filepath"
)

// DeleteVirtualHost removes the site identified by serverName: its
// sites-enabled symlink (if present) and its sites-available config file.
// It returns ErrSiteNotFound if no config file exists for serverName.
func (m *Manager) DeleteVirtualHost(ctx context.Context, serverName string) error {
	server, err := m.Detect(ctx)
	if err != nil {
		return err
	}
	if !server.Available || server.SitesAvailableDir == "" || server.SitesEnabledDir == "" {
		return ErrNotAvailable
	}

	fileName := siteFileName(serverName)
	availablePath := filepath.Join(server.SitesAvailableDir, fileName)
	enabledPath := filepath.Join(server.SitesEnabledDir, fileName)

	exists, err := m.host.FileExists(availablePath)
	if err != nil {
		return fmt.Errorf("nginx: check existing site: %w", err)
	}
	if !exists {
		return ErrSiteNotFound
	}

	enabled, err := m.host.FileExists(enabledPath)
	if err != nil {
		return fmt.Errorf("nginx: check enabled site: %w", err)
	}
	if enabled {
		if err := m.host.RemoveFile(enabledPath); err != nil {
			return fmt.Errorf("nginx: disable site: %w", err)
		}
	}

	if err := m.host.RemoveFile(availablePath); err != nil {
		return fmt.Errorf("nginx: remove site config: %w", err)
	}

	return nil
}
