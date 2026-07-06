package services

import (
	"context"
	"sync"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
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

	// ActiveSession reports the most recently started or refreshed session
	// tracked in memory for appID, if any. It performs no I/O — it is not
	// a live liveness check, just whatever Start/Stop/RefreshSession last
	// observed — so a process that exited on its own is only reflected
	// here after an explicit RefreshSession. This is what lets a screen
	// other than the one that started a session (e.g. the dashboard) know
	// one exists.
	ActiveSession(appID string) (RunSession, bool)
}

type executionService struct {
	repo        repositories.ApplicationRepository
	resolver    *execution.Resolver
	monitor     *monitor.Monitor
	sessionRepo repositories.SessionRepository

	mu       sync.Mutex
	sessions map[string]execution.Session
}

// NewExecutionService builds the default ExecutionService, backed by repo,
// resolver, monitor and sessionRepo. It immediately rehydrates its
// in-memory session registry from sessionRepo, keeping only the sessions
// whose process the Runtime Monitor confirms is still actually alive right
// now — a one-time check, not periodic polling — and discarding (and
// un-persisting) the rest. This is what lets ActiveSession still report a
// session that was started in a previous run of IAMXFREE, as long as its
// process is still running.
func NewExecutionService(repo repositories.ApplicationRepository, resolver *execution.Resolver, runtimeMonitor *monitor.Monitor, sessionRepo repositories.SessionRepository) ExecutionService {
	svc := &executionService{
		repo:        repo,
		resolver:    resolver,
		monitor:     runtimeMonitor,
		sessionRepo: sessionRepo,
		sessions:    make(map[string]execution.Session),
	}
	svc.hydrateSessions(context.Background())
	return svc
}

func (s *executionService) hydrateSessions(ctx context.Context) {
	persisted, err := s.sessionRepo.List(ctx)
	if err != nil {
		return
	}

	for appID, session := range persisted {
		snap, err := s.monitor.Snapshot(session)
		if err != nil || snap.Process.State != monitor.ProcessStateRunning {
			_ = s.sessionRepo.Delete(ctx, appID)
			continue
		}
		s.sessions[appID] = session

		if app, err := s.repo.FindByID(ctx, appID); err == nil {
			s.syncApplicationStatus(ctx, app, execution.StatusRunning)
		}
	}
}

// trackSession records session as the one currently tracked for appID, both
// in memory and persisted via sessionRepo. Persisting is best-effort: a
// failure to write the file doesn't undo the fact that the process is
// genuinely running — it only means ActiveSession might forget about it if
// IAMXFREE is restarted before the next successful save.
func (s *executionService) trackSession(ctx context.Context, appID string, session execution.Session) {
	s.mu.Lock()
	s.sessions[appID] = session
	s.mu.Unlock()

	_ = s.sessionRepo.Save(ctx, appID, session)
}

func (s *executionService) untrackSession(ctx context.Context, appID string) {
	s.mu.Lock()
	delete(s.sessions, appID)
	s.mu.Unlock()

	_ = s.sessionRepo.Delete(ctx, appID)
}

// syncApplicationStatus best-effort persists app's Status to match status —
// the session's lifecycle state — so the badge shown on the dashboard and
// detail screens reflects reality instead of whatever it was when the
// application was registered (or last edited). A failure to persist here
// doesn't affect the caller: the process itself already started, stopped,
// or was refreshed successfully regardless.
func (s *executionService) syncApplicationStatus(ctx context.Context, app *models.Application, status execution.Status) {
	app.Status = applicationStatusFor(status)
	app.Touch()
	_ = s.repo.Update(ctx, app)
}

// applicationStatusFor maps a Session's technology-agnostic lifecycle
// Status onto the coarser models.ApplicationStatus shown throughout the
// TUI. Starting/stopping/stopped all collapse to StatusStopped — the
// dashboard has no notion of those in-between states — while Failed maps
// to StatusError so a crashed process reads differently from one that was
// cleanly stopped.
func applicationStatusFor(status execution.Status) models.ApplicationStatus {
	switch status {
	case execution.StatusRunning:
		return models.StatusRunning
	case execution.StatusFailed:
		return models.StatusError
	default:
		return models.StatusStopped
	}
}

func (s *executionService) ActiveSession(appID string) (RunSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[appID]
	if !ok {
		return RunSession{}, false
	}
	return toRunSession(session), true
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
	s.trackSession(ctx, appID, session)
	s.syncApplicationStatus(ctx, app, session.Status)
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

	if err := strategy.Stop(ctx, app, fromRunSession(session)); err != nil {
		return err
	}
	s.untrackSession(ctx, appID)
	s.syncApplicationStatus(ctx, app, execution.StatusStopped)
	return nil
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
	s.trackSession(ctx, appID, updated)
	s.syncApplicationStatus(ctx, app, updated.Status)
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
