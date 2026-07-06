package runtimehosttest_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestFakeHostLookPathDefaultsToNotFound(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	availability, err := host.LookPath("npm")
	if err != nil {
		t.Fatalf("LookPath() error = %v", err)
	}
	if availability.Found() {
		t.Fatal("expected an unconfigured tool to report not found")
	}
}

func TestFakeHostLookPathConfigured(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithLookPath("npm", runtimehost.ToolAvailability{Name: "npm", Path: "/usr/bin/npm", Status: runtimehost.AvailabilityFound})

	availability, err := host.LookPath("npm")
	if err != nil {
		t.Fatalf("LookPath() error = %v", err)
	}
	if !availability.Found() || availability.Path != "/usr/bin/npm" {
		t.Fatalf("LookPath() = %+v, want a found tool at /usr/bin/npm", availability)
	}
}

func TestFakeHostVersionDefaultsToUnavailable(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	info, err := host.Version(context.Background(), "npm", []string{"--version"})
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	if info.Available {
		t.Fatal("expected an unconfigured tool to be unavailable")
	}
}

func TestFakeHostVersionConfigured(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithVersion("npm", runtimehost.ToolInfo{Name: "npm", Available: true, Version: "10.2.0"})

	info, err := host.Version(context.Background(), "npm", []string{"--version"})
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	if !info.Available || info.Version != "10.2.0" {
		t.Fatalf("Version() = %+v, want version 10.2.0", info)
	}
}

func TestFakeHostRunUnconfiguredReturnsError(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	_, err := host.Run(context.Background(), runtimehost.Command{Name: "npm", Args: []string{"install"}})
	if err == nil {
		t.Fatal("expected an error for an unconfigured command")
	}
}

func TestFakeHostRunConfigured(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("npm", []string{"install"}, runtimehost.CommandResult{ExitCode: 0, Stdout: "added 42 packages"}, nil)

	result, err := host.RunCaptured(context.Background(), runtimehost.Command{Name: "npm", Args: []string{"install"}})
	if err != nil {
		t.Fatalf("RunCaptured() error = %v", err)
	}
	if result.Stdout != "added 42 packages" {
		t.Fatalf("Stdout = %q, want %q", result.Stdout, "added 42 packages")
	}
}

func TestFakeHostRunConfiguredError(t *testing.T) {
	wantErr := errors.New("boom")
	host := runtimehosttest.NewFakeHost().
		WithRunResult("npm", []string{"install"}, runtimehost.CommandResult{ExitCode: 1}, wantErr)

	_, err := host.Run(context.Background(), runtimehost.Command{Name: "npm", Args: []string{"install"}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}

func TestFakeHostWorkingDirDefault(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	dir, err := host.WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() error = %v", err)
	}
	if dir != "/" {
		t.Fatalf("WorkingDir() = %q, want %q", dir, "/")
	}
}

func TestFakeHostWorkingDirConfigured(t *testing.T) {
	host := runtimehosttest.NewFakeHost().WithWorkingDir("/srv/apps/my-api")

	dir, err := host.WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() error = %v", err)
	}
	if dir != "/srv/apps/my-api" {
		t.Fatalf("WorkingDir() = %q, want %q", dir, "/srv/apps/my-api")
	}
}

func TestFakeHostFileAndDirExists(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithFile("/srv/apps/my-api/package.json").
		WithDir("/srv/apps/my-api")

	if ok, _ := host.FileExists("/srv/apps/my-api/package.json"); !ok {
		t.Error("expected the configured file to exist")
	}
	if ok, _ := host.FileExists("/srv/apps/my-api/missing.json"); ok {
		t.Error("expected an unconfigured file to not exist")
	}
	if ok, _ := host.DirExists("/srv/apps/my-api"); !ok {
		t.Error("expected the configured directory to exist")
	}
	if ok, _ := host.DirExists("/srv/apps/other"); ok {
		t.Error("expected an unconfigured directory to not exist")
	}
}

func TestFakeHostSatisfiesHostInterface(t *testing.T) {
	var _ runtimehost.Host = runtimehosttest.NewFakeHost()
}

func TestFakeHostReadFileConfigured(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithReadFile("/srv/apps/my-api/package.json", []byte(`{"name":"my-api"}`), nil)

	data, err := host.ReadFile("/srv/apps/my-api/package.json")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != `{"name":"my-api"}` {
		t.Fatalf("ReadFile() = %q, want %q", data, `{"name":"my-api"}`)
	}
}

func TestFakeHostReadFileUnconfiguredReturnsNotExist(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	if _, err := host.ReadFile("/srv/apps/my-api/package.json"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ReadFile() error = %v, want os.ErrNotExist", err)
	}
}

func TestFakeHostReadFileConfiguredError(t *testing.T) {
	wantErr := errors.New("permission denied")
	host := runtimehosttest.NewFakeHost().
		WithReadFile("/srv/apps/my-api/package.json", nil, wantErr)

	if _, err := host.ReadFile("/srv/apps/my-api/package.json"); !errors.Is(err, wantErr) {
		t.Fatalf("ReadFile() error = %v, want %v", err, wantErr)
	}
}

func TestFakeHostStartProcessConfigured(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithStartProcess("npm", []string{"start"}, 4242, nil)

	pid, err := host.StartProcess(context.Background(), runtimehost.Command{Name: "npm", Args: []string{"start"}})
	if err != nil {
		t.Fatalf("StartProcess() error = %v", err)
	}
	if pid != 4242 {
		t.Fatalf("StartProcess() pid = %d, want 4242", pid)
	}

	running, err := host.IsProcessRunning(4242)
	if err != nil {
		t.Fatalf("IsProcessRunning() error = %v", err)
	}
	if !running {
		t.Fatal("expected a successfully started process to report running")
	}
}

func TestFakeHostStartProcessUnconfiguredReturnsError(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	if _, err := host.StartProcess(context.Background(), runtimehost.Command{Name: "npm", Args: []string{"start"}}); err == nil {
		t.Fatal("expected an error for an unconfigured command")
	}
}

func TestFakeHostStartProcessConfiguredError(t *testing.T) {
	wantErr := errors.New("boom")
	host := runtimehosttest.NewFakeHost().
		WithStartProcess("npm", []string{"start"}, 0, wantErr)

	if _, err := host.StartProcess(context.Background(), runtimehost.Command{Name: "npm", Args: []string{"start"}}); !errors.Is(err, wantErr) {
		t.Fatalf("StartProcess() error = %v, want %v", err, wantErr)
	}
}

func TestFakeHostIsProcessRunningDefaultsToFalse(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	running, err := host.IsProcessRunning(1234)
	if err != nil {
		t.Fatalf("IsProcessRunning() error = %v", err)
	}
	if running {
		t.Fatal("expected an unconfigured PID to report not running")
	}
}

func TestFakeHostStopProcessMarksNotRunning(t *testing.T) {
	host := runtimehosttest.NewFakeHost().WithRunningPID(4242, true)

	if err := host.StopProcess(4242); err != nil {
		t.Fatalf("StopProcess() error = %v", err)
	}

	running, _ := host.IsProcessRunning(4242)
	if running {
		t.Fatal("expected the process to no longer be running after StopProcess")
	}
	if !host.Stopped(4242) {
		t.Fatal("expected Stopped(4242) to be true")
	}
}

func TestFakeHostStopProcessConfiguredError(t *testing.T) {
	wantErr := errors.New("no such process")
	host := runtimehosttest.NewFakeHost().
		WithRunningPID(4242, true).
		WithStopError(4242, wantErr)

	if err := host.StopProcess(4242); !errors.Is(err, wantErr) {
		t.Fatalf("StopProcess() error = %v, want %v", err, wantErr)
	}
	if !host.Stopped(4242) {
		t.Fatal("expected Stopped(4242) to be true even when the stop failed")
	}
}
