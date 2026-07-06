package execution

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// nodeStrategy is the reference Strategy implementation: Node.js managed
// through npm. Every future technology (Python, Go, Docker, systemd, ...)
// follows the shape established here. It talks to the operating system
// exclusively through runtimehost.Host — never os/exec, never the
// filesystem directly.
type nodeStrategy struct {
	host runtimehost.Host
}

// NewNodeStrategy returns a Strategy for Node.js applications managed with
// npm, backed by host.
func NewNodeStrategy(host runtimehost.Host) Strategy {
	return &nodeStrategy{host: host}
}

var _ Strategy = (*nodeStrategy)(nil)

func (s *nodeStrategy) Metadata() Metadata {
	return Metadata{
		Name:              "Node.js (npm)",
		SupportedRuntimes: []models.Runtime{models.RuntimeNode},
		Requirements:      []string{"node", "npm"},
		Capabilities:      []Capability{CapabilityStart, CapabilityStop},
	}
}

// CanHandle claims only Node applications explicitly configured to use
// npm — other Node package managers (pnpm, yarn, bun) get their own
// Strategy later, and an application with no package manager configured is
// never assumed to be npm.
func (s *nodeStrategy) CanHandle(app *models.Application) bool {
	return app.Runtime == models.RuntimeNode && app.Config.PackageManager == "npm"
}

// HealthCheck is implemented in node_strategy_health.go.

func (s *nodeStrategy) Readiness(ctx context.Context, app *models.Application) (Readiness, error) {
	health, err := s.HealthCheck(ctx, app)
	if err != nil {
		return Readiness{}, err
	}
	return DeriveReadiness(health), nil
}

// Start and Stop are implemented in node_strategy_run.go.
// Logs is implemented in node_strategy_logs.go.

func (s *nodeStrategy) Install(context.Context, *models.Application) error {
	return ErrNotImplemented
}

func (s *nodeStrategy) Build(context.Context, *models.Application) error {
	return ErrNotImplemented
}

func (s *nodeStrategy) Restart(context.Context, *models.Application) error {
	return ErrNotImplemented
}

func (s *nodeStrategy) Update(context.Context, *models.Application) error {
	return ErrNotImplemented
}
