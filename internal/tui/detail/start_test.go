package detail

import (
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestStartKeyTriggersStartCmd(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m2, cmd := m.handleKey(keyMsg("s"))
	if cmd == nil {
		t.Fatal("expected pressing s to return a command")
	}
	if m2.status != "Starting…" {
		t.Fatalf("status = %q, want %q", m2.status, "Starting…")
	}
}

func TestStartedMsgUpdatesSession(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	session := services.RunSession{PID: 4242, Status: "running", WorkingDir: "/srv/apps/my-api", Runtime: app.Runtime}
	m, _ = update(t, m, startedMsg{session: session})

	if !m.hasSession {
		t.Fatal("expected hasSession to be true after startedMsg")
	}
	if m.session.PID != 4242 {
		t.Errorf("session.PID = %d, want 4242", m.session.PID)
	}
	if m.status == "" {
		t.Fatal("expected a status message after starting")
	}
}

func TestStartFailedMsgSetsError(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	wantErr := errors.New("application is not ready")
	m, _ = update(t, m, startFailedMsg{err: wantErr})

	if !errors.Is(m.statusErr, wantErr) {
		t.Fatalf("statusErr = %v, want %v", m.statusErr, wantErr)
	}
	if m.hasSession {
		t.Fatal("expected hasSession to remain false after a failed start")
	}
}

func TestStartKeyRefusesWhenSessionAlreadyTracked(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	m2, cmd := m.handleKey(keyMsg("s"))
	if cmd != nil {
		t.Fatal("expected no command when a session is already tracked")
	}
	if m2.status == "" {
		t.Fatal("expected a status message explaining why start was refused")
	}
}

func TestStartCmdReturnsStartedMsgOnSuccess(t *testing.T) {
	app := newTestApp()
	wantSession := services.RunSession{PID: 99, Status: "running"}
	execSvc := &fakeExecutionService{startSession: wantSession}
	m := New(&fakeAppService{app: app}, execSvc, app.ID)

	msg := m.startCmd()()
	got, ok := msg.(startedMsg)
	if !ok {
		t.Fatalf("expected startedMsg, got %T", msg)
	}
	if got.session.PID != 99 {
		t.Errorf("PID = %d, want 99", got.session.PID)
	}
}

func TestStartCmdReturnsStartFailedMsgOnError(t *testing.T) {
	app := newTestApp()
	wantErr := errors.New("not ready")
	execSvc := &fakeExecutionService{startErr: wantErr}
	m := New(&fakeAppService{app: app}, execSvc, app.ID)

	msg := m.startCmd()()
	got, ok := msg.(startFailedMsg)
	if !ok {
		t.Fatalf("expected startFailedMsg, got %T", msg)
	}
	if !errors.Is(got.err, wantErr) {
		t.Errorf("err = %v, want %v", got.err, wantErr)
	}
}
