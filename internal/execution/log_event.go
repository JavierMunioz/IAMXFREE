package execution

import "time"

// LogEventType classifies a LogEvent so a consumer (the TUI, most notably)
// can style or filter it without parsing the content.
type LogEventType string

const (
	// LogEventStdout is a line the application wrote to its standard output.
	LogEventStdout LogEventType = "stdout"

	// LogEventStderr is a line the application wrote to its standard error.
	LogEventStderr LogEventType = "stderr"

	// LogEventSystem is a message IAMXFREE itself produced about the
	// stream (e.g. "capture started"), not something the application wrote.
	LogEventSystem LogEventType = "system"

	// LogEventError reports a problem with the stream itself, not
	// something the application logged.
	LogEventError LogEventType = "error"

	// LogEventEOF marks the natural end of the stream — the underlying
	// process exited.
	LogEventEOF LogEventType = "eof"
)

// LogEvent is one entry in a Strategy's log stream — never bare text: every
// event carries when it happened and what kind it is, so a consumer can
// style or filter it without parsing content.
type LogEvent struct {
	Timestamp time.Time
	Type      LogEventType
	Content   string
}
