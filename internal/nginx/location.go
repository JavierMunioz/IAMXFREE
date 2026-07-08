package nginx

// LocationKind names what a Location does with matching requests.
// ReverseProxy is the only kind implemented this iteration; PHP, static
// file serving and others are future work, added here as new constants
// without changing Location's shape.
type LocationKind string

const (
	LocationKindReverseProxy LocationKind = "reverse_proxy"
)

// Location is one `location` block inside a VirtualHost. Path is the
// request path it matches (e.g. "/"). Which of the kind-specific fields is
// populated is determined by Kind — today only Upstream, for
// LocationKindReverseProxy.
type Location struct {
	Path     string
	Kind     LocationKind
	Upstream *Upstream
}
