package execution_test

import (
	"context"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func TestNodeStrategyMetadata(t *testing.T) {
	strategy := execution.NewNodeStrategy(runtimehosttest.NewFakeHost())

	meta := strategy.Metadata()
	if meta.Name != "Node.js (npm)" {
		t.Errorf("Name = %q, want %q", meta.Name, "Node.js (npm)")
	}
	if len(meta.SupportedRuntimes) != 1 || meta.SupportedRuntimes[0] != models.RuntimeNode {
		t.Errorf("SupportedRuntimes = %v, want [node]", meta.SupportedRuntimes)
	}
	wantCapabilities := map[execution.Capability]bool{execution.CapabilityStart: true, execution.CapabilityStop: true}
	if len(meta.Capabilities) != len(wantCapabilities) {
		t.Fatalf("Capabilities = %v, want start and stop only", meta.Capabilities)
	}
	for _, c := range meta.Capabilities {
		if !wantCapabilities[c] {
			t.Errorf("unexpected capability %q", c)
		}
	}
}

func TestNodeStrategyCanHandle(t *testing.T) {
	strategy := execution.NewNodeStrategy(runtimehosttest.NewFakeHost())

	nodeNpmApp := models.NewApplication("my-api", models.ApplicationTypeAPI)
	nodeNpmApp.Runtime = models.RuntimeNode
	nodeNpmApp.Config.PackageManager = "npm"
	if !strategy.CanHandle(nodeNpmApp) {
		t.Error("expected CanHandle to be true for node+npm")
	}

	nodePnpmApp := models.NewApplication("my-api", models.ApplicationTypeAPI)
	nodePnpmApp.Runtime = models.RuntimeNode
	nodePnpmApp.Config.PackageManager = "pnpm"
	if strategy.CanHandle(nodePnpmApp) {
		t.Error("expected CanHandle to be false for node+pnpm (a different strategy's job)")
	}

	nodeUnknownPMApp := models.NewApplication("my-api", models.ApplicationTypeAPI)
	nodeUnknownPMApp.Runtime = models.RuntimeNode
	if strategy.CanHandle(nodeUnknownPMApp) {
		t.Error("expected CanHandle to be false when no package manager is configured (never assume npm)")
	}

	pythonApp := models.NewApplication("my-api", models.ApplicationTypeAPI)
	pythonApp.Runtime = models.RuntimePython
	pythonApp.Config.PackageManager = "npm" // nonsensical, but should still not match
	if strategy.CanHandle(pythonApp) {
		t.Error("expected CanHandle to be false for a non-Node runtime")
	}
}

func TestNodeStrategyUnimplementedMethodsReportNotImplemented(t *testing.T) {
	strategy := execution.NewNodeStrategy(runtimehosttest.NewFakeHost())
	app := models.NewApplication("my-api", models.ApplicationTypeAPI)
	ctx := context.Background()

	if err := strategy.Install(ctx, app); err != execution.ErrNotImplemented {
		t.Errorf("Install() error = %v, want ErrNotImplemented", err)
	}
	if err := strategy.Build(ctx, app); err != execution.ErrNotImplemented {
		t.Errorf("Build() error = %v, want ErrNotImplemented", err)
	}
	if err := strategy.Restart(ctx, app); err != execution.ErrNotImplemented {
		t.Errorf("Restart() error = %v, want ErrNotImplemented", err)
	}
	if err := strategy.Update(ctx, app); err != execution.ErrNotImplemented {
		t.Errorf("Update() error = %v, want ErrNotImplemented", err)
	}
}
