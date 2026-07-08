package nginx

import (
	"context"
	"fmt"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// Reload asks the running Nginx process to reload its configuration via
// `nginx -s reload` — a graceful reload that picks up config changes
// without dropping connections or restarting the process. A failure here
// is always a real error: there is no valid negative outcome for "reload
// the server", unlike this package's read-only checks.
func (m *Manager) Reload(ctx context.Context) error {
	result, err := m.host.RunCaptured(ctx, runtimehost.Command{Name: "nginx", Args: []string{"-s", "reload"}})
	if err != nil {
		output := strings.TrimSpace(result.Stdout + result.Stderr)
		return fmt.Errorf("nginx: reload failed: %s: %w", output, err)
	}
	return nil
}
