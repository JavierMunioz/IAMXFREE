package detail

import (
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestRefreshKeyReturnsCommandWithoutSession(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m2, cmd := m.handleKey(keyMsg("f5"))
	if cmd == nil {
		t.Fatal("expected f5 to return a command even without a tracked session")
	}
	if m2.status != "Refreshing…" {
		t.Fatalf("status = %q, want %q", m2.status, "Refreshing…")
	}
}

func TestRefreshCmdOnlyChecksHealthWithoutSession(t *testing.T) {
	app := newTestApp()
	appSvc := &fakeAppService{app: app, health: services.ExecutionHealth{StrategyName: "Node.js (npm)", Healthy: true}}
	m := New(appSvc, &fakeExecutionService{}, app.ID)

	msg := m.refreshCmd()()
	if _, ok := msg.(healthLoadedMsg); !ok {
		t.Fatalf("expected healthLoadedMsg when no session is tracked, got %T", msg)
	}
}

func TestSessionRefreshedMsgUpdatesSession(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242, Status: "running"}

	m, _ = update(t, m, sessionRefreshedMsg{session: services.RunSession{PID: 4242, Status: "stopped"}})
	if m.session.Status != "stopped" {
		t.Fatalf("session.Status = %q, want %q", m.session.Status, "stopped")
	}
	if m.status == "" {
		t.Fatal("expected a status message after refreshing")
	}
}

func TestSessionRefreshFailedMsgSetsError(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	wantErr := errors.New("lookup failed")
	m, _ = update(t, m, sessionRefreshFailedMsg{err: wantErr})
	if !errors.Is(m.statusErr, wantErr) {
		t.Fatalf("statusErr = %v, want %v", m.statusErr, wantErr)
	}
}

func TestRefreshSessionCmdReturnsRefreshedMsgOnSuccess(t *testing.T) {
	app := newTestApp()
	execSvc := &fakeExecutionService{refreshed: services.RunSession{PID: 4242, Status: "stopped"}}
	m := New(&fakeAppService{app: app}, execSvc, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242, Status: "running"}

	msg := m.refreshSessionCmd()()
	got, ok := msg.(sessionRefreshedMsg)
	if !ok {
		t.Fatalf("expected sessionRefreshedMsg, got %T", msg)
	}
	if got.session.Status != "stopped" {
		t.Errorf("Status = %q, want %q", got.session.Status, "stopped")
	}
}

func TestRefreshSessionCmdReturnsFailedMsgOnError(t *testing.T) {
	app := newTestApp()
	wantErr := errors.New("lookup failed")
	execSvc := &fakeExecutionService{refreshErr: wantErr}
	m := New(&fakeAppService{app: app}, execSvc, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	msg := m.refreshSessionCmd()()
	got, ok := msg.(sessionRefreshFailedMsg)
	if !ok {
		t.Fatalf("expected sessionRefreshFailedMsg, got %T", msg)
	}
	if !errors.Is(got.err, wantErr) {
		t.Errorf("err = %v, want %v", got.err, wantErr)
	}
}
