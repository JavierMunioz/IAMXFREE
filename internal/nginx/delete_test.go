package nginx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
)

func TestDeleteVirtualHostRemovesFileAndSymlink(t *testing.T) {
	host := availableHost().
		WithFile("/etc/nginx/sites-available/example.com.conf").
		WithFile("/etc/nginx/sites-enabled/example.com.conf")
	manager := nginx.NewManager(host)

	if err := manager.DeleteVirtualHost(context.Background(), "example.com"); err != nil {
		t.Fatalf("DeleteVirtualHost() error = %v", err)
	}

	if !host.Removed("/etc/nginx/sites-enabled/example.com.conf") {
		t.Error("expected the sites-enabled symlink to be removed")
	}
	if !host.Removed("/etc/nginx/sites-available/example.com.conf") {
		t.Error("expected the sites-available config file to be removed")
	}
}

func TestDeleteVirtualHostAlreadyDisabled(t *testing.T) {
	host := availableHost().WithFile("/etc/nginx/sites-available/example.com.conf")
	manager := nginx.NewManager(host)

	if err := manager.DeleteVirtualHost(context.Background(), "example.com"); err != nil {
		t.Fatalf("DeleteVirtualHost() error = %v", err)
	}
	if !host.Removed("/etc/nginx/sites-available/example.com.conf") {
		t.Error("expected the sites-available config file to be removed")
	}
}

func TestDeleteVirtualHostNotFound(t *testing.T) {
	host := availableHost()
	manager := nginx.NewManager(host)

	err := manager.DeleteVirtualHost(context.Background(), "example.com")
	if !errors.Is(err, nginx.ErrSiteNotFound) {
		t.Fatalf("DeleteVirtualHost() error = %v, want ErrSiteNotFound", err)
	}
}
