package execution

import (
	"context"
	"errors"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// ErrNotImplemented is returned by a Strategy's lifecycle methods until real
// execution logic is written for it. It exists so every strategy reports
// "not implemented" the same way instead of each inventing its own error or
// panicking.
var ErrNotImplemented = errors.New("execution: not implemented yet")

// Capability names one lifecycle operation a Strategy actually supports.
// The dashboard will use this, once execution is wired up, to show which
// actions make sense for a given application instead of offering ones that
// would just fail.
type Capability string

const (
	CapabilityInstall Capability = "install"
	CapabilityBuild   Capability = "build"
	CapabilityStart   Capability = "start"
	CapabilityStop    Capability = "stop"
	CapabilityRestart Capability = "restart"
	CapabilityUpdate  Capability = "update"
)

// Metadata describes a Strategy for display and discovery purposes. It is
// what the dashboard will eventually show to explain how an application is
// managed, independent of whether any lifecycle method has ever been
// called.
type Metadata struct {
	Name                string
	SupportedRuntimes   []models.Runtime
	SupportedFrameworks []models.Framework
	Requirements        []string
	Capabilities        []Capability
}

// Strategy describes how one kind of application is installed, built and
// run. Each supported technology implements its own Strategy; none exist
// yet — this package only defines the shape they will take.
type Strategy interface {
	// Metadata describes this strategy: its name, what it supports, and
	// what it requires to be available on the VPS.
	Metadata() Metadata

	// CanHandle reports whether this strategy knows how to manage app. It
	// is a pure decision based on the application's own data — it performs
	// no I/O and starts no process.
	CanHandle(app *models.Application) bool

	Install(ctx context.Context, app *models.Application) error
	Build(ctx context.Context, app *models.Application) error
	Start(ctx context.Context, app *models.Application) error
	Stop(ctx context.Context, app *models.Application) error
	Restart(ctx context.Context, app *models.Application) error
	Update(ctx context.Context, app *models.Application) error
}
