package services

import (
	"sync"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
)

// logStreamAdapter translates an execution.LogStream into the
// services-level LogStream shape, so ExecutionService's callers never need
// to import internal/execution.
type logStreamAdapter struct {
	stream    execution.LogStream
	events    chan LogEvent
	closed    chan struct{}
	closeOnce sync.Once
}

var _ LogStream = (*logStreamAdapter)(nil)

func newLogStreamAdapter(stream execution.LogStream) *logStreamAdapter {
	a := &logStreamAdapter{
		stream: stream,
		events: make(chan LogEvent),
		closed: make(chan struct{}),
	}
	go a.run()
	return a
}

func (a *logStreamAdapter) run() {
	defer close(a.events)
	for e := range a.stream.Events() {
		event := LogEvent{Timestamp: e.Timestamp, Type: string(e.Type), Content: e.Content}
		select {
		case a.events <- event:
		case <-a.closed:
			return
		}
	}
}

func (a *logStreamAdapter) Events() <-chan LogEvent { return a.events }

func (a *logStreamAdapter) Err() error { return a.stream.Err() }

func (a *logStreamAdapter) Close() error {
	a.closeOnce.Do(func() { close(a.closed) })
	return a.stream.Close()
}
