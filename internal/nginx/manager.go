package nginx

import "github.com/JavierMunioz/IAMXFREE/internal/runtimehost"

// Manager administers Nginx virtual hosts on the host it is given. It is
// the only entry point this package exposes for detecting, validating,
// and creating/updating/deleting/listing sites — every operation goes
// through the runtimehost.Host it was built with, never the OS directly.
type Manager struct {
	host runtimehost.Host
}

// NewManager returns a Manager that operates through host.
func NewManager(host runtimehost.Host) *Manager {
	return &Manager{host: host}
}
