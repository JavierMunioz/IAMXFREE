package deployment

import (
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
)

// virtualHostFor builds the nginx.VirtualHost that should exist for app's
// configured domain, reverse-proxying to port on localhost. It is the
// single source of truth a zero-downtime deployment uses both to switch
// the upstream to a candidate port and to restore it if that switch needs
// to be compensated — the same construction, called with a different
// port, passed straight to the existing UpdateVirtualHost (no new Nginx
// capability needed: re-rendering the whole site from this desired state
// is how the package already works).
func virtualHostFor(app *models.Application, port int) nginx.VirtualHost {
	return nginx.VirtualHost{
		ServerName: app.Config.Domain,
		Listen:     80,
		Locations: []nginx.Location{
			{
				Path: "/",
				Kind: nginx.LocationKindReverseProxy,
				Upstream: &nginx.Upstream{
					Host: "localhost",
					Port: port,
				},
			},
		},
	}
}
