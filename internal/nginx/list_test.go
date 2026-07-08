package nginx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestListVirtualHosts(t *testing.T) {
	host := availableHost().
		WithReadDir("/etc/nginx/sites-available", []string{"zeta.com.conf", "alpha.com.conf", "README"}, nil).
		WithFile("/etc/nginx/sites-enabled/alpha.com.conf")
	manager := nginx.NewManager(host)

	sites, err := manager.ListVirtualHosts(context.Background())
	if err != nil {
		t.Fatalf("ListVirtualHosts() error = %v", err)
	}
	if len(sites) != 2 {
		t.Fatalf("ListVirtualHosts() = %v, want 2 sites (non-.conf entries excluded)", sites)
	}

	if sites[0].ServerName != "alpha.com" || !sites[0].Enabled {
		t.Errorf("sites[0] = %+v, want alpha.com enabled", sites[0])
	}
	if sites[1].ServerName != "zeta.com" || sites[1].Enabled {
		t.Errorf("sites[1] = %+v, want zeta.com disabled", sites[1])
	}
}

func TestListVirtualHostsEmpty(t *testing.T) {
	host := availableHost()
	manager := nginx.NewManager(host)

	sites, err := manager.ListVirtualHosts(context.Background())
	if err != nil {
		t.Fatalf("ListVirtualHosts() error = %v", err)
	}
	if len(sites) != 0 {
		t.Fatalf("ListVirtualHosts() = %v, want empty", sites)
	}
}

func TestListVirtualHostsNginxNotAvailable(t *testing.T) {
	host := runtimehosttest.NewFakeHost()
	manager := nginx.NewManager(host)

	_, err := manager.ListVirtualHosts(context.Background())
	if !errors.Is(err, nginx.ErrNotAvailable) {
		t.Fatalf("ListVirtualHosts() error = %v, want ErrNotAvailable", err)
	}
}
