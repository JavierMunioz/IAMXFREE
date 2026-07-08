package nginx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestReloadSuccess(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("nginx", []string{"-s", "reload"}, runtimehost.CommandResult{ExitCode: 0}, nil)
	manager := nginx.NewManager(host)

	if err := manager.Reload(context.Background()); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}
}

func TestReloadFailure(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithRunResult("nginx", []string{"-s", "reload"},
			runtimehost.CommandResult{ExitCode: 1, Stderr: "nginx: [error] invalid PID"},
			&runtimehost.ExecutionError{Command: "nginx", ExitCode: 1, Err: errors.New("exit status 1")},
		)
	manager := nginx.NewManager(host)

	if err := manager.Reload(context.Background()); err == nil {
		t.Fatal("expected an error for a failed reload")
	}
}
