package git

import "github.com/JavierMunioz/IAMXFREE/internal/runtimehost"

// Manager inspects Git repositories on the host it is given. Every
// operation runs through the runtimehost.Host it was built with — never
// exec.Command directly.
type Manager struct {
	host runtimehost.Host
}

// NewManager returns a Manager that operates through host.
func NewManager(host runtimehost.Host) *Manager {
	return &Manager{host: host}
}
