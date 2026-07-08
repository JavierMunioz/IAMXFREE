package nginx_test

import (
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
)

func TestRenderReverseProxy(t *testing.T) {
	vh := nginx.VirtualHost{
		ServerName:    "example.com",
		ServerAliases: []string{"www.example.com"},
		Listen:        80,
		Locations: []nginx.Location{
			{
				Path:     "/",
				Kind:     nginx.LocationKindReverseProxy,
				Upstream: &nginx.Upstream{Host: "localhost", Port: 3000},
			},
		},
	}

	out, err := nginx.Render(vh)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{
		"server {",
		"listen 80;",
		"server_name example.com www.example.com;",
		"location / {",
		"proxy_pass http://localhost:3000;",
		"proxy_set_header Host $host;",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() output missing %q, got:\n%s", want, out)
		}
	}
}

func TestRenderUnknownLocationKind(t *testing.T) {
	vh := nginx.VirtualHost{
		ServerName: "example.com",
		Listen:     80,
		Locations: []nginx.Location{
			{Path: "/", Kind: "php"},
		},
	}

	_, err := nginx.Render(vh)
	if err == nil {
		t.Fatal("expected an error for an unregistered location kind")
	}
}
