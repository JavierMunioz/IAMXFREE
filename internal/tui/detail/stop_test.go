package detail

import (
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestStopKeyRefusesWhenNoSession(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m2, cmd := m.handleKey(keyMsg("x"))
	if cmd != nil {
		t.Fatal("expected no command when there is no active session")
	}
	if m2.status == "" {
		t.Fatal("expected a status message explaining why stop was refused")
	}
}

func TestStopKeyTriggersStopCmdWhenSessionTracked(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	m2, cmd := m.handleKey(keyMsg("x"))
	if cmd == nil {
		t.Fatal("expected pressing x to return a command")
	}
	if m2.status != "Stopping…" {
		t.Fatalf("status = %q, want %q", m2.status, "Stopping…")
	}
}

func TestStoppedMsgClearsSession(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	m, _ = update(t, m, stoppedMsg{})
	if m.hasSession {
		t.Fatal("expected hasSession to be false after stoppedMsg")
	}
	if m.status == "" {
		t.Fatal("expected a status message after stopping")
	}
}

func TestStopFailedMsgKeepsSessionTracked(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	wantErr := errors.New("no such process")
	m, _ = update(t, m, stopFailedMsg{err: wantErr})

	if !errors.Is(m.statusErr, wantErr) {
		t.Fatalf("statusErr = %v, want %v", m.statusErr, wantErr)
	}
	if !m.hasSession {
		t.Fatal("expected the session to remain tracked after a failed stop")
	}
}

func TestStopCmdReturnsStoppedMsgOnSuccess(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	if _, ok := m.stopCmd()().(stoppedMsg); !ok {
		t.Fatalf("expected stoppedMsg, got %T", m.stopCmd()())
	}
}

func TestStopCmdReturnsStopFailedMsgOnError(t *testing.T) {
	app := newTestApp()
	wantErr := errors.New("no such process")
	execSvc := &fakeExecutionService{stopErr: wantErr}
	m := New(&fakeAppService{app: app}, execSvc, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	msg := m.stopCmd()()
	got, ok := msg.(stopFailedMsg)
	if !ok {
		t.Fatalf("expected stopFailedMsg, got %T", msg)
	}
	if !errors.Is(got.err, wantErr) {
		t.Errorf("err = %v, want %v", got.err, wantErr)
	}
}
