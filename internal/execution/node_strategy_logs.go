package execution

import (
	"context"
	"sync"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// Logs attaches to the output already being captured for session.PID —
// never starting a new process — and adapts it into the execution-level
// LogEvent shape.
func (s *nodeStrategy) Logs(_ context.Context, _ *models.Application, session Session) (LogStream, error) {
	output, err := s.host.StreamOutput(session.PID)
	if err != nil {
		return nil, err
	}
	return newNodeLogStream(output), nil
}

// nodeLogStream adapts a runtimehost.OutputStream into an
// execution.LogStream, appending a final Error or EOF event once the
// underlying output ends.
type nodeLogStream struct {
	output    runtimehost.OutputStream
	events    chan LogEvent
	closed    chan struct{}
	closeOnce sync.Once
}

var _ LogStream = (*nodeLogStream)(nil)

func newNodeLogStream(output runtimehost.OutputStream) *nodeLogStream {
	s := &nodeLogStream{
		output: output,
		events: make(chan LogEvent),
		closed: make(chan struct{}),
	}
	go s.run()
	return s
}

func (s *nodeLogStream) run() {
	defer close(s.events)

	for chunk := range s.output.Chunks() {
		event := LogEvent{
			Timestamp: chunk.Time,
			Type:      logEventTypeForChunk(chunk.Stream),
			Content:   chunk.Line,
		}
		select {
		case s.events <- event:
		case <-s.closed:
			return
		}
	}

	final := LogEvent{Timestamp: time.Now().UTC(), Type: LogEventEOF, Content: "process exited"}
	if err := s.output.Err(); err != nil {
		final = LogEvent{Timestamp: time.Now().UTC(), Type: LogEventError, Content: err.Error()}
	}
	select {
	case s.events <- final:
	case <-s.closed:
	}
}

func logEventTypeForChunk(kind runtimehost.OutputStreamKind) LogEventType {
	if kind == runtimehost.OutputStderr {
		return LogEventStderr
	}
	return LogEventStdout
}

func (s *nodeLogStream) Events() <-chan LogEvent { return s.events }

func (s *nodeLogStream) Err() error { return s.output.Err() }

func (s *nodeLogStream) Close() error {
	s.closeOnce.Do(func() { close(s.closed) })
	return s.output.Close()
}
