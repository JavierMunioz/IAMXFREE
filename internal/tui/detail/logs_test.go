package detail

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestLogsKeyRefusesWhenNoSession(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m2, cmd := m.handleKey(keyMsg("l"))
	if cmd != nil {
		t.Fatal("expected no command when there is no active session")
	}
	if m2.status == "" {
		t.Fatal("expected a status message explaining why logs was refused")
	}
}

func TestLogsKeyEmitsOpenLogsMsgWhenSessionTracked(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	_, cmd := m.handleKey(keyMsg("l"))
	if cmd == nil {
		t.Fatal("expected pressing l to return a command")
	}
	msg, ok := cmd().(OpenLogsMsg)
	if !ok {
		t.Fatalf("expected OpenLogsMsg, got %T", cmd())
	}
	if msg.AppID != app.ID {
		t.Errorf("AppID = %q, want %q", msg.AppID, app.ID)
	}
	if msg.Session.PID != 4242 {
		t.Errorf("Session.PID = %d, want 4242", msg.Session.PID)
	}
}
