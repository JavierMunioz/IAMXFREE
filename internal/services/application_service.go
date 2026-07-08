package services

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// ApplicationService defines the business operations available for managing
// applications. It sits between presentation (TUI/CLI) and persistence: it
// validates, decides how conflicts are handled, and will coordinate the
// process/reverse-proxy/SSL managers once they exist. Presentation code
// depends only on this interface, never on ApplicationRepository directly.
type ApplicationService interface {
	Register(ctx context.Context, app *models.Application) error
	Get(ctx context.Context, id string) (*models.Application, error)
	List(ctx context.Context) ([]*models.Application, error)
	UpdateConfig(ctx context.Context, id string, config models.DeploymentConfig) (*models.Application, error)
	ChangeStatus(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error)
	Remove(ctx context.Context, id string) error

	// ResolveExecutionStrategy answers "which strategy will manage this
	// application" without running anything: it returns the metadata of
	// the execution.Strategy the resolver would pick, or an error if none
	// is registered for the application's runtime/framework yet.
	ResolveExecutionStrategy(ctx context.Context, id string) (execution.Metadata, error)

	// CheckExecutionHealth resolves the application's execution strategy
	// (if any) and runs its HealthCheck, summarizing the result for
	// display (e.g. on a dashboard card) without exposing execution types
	// to callers.
	CheckExecutionHealth(ctx context.Context, id string) (ExecutionHealth, error)

	// CheckGitStatus inspects the application's source directory
	// (Source.LocalPath) for a Git repository, summarizing its state for
	// display (e.g. on a dashboard card or a detail panel) without
	// exposing git types to callers. IsRepo == false is a normal outcome
	// — no repository configured, or the directory isn't one — never an
	// error by itself.
	CheckGitStatus(ctx context.Context, id string) (GitStatus, error)
}

// ExecutionHealth summarizes how an application would be managed and
// whether its prerequisites are currently satisfied.
type ExecutionHealth struct {
	StrategyName string
	Healthy      bool
}

// GitStatus summarizes an application's Git repository state. Every field
// besides IsRepo is zero-valued when IsRepo is false.
type GitStatus struct {
	IsRepo bool

	Branch   string
	Detached bool

	CommitSHA     string
	CommitMessage string

	Clean     bool
	Modified  int
	Untracked int
	Ahead     int
	Behind    int

	RemoteName string
	RemoteURL  string
}
