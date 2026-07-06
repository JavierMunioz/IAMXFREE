package services

import (
	"context"

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
}
