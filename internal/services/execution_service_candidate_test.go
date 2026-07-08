package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestExecutionServiceStartCandidateDoesNotDisturbActiveSession(t *testing.T) {
	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 1000, Status: execution.StatusRunning, Port: 3000},
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if _, err := svc.Start(context.Background(), app.ID); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	strategy.startSession = execution.Session{PID: 2000, Status: execution.StatusRunning, Port: 3001}
	candidate, err := svc.StartCandidate(context.Background(), app.ID, 3001)
	if err != nil {
		t.Fatalf("StartCandidate() error = %v", err)
	}
	if candidate.PID != 2000 || candidate.Port != 3001 {
		t.Errorf("candidate = %+v, want PID=2000 Port=3001", candidate)
	}

	active, ok := svc.ActiveSession(app.ID)
	if !ok {
		t.Fatal("expected the active session to still be tracked")
	}
	if active.PID != 1000 {
		t.Errorf("ActiveSession().PID = %d, want 1000 (untouched by StartCandidate)", active.PID)
	}

	got, ok := svc.CandidateSession(app.ID)
	if !ok {
		t.Fatal("expected CandidateSession to report the started candidate")
	}
	if got.PID != 2000 {
		t.Errorf("CandidateSession().PID = %d, want 2000", got.PID)
	}
}

func TestExecutionServiceCandidateSessionUnknownAppReportsNotFound(t *testing.T) {
	strategy := &fakeStrategy{runtime: models.RuntimeNode}
	svc, _ := newExecutionServiceWithStrategy(t, strategy)

	if _, ok := svc.CandidateSession("never-started"); ok {
		t.Fatal("expected CandidateSession to report false for an app with no candidate")
	}
}

func TestExecutionServicePromoteCandidateReplacesActiveSession(t *testing.T) {
	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 1000, Status: execution.StatusRunning, Port: 3000},
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if _, err := svc.Start(context.Background(), app.ID); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	strategy.startSession = execution.Session{PID: 2000, Status: execution.StatusRunning, Port: 3001}
	if _, err := svc.StartCandidate(context.Background(), app.ID, 3001); err != nil {
		t.Fatalf("StartCandidate() error = %v", err)
	}

	if err := svc.PromoteCandidate(context.Background(), app.ID); err != nil {
		t.Fatalf("PromoteCandidate() error = %v", err)
	}

	active, ok := svc.ActiveSession(app.ID)
	if !ok {
		t.Fatal("expected an active session after promotion")
	}
	if active.PID != 2000 || active.Port != 3001 {
		t.Errorf("ActiveSession() = %+v, want PID=2000 Port=3001 (promoted candidate)", active)
	}

	if _, ok := svc.CandidateSession(app.ID); ok {
		t.Fatal("expected no candidate session tracked after promotion")
	}
}

func TestExecutionServicePromoteCandidateErrorsWithoutACandidate(t *testing.T) {
	strategy := &fakeStrategy{runtime: models.RuntimeNode}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if err := svc.PromoteCandidate(context.Background(), app.ID); err == nil {
		t.Fatal("expected an error when no candidate session is tracked")
	}
}

func TestExecutionServiceStopCandidateStopsAndUntracks(t *testing.T) {
	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 2000, Status: execution.StatusRunning, Port: 3001},
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	candidate, err := svc.StartCandidate(context.Background(), app.ID, 3001)
	if err != nil {
		t.Fatalf("StartCandidate() error = %v", err)
	}

	if err := svc.StopCandidate(context.Background(), app.ID, candidate); err != nil {
		t.Fatalf("StopCandidate() error = %v", err)
	}

	if _, ok := svc.CandidateSession(app.ID); ok {
		t.Fatal("expected no candidate session tracked after StopCandidate")
	}
}

func TestExecutionServiceStopCandidatePropagatesStrategyError(t *testing.T) {
	wantErr := errors.New("kill failed")
	strategy := &fakeStrategy{
		runtime:      models.RuntimeNode,
		startSession: execution.Session{PID: 2000, Status: execution.StatusRunning, Port: 3001},
		stopErr:      wantErr,
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	candidate, err := svc.StartCandidate(context.Background(), app.ID, 3001)
	if err != nil {
		t.Fatalf("StartCandidate() error = %v", err)
	}

	if err := svc.StopCandidate(context.Background(), app.ID, candidate); !errors.Is(err, wantErr) {
		t.Fatalf("StopCandidate() error = %v, want %v", err, wantErr)
	}
	if _, ok := svc.CandidateSession(app.ID); !ok {
		t.Fatal("expected the candidate to remain tracked when stopping it failed")
	}
}

func TestExecutionServiceCheckStatusDoesNotTrackAnything(t *testing.T) {
	strategy := &fakeStrategy{
		runtime:       models.RuntimeNode,
		statusSession: execution.Session{PID: 4242, Status: execution.StatusStopped},
	}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	updated, err := svc.CheckStatus(context.Background(), app.ID, services.RunSession{PID: 4242})
	if err != nil {
		t.Fatalf("CheckStatus() error = %v", err)
	}
	if updated.Status != string(execution.StatusStopped) {
		t.Errorf("Status = %q, want %q", updated.Status, execution.StatusStopped)
	}
	if _, ok := svc.ActiveSession(app.ID); ok {
		t.Fatal("expected CheckStatus to never track an active session")
	}
	if _, ok := svc.CandidateSession(app.ID); ok {
		t.Fatal("expected CheckStatus to never track a candidate session")
	}
}

func TestExecutionServiceStopSessionDoesNotTouchTrackedRegistry(t *testing.T) {
	strategy := &fakeStrategy{runtime: models.RuntimeNode}
	svc, app := newExecutionServiceWithStrategy(t, strategy)

	if err := svc.StopSession(context.Background(), app.ID, services.RunSession{PID: 9999}); err != nil {
		t.Fatalf("StopSession() error = %v", err)
	}
	if strategy.stopErr != nil {
		t.Fatalf("unexpected strategy error: %v", strategy.stopErr)
	}
}
