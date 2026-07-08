package nginx_test

import (
	"context"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestDetectNotInstalled(t *testing.T) {
	host := runtimehosttest.NewFakeHost()
	manager := nginx.NewManager(host)

	server, err := manager.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if server.Available {
		t.Fatal("expected Available = false when nginx is not on PATH")
	}
}

func TestDetectInstalled(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithVersion("nginx", runtimehost.ToolInfo{
			Name:      "nginx",
			Path:      "/usr/sbin/nginx",
			Available: true,
			Version:   "nginx version: nginx/1.24.0 (Ubuntu)",
		})
	manager := nginx.NewManager(host)

	server, err := manager.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if !server.Available {
		t.Fatal("expected Available = true")
	}
	if server.BinaryPath != "/usr/sbin/nginx" {
		t.Errorf("BinaryPath = %q, want %q", server.BinaryPath, "/usr/sbin/nginx")
	}
	if server.Version != "1.24.0" {
		t.Errorf("Version = %q, want %q", server.Version, "1.24.0")
	}
}

func TestDetectVersionUnparseable(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithVersion("nginx", runtimehost.ToolInfo{
			Name:      "nginx",
			Available: true,
			Version:   "garbage output",
		})
	manager := nginx.NewManager(host)

	server, err := manager.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if !server.Available {
		t.Fatal("expected Available = true even when version could not be parsed")
	}
	if server.Version != "" {
		t.Errorf("Version = %q, want empty", server.Version)
	}
	if len(server.Notes) == 0 {
		t.Error("expected a note explaining the version could not be parsed")
	}
}

func TestDetectConfigPaths(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithVersion("nginx", runtimehost.ToolInfo{Name: "nginx", Available: true, Version: "nginx/1.24.0"}).
		WithDir("/etc/nginx").
		WithFile("/etc/nginx/nginx.conf").
		WithDir("/etc/nginx/sites-available").
		WithDir("/etc/nginx/sites-enabled")
	manager := nginx.NewManager(host)

	server, err := manager.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if server.ConfigRoot != "/etc/nginx" {
		t.Errorf("ConfigRoot = %q, want %q", server.ConfigRoot, "/etc/nginx")
	}
	if server.MainConfigFile != "/etc/nginx/nginx.conf" {
		t.Errorf("MainConfigFile = %q, want %q", server.MainConfigFile, "/etc/nginx/nginx.conf")
	}
	if server.SitesAvailableDir != "/etc/nginx/sites-available" {
		t.Errorf("SitesAvailableDir = %q, want %q", server.SitesAvailableDir, "/etc/nginx/sites-available")
	}
	if server.SitesEnabledDir != "/etc/nginx/sites-enabled" {
		t.Errorf("SitesEnabledDir = %q, want %q", server.SitesEnabledDir, "/etc/nginx/sites-enabled")
	}
	if len(server.Notes) != 0 {
		t.Errorf("Notes = %v, want empty", server.Notes)
	}
}

func TestDetectConfigPathsMissing(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithVersion("nginx", runtimehost.ToolInfo{Name: "nginx", Available: true, Version: "nginx/1.24.0"})
	manager := nginx.NewManager(host)

	server, err := manager.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if server.ConfigRoot != "" {
		t.Errorf("ConfigRoot = %q, want empty", server.ConfigRoot)
	}
	if len(server.Notes) == 0 {
		t.Error("expected a note explaining the config root was not found")
	}
}

func TestDetectVersionCommandFailed(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithVersion("nginx", runtimehost.ToolInfo{
			Name:       "nginx",
			Available:  true,
			VersionErr: context.DeadlineExceeded,
		})
	manager := nginx.NewManager(host)

	server, err := manager.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if !server.Available {
		t.Fatal("expected Available = true even when the version command failed")
	}
	if len(server.Notes) == 0 {
		t.Error("expected a note explaining the version command failed")
	}
}
