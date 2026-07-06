package runtimehost

import (
	"fmt"
	"strings"
	"time"
)

// Command describes what to run: an executable name (resolved via PATH
// unless it's already an absolute/relative path), its arguments, and
// optionally the working directory to run it in. An empty Dir means the
// Host's own current working directory.
type Command struct {
	Name string
	Args []string
	Dir  string
}

// CommandResult is what running a Command produced. Stdout/Stderr are only
// populated when the command was run through Host.RunCaptured; Host.Run
// leaves them empty since output is inherited by the parent process instead
// of captured.
type CommandResult struct {
	Command  string
	Args     []string
	Dir      string
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Availability states whether a tool could be located on PATH. "Not found"
// is a normal, expected outcome — it is never reported as an error.
type Availability string

const (
	AvailabilityFound    Availability = "found"
	AvailabilityNotFound Availability = "not_found"
)

// ToolAvailability is the result of looking a tool up on PATH.
type ToolAvailability struct {
	Name   string
	Path   string
	Status Availability
}

// Found reports whether the tool was located.
func (a ToolAvailability) Found() bool {
	return a.Status == AvailabilityFound
}

// ToolInfo describes an installed tool: where it lives and what version it
// reports. Version is empty when it could not be determined; VersionErr
// explains why without failing the whole call — a tool that exists but
// whose version command failed is still a "found" tool, just one whose
// version is unknown.
type ToolInfo struct {
	Name       string
	Path       string
	Available  bool
	Version    string
	VersionErr error
}

// ExecutionError is returned when a Command did not succeed. It carries
// enough structured detail — command, args, exit code, captured stderr —
// that a caller can decide what happened without re-parsing a generic error
// string.
type ExecutionError struct {
	Command  string
	Args     []string
	ExitCode int
	Stderr   string
	Err      error
}

func (e *ExecutionError) Error() string {
	invocation := strings.TrimSpace(e.Command + " " + strings.Join(e.Args, " "))
	if stderr := strings.TrimSpace(e.Stderr); stderr != "" {
		return fmt.Sprintf("runtimehost: %s: exit %d: %s", invocation, e.ExitCode, stderr)
	}
	return fmt.Sprintf("runtimehost: %s: exit %d: %v", invocation, e.ExitCode, e.Err)
}

func (e *ExecutionError) Unwrap() error { return e.Err }
