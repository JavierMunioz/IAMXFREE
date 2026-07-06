package runtimehost

import "context"

// Host is the only way anything in IAMXFREE talks to the operating system.
// It exposes the common operations execution strategies will need — none of
// them run a persistent process; that is a concern for a later iteration.
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
}
