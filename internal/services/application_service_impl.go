package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories"
)

// ErrApplicationNameTaken is returned by Register when an application with
// the same name is already persisted.
var ErrApplicationNameTaken = errors.New("application name is already taken")

type applicationService struct {
	repo       repositories.ApplicationRepository
	resolver   *execution.Resolver
	gitManager *git.Manager
}

// NewApplicationService builds the default ApplicationService, backed by
// repo, resolver and gitManager. Swapping repo's concrete implementation
// (JSON, SQLite, ...) never requires changing this type, and neither does
// registering a new execution.Strategy with resolver.
func NewApplicationService(repo repositories.ApplicationRepository, resolver *execution.Resolver, gitManager *git.Manager) ApplicationService {
	return &applicationService{repo: repo, resolver: resolver, gitManager: gitManager}
}

func (s *applicationService) Register(ctx context.Context, app *models.Application) error {
	if err := app.Validate(); err != nil {
		return err
	}

	_, err := s.repo.FindByName(ctx, app.Name)
	switch {
	case err == nil:
		return fmt.Errorf("%q: %w", app.Name, ErrApplicationNameTaken)
	case errors.Is(err, repositories.ErrApplicationNotFound):
		// no conflict; continue
	default:
		return err
	}

	return s.repo.Create(ctx, app)
}

func (s *applicationService) Get(ctx context.Context, id string) (*models.Application, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *applicationService) List(ctx context.Context) ([]*models.Application, error) {
	return s.repo.List(ctx)
}

func (s *applicationService) UpdateConfig(ctx context.Context, id string, config models.DeploymentConfig) (*models.Application, error) {
	app, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	app.Config = config
	app.Touch()

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *applicationService) ChangeStatus(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error) {
	app, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	app.Status = status
	app.Touch()

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *applicationService) Remove(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *applicationService) ResolveExecutionStrategy(ctx context.Context, id string) (execution.Metadata, error) {
	app, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return execution.Metadata{}, err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return execution.Metadata{}, err
	}

	return strategy.Metadata(), nil
}

func (s *applicationService) CheckExecutionHealth(ctx context.Context, id string) (ExecutionHealth, error) {
	app, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return ExecutionHealth{}, err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return ExecutionHealth{}, err
	}

	health, err := strategy.HealthCheck(ctx, app)
	if err != nil {
		return ExecutionHealth{}, err
	}

	return ExecutionHealth{
		StrategyName: strategy.Metadata().Name,
		Healthy:      health.Healthy(),
	}, nil
}

func (s *applicationService) CheckGitStatus(ctx context.Context, id string) (GitStatus, error) {
	app, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return GitStatus{}, err
	}

	if strings.TrimSpace(app.Source.LocalPath) == "" {
		return GitStatus{}, nil
	}

	repo, err := s.gitManager.Inspect(ctx, app.Source.LocalPath)
	if err != nil {
		return GitStatus{}, err
	}
	if !repo.IsRepo {
		return GitStatus{}, nil
	}

	status := GitStatus{
		IsRepo:        true,
		Branch:        repo.Branch.Name,
		Detached:      repo.Branch.Detached,
		CommitSHA:     repo.Commit.ShortSHA,
		CommitMessage: repo.Commit.Message,
		Clean:         repo.Status.WorkingTree.Clean,
		Modified:      len(repo.Status.WorkingTree.Modified),
		Untracked:     len(repo.Status.WorkingTree.Untracked),
		Ahead:         repo.Status.Ahead,
		Behind:        repo.Status.Behind,
	}
	if len(repo.Remotes) > 0 {
		status.RemoteName = repo.Remotes[0].Name
		status.RemoteURL = repo.Remotes[0].URL
	}

	return status, nil
}
