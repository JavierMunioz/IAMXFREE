package nginx_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
)

func validVirtualHost() nginx.VirtualHost {
	return nginx.VirtualHost{
		ServerName: "example.com",
		Listen:     80,
		Locations: []nginx.Location{
			{
				Path:     "/",
				Kind:     nginx.LocationKindReverseProxy,
				Upstream: &nginx.Upstream{Host: "localhost", Port: 3000},
			},
		},
	}
}

func TestValidateVirtualHostValid(t *testing.T) {
	result := nginx.ValidateVirtualHost(validVirtualHost())
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", result.Errors)
	}
}

func TestValidateVirtualHostMissingServerName(t *testing.T) {
	vh := validVirtualHost()
	vh.ServerName = ""

	result := nginx.ValidateVirtualHost(vh)
	if result.Valid {
		t.Fatal("expected invalid for empty server_name")
	}
}

func TestValidateVirtualHostBadServerName(t *testing.T) {
	vh := validVirtualHost()
	vh.ServerName = "not a domain!"

	result := nginx.ValidateVirtualHost(vh)
	if result.Valid {
		t.Fatal("expected invalid for a malformed server_name")
	}
}

func TestValidateVirtualHostBadListenPort(t *testing.T) {
	vh := validVirtualHost()
	vh.Listen = 70000

	result := nginx.ValidateVirtualHost(vh)
	if result.Valid {
		t.Fatal("expected invalid for an out-of-range listen port")
	}
}

func TestValidateVirtualHostNoLocations(t *testing.T) {
	vh := validVirtualHost()
	vh.Locations = nil

	result := nginx.ValidateVirtualHost(vh)
	if result.Valid {
		t.Fatal("expected invalid for zero locations")
	}
}

func TestValidateVirtualHostDuplicateLocationPaths(t *testing.T) {
	vh := validVirtualHost()
	vh.Locations = append(vh.Locations, nginx.Location{
		Path:     "/",
		Kind:     nginx.LocationKindReverseProxy,
		Upstream: &nginx.Upstream{Host: "localhost", Port: 3001},
	})

	result := nginx.ValidateVirtualHost(vh)
	if result.Valid {
		t.Fatal("expected invalid for duplicate location paths")
	}
}

func TestValidateVirtualHostLocationPathMustStartWithSlash(t *testing.T) {
	vh := validVirtualHost()
	vh.Locations[0].Path = "api"

	result := nginx.ValidateVirtualHost(vh)
	if result.Valid {
		t.Fatal("expected invalid for a location path not starting with /")
	}
}

func TestValidateVirtualHostReverseProxyMissingUpstream(t *testing.T) {
	vh := validVirtualHost()
	vh.Locations[0].Upstream = nil

	result := nginx.ValidateVirtualHost(vh)
	if result.Valid {
		t.Fatal("expected invalid for a reverse proxy location with no upstream")
	}
}

func TestValidateVirtualHostReverseProxyBadUpstreamPort(t *testing.T) {
	vh := validVirtualHost()
	vh.Locations[0].Upstream.Port = 0

	result := nginx.ValidateVirtualHost(vh)
	if result.Valid {
		t.Fatal("expected invalid for an out-of-range upstream port")
	}
}

func TestValidateVirtualHostUnsupportedLocationKind(t *testing.T) {
	vh := validVirtualHost()
	vh.Locations[0].Kind = "php"

	result := nginx.ValidateVirtualHost(vh)
	if result.Valid {
		t.Fatal("expected invalid for an unsupported location kind")
	}
}
