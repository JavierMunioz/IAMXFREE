package runtimehost

import "context"

// Host is the only way anything in IAMXFREE talks to the operating system.
// It exposes the common operations execution strategies will need,
// including starting and stopping a long-running process — but not yet
// supervising it (no automatic restarts; that is a concern for a later
// iteration).
type Host interface {
	// LookPath reports whether name can be found as an executable on PATH.
	// "Not found" is reported through the returned ToolAvailability, not as
	// an error.
	LookPath(name string) (ToolAvailability, error)

	// Version runs name with versionArgs (e.g. []string{"--version"}) and
	// returns structured information about it. If name is not on PATH,
	// ToolInfo.Available is false and no command runs. If the tool exists
	// but running it fails, that failure is captured in ToolInfo.VersionErr
	// rather than returned as err.
	Version(ctx context.Context, name string, versionArgs []string) (ToolInfo, error)

	// Run executes cmd synchronously, inheriting stdio — its output is not
	// captured, so the returned CommandResult's Stdout/Stderr stay empty.
	Run(ctx context.Context, cmd Command) (CommandResult, error)

	// RunCaptured executes cmd synchronously and captures its stdout and
	// stderr into the returned CommandResult.
	RunCaptured(ctx context.Context, cmd Command) (CommandResult, error)

	// WorkingDir returns the host's current working directory.
	WorkingDir() (string, error)

	// FileExists reports whether path exists and is a regular file.
	FileExists(path string) (bool, error)

	// DirExists reports whether path exists and is a directory.
	DirExists(path string) (bool, error)

	// ReadFile reads and returns the contents of the file at path. Callers
	// needing to inspect a project's files (e.g. an execution strategy
	// checking package.json) use this instead of touching the filesystem
	// directly.
	ReadFile(path string) ([]byte, error)

	// StartProcess starts cmd as a background process and returns its PID
	// immediately, without waiting for it to exit. Unlike Run/RunCaptured,
	// this is how a long-running process (e.g. a web server) is started
	// and left running.
	StartProcess(ctx context.Context, cmd Command) (int, error)

	// IsProcessRunning reports whether pid refers to a currently running
	// process.
	IsProcessRunning(pid int) (bool, error)

	// StopProcess asks pid to terminate gracefully.
	StopProcess(pid int) error

	// StreamOutput returns the captured stdout/stderr for a process
	// previously started with StartProcess. Capture begins at StartProcess
	// time — this never starts a new process, only attaches to output
	// already being captured; backlog is bounded, so very old output may
	// already be gone. Fails if pid was never started by this Host.
	StreamOutput(pid int) (OutputStream, error)

	// ProcessResources queries the operating system for pid's current
	// resource usage: CPU time and memory footprint. It never fabricates a
	// value — any metric that could not be determined (e.g. running on a
	// platform without /proc, or pid no longer existing) is reported as
	// unavailable through ProcessResources' Known flags rather than a
	// fabricated zero. This is a read-only observation; it never affects
	// pid.
	ProcessResources(pid int) (ProcessResources, error)
}
