package services

import "time"

// LogEvent is a plain projection of execution.LogEvent so callers never
// need to import internal/execution just to consume a log stream.
type LogEvent struct {
	Timestamp time.Time
	Type      string
	Content   string
}

// LogStream is a live flow of LogEvents opened by
// ExecutionService.OpenLogs. Whatever consumes it (the TUI, today) decides
// how to display it.
type LogStream interface {
	// Events delivers log events as they occur. The channel is closed when
	// the stream ends — either because the underlying process exited or
	// because Close was called.
	Events() <-chan LogEvent

	// Err reports why the stream ended, if abnormally. Only meaningful
	// after Events() is closed.
	Err() error

	// Close stops delivering events on this stream. It never affects the
	// underlying process.
	Close() error
}
