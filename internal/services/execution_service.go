package services

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories"
)

// ExecutionService is how the rest of IAMXFREE controls a running
// application — starting it, stopping it, and checking whether a
// previously-started session is still alive. It resolves each application's
// execution.Strategy the same way ApplicationService does, but callers
// depend on this interface instead of internal/execution directly.
type ExecutionService interface {
	// Start resolves the application's execution strategy and starts its
	// configured start command, returning the resulting session.
	Start(ctx context.Context, appID string) (RunSession, error)

	// Stop terminates the process described by session.
	Stop(ctx context.Context, appID string, session RunSession) error

	// RefreshSession re-checks whether session's process is still alive,
	// returning an updated RunSession.
	RefreshSession(ctx context.Context, appID string, session RunSession) (RunSession, error)
}

type executionService struct {
	repo     repositories.ApplicationRepository
	resolver *execution.Resolver
}

// NewExecutionService builds the default ExecutionService, backed by repo
// and resolver.
func NewExecutionService(repo repositories.ApplicationRepository, resolver *execution.Resolver) ExecutionService {
	return &executionService{repo: repo, resolver: resolver}
}

func (s *executionService) Start(ctx context.Context, appID string) (RunSession, error) {
	app, err := s.repo.FindByID(ctx, appID)
	if err != nil {
		return RunSession{}, err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return RunSession{}, err
	}

	session, err := strategy.Start(ctx, app)
	if err != nil {
		return RunSession{}, err
	}
	return toRunSession(session), nil
}

func (s *executionService) Stop(ctx context.Context, appID string, session RunSession) error {
	app, err := s.repo.FindByID(ctx, appID)
	if err != nil {
		return err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return err
	}

	return strategy.Stop(ctx, app, fromRunSession(session))
}

func (s *executionService) RefreshSession(ctx context.Context, appID string, session RunSession) (RunSession, error) {
	app, err := s.repo.FindByID(ctx, appID)
	if err != nil {
		return RunSession{}, err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return RunSession{}, err
	}

	updated, err := strategy.Status(ctx, app, fromRunSession(session))
	if err != nil {
		return RunSession{}, err
	}
	return toRunSession(updated), nil
}

func toRunSession(session execution.Session) RunSession {
	return RunSession{
		PID:        session.PID,
		StartedAt:  session.StartedAt,
		Command:    session.Command,
		Args:       session.Args,
		WorkingDir: session.WorkingDir,
		Status:     string(session.Status),
		Runtime:    session.Runtime,
	}
}

func fromRunSession(rs RunSession) execution.Session {
	return execution.Session{
		PID:        rs.PID,
		StartedAt:  rs.StartedAt,
		Command:    rs.Command,
		Args:       rs.Args,
		WorkingDir: rs.WorkingDir,
		Status:     execution.Status(rs.Status),
		Runtime:    rs.Runtime,
	}
}
