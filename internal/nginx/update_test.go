package nginx_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
)

func TestUpdateVirtualHostOverwritesContent(t *testing.T) {
	host := availableHost().WithFile("/etc/nginx/sites-available/example.com.conf")
	manager := nginx.NewManager(host)

	vh := validVirtualHost()
	vh.Locations[0].Upstream.Port = 4000

	if err := manager.UpdateVirtualHost(context.Background(), vh); err != nil {
		t.Fatalf("UpdateVirtualHost() error = %v", err)
	}

	content, ok := host.WrittenFile("/etc/nginx/sites-available/example.com.conf")
	if !ok {
		t.Fatal("expected the config file to be rewritten")
	}
	if !strings.Contains(string(content), "4000") {
		t.Errorf("expected rewritten content to reference the new port, got:\n%s", content)
	}
}

func TestUpdateVirtualHostNotFound(t *testing.T) {
	host := availableHost()
	manager := nginx.NewManager(host)

	err := manager.UpdateVirtualHost(context.Background(), validVirtualHost())
	if !errors.Is(err, nginx.ErrSiteNotFound) {
		t.Fatalf("UpdateVirtualHost() error = %v, want ErrSiteNotFound", err)
	}
}

func TestUpdateVirtualHostReEnablesDisabledSite(t *testing.T) {
	host := availableHost().WithFile("/etc/nginx/sites-available/example.com.conf")
	manager := nginx.NewManager(host)

	if err := manager.UpdateVirtualHost(context.Background(), validVirtualHost()); err != nil {
		t.Fatalf("UpdateVirtualHost() error = %v", err)
	}

	target, ok := host.SymlinkTarget("/etc/nginx/sites-enabled/example.com.conf")
	if !ok || target != "/etc/nginx/sites-available/example.com.conf" {
		t.Errorf("SymlinkTarget() = (%q, %v), want the site re-enabled", target, ok)
	}
}

func TestUpdateVirtualHostLeavesEnabledSiteAlone(t *testing.T) {
	host := availableHost().
		WithFile("/etc/nginx/sites-available/example.com.conf").
		WithFile("/etc/nginx/sites-enabled/example.com.conf")
	manager := nginx.NewManager(host)

	if err := manager.UpdateVirtualHost(context.Background(), validVirtualHost()); err != nil {
		t.Fatalf("UpdateVirtualHost() error = %v", err)
	}

	if _, ok := host.SymlinkTarget("/etc/nginx/sites-enabled/example.com.conf"); ok {
		t.Error("expected no new Symlink call for an already-enabled site")
	}
}
