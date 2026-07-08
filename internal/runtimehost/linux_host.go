package runtimehost

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// LinuxHost implements Host using the real operating system: os/exec for
// commands, os for filesystem checks. It is the implementation IAMXFREE
// runs with on a Linux VPS.
type LinuxHost struct {
	mu        sync.Mutex
	processes map[int]*os.Process
	outputs   map[int]*outputBuffer
}

// NewLinuxHost returns a Host backed by the real operating system.
func NewLinuxHost() *LinuxHost {
	return &LinuxHost{
		processes: make(map[int]*os.Process),
		outputs:   make(map[int]*outputBuffer),
	}
}

var _ Host = (*LinuxHost)(nil)

func (h *LinuxHost) LookPath(name string) (ToolAvailability, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return ToolAvailability{Name: name, Status: AvailabilityNotFound}, nil
	}
	return ToolAvailability{Name: name, Path: path, Status: AvailabilityFound}, nil
}

func (h *LinuxHost) Version(ctx context.Context, name string, versionArgs []string) (ToolInfo, error) {
	availability, err := h.LookPath(name)
	if err != nil {
		return ToolInfo{}, err
	}
	if !availability.Found() {
		return ToolInfo{Name: name, Available: false}, nil
	}

	info := ToolInfo{Name: name, Path: availability.Path, Available: true}

	result, runErr := h.RunCaptured(ctx, Command{Name: name, Args: versionArgs})
	if runErr != nil {
		info.VersionErr = runErr
		return info, nil
	}

	version := strings.TrimSpace(result.Stdout)
	if version == "" {
		// Some tools (notably older Java) print version info to stderr.
		version = strings.TrimSpace(result.Stderr)
	}
	info.Version = version
	return info, nil
}

func (h *LinuxHost) Run(ctx context.Context, cmd Command) (CommandResult, error) {
	return h.run(ctx, cmd, false)
}

func (h *LinuxHost) RunCaptured(ctx context.Context, cmd Command) (CommandResult, error) {
	return h.run(ctx, cmd, true)
}

func (h *LinuxHost) run(ctx context.Context, cmd Command, capture bool) (CommandResult, error) {
	execCmd := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	execCmd.Dir = cmd.Dir

	var stdout, stderr bytes.Buffer
	if capture {
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr
	} else {
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
	}

	start := time.Now()
	runErr := execCmd.Run()
	duration := time.Since(start)

	// ProcessState is nil when the command never actually started (e.g. the
	// executable could not be found or executed); -1 signals that case.
	exitCode := -1
	if execCmd.ProcessState != nil {
		exitCode = execCmd.ProcessState.ExitCode()
	}

	result := CommandResult{
		Command:  cmd.Name,
		Args:     cmd.Args,
		Dir:      cmd.Dir,
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if runErr != nil {
		return result, &ExecutionError{
			Command:  cmd.Name,
			Args:     cmd.Args,
			ExitCode: exitCode,
			Stderr:   result.Stderr,
			Err:      runErr,
		}
	}

	return result, nil
}

func (h *LinuxHost) WorkingDir() (string, error) {
	return os.Getwd()
}

func (h *LinuxHost) FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

func (h *LinuxHost) DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (h *LinuxHost) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (h *LinuxHost) WriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0o644)
}

func (h *LinuxHost) RemoveFile(path string) error {
	return os.Remove(path)
}

func (h *LinuxHost) ReadDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, nil
}

func (h *LinuxHost) Symlink(target, linkPath string) error {
	return os.Symlink(target, linkPath)
}

// StartProcess starts cmd in the background. It tracks the resulting
// *os.Process internally and reaps it in a goroutine once it exits, so a
// long-running child never becomes a zombie even though nothing calls
// Wait() on it explicitly. It also captures the process's stdout/stderr
// into a bounded per-PID buffer from the moment it starts, so a later
// StreamOutput call can replay everything captured so far.
func (h *LinuxHost) StartProcess(_ context.Context, cmd Command) (int, error) {
	execCmd := exec.Command(cmd.Name, cmd.Args...)
	execCmd.Dir = cmd.Dir

	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return 0, &ExecutionError{Command: cmd.Name, Args: cmd.Args, Err: err}
	}
	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return 0, &ExecutionError{Command: cmd.Name, Args: cmd.Args, Err: err}
	}

	if err := execCmd.Start(); err != nil {
		return 0, &ExecutionError{
			Command: cmd.Name,
			Args:    cmd.Args,
			Err:     err,
		}
	}

	pid := execCmd.Process.Pid
	output := newOutputBuffer(defaultOutputBufferCapacity)

	h.mu.Lock()
	h.processes[pid] = execCmd.Process
	h.outputs[pid] = output
	h.mu.Unlock()

	var scanning sync.WaitGroup
	scanning.Add(2)
	go scanOutputInto(&scanning, stdout, OutputStdout, output)
	go scanOutputInto(&scanning, stderr, OutputStderr, output)

	go func() {
		// StdoutPipe/StderrPipe's docs require every read to complete before
		// Wait is called, so the scanners must finish first.
		scanning.Wait()
		waitErr := execCmd.Wait()
		output.finish(waitErr)

		h.mu.Lock()
		delete(h.processes, pid)
		h.mu.Unlock()
	}()

	return pid, nil
}

// scanOutputInto reads r line by line, appending each line to output as an
// OutputChunk of kind, until r is exhausted (the process exited and closed
// its pipes).
func scanOutputInto(wg *sync.WaitGroup, r io.Reader, kind OutputStreamKind, output *outputBuffer) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		output.append(OutputChunk{Stream: kind, Line: scanner.Text(), Time: time.Now()})
	}
}

// IsProcessRunning first checks whether this Host itself is still tracking
// pid (started it and hasn't observed it exit); if not, it falls back to a
// best-effort OS-level liveness check, so a PID from a previous IAMXFREE
// session can still be queried.
func (h *LinuxHost) IsProcessRunning(pid int) (bool, error) {
	h.mu.Lock()
	_, tracked := h.processes[pid]
	h.mu.Unlock()
	if tracked {
		return true, nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false, nil
	}
	return true, nil
}

func (h *LinuxHost) StopProcess(pid int) error {
	h.mu.Lock()
	process, tracked := h.processes[pid]
	h.mu.Unlock()

	if !tracked {
		var err error
		process, err = os.FindProcess(pid)
		if err != nil {
			return err
		}
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return &ExecutionError{Command: "stop", Err: err}
	}
	return nil
}

func (h *LinuxHost) StreamOutput(pid int) (OutputStream, error) {
	h.mu.Lock()
	output, ok := h.outputs[pid]
	h.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("runtimehost: no output captured for pid %d", pid)
	}
	return output.subscribe(), nil
}
