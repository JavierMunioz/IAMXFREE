package runtimehost

import "time"

// OutputStreamKind identifies which of a process's output streams a chunk
// came from.
type OutputStreamKind string

const (
	OutputStdout OutputStreamKind = "stdout"
	OutputStderr OutputStreamKind = "stderr"
)

// OutputChunk is one line of captured process output.
type OutputChunk struct {
	Stream OutputStreamKind
	Line   string
	Time   time.Time
}

// OutputStream delivers captured output chunks for a process previously
// started with StartProcess: first any buffered backlog (bounded — old
// output is evicted, never stored indefinitely), then live chunks as they
// arrive, until the process exits or the stream is closed.
type OutputStream interface {
	// Chunks returns the channel chunks are delivered on. It is closed when
	// the stream ends — either because the underlying process exited or
	// because Close was called.
	Chunks() <-chan OutputChunk

	// Err reports why the stream ended, if abnormally. Only meaningful
	// after Chunks() is closed.
	Err() error

	// Close stops delivering chunks to this stream. It never affects the
	// underlying process.
	Close() error
}
