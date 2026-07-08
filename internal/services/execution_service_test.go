package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/monitor"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories/jsonstore"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func newExecutionServiceWithStrategy(t *testing.T, strategy execution.Strategy) (services.ExecutionService, *models.Application) {
	t.Helper()
	return newExecutionServiceWithHost(t, strategy, runtimehosttest.NewFakeHost())
}

func newExecutionServiceWithHost(t *testing.T, strategy execution.Strategy, host runtimehost.Host) (services.ExecutionService, *models.Application) {
	t.Helper()
	return newExecutionServiceWithHostAndSessions(t, strategy, host, newSessionRepo(t))
}

func newSessionRepo(t *testing.T) repositories.SessionRepository {
	t.Helper()
	repo, err := jsonstore.NewSessionRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewSessionRepository() error = %v", err)
	}
	return repo
}

func newExecutionServiceWithHostAndSessions(t *testing.T, strategy execution.Strategy, host runtimehost.Host, sessionRepo repositories.SessionRepository) (services.ExecutionService, *models.Application) {
	t.Helper()

	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}

	registry := execution.NewRegistry()
	registry.Register(strategy)
	resolver := execution.NewResolver(registry)

	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	if err := repo.Create(context.Background(), app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	return services.NewExecutionService(repo, resolver, monitor.New(host), sessionRepo), app
}

func TestExecutionServiceInstall(t *testing.T) {
	strategy := &fakeStrategy{runtime: models.RuntimeNode}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if err := svc.Install(context.Background(), app.ID); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
}

func TestExecutionServiceInstallPropagatesStrategyError(t *testing.T) {
	wantErr := errors.New("npm not found")
	strategy := &fakeStrategy{runtime: models.RuntimeNode, installErr: wantErr}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if err := svc.Install(context.Background(), app.ID); !errors.Is(err, wantErr) {
		t.Fatalf("Install() error = %v, want %v", err, wantErr)
	}
}

func TestExecutionServiceInstallUnknownApplication(t *testing.T) {
	svc, _ := newExecutionServiceWithStrategy(t, &fakeStrategy{runtime: models.RuntimeNode})

	if err := svc.Install(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("expected an error for an unknown application")
	}
}

func TestExecutionServiceBuild(t *testing.T) {
	strategy := &fakeStrategy{runtime: models.RuntimeNode}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if err := svc.Build(context.Background(), app.ID); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestExecutionServiceBuildPropagatesStrategyError(t *testing.T) {
	wantErr := errors.New("build script failed")
	strategy := &fakeStrategy{runtime: models.RuntimeNode, buildErr: wantErr}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if err := svc.Build(context.Background(), app.ID); !errors.Is(err, wantErr) {
		t.Fatalf("Build() error = %v, want %v", err, wantErr)
	}
}

func TestExecutionServiceStartReturnsRunSession(t *testing.T) {
	startedAt := time.Now().UTC()
	strategy := &fakeStrategy{
		runtime: models.RuntimeNode,
		startSession: execution.Session{
			PID:        4242,
			StartedAt:  startedAt,
			Command:    "npm",
			Args:       []string{"start"},
			WorkingDir: "/srv/apps/my-api",
			Status:     execution.StatusRunning,
			Runtime:    models.RuntimeNode,
		},
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	session, err := svc.Start(context.Background(), app.ID)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if session.PID != 4242 {
		t.Errorf("PID = %d, want 4242", session.PID)
	}
	if session.Command != "npm" || len(session.Args) != 1 || session.Args[0] != "start" {
		t.Errorf("Command/Args = %q/%v", session.Command, session.Args)
	}
	if session.Status != string(execution.StatusRunning) {
		t.Errorf("Status = %q, want %q", session.Status, execution.StatusRunning)
	}
	if session.Runtime != models.RuntimeNode {
		t.Errorf("Runtime = %q, want %q", session.Runtime, models.RuntimeNode)
	}
}

func TestExecutionServiceStartPropagatesStrategyError(t *testing.T) {
	wantErr := errors.New("not ready")
	strategy := &fakeStrategy{runtime: models.RuntimeNode, startErr: wantErr}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if _, err := svc.Start(context.Background(), app.ID); !errors.Is(err, wantErr) {
		t.Fatalf("Start() error = %v, want %v", err, wantErr)
	}
}

func TestExecutionServiceStartUnknownApplication(t *testing.T) {
	svc, _ := newExecutionServiceWithStrategy(t, &fakeStrategy{runtime: models.RuntimeNode})

	if _, err := svc.Start(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("expected an error for an unknown application")
	}
}

func TestExecutionServiceStop(t *testing.T) {
	strategy := &fakeStrategy{runtime: models.RuntimeNode}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	session := services.RunSession{PID: 4242}
	if err := svc.Stop(context.Background(), app.ID, session); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestExecutionServiceStopPropagatesStrategyError(t *testing.T) {
	wantErr := errors.New("no such process")
	strategy := &fakeStrategy{runtime: models.RuntimeNode, stopErr: wantErr}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if err := svc.Stop(context.Background(), app.ID, services.RunSession{PID: 4242}); !errors.Is(err, wantErr) {
		t.Fatalf("Stop() error = %v, want %v", err, wantErr)
	}
}

func TestExecutionServiceRefreshSession(t *testing.T) {
	strategy := &fakeStrategy{
		runtime: models.RuntimeNode,
		statusSession: execution.Session{
			PID:    4242,
			Status: execution.StatusStopped,
		},
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	refreshed, err := svc.RefreshSession(context.Background(), app.ID, services.RunSession{PID: 4242})
	if err != nil {
		t.Fatalf("RefreshSession() error = %v", err)
	}
	if refreshed.Status != string(execution.StatusStopped) {
		t.Errorf("Status = %q, want %q", refreshed.Status, execution.StatusStopped)
	}
}

func TestExecutionServiceRefreshSessionPropagatesStrategyError(t *testing.T) {
	wantErr := errors.New("lookup failed")
	strategy := &fakeStrategy{runtime: models.RuntimeNode, statusErr: wantErr}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if _, err := svc.RefreshSession(context.Background(), app.ID, services.RunSession{PID: 4242}); !errors.Is(err, wantErr) {
		t.Fatalf("RefreshSession() error = %v, want %v", err, wantErr)
	}
}

// fakeLogStream is a minimal execution.LogStream test double.
type fakeLogStream struct {
	events chan execution.LogEvent
	err    error
}

func newFakeLogStream(logEvents []execution.LogEvent, err error) *fakeLogStream {
	ch := make(chan execution.LogEvent, len(logEvents))
	for _, e := range logEvents {
		ch <- e
	}
	close(ch)
	return &fakeLogStream{events: ch, err: err}
}

func (s *fakeLogStream) Events() <-chan execution.LogEvent { return s.events }
func (s *fakeLogStream) Err() error                        { return s.err }
func (s *fakeLogStream) Close() error                      { return nil }

func TestExecutionServiceOpenLogsAdaptsEvents(t *testing.T) {
	logEvents := []execution.LogEvent{
		{Timestamp: time.Now(), Type: execution.LogEventStdout, Content: "listening on :3000"},
		{Timestamp: time.Now(), Type: execution.LogEventEOF, Content: "process exited"},
	}
	strategy := &fakeStrategy{runtime: models.RuntimeNode, logStream: newFakeLogStream(logEvents, nil)}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	stream, err := svc.OpenLogs(context.Background(), app.ID, services.RunSession{PID: 4242})
	if err != nil {
		t.Fatalf("OpenLogs() error = %v", err)
	}

	var got []services.LogEvent
	for e := range stream.Events() {
		got = append(got, e)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2: %+v", len(got), got)
	}
	if got[0].Type != string(execution.LogEventStdout) || got[0].Content != "listening on :3000" {
		t.Errorf("got[0] = %+v, want stdout %q", got[0], "listening on :3000")
	}
	if got[1].Type != string(execution.LogEventEOF) {
		t.Errorf("got[1].Type = %q, want %q", got[1].Type, execution.LogEventEOF)
	}
	if stream.Err() != nil {
		t.Fatalf("Err() = %v, want nil", stream.Err())
	}
}

func TestExecutionServiceOpenLogsPropagatesStrategyError(t *testing.T) {
	wantErr := errors.New("no output captured")
	strategy := &fakeStrategy{runtime: models.RuntimeNode, logsErr: wantErr}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if _, err := svc.OpenLogs(context.Background(), app.ID, services.RunSession{PID: 4242}); !errors.Is(err, wantErr) {
		t.Fatalf("OpenLogs() error = %v, want %v", err, wantErr)
	}
}

func TestExecutionServiceOpenLogsUnknownApplication(t *testing.T) {
	svc, _ := newExecutionServiceWithStrategy(t, &fakeStrategy{runtime: models.RuntimeNode})

	if _, err := svc.OpenLogs(context.Background(), "does-not-exist", services.RunSession{PID: 4242}); err == nil {
		t.Fatal("expected an error for an unknown application")
	}
}

func TestExecutionServiceSnapshotReportsRunningProcess(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunningPID(4242, true).
		WithProcessResources(4242, runtimehost.ProcessResources{
			CPUPercent:      5,
			CPUPercentKnown: true,
			MemoryRSSBytes:  64 * 1024 * 1024,
			MemoryRSSKnown:  true,
		})
	svc, _ := newExecutionServiceWithHost(t, &fakeStrategy{runtime: models.RuntimeNode}, host)

	session := services.RunSession{PID: 4242, Command: "npm", WorkingDir: "/srv/apps/my-api"}
	snap, err := svc.Snapshot(context.Background(), session)
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if snap.State != "running" {
		t.Errorf("State = %q, want %q", snap.State, "running")
	}
	if !snap.CPUPercent.Available || snap.CPUPercent.Value != 5 {
		t.Errorf("CPUPercent = %+v, want available 5", snap.CPUPercent)
	}
	if !snap.MemoryRSSBytes.Available || snap.MemoryRSSBytes.Value != 64*1024*1024 {
		t.Errorf("MemoryRSSBytes = %+v, want available %d", snap.MemoryRSSBytes, 64*1024*1024)
	}
	if snap.MemoryVSZBytes.Available {
		t.Error("expected MemoryVSZBytes to be unavailable since the host never reported it")
	}
	if snap.WorkingDir != "/srv/apps/my-api" {
		t.Errorf("WorkingDir = %q, want %q", snap.WorkingDir, "/srv/apps/my-api")
	}
}

func TestExecutionServiceSnapshotReportsStoppedProcess(t *testing.T) {
	host := runtimehosttest.NewFakeHost().WithRunningPID(4242, false)
	svc, _ := newExecutionServiceWithHost(t, &fakeStrategy{runtime: models.RuntimeNode}, host)

	snap, err := svc.Snapshot(context.Background(), services.RunSession{PID: 4242})
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if snap.State != "stopped" {
		t.Errorf("State = %q, want %q", snap.State, "stopped")
	}
}

func TestExecutionServiceActiveSessionUnknownAppReportsNotFound(t *testing.T) {
	svc, _ := newExecutionServiceWithStrategy(t, &fakeStrategy{runtime: models.RuntimeNode})

	if _, ok := svc.ActiveSession("never-started"); ok {
		t.Fatal("expected ActiveSession to report false for an app that was never started")
	}
}

func TestExecutionServiceStartTracksActiveSession(t *testing.T) {
	strategy := &fakeStrategy{
		runtime: models.RuntimeNode,
		startSession: execution.Session{
			PID:     4242,
			Status:  execution.StatusRunning,
			Runtime: models.RuntimeNode,
		},
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if _, err := svc.Start(context.Background(), app.ID); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	session, ok := svc.ActiveSession(app.ID)
	if !ok {
		t.Fatal("expected ActiveSession to report true after Start")
	}
	if session.PID != 4242 {
		t.Errorf("PID = %d, want 4242", session.PID)
	}
}

func TestExecutionServiceStopUntracksActiveSession(t *testing.T) {
	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 4242, Status: execution.StatusRunning},
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	session, err := svc.Start(context.Background(), app.ID)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := svc.Stop(context.Background(), app.ID, session); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	if _, ok := svc.ActiveSession(app.ID); ok {
		t.Fatal("expected ActiveSession to report false after Stop")
	}
}

func TestExecutionServiceStopKeepsTrackingOnFailure(t *testing.T) {
	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 4242, Status: execution.StatusRunning},
		stopErr:      errors.New("no such process"),
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	session, err := svc.Start(context.Background(), app.ID)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := svc.Stop(context.Background(), app.ID, session); err == nil {
		t.Fatal("expected Stop to propagate the strategy's error")
	}

	if _, ok := svc.ActiveSession(app.ID); !ok {
		t.Fatal("expected ActiveSession to still report true after a failed Stop")
	}
}

func TestExecutionServiceRefreshSessionUpdatesActiveSession(t *testing.T) {
	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 4242, Status: execution.StatusRunning},
		statusSession: execution.Session{
			PID:    4242,
			Status: execution.StatusStopped,
		},
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	session, err := svc.Start(context.Background(), app.ID)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if _, err := svc.RefreshSession(context.Background(), app.ID, session); err != nil {
		t.Fatalf("RefreshSession() error = %v", err)
	}

	refreshed, ok := svc.ActiveSession(app.ID)
	if !ok {
		t.Fatal("expected ActiveSession to still report true after RefreshSession")
	}
	if refreshed.Status != string(execution.StatusStopped) {
		t.Errorf("Status = %q, want %q", refreshed.Status, execution.StatusStopped)
	}
}

func TestExecutionServiceStartPersistsSession(t *testing.T) {
	sessionRepo := newSessionRepo(t)
	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 4242, Status: execution.StatusRunning},
	}
	svc, app := newExecutionServiceWithHostAndSessions(t, strategy, runtimehosttest.NewFakeHost(), sessionRepo)

	if _, err := svc.Start(context.Background(), app.ID); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	persisted, err := sessionRepo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if persisted[app.ID].PID != 4242 {
		t.Fatalf("persisted session PID = %d, want 4242", persisted[app.ID].PID)
	}
}

func TestExecutionServiceStopRemovesPersistedSession(t *testing.T) {
	sessionRepo := newSessionRepo(t)
	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 4242, Status: execution.StatusRunning},
	}
	svc, app := newExecutionServiceWithHostAndSessions(t, strategy, runtimehosttest.NewFakeHost(), sessionRepo)

	session, err := svc.Start(context.Background(), app.ID)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := svc.Stop(context.Background(), app.ID, session); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	persisted, err := sessionRepo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if _, present := persisted[app.ID]; present {
		t.Fatal("expected the persisted session to be removed after Stop")
	}
}

func TestNewExecutionServiceHydratesAStillRunningSessionFromDisk(t *testing.T) {
	ctx := context.Background()
	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
	resolver := execution.NewResolver(execution.NewRegistry())

	sessionRepo := newSessionRepo(t)
	if err := sessionRepo.Save(ctx, "app-1", execution.Session{PID: 4242, Status: execution.StatusRunning}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	host := runtimehosttest.NewFakeHost().WithRunningPID(4242, true)
	svc := services.NewExecutionService(repo, resolver, monitor.New(host), sessionRepo)

	session, ok := svc.ActiveSession("app-1")
	if !ok {
		t.Fatal("expected the still-running persisted session to be rehydrated")
	}
	if session.PID != 4242 {
		t.Errorf("PID = %d, want 4242", session.PID)
	}
}

func TestNewExecutionServiceDiscardsADeadSessionFromDisk(t *testing.T) {
	ctx := context.Background()
	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
	resolver := execution.NewResolver(execution.NewRegistry())

	sessionRepo := newSessionRepo(t)
	if err := sessionRepo.Save(ctx, "app-1", execution.Session{PID: 4242, Status: execution.StatusRunning}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// No WithRunningPID configured — the FakeHost reports 4242 as not
	// running, simulating a process that exited while IAMXFREE was closed.
	host := runtimehosttest.NewFakeHost()
	svc := services.NewExecutionService(repo, resolver, monitor.New(host), sessionRepo)

	if _, ok := svc.ActiveSession("app-1"); ok {
		t.Fatal("expected a dead session to not be rehydrated")
	}

	persisted, err := sessionRepo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if _, present := persisted["app-1"]; present {
		t.Fatal("expected the dead session to be removed from the repository too")
	}
}

func TestExecutionServiceStartSyncsApplicationStatusToRunning(t *testing.T) {
	ctx := context.Background()
	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	app.Status = models.StatusStopped
	if err := repo.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 4242, Status: execution.StatusRunning},
	}
	registry := execution.NewRegistry()
	registry.Register(strategy)
	resolver := execution.NewResolver(registry)
	svc := services.NewExecutionService(repo, resolver, monitor.New(runtimehosttest.NewFakeHost()), newSessionRepo(t))

	if _, err := svc.Start(ctx, app.ID); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	got, err := repo.FindByID(ctx, app.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != models.StatusRunning {
		t.Fatalf("Status = %q, want %q", got.Status, models.StatusRunning)
	}
}

func TestExecutionServiceStopSyncsApplicationStatusToStopped(t *testing.T) {
	ctx := context.Background()
	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	app.Status = models.StatusRunning
	if err := repo.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	strategy := &fakeStrategy{runtime: models.RuntimeNode}
	registry := execution.NewRegistry()
	registry.Register(strategy)
	resolver := execution.NewResolver(registry)
	svc := services.NewExecutionService(repo, resolver, monitor.New(runtimehosttest.NewFakeHost()), newSessionRepo(t))

	if err := svc.Stop(ctx, app.ID, services.RunSession{PID: 4242}); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	got, err := repo.FindByID(ctx, app.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != models.StatusStopped {
		t.Fatalf("Status = %q, want %q", got.Status, models.StatusStopped)
	}
}

func TestExecutionServiceRefreshSessionSyncsApplicationStatus(t *testing.T) {
	ctx := context.Background()
	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Runtime = models.RuntimeNode
	app.Status = models.StatusRunning
	if err := repo.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	strategy := &fakeStrategy{
		runtime:       models.RuntimeNode,
		statusSession: execution.Session{PID: 4242, Status: execution.StatusStopped},
	}
	registry := execution.NewRegistry()
	registry.Register(strategy)
	resolver := execution.NewResolver(registry)
	svc := services.NewExecutionService(repo, resolver, monitor.New(runtimehosttest.NewFakeHost()), newSessionRepo(t))

	if _, err := svc.RefreshSession(ctx, app.ID, services.RunSession{PID: 4242}); err != nil {
		t.Fatalf("RefreshSession() error = %v", err)
	}

	got, err := repo.FindByID(ctx, app.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != models.StatusStopped {
		t.Fatalf("Status = %q, want %q — RefreshSession must notice the process died and correct the badge", got.Status, models.StatusStopped)
	}
}

func TestNewExecutionServiceHydrationSyncsApplicationStatusToRunning(t *testing.T) {
	ctx := context.Background()
	repo, err := jsonstore.NewApplicationRepository(t.TempDir())
	if err != nil {
		t.Fatalf("NewApplicationRepository() error = %v", err)
	}
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	app.Status = models.StatusStopped
	if err := repo.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	resolver := execution.NewResolver(execution.NewRegistry())

	sessionRepo := newSessionRepo(t)
	if err := sessionRepo.Save(ctx, app.ID, execution.Session{PID: 4242, Status: execution.StatusRunning}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	host := runtimehosttest.NewFakeHost().WithRunningPID(4242, true)
	services.NewExecutionService(repo, resolver, monitor.New(host), sessionRepo)

	got, err := repo.FindByID(ctx, app.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != models.StatusRunning {
		t.Fatalf("Status = %q, want %q — a still-running rehydrated session must correct a stale \"stopped\" badge on startup", got.Status, models.StatusRunning)
	}
}
