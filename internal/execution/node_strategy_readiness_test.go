package execution_test

import (
	"context"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestNodeStrategyReadinessReadyWhenHealthy(t *testing.T) {
	strategy := execution.NewNodeStrategy(healthyFakeHost())

	readiness, err := strategy.Readiness(context.Background(), healthyNodeApp())
	if err != nil {
		t.Fatalf("Readiness() error = %v", err)
	}
	if !readiness.Ready {
		t.Fatalf("expected Ready to be true, got %+v", readiness)
	}
}

func TestNodeStrategyReadinessMissingDependency(t *testing.T) {
	host := healthyFakeHost().WithLookPath("node", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityNotFound})
	strategy := execution.NewNodeStrategy(host)

	readiness, err := strategy.Readiness(context.Background(), healthyNodeApp())
	if err != nil {
		t.Fatalf("Readiness() error = %v", err)
	}
	if readiness.Ready {
		t.Fatal("expected Ready to be false when node is missing")
	}
	if len(readiness.MissingDependencies) != 1 {
		t.Fatalf("MissingDependencies = %v, want exactly one entry", readiness.MissingDependencies)
	}
}

func TestNodeStrategyReadinessNoScriptsIsOnlyAWarning(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithLookPath("node", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithLookPath("npm", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithDir("/srv/apps/my-api").
		WithFile("/srv/apps/my-api/package.json").
		WithReadFile("/srv/apps/my-api/package.json", []byte(`{"name":"my-api"}`), nil)
	strategy := execution.NewNodeStrategy(host)

	readiness, err := strategy.Readiness(context.Background(), healthyNodeApp())
	if err != nil {
		t.Fatalf("Readiness() error = %v", err)
	}
	if !readiness.Ready {
		t.Fatalf("expected Ready to still be true with only a scripts warning, got %+v", readiness)
	}
	if len(readiness.Warnings) != 1 {
		t.Fatalf("Warnings = %v, want exactly one entry", readiness.Warnings)
	}
}
