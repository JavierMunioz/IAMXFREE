package logs

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// fakeExecutionService is a minimal services.ExecutionService test double.
type fakeExecutionService struct {
	logStream services.LogStream
	logsErr   error
}

func (f *fakeExecutionService) Install(context.Context, string) error { return nil }
func (f *fakeExecutionService) Build(context.Context, string) error   { return nil }
func (f *fakeExecutionService) Start(context.Context, string) (services.RunSession, error) {
	return services.RunSession{}, nil
}
func (f *fakeExecutionService) Stop(context.Context, string, services.RunSession) error { return nil }
func (f *fakeExecutionService) RefreshSession(context.Context, string, services.RunSession) (services.RunSession, error) {
	return services.RunSession{}, nil
}
func (f *fakeExecutionService) OpenLogs(context.Context, string, services.RunSession) (services.LogStream, error) {
	return f.logStream, f.logsErr
}
func (f *fakeExecutionService) Snapshot(context.Context, services.RunSession) (services.RuntimeSnapshot, error) {
	return services.RuntimeSnapshot{}, nil
}
func (f *fakeExecutionService) ActiveSession(string) (services.RunSession, bool) {
	return services.RunSession{}, false
}
func (f *fakeExecutionService) StartCandidate(context.Context, string, int) (services.RunSession, error) {
	return services.RunSession{}, nil
}
func (f *fakeExecutionService) CandidateSession(string) (services.RunSession, bool) {
	return services.RunSession{}, false
}
func (f *fakeExecutionService) StopCandidate(context.Context, string, services.RunSession) error {
	return nil
}
func (f *fakeExecutionService) PromoteCandidate(context.Context, string) error { return nil }
func (f *fakeExecutionService) CheckStatus(context.Context, string, services.RunSession) (services.RunSession, error) {
	return services.RunSession{}, nil
}
func (f *fakeExecutionService) StopSession(context.Context, string, services.RunSession) error {
	return nil
}

// fakeLogStream is a minimal services.LogStream test double, replaying a
// fixed slice of events and then closing.
type fakeLogStream struct {
	ch     chan services.LogEvent
	err    error
	closed bool
}

func newFakeLogStream(events []services.LogEvent, err error) *fakeLogStream {
	ch := make(chan services.LogEvent, len(events))
	for _, e := range events {
		ch <- e
	}
	close(ch)
	return &fakeLogStream{ch: ch, err: err}
}

func (s *fakeLogStream) Events() <-chan services.LogEvent { return s.ch }
func (s *fakeLogStream) Err() error                       { return s.err }
func (s *fakeLogStream) Close() error {
	s.closed = true
	return nil
}

func update(t *testing.T, m Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(msg)
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update() returned %T, want Model", updated)
	}
	return next, cmd
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestInitReturnsOpenLogsCommand(t *testing.T) {
	m := New(&fakeExecutionService{}, "app-1", services.RunSession{PID: 4242})
	if cmd := m.Init(); cmd == nil {
		t.Fatal("expected Init to return a command")
	}
}

func TestLogsOpenedTriggersReadLoop(t *testing.T) {
	stream := newFakeLogStream([]services.LogEvent{{Type: "stdout", Content: "hello"}}, nil)
	m := New(&fakeExecutionService{}, "app-1", services.RunSession{})

	m, cmd := update(t, m, logsOpenedMsg{stream: stream})
	if m.stream == nil {
		t.Fatal("expected the stream to be stored")
	}
	if cmd == nil {
		t.Fatal("expected a command to read the first event")
	}

	msg := cmd()
	event, ok := msg.(logEventMsg)
	if !ok {
		t.Fatalf("expected logEventMsg, got %T", msg)
	}
	if event.event.Content != "hello" {
		t.Errorf("event.Content = %q, want %q", event.event.Content, "hello")
	}
}

func TestLogEventMsgAppendsAndKeepsReading(t *testing.T) {
	stream := newFakeLogStream(nil, nil)
	m := New(&fakeExecutionService{}, "app-1", services.RunSession{})
	m.stream = stream

	m, cmd := update(t, m, logEventMsg{event: services.LogEvent{Type: "stdout", Content: "line1"}})
	if m.buffer.len() != 1 {
		t.Fatalf("buffer.len() = %d, want 1", m.buffer.len())
	}
	if cmd == nil {
		t.Fatal("expected a command to keep reading the next event")
	}
}

func TestLogStreamEndedMsgSetsFlag(t *testing.T) {
	m := New(&fakeExecutionService{}, "app-1", services.RunSession{})
	m, _ = update(t, m, logStreamEndedMsg{})
	if !m.streamEnded {
		t.Fatal("expected streamEnded to be true")
	}
}

func TestLogsOpenFailedSetsOpenErr(t *testing.T) {
	wantErr := errors.New("no output captured")
	m := New(&fakeExecutionService{}, "app-1", services.RunSession{})
	m, _ = update(t, m, logsOpenFailedMsg{err: wantErr})

	if !errors.Is(m.openErr, wantErr) {
		t.Fatalf("openErr = %v, want %v", m.openErr, wantErr)
	}
	if !strings.Contains(m.View(), "no output captured") {
		t.Fatalf("expected the view to show the open error, got:\n%s", m.View())
	}
}

func TestViewRendersDistinctPrefixesPerEventType(t *testing.T) {
	m := New(&fakeExecutionService{}, "app-1", services.RunSession{})
	for _, e := range []services.LogEvent{
		{Type: "stdout", Content: "out line"},
		{Type: "stderr", Content: "err line"},
		{Type: "system", Content: "sys line"},
		{Type: "error", Content: "err detail"},
		{Type: "eof", Content: "process exited"},
	} {
		m.buffer.append(e)
	}
	m.height = 24

	view := stripANSI(t, m.View())
	for _, want := range []string{"out out line", "err err line", "sys sys line", "ERR err detail", "end process exited"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestBackKeyClosesStreamAndEmitsBackMsg(t *testing.T) {
	stream := newFakeLogStream(nil, nil)
	m := New(&fakeExecutionService{}, "app-1", services.RunSession{})
	m.stream = stream

	m2, cmd := m.handleKey(keyMsg("esc"))
	if !stream.closed {
		t.Fatal("expected esc to close the stream")
	}
	if cmd == nil {
		t.Fatal("expected esc to return a command")
	}
	if _, ok := cmd().(BackMsg); !ok {
		t.Fatalf("expected BackMsg, got %T", cmd())
	}
	_ = m2
}

func TestQuitReturnsCommand(t *testing.T) {
	m := New(&fakeExecutionService{}, "app-1", services.RunSession{})
	_, cmd := m.handleKey(keyMsg("q"))
	if cmd == nil {
		t.Fatal("expected quit to return a command")
	}
}
