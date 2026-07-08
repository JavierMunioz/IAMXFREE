package nginx

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ListVirtualHosts lists every site with a config file in
// SitesAvailableDir, reporting whether each is currently enabled (has a
// matching symlink in SitesEnabledDir). Results are sorted by ServerName.
func (m *Manager) ListVirtualHosts(ctx context.Context) ([]SiteSummary, error) {
	server, err := m.Detect(ctx)
	if err != nil {
		return nil, err
	}
	if !server.Available || server.SitesAvailableDir == "" {
		return nil, ErrNotAvailable
	}

	entries, err := m.host.ReadDir(server.SitesAvailableDir)
	if err != nil {
		return nil, fmt.Errorf("nginx: list sites-available: %w", err)
	}

	sites := make([]SiteSummary, 0, len(entries))
	for _, entry := range entries {
		if !strings.HasSuffix(entry, ".conf") {
			continue
		}

		summary := SiteSummary{
			ServerName: strings.TrimSuffix(entry, ".conf"),
			FilePath:   filepath.Join(server.SitesAvailableDir, entry),
		}

		if server.SitesEnabledDir != "" {
			enabled, err := m.host.FileExists(filepath.Join(server.SitesEnabledDir, entry))
			if err != nil {
				return nil, fmt.Errorf("nginx: check enabled site %q: %w", entry, err)
			}
			summary.Enabled = enabled
		}

		sites = append(sites, summary)
	}

	sort.Slice(sites, func(i, j int) bool { return sites[i].ServerName < sites[j].ServerName })

	return sites, nil
}
