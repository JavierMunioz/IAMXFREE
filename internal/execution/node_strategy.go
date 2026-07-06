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

// HealthCheck, Readiness, Start and Stop are filled in incrementally by
// later commits; for now they report not implemented like every other
// still-unimplemented lifecycle method.

func (s *nodeStrategy) HealthCheck(context.Context, *models.Application) (HealthCheck, error) {
	return HealthCheck{}, ErrNotImplemented
}

func (s *nodeStrategy) Readiness(context.Context, *models.Application) (Readiness, error) {
	return Readiness{}, ErrNotImplemented
}

func (s *nodeStrategy) Start(context.Context, *models.Application) (Session, error) {
	return Session{}, ErrNotImplemented
}

func (s *nodeStrategy) Stop(context.Context, *models.Application, Session) error {
	return ErrNotImplemented
}

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
