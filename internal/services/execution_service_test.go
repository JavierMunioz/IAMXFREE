package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories/jsonstore"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func newExecutionServiceWithStrategy(t *testing.T, strategy execution.Strategy) (services.ExecutionService, *models.Application) {
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

	return services.NewExecutionService(repo, resolver), app
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
