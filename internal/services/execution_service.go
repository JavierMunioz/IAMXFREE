package services

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/monitor"
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

	// OpenLogs opens a live stream of session's captured output. It never
	// starts a new process — it attaches to output already being captured
	// for the session.
	OpenLogs(ctx context.Context, appID string, session RunSession) (LogStream, error)

	// Snapshot observes session's process right now — its real OS-level
	// state and resource usage — via the Runtime Monitor. Unlike
	// Start/Stop/RefreshSession, it never resolves an execution.Strategy:
	// the Runtime Monitor talks to runtimehost.Host directly, the same way
	// regardless of which strategy started the process. It is always
	// on-demand; nothing calls this automatically.
	Snapshot(ctx context.Context, session RunSession) (RuntimeSnapshot, error)
}

type executionService struct {
	repo     repositories.ApplicationRepository
	resolver *execution.Resolver
	monitor  *monitor.Monitor
}

// NewExecutionService builds the default ExecutionService, backed by repo,
// resolver and monitor.
func NewExecutionService(repo repositories.ApplicationRepository, resolver *execution.Resolver, runtimeMonitor *monitor.Monitor) ExecutionService {
	return &executionService{repo: repo, resolver: resolver, monitor: runtimeMonitor}
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

func (s *executionService) OpenLogs(ctx context.Context, appID string, session RunSession) (LogStream, error) {
	app, err := s.repo.FindByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	strategy, err := s.resolver.Resolve(app)
	if err != nil {
		return nil, err
	}

	stream, err := strategy.Logs(ctx, app, fromRunSession(session))
	if err != nil {
		return nil, err
	}
	return newLogStreamAdapter(stream), nil
}

func (s *executionService) Snapshot(_ context.Context, session RunSession) (RuntimeSnapshot, error) {
	snap, err := s.monitor.Snapshot(fromRunSession(session))
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	return toRuntimeSnapshot(snap), nil
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
