package runtimehost_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// These tests exercise LinuxHost against POSIX-guaranteed utilities (sh,
// echo, true, false) rather than any tool a developer might or might not
// have installed, so they stay deterministic across machines.

func TestLinuxHostLookPathFound(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	availability, err := host.LookPath("sh")
	if err != nil {
		t.Fatalf("LookPath() error = %v", err)
	}
	if !availability.Found() {
		t.Fatal("expected sh to be found on PATH")
	}
	if availability.Path == "" {
		t.Fatal("expected a resolved path")
	}
}

func TestLinuxHostLookPathNotFound(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	availability, err := host.LookPath("definitely-not-a-real-command-xyz123")
	if err != nil {
		t.Fatalf("LookPath() error = %v", err)
	}
	if availability.Found() {
		t.Fatal("expected the tool to not be found")
	}
}

func TestLinuxHostRunCapturedSuccess(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	result, err := host.RunCaptured(context.Background(), runtimehost.Command{
		Name: "sh",
		Args: []string{"-c", "echo hello"},
	})
	if err != nil {
		t.Fatalf("RunCaptured() error = %v", err)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "hello\n")
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestLinuxHostRunCapturedFailure(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	_, err := host.RunCaptured(context.Background(), runtimehost.Command{
		Name: "sh",
		Args: []string{"-c", "echo oops 1>&2; exit 3"},
	})
	if err == nil {
		t.Fatal("expected an error for a nonzero exit code")
	}

	var execErr *runtimehost.ExecutionError
	if !errors.As(err, &execErr) {
		t.Fatalf("expected an *ExecutionError, got %T", err)
	}
	if execErr.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", execErr.ExitCode)
	}
	if !strings.Contains(execErr.Stderr, "oops") {
		t.Errorf("Stderr = %q, want it to contain %q", execErr.Stderr, "oops")
	}
}

func TestLinuxHostRunDoesNotCaptureOutput(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	result, err := host.Run(context.Background(), runtimehost.Command{
		Name: "sh",
		Args: []string{"-c", "echo hello"},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Stdout != "" {
		t.Errorf("Stdout = %q, want empty (Run does not capture output)", result.Stdout)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestLinuxHostRunRespectsDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "marker.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	host := runtimehost.NewLinuxHost()
	result, err := host.RunCaptured(context.Background(), runtimehost.Command{
		Name: "ls",
		Dir:  dir,
	})
	if err != nil {
		t.Fatalf("RunCaptured() error = %v", err)
	}
	if !strings.Contains(result.Stdout, "marker.txt") {
		t.Errorf("Stdout = %q, want it to list marker.txt from Dir", result.Stdout)
	}
}

func TestLinuxHostRunRespectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	host := runtimehost.NewLinuxHost()
	_, err := host.RunCaptured(ctx, runtimehost.Command{Name: "sh", Args: []string{"-c", "echo hi"}})
	if err == nil {
		t.Fatal("expected an error when the context is already cancelled")
	}
}

func TestLinuxHostVersion(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	info, err := host.Version(context.Background(), "sh", []string{"-c", "echo v1.2.3"})
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	if !info.Available {
		t.Fatal("expected Available to be true")
	}
	if info.Version != "v1.2.3" {
		t.Errorf("Version = %q, want %q", info.Version, "v1.2.3")
	}
	if info.VersionErr != nil {
		t.Errorf("VersionErr = %v, want nil", info.VersionErr)
	}
}

func TestLinuxHostVersionToolNotFound(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	info, err := host.Version(context.Background(), "definitely-not-a-real-command-xyz123", []string{"--version"})
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	if info.Available {
		t.Fatal("expected Available to be false")
	}
	if info.Version != "" {
		t.Errorf("Version = %q, want empty", info.Version)
	}
}

func TestLinuxHostVersionCommandFails(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	info, err := host.Version(context.Background(), "sh", []string{"-c", "exit 1"})
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	if !info.Available {
		t.Fatal("expected Available to be true (sh does exist)")
	}
	if info.VersionErr == nil {
		t.Fatal("expected VersionErr to be set")
	}
	if info.Version != "" {
		t.Errorf("Version = %q, want empty", info.Version)
	}
}

func TestLinuxHostWorkingDir(t *testing.T) {
	host := runtimehost.NewLinuxHost()

	got, err := host.WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() error = %v", err)
	}
	want, _ := os.Getwd()
	if got != want {
		t.Errorf("WorkingDir() = %q, want %q", got, want)
	}
}

func TestLinuxHostFileExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	host := runtimehost.NewLinuxHost()

	if ok, err := host.FileExists(file); err != nil || !ok {
		t.Errorf("FileExists(file) = %v, %v, want true, nil", ok, err)
	}
	if ok, err := host.FileExists(dir); err != nil || ok {
		t.Errorf("FileExists(dir) = %v, %v, want false, nil (it's a directory)", ok, err)
	}
	if ok, err := host.FileExists(filepath.Join(dir, "missing.txt")); err != nil || ok {
		t.Errorf("FileExists(missing) = %v, %v, want false, nil", ok, err)
	}
}

func TestLinuxHostDirExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	host := runtimehost.NewLinuxHost()

	if ok, err := host.DirExists(dir); err != nil || !ok {
		t.Errorf("DirExists(dir) = %v, %v, want true, nil", ok, err)
	}
	if ok, err := host.DirExists(file); err != nil || ok {
		t.Errorf("DirExists(file) = %v, %v, want false, nil (it's a file)", ok, err)
	}
	if ok, err := host.DirExists(filepath.Join(dir, "missing")); err != nil || ok {
		t.Errorf("DirExists(missing) = %v, %v, want false, nil", ok, err)
	}
}

func TestLinuxHostReadFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "package.json")
	if err := os.WriteFile(file, []byte(`{"name":"my-api"}`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	host := runtimehost.NewLinuxHost()
	data, err := host.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != `{"name":"my-api"}` {
		t.Errorf("ReadFile() = %q, want %q", data, `{"name":"my-api"}`)
	}
}

func TestLinuxHostReadFileMissing(t *testing.T) {
	host := runtimehost.NewLinuxHost()
	if _, err := host.ReadFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected an error for a missing file")
	}
}
