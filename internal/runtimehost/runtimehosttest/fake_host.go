package runtimehosttest

import (
	"context"
	"fmt"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// FakeHost is a deterministic test double for runtimehost.Host. Configure
// canned responses via its With* methods before use; anything not
// configured returns a clear "not found"/empty result instead of touching
// the real operating system, so tests never depend on the machine they run
// on.
type FakeHost struct {
	lookPaths  map[string]runtimehost.ToolAvailability
	versions   map[string]runtimehost.ToolInfo
	runResults map[string]runtimehost.CommandResult
	runErrors  map[string]error
	workingDir string
	files      map[string]bool
	dirs       map[string]bool
}

// NewFakeHost returns an empty FakeHost; every operation reports "not
// found"/empty until configured with a With* method.
func NewFakeHost() *FakeHost {
	return &FakeHost{
		lookPaths:  make(map[string]runtimehost.ToolAvailability),
		versions:   make(map[string]runtimehost.ToolInfo),
		runResults: make(map[string]runtimehost.CommandResult),
		runErrors:  make(map[string]error),
		files:      make(map[string]bool),
		dirs:       make(map[string]bool),
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

func commandKey(name string, args []string) string {
	return strings.TrimSpace(name + " " + strings.Join(args, " "))
}
