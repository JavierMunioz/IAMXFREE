package nginx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func availableHost() *runtimehosttest.FakeHost {
	return runtimehosttest.NewFakeHost().
		WithVersion("nginx", runtimehost.ToolInfo{Name: "nginx", Available: true, Version: "nginx/1.24.0"}).
		WithDir("/etc/nginx").
		WithFile("/etc/nginx/nginx.conf").
		WithDir("/etc/nginx/sites-available").
		WithDir("/etc/nginx/sites-enabled")
}

func TestCreateVirtualHostWritesAndEnables(t *testing.T) {
	host := availableHost()
	manager := nginx.NewManager(host)

	err := manager.CreateVirtualHost(context.Background(), validVirtualHost())
	if err != nil {
		t.Fatalf("CreateVirtualHost() error = %v", err)
	}

	content, ok := host.WrittenFile("/etc/nginx/sites-available/example.com.conf")
	if !ok {
		t.Fatal("expected a config file to be written to sites-available")
	}
	if len(content) == 0 {
		t.Error("expected non-empty rendered content")
	}

	target, ok := host.SymlinkTarget("/etc/nginx/sites-enabled/example.com.conf")
	if !ok || target != "/etc/nginx/sites-available/example.com.conf" {
		t.Errorf("SymlinkTarget() = (%q, %v), want the sites-available path", target, ok)
	}
}

func TestCreateVirtualHostAlreadyExists(t *testing.T) {
	host := availableHost().WithFile("/etc/nginx/sites-available/example.com.conf")
	manager := nginx.NewManager(host)

	err := manager.CreateVirtualHost(context.Background(), validVirtualHost())
	if !errors.Is(err, nginx.ErrSiteAlreadyExists) {
		t.Fatalf("CreateVirtualHost() error = %v, want ErrSiteAlreadyExists", err)
	}
}

func TestCreateVirtualHostInvalid(t *testing.T) {
	host := availableHost()
	manager := nginx.NewManager(host)

	vh := validVirtualHost()
	vh.ServerName = ""

	err := manager.CreateVirtualHost(context.Background(), vh)
	if err == nil {
		t.Fatal("expected an error for an invalid virtual host")
	}
	if _, ok := host.WrittenFile("/etc/nginx/sites-available/.conf"); ok {
		t.Error("expected nothing to be written for an invalid virtual host")
	}
}

func TestCreateVirtualHostNginxNotAvailable(t *testing.T) {
	host := runtimehosttest.NewFakeHost()
	manager := nginx.NewManager(host)

	err := manager.CreateVirtualHost(context.Background(), validVirtualHost())
	if !errors.Is(err, nginx.ErrNotAvailable) {
		t.Fatalf("CreateVirtualHost() error = %v, want ErrNotAvailable", err)
	}
}
