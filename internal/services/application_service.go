package services

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// ApplicationService will define the business operations available for
// managing applications, coordinating an ApplicationRepository with the
// process/reverse-proxy/SSL managers once they exist. It is declared now,
// without an implementation, so internal/core and internal/tui can be
// designed against a stable contract before that logic is built.
type ApplicationService interface {
	Register(ctx context.Context, app *models.Application) error
	Get(ctx context.Context, id string) (*models.Application, error)
	List(ctx context.Context) ([]*models.Application, error)
	UpdateConfig(ctx context.Context, id string, config models.DeploymentConfig) (*models.Application, error)
	ChangeStatus(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error)
	Remove(ctx context.Context, id string) error
}
