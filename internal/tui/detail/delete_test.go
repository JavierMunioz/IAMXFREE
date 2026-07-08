package detail

import (
	"errors"
	"testing"
)

func TestDeleteKeyArmsConfirmation(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m2, cmd := m.handleKey(keyMsg("d"))
	if cmd != nil {
		t.Fatal("expected the first d press to only arm confirmation, no command yet")
	}
	if !m2.confirmingDelete {
		t.Fatal("expected confirmingDelete to be true after the first d press")
	}
	if m2.status == "" {
		t.Fatal("expected a status message explaining the confirmation is pending")
	}
}

func TestDeleteKeyConfirmedTriggersDeleteCmd(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.app = app
	m.confirmingDelete = true

	m2, cmd := m.handleKey(keyMsg("d"))
	if cmd == nil {
		t.Fatal("expected the second d press to return a command")
	}
	if m2.confirmingDelete {
		t.Fatal("expected confirmingDelete to be cleared once confirmed")
	}
	if m2.status != "Deleting…" {
		t.Fatalf("status = %q, want %q", m2.status, "Deleting…")
	}
}

func TestAnyOtherKeyCancelsDeleteConfirmation(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.confirmingDelete = true

	m2, cmd := m.handleKey(keyMsg("x"))
	if cmd != nil {
		t.Fatal("expected cancelling the confirmation to return no command")
	}
	if m2.confirmingDelete {
		t.Fatal("expected confirmingDelete to be cleared after cancelling")
	}
	if m2.status != "Delete cancelled." {
		t.Fatalf("status = %q, want %q", m2.status, "Delete cancelled.")
	}
}

func TestDeleteCmdReturnsDeletedMsgOnSuccess(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.app = app

	msg, ok := m.deleteCmd()().(DeletedMsg)
	if !ok {
		t.Fatalf("expected DeletedMsg, got %T", m.deleteCmd()())
	}
	if msg.AppID != app.ID || msg.Name != app.Name {
		t.Fatalf("msg = %+v, want AppID=%q Name=%q", msg, app.ID, app.Name)
	}
}

func TestDeleteCmdReturnsDeleteFailedMsgOnError(t *testing.T) {
	app := newTestApp()
	wantErr := errors.New("database is locked")
	m := New(&fakeAppService{app: app, removeErr: wantErr}, &fakeExecutionService{}, app.ID)
	m.app = app

	msg := m.deleteCmd()()
	got, ok := msg.(deleteFailedMsg)
	if !ok {
		t.Fatalf("expected deleteFailedMsg, got %T", msg)
	}
	if !errors.Is(got.err, wantErr) {
		t.Errorf("err = %v, want %v", got.err, wantErr)
	}
}

func TestDeleteFailedMsgSetsError(t *testing.T) {
	app := newTestApp()
	wantErr := errors.New("database is locked")
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m, _ = update(t, m, deleteFailedMsg{err: wantErr})
	if !errors.Is(m.statusErr, wantErr) {
		t.Fatalf("statusErr = %v, want %v", m.statusErr, wantErr)
	}
}
