package nginx

import (
	"context"
	"regexp"
)

var nginxVersionPattern = regexp.MustCompile(`nginx/(\d+\.\d+\.\d+)`)

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

	return server, nil
}
