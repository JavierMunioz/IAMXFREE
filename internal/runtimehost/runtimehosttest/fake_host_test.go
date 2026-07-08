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

func TestFakeHostWriteFile(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	if err := host.WriteFile("/etc/nginx/sites-available/example.conf", []byte("server {}")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	content, ok := host.WrittenFile("/etc/nginx/sites-available/example.conf")
	if !ok {
		t.Fatal("expected WrittenFile to report the file as written")
	}
	if string(content) != "server {}" {
		t.Fatalf("WrittenFile() = %q, want %q", content, "server {}")
	}
	if ok, _ := host.FileExists("/etc/nginx/sites-available/example.conf"); !ok {
		t.Error("expected a written file to also report as existing")
	}
}

func TestFakeHostWriteFileConfiguredError(t *testing.T) {
	wantErr := errors.New("disk full")
	host := runtimehosttest.NewFakeHost().
		WithWriteFileError("/etc/nginx/sites-available/example.conf", wantErr)

	err := host.WriteFile("/etc/nginx/sites-available/example.conf", []byte("server {}"))
	if !errors.Is(err, wantErr) {
		t.Fatalf("WriteFile() error = %v, want %v", err, wantErr)
	}
	if _, ok := host.WrittenFile("/etc/nginx/sites-available/example.conf"); ok {
		t.Fatal("expected a failed write to not be recorded")
	}
}

func TestFakeHostRemoveFile(t *testing.T) {
	host := runtimehosttest.NewFakeHost().WithFile("/etc/nginx/sites-available/example.conf")

	if err := host.RemoveFile("/etc/nginx/sites-available/example.conf"); err != nil {
		t.Fatalf("RemoveFile() error = %v", err)
	}
	if !host.Removed("/etc/nginx/sites-available/example.conf") {
		t.Fatal("expected Removed to report true after RemoveFile")
	}
	if ok, _ := host.FileExists("/etc/nginx/sites-available/example.conf"); ok {
		t.Error("expected a removed file to no longer exist")
	}
}

func TestFakeHostRemoveFileConfiguredError(t *testing.T) {
	wantErr := errors.New("permission denied")
	host := runtimehosttest.NewFakeHost().
		WithFile("/etc/nginx/sites-available/example.conf").
		WithRemoveFileError("/etc/nginx/sites-available/example.conf", wantErr)

	err := host.RemoveFile("/etc/nginx/sites-available/example.conf")
	if !errors.Is(err, wantErr) {
		t.Fatalf("RemoveFile() error = %v, want %v", err, wantErr)
	}
	if ok, _ := host.FileExists("/etc/nginx/sites-available/example.conf"); !ok {
		t.Error("expected a failed removal to leave the file in place")
	}
}

func TestFakeHostReadDir(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithReadDir("/etc/nginx/sites-available", []string{"example.com.conf", "api.example.com.conf"}, nil)

	entries, err := host.ReadDir("/etc/nginx/sites-available")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 2 || entries[0] != "example.com.conf" || entries[1] != "api.example.com.conf" {
		t.Fatalf("ReadDir() = %v, want [example.com.conf api.example.com.conf]", entries)
	}
}

func TestFakeHostReadDirUnconfiguredReturnsEmpty(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	entries, err := host.ReadDir("/etc/nginx/sites-available")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("ReadDir() = %v, want empty", entries)
	}
}

func TestFakeHostSymlink(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	err := host.Symlink(
		"/etc/nginx/sites-available/example.com.conf",
		"/etc/nginx/sites-enabled/example.com.conf",
	)
	if err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	target, ok := host.SymlinkTarget("/etc/nginx/sites-enabled/example.com.conf")
	if !ok || target != "/etc/nginx/sites-available/example.com.conf" {
		t.Fatalf("SymlinkTarget() = (%q, %v), want (%q, true)", target, ok, "/etc/nginx/sites-available/example.com.conf")
	}
	if ok, _ := host.FileExists("/etc/nginx/sites-enabled/example.com.conf"); !ok {
		t.Error("expected the symlink path to report as existing")
	}
}

func TestFakeHostSymlinkAlreadyExists(t *testing.T) {
	host := runtimehosttest.NewFakeHost()
	_ = host.Symlink("/etc/nginx/sites-available/example.com.conf", "/etc/nginx/sites-enabled/example.com.conf")

	err := host.Symlink("/etc/nginx/sites-available/example.com.conf", "/etc/nginx/sites-enabled/example.com.conf")
	if err == nil {
		t.Fatal("expected Symlink to fail when linkPath already exists")
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

func TestFakeHostStartProcessRecordsEnv(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithStartProcess("npm", []string{"start"}, 4242, nil)

	_, err := host.StartProcess(context.Background(), runtimehost.Command{
		Name: "npm", Args: []string{"start"}, Env: []string{"PORT=3001"},
	})
	if err != nil {
		t.Fatalf("StartProcess() error = %v", err)
	}

	env, ok := host.StartedEnv("npm", []string{"start"})
	if !ok {
		t.Fatal("expected StartedEnv to report the command was started")
	}
	if len(env) != 1 || env[0] != "PORT=3001" {
		t.Errorf("StartedEnv() = %v, want [PORT=3001]", env)
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

func TestFakeHostStreamOutputReplaysConfiguredChunks(t *testing.T) {
	want := []runtimehost.OutputChunk{
		{Stream: runtimehost.OutputStdout, Line: "listening on :3000"},
		{Stream: runtimehost.OutputStderr, Line: "warning: deprecated flag"},
	}
	host := runtimehosttest.NewFakeHost().WithOutputStream(4242, want, nil)

	stream, err := host.StreamOutput(4242)
	if err != nil {
		t.Fatalf("StreamOutput() error = %v", err)
	}

	var got []runtimehost.OutputChunk
	for c := range stream.Chunks() {
		got = append(got, c)
	}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Stream != want[i].Stream || got[i].Line != want[i].Line {
			t.Errorf("got[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
	if stream.Err() != nil {
		t.Fatalf("Err() = %v, want nil", stream.Err())
	}
}

func TestFakeHostStreamOutputConfiguredError(t *testing.T) {
	wantErr := errors.New("process crashed")
	host := runtimehosttest.NewFakeHost().WithOutputStream(4242, nil, wantErr)

	stream, err := host.StreamOutput(4242)
	if err != nil {
		t.Fatalf("StreamOutput() error = %v", err)
	}
	for range stream.Chunks() {
	}
	if !errors.Is(stream.Err(), wantErr) {
		t.Fatalf("Err() = %v, want %v", stream.Err(), wantErr)
	}
}

func TestFakeHostStreamOutputUnconfiguredReturnsError(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	if _, err := host.StreamOutput(4242); err == nil {
		t.Fatal("expected an error for an unconfigured pid")
	}
}

func TestFakeHostProcessResourcesConfigured(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithProcessResources(4242, runtimehost.ProcessResources{
			CPUPercent:      12.5,
			CPUPercentKnown: true,
			MemoryRSSBytes:  1024 * 1024,
			MemoryRSSKnown:  true,
		})

	resources, err := host.ProcessResources(4242)
	if err != nil {
		t.Fatalf("ProcessResources() error = %v", err)
	}
	if !resources.CPUPercentKnown || resources.CPUPercent != 12.5 {
		t.Errorf("CPUPercent = %v, known=%v, want 12.5, true", resources.CPUPercent, resources.CPUPercentKnown)
	}
	if !resources.MemoryRSSKnown || resources.MemoryRSSBytes != 1024*1024 {
		t.Errorf("MemoryRSSBytes = %d, known=%v, want %d, true", resources.MemoryRSSBytes, resources.MemoryRSSKnown, 1024*1024)
	}
	if resources.MemoryVSZKnown {
		t.Error("expected MemoryVSZKnown to be false since it was never configured")
	}
}

func TestFakeHostProcessResourcesUnconfiguredReportsUnavailable(t *testing.T) {
	host := runtimehosttest.NewFakeHost()

	resources, err := host.ProcessResources(4242)
	if err != nil {
		t.Fatalf("ProcessResources() error = %v, want nil", err)
	}
	if resources.CPUPercentKnown || resources.MemoryRSSKnown || resources.MemoryVSZKnown {
		t.Fatalf("expected every metric to be unavailable, got %+v", resources)
	}
}
