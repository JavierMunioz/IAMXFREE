package execution_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestNodeStrategyStartSuccess(t *testing.T) {
	host := healthyFakeHost().WithStartProcess("npm", []string{"start"}, 4242, nil)
	strategy := execution.NewNodeStrategy(host)

	session, err := strategy.Start(context.Background(), healthyNodeApp(), 3000)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if session.PID != 4242 {
		t.Errorf("PID = %d, want 4242", session.PID)
	}
	if session.Command != "npm" || len(session.Args) != 1 || session.Args[0] != "start" {
		t.Errorf("Command/Args = %q/%v, want npm/[start]", session.Command, session.Args)
	}
	if session.WorkingDir != "/srv/apps/my-api" {
		t.Errorf("WorkingDir = %q, want %q", session.WorkingDir, "/srv/apps/my-api")
	}
	if session.Status != execution.StatusRunning {
		t.Errorf("Status = %q, want %q", session.Status, execution.StatusRunning)
	}
	if session.Runtime != models.RuntimeNode {
		t.Errorf("Runtime = %q, want %q", session.Runtime, models.RuntimeNode)
	}
	if session.Port != 3000 {
		t.Errorf("Port = %d, want 3000", session.Port)
	}
	if session.StartedAt.IsZero() {
		t.Error("expected StartedAt to be set")
	}
}

func TestNodeStrategyStartSetsPortEnv(t *testing.T) {
	host := healthyFakeHost().WithStartProcess("npm", []string{"start"}, 4242, nil)
	strategy := execution.NewNodeStrategy(host)

	if _, err := strategy.Start(context.Background(), healthyNodeApp(), 3001); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	env, ok := host.StartedEnv("npm", []string{"start"})
	if !ok {
		t.Fatal("expected StartProcess to have been called")
	}
	if len(env) != 1 || env[0] != "PORT=3001" {
		t.Errorf("Env = %v, want [PORT=3001]", env)
	}
}

func TestNodeStrategyStartRefusesWhenNotReady(t *testing.T) {
	host := healthyFakeHost().WithLookPath("node", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityNotFound})
	strategy := execution.NewNodeStrategy(host)

	_, err := strategy.Start(context.Background(), healthyNodeApp(), 3000)
	if err == nil {
		t.Fatal("expected an error when the application is not ready")
	}
}

func TestNodeStrategyStartPropagatesProcessError(t *testing.T) {
	wantErr := errors.New("exec format error")
	host := healthyFakeHost().WithStartProcess("npm", []string{"start"}, 0, wantErr)
	strategy := execution.NewNodeStrategy(host)

	_, err := strategy.Start(context.Background(), healthyNodeApp(), 3000)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Start() error = %v, want %v", err, wantErr)
	}
}

func TestNodeStrategyStartSplitsMultiWordCommand(t *testing.T) {
	host := healthyFakeHost()
	app := healthyNodeApp()
	app.Config.StartCommand = "node server.js --port 3000"
	host = host.WithStartProcess("node", []string{"server.js", "--port", "3000"}, 99, nil)
	strategy := execution.NewNodeStrategy(host)

	session, err := strategy.Start(context.Background(), app, 3000)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if session.Command != "node" {
		t.Errorf("Command = %q, want %q", session.Command, "node")
	}
	wantArgs := []string{"server.js", "--port", "3000"}
	if len(session.Args) != len(wantArgs) {
		t.Fatalf("Args = %v, want %v", session.Args, wantArgs)
	}
	for i, arg := range wantArgs {
		if session.Args[i] != arg {
			t.Errorf("Args[%d] = %q, want %q", i, session.Args[i], arg)
		}
	}
}

func TestSplitCommandEmptyString(t *testing.T) {
	// Exercised indirectly through Start, but an empty configured command
	// should never reach here since commands_configured would already fail
	// Readiness — this documents that expectation.
	host := runtimehosttest.NewFakeHost().
		WithLookPath("node", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithLookPath("npm", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithDir("/srv/apps/my-api").
		WithFile("/srv/apps/my-api/package.json").
		WithReadFile("/srv/apps/my-api/package.json", []byte(`{"scripts":{"start":"node server.js"}}`), nil)
	strategy := execution.NewNodeStrategy(host)

	app := healthyNodeApp()
	app.Config.StartCommand = ""

	if _, err := strategy.Start(context.Background(), app, 3000); err == nil {
		t.Fatal("expected Start() to fail when no start command is configured")
	}
}
