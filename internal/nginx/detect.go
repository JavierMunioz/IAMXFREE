package nginx

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
)

var nginxVersionPattern = regexp.MustCompile(`nginx/(\d+\.\d+\.\d+)`)

// defaultConfigRoots are the directories checked, in order, for an Nginx
// installation's configuration — "/etc/nginx" on every Debian/Ubuntu/RHEL
// distribution this manager targets.
var defaultConfigRoots = []string{"/etc/nginx"}

// Detect probes the Manager's host for an Nginx installation: whether the
// binary is on PATH and what version it reports. Available == false is a
// normal outcome (Nginx is simply not installed), never an error.
func (m *Manager) Detect(ctx context.Context) (Server, error) {
	info, err := m.host.Version(ctx, "nginx", []string{"-v"})
	if err != nil {
		return Server{}, err
	}
	if !info.Available {
		return Server{Available: false}, nil
	}

	server := Server{Available: true, BinaryPath: info.Path}

	if info.VersionErr != nil {
		server.Notes = append(server.Notes, "nginx -v failed: "+info.VersionErr.Error())
		return server, nil
	}

	if match := nginxVersionPattern.FindStringSubmatch(info.Version); match != nil {
		server.Version = match[1]
	} else {
		server.Notes = append(server.Notes, "could not parse version from: "+info.Version)
	}

	m.detectConfigPaths(&server)

	return server, nil
}

// detectConfigPaths locates server's config root and, within it, its main
// config file and sites-available/sites-enabled directories. Anything not
// found is left empty with a note explaining why, rather than guessed at.
func (m *Manager) detectConfigPaths(server *Server) {
	for _, root := range defaultConfigRoots {
		if ok, _ := m.host.DirExists(root); ok {
			server.ConfigRoot = root
			break
		}
	}
	if server.ConfigRoot == "" {
		server.Notes = append(server.Notes, "config root not found (checked: "+strings.Join(defaultConfigRoots, ", ")+")")
		return
	}

	mainConfig := filepath.Join(server.ConfigRoot, "nginx.conf")
	if ok, _ := m.host.FileExists(mainConfig); ok {
		server.MainConfigFile = mainConfig
	} else {
		server.Notes = append(server.Notes, "main config file not found at "+mainConfig)
	}

	sitesAvailable := filepath.Join(server.ConfigRoot, "sites-available")
	if ok, _ := m.host.DirExists(sitesAvailable); ok {
		server.SitesAvailableDir = sitesAvailable
	} else {
		server.Notes = append(server.Notes, "sites-available directory not found at "+sitesAvailable)
	}

	sitesEnabled := filepath.Join(server.ConfigRoot, "sites-enabled")
	if ok, _ := m.host.DirExists(sitesEnabled); ok {
		server.SitesEnabledDir = sitesEnabled
	} else {
		server.Notes = append(server.Notes, "sites-enabled directory not found at "+sitesEnabled)
	}
}
