package nginx

import "fmt"

// Upstream is the single backend a reverse-proxy Location forwards traffic
// to (e.g. "localhost:3000"). Multiple upstream servers and load balancing
// are not supported this iteration — that is a single target, typed
// instead of a free-form string.
type Upstream struct {
	Host string
	Port int
}

// Address formats u as "host:port", the form Nginx's proxy_pass expects.
func (u Upstream) Address() string {
	return fmt.Sprintf("%s:%d", u.Host, u.Port)
}
