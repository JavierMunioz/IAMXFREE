package runtimehosttest

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// FakeHost is a deterministic test double for runtimehost.Host. Configure
// canned responses via its With* methods before use; anything not
// configured returns a clear "not found"/empty result instead of touching
// the real operating system, so tests never depend on the machine they run
// on.
type FakeHost struct {
	lookPaths    map[string]runtimehost.ToolAvailability
	versions     map[string]runtimehost.ToolInfo
	runResults   map[string]runtimehost.CommandResult
	runErrors    map[string]error
	workingDir   string
	files        map[string]bool
	dirs         map[string]bool
	fileContents map[string][]byte
	fileErrors   map[string]error

	startProcessResults map[string]startProcessResult
	runningPIDs         map[int]bool
	stopErrors          map[int]error
	stoppedPIDs         map[int]bool

	outputStreams map[int]outputStreamResult
}

type outputStreamResult struct {
	chunks []runtimehost.OutputChunk
	err    error
}

type startProcessResult struct {
	pid int
	err error
}

// NewFakeHost returns an empty FakeHost; every operation reports "not
// found"/empty until configured with a With* method.
func NewFakeHost() *FakeHost {
	return &FakeHost{
		lookPaths:           make(map[string]runtimehost.ToolAvailability),
		versions:            make(map[string]runtimehost.ToolInfo),
		runResults:          make(map[string]runtimehost.CommandResult),
		runErrors:           make(map[string]error),
		files:               make(map[string]bool),
		dirs:                make(map[string]bool),
		fileContents:        make(map[string][]byte),
		fileErrors:          make(map[string]error),
		startProcessResults: make(map[string]startProcessResult),
		runningPIDs:         make(map[int]bool),
		stopErrors:          make(map[int]error),
		stoppedPIDs:         make(map[int]bool),
		outputStreams:       make(map[int]outputStreamResult),
	}
}

var _ runtimehost.Host = (*FakeHost)(nil)

// WithLookPath makes LookPath(name) return availability.
func (f *FakeHost) WithLookPath(name string, availability runtimehost.ToolAvailability) *FakeHost {
	f.lookPaths[name] = availability
	return f
}

// WithVersion makes Version(_, name, _) return info.
func (f *FakeHost) WithVersion(name string, info runtimehost.ToolInfo) *FakeHost {
	f.versions[name] = info
	return f
}

// WithRunResult makes Run/RunCaptured return result and err for a command
// matching name and args exactly.
func (f *FakeHost) WithRunResult(name string, args []string, result runtimehost.CommandResult, err error) *FakeHost {
	key := commandKey(name, args)
	f.runResults[key] = result
	f.runErrors[key] = err
	return f
}

// WithWorkingDir makes WorkingDir() return dir.
func (f *FakeHost) WithWorkingDir(dir string) *FakeHost {
	f.workingDir = dir
	return f
}

// WithFile makes FileExists(path) return true.
func (f *FakeHost) WithFile(path string) *FakeHost {
	f.files[path] = true
	return f
}

// WithDir makes DirExists(path) return true.
func (f *FakeHost) WithDir(path string) *FakeHost {
	f.dirs[path] = true
	return f
}

// WithReadFile makes ReadFile(path) return content and err.
func (f *FakeHost) WithReadFile(path string, content []byte, err error) *FakeHost {
	f.fileContents[path] = content
	f.fileErrors[path] = err
	return f
}

// WithStartProcess makes StartProcess return pid and err for a command
// matching name and args exactly. A nil err also marks pid as running, so
// a subsequent IsProcessRunning(pid) reports true until WithRunningPID or
// StopProcess says otherwise.
func (f *FakeHost) WithStartProcess(name string, args []string, pid int, err error) *FakeHost {
	f.startProcessResults[commandKey(name, args)] = startProcessResult{pid: pid, err: err}
	if err == nil {
		f.runningPIDs[pid] = true
	}
	return f
}

// WithRunningPID makes IsProcessRunning(pid) report running.
func (f *FakeHost) WithRunningPID(pid int, running bool) *FakeHost {
	f.runningPIDs[pid] = running
	return f
}

// WithStopError makes StopProcess(pid) return err instead of succeeding.
func (f *FakeHost) WithStopError(pid int, err error) *FakeHost {
	f.stopErrors[pid] = err
	return f
}

// Stopped reports whether StopProcess(pid) was called, regardless of
// whether it succeeded — useful for asserting a strategy attempted to stop
// the right process.
func (f *FakeHost) Stopped(pid int) bool {
	return f.stoppedPIDs[pid]
}

// WithOutputStream makes StreamOutput(pid) return a stream that replays
// chunks in order and then ends, reporting err (nil for a clean end).
func (f *FakeHost) WithOutputStream(pid int, chunks []runtimehost.OutputChunk, err error) *FakeHost {
	f.outputStreams[pid] = outputStreamResult{chunks: chunks, err: err}
	return f
}

func (f *FakeHost) LookPath(name string) (runtimehost.ToolAvailability, error) {
	if avail, ok := f.lookPaths[name]; ok {
		return avail, nil
	}
	return runtimehost.ToolAvailability{Name: name, Status: runtimehost.AvailabilityNotFound}, nil
}

func (f *FakeHost) Version(_ context.Context, name string, _ []string) (runtimehost.ToolInfo, error) {
	if info, ok := f.versions[name]; ok {
		return info, nil
	}
	return runtimehost.ToolInfo{Name: name, Available: false}, nil
}

func (f *FakeHost) Run(_ context.Context, cmd runtimehost.Command) (runtimehost.CommandResult, error) {
	return f.lookupResult(cmd)
}

func (f *FakeHost) RunCaptured(_ context.Context, cmd runtimehost.Command) (runtimehost.CommandResult, error) {
	return f.lookupResult(cmd)
}

func (f *FakeHost) lookupResult(cmd runtimehost.Command) (runtimehost.CommandResult, error) {
	key := commandKey(cmd.Name, cmd.Args)
	result, ok := f.runResults[key]
	if !ok {
		return runtimehost.CommandResult{}, fmt.Errorf("runtimehosttest: no result configured for %q", key)
	}
	return result, f.runErrors[key]
}

func (f *FakeHost) WorkingDir() (string, error) {
	if f.workingDir != "" {
		return f.workingDir, nil
	}
	return "/", nil
}

func (f *FakeHost) FileExists(path string) (bool, error) {
	return f.files[path], nil
}

func (f *FakeHost) DirExists(path string) (bool, error) {
	return f.dirs[path], nil
}

func (f *FakeHost) ReadFile(path string) ([]byte, error) {
	if err, ok := f.fileErrors[path]; ok && err != nil {
		return nil, err
	}
	if content, ok := f.fileContents[path]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}

func (f *FakeHost) StartProcess(_ context.Context, cmd runtimehost.Command) (int, error) {
	key := commandKey(cmd.Name, cmd.Args)
	result, ok := f.startProcessResults[key]
	if !ok {
		return 0, fmt.Errorf("runtimehosttest: no start-process result configured for %q", key)
	}
	return result.pid, result.err
}

func (f *FakeHost) IsProcessRunning(pid int) (bool, error) {
	return f.runningPIDs[pid], nil
}

func (f *FakeHost) StopProcess(pid int) error {
	f.stoppedPIDs[pid] = true
	if err, ok := f.stopErrors[pid]; ok && err != nil {
		return err
	}
	f.runningPIDs[pid] = false
	return nil
}

func (f *FakeHost) StreamOutput(pid int) (runtimehost.OutputStream, error) {
	result, ok := f.outputStreams[pid]
	if !ok {
		return nil, fmt.Errorf("runtimehosttest: no output stream configured for pid %d", pid)
	}
	return newFakeOutputStream(result.chunks, result.err), nil
}

// fakeOutputStream replays a fixed, pre-buffered slice of chunks and then
// ends — deterministic and dependency-free, unlike LinuxHost's live capture.
type fakeOutputStream struct {
	ch  chan runtimehost.OutputChunk
	err error
}

func newFakeOutputStream(chunks []runtimehost.OutputChunk, err error) *fakeOutputStream {
	ch := make(chan runtimehost.OutputChunk, len(chunks))
	for _, c := range chunks {
		ch <- c
	}
	close(ch)
	return &fakeOutputStream{ch: ch, err: err}
}

func (s *fakeOutputStream) Chunks() <-chan runtimehost.OutputChunk { return s.ch }
func (s *fakeOutputStream) Err() error                             { return s.err }
func (s *fakeOutputStream) Close() error                           { return nil }

func commandKey(name string, args []string) string {
	return strings.TrimSpace(name + " " + strings.Join(args, " "))
}
