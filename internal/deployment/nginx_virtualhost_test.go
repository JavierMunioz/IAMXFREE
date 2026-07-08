package deployment

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
)

func TestVirtualHostForBuildsAReverseProxyToPort(t *testing.T) {
	app := &models.Application{Config: models.DeploymentConfig{Domain: "example.com"}}

	vh := virtualHostFor(app, 3001)

	if vh.ServerName != "example.com" {
		t.Errorf("ServerName = %q, want %q", vh.ServerName, "example.com")
	}
	if vh.Listen != 80 {
		t.Errorf("Listen = %d, want 80", vh.Listen)
	}
	if len(vh.Locations) != 1 {
		t.Fatalf("len(Locations) = %d, want 1", len(vh.Locations))
	}
	loc := vh.Locations[0]
	if loc.Path != "/" || loc.Kind != nginx.LocationKindReverseProxy {
		t.Errorf("Location = %+v, want Path=/ Kind=reverse_proxy", loc)
	}
	if loc.Upstream == nil || loc.Upstream.Host != "localhost" || loc.Upstream.Port != 3001 {
		t.Errorf("Upstream = %+v, want localhost:3001", loc.Upstream)
	}
}

func TestVirtualHostForProducesAValidVirtualHost(t *testing.T) {
	app := &models.Application{Config: models.DeploymentConfig{Domain: "example.com"}}

	result := nginx.ValidateVirtualHost(virtualHostFor(app, 3001))
	if !result.Valid {
		t.Errorf("expected a valid VirtualHost, got errors: %v", result.Errors)
	}
}

func TestVirtualHostForDifferentPortsOnlyDifferInUpstreamPort(t *testing.T) {
	app := &models.Application{Config: models.DeploymentConfig{Domain: "example.com"}}

	active := virtualHostFor(app, 3000)
	candidate := virtualHostFor(app, 3001)

	if active.ServerName != candidate.ServerName || active.Listen != candidate.Listen {
		t.Fatal("expected everything but the upstream port to be identical")
	}
	if active.Locations[0].Upstream.Port == candidate.Locations[0].Upstream.Port {
		t.Fatal("expected the two VirtualHosts to target different upstream ports")
	}
}
