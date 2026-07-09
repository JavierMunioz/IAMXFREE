package detail

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

func TestRefreshCmdOnlyChecksHealthAndGitWithoutSession(t *testing.T) {
	app := newTestApp()
	appSvc := &fakeAppService{app: app, health: services.ExecutionHealth{StrategyName: "Node.js (npm)", Healthy: true}}
	m := New(appSvc, &fakeExecutionService{}, app.ID)

	batch, ok := m.refreshCmd()().(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected a tea.BatchMsg, got %T", m.refreshCmd()())
	}

	var sawHealth, sawGit, sawCandidate bool
	for _, cmd := range batch {
		switch cmd().(type) {
		case healthLoadedMsg:
			sawHealth = true
		case gitStatusLoadedMsg, gitStatusUnavailableMsg:
			sawGit = true
		case candidateSessionLoadedMsg:
			sawCandidate = true
		}
	}
	if !sawHealth {
		t.Error("expected the batch to include a command producing healthLoadedMsg")
	}
	if !sawGit {
		t.Error("expected the batch to include a command producing a Git status message")
	}
	if !sawCandidate {
		t.Error("expected the batch to include a command producing candidateSessionLoadedMsg")
	}
	if len(batch) != 3 {
		t.Fatalf("expected exactly 3 commands without a tracked session, got %d", len(batch))
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

func TestRefreshCmdAlsoFetchesSnapshotWithSession(t *testing.T) {
	app := newTestApp()
	execSvc := &fakeExecutionService{snapshot: services.RuntimeSnapshot{PID: 4242, State: "running"}}
	m := New(&fakeAppService{app: app}, execSvc, app.ID)
	m.hasSession = true
	m.session = services.RunSession{PID: 4242}

	batch := m.refreshCmd()
	if batch == nil {
		t.Fatal("expected refreshCmd to return a command")
	}

	msg := batch()
	batchMsg, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected a tea.BatchMsg, got %T", msg)
	}

	var sawSnapshot bool
	for _, cmd := range batchMsg {
		if _, ok := cmd().(snapshotLoadedMsg); ok {
			sawSnapshot = true
		}
	}
	if !sawSnapshot {
		t.Fatal("expected refreshCmd's batch to include a snapshotLoadedMsg-producing command")
	}
}

func TestSnapshotCmdReturnsSnapshotLoadedMsgOnSuccess(t *testing.T) {
	app := newTestApp()
	want := services.RuntimeSnapshot{PID: 4242, State: "running", CPUPercent: services.Metric{Value: 5, Available: true}}
	execSvc := &fakeExecutionService{snapshot: want}
	m := New(&fakeAppService{app: app}, execSvc, app.ID)
	m.session = services.RunSession{PID: 4242}

	msg := m.snapshotCmd()()
	got, ok := msg.(snapshotLoadedMsg)
	if !ok {
		t.Fatalf("expected snapshotLoadedMsg, got %T", msg)
	}
	if got.snapshot.PID != 4242 || !got.snapshot.CPUPercent.Available {
		t.Errorf("snapshot = %+v, unexpected", got.snapshot)
	}
}

func TestSnapshotCmdReturnsFailedMsgOnError(t *testing.T) {
	app := newTestApp()
	wantErr := errors.New("lookup failed")
	execSvc := &fakeExecutionService{snapshotErr: wantErr}
	m := New(&fakeAppService{app: app}, execSvc, app.ID)
	m.session = services.RunSession{PID: 4242}

	msg := m.snapshotCmd()()
	got, ok := msg.(snapshotFailedMsg)
	if !ok {
		t.Fatalf("expected snapshotFailedMsg, got %T", msg)
	}
	if !errors.Is(got.err, wantErr) {
		t.Errorf("err = %v, want %v", got.err, wantErr)
	}
}

func TestSnapshotLoadedMsgUpdatesModel(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)

	m, _ = update(t, m, snapshotLoadedMsg{snapshot: services.RuntimeSnapshot{PID: 4242}})
	if !m.hasSnapshot || m.snapshot.PID != 4242 {
		t.Fatalf("expected the snapshot to be recorded, got hasSnapshot=%v snapshot=%+v", m.hasSnapshot, m.snapshot)
	}
}

func TestStartedMsgClearsPreviousSnapshot(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSnapshot = true
	m.snapshot = services.RuntimeSnapshot{PID: 1}

	m, _ = update(t, m, startedMsg{session: services.RunSession{PID: 4242, Status: "running"}})
	if m.hasSnapshot {
		t.Fatal("expected a fresh start to clear any previously observed snapshot")
	}
}

func TestStoppedMsgClearsSnapshot(t *testing.T) {
	app := newTestApp()
	m := New(&fakeAppService{app: app}, &fakeExecutionService{}, app.ID)
	m.hasSnapshot = true
	m.snapshot = services.RuntimeSnapshot{PID: 4242}

	m, _ = update(t, m, stoppedMsg{})
	if m.hasSnapshot {
		t.Fatal("expected stopping the session to clear its snapshot")
	}
}
