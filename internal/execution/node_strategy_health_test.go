package execution_test

import (
	"context"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
)

func healthyNodeApp() *models.Application {
	app := models.NewApplication("my-api", models.ApplicationTypeBackend)
	app.Runtime = models.RuntimeNode
	app.Source.LocalPath = "/srv/apps/my-api"
	app.Config.PackageManager = "npm"
	app.Config.StartCommand = "npm start"
	return app
}

func healthyFakeHost() *runtimehosttest.FakeHost {
	return runtimehosttest.NewFakeHost().
		WithLookPath("node", runtimehost.ToolAvailability{Name: "node", Path: "/usr/bin/node", Status: runtimehost.AvailabilityFound}).
		WithLookPath("npm", runtimehost.ToolAvailability{Name: "npm", Path: "/usr/bin/npm", Status: runtimehost.AvailabilityFound}).
		WithDir("/srv/apps/my-api").
		WithFile("/srv/apps/my-api/package.json").
		WithReadFile("/srv/apps/my-api/package.json", []byte(`{"scripts":{"start":"node server.js"}}`), nil)
}

func findItem(t *testing.T, health execution.HealthCheck, name execution.HealthCheckName) execution.HealthCheckItem {
	t.Helper()
	for _, item := range health.Items {
		if item.Name == name {
			return item
		}
	}
	t.Fatalf("no health check item named %q", name)
	return execution.HealthCheckItem{}
}

func TestNodeStrategyHealthCheckAllPass(t *testing.T) {
	strategy := execution.NewNodeStrategy(healthyFakeHost())

	health, err := strategy.HealthCheck(context.Background(), healthyNodeApp())
	if err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}
	if !health.Healthy() {
		t.Fatalf("expected a healthy result, got failures: %+v", health.Failed())
	}
	if len(health.Items) != 7 {
		t.Fatalf("len(Items) = %d, want 7", len(health.Items))
	}
}

func TestNodeStrategyHealthCheckNodeMissing(t *testing.T) {
	host := healthyFakeHost().WithLookPath("node", runtimehost.ToolAvailability{Name: "node", Status: runtimehost.AvailabilityNotFound})
	strategy := execution.NewNodeStrategy(host)

	health, err := strategy.HealthCheck(context.Background(), healthyNodeApp())
	if err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}
	if health.Healthy() {
		t.Fatal("expected an unhealthy result when node is missing")
	}
	if findItem(t, health, execution.HealthCheckRuntimeInstalled).Status != execution.HealthStatusFail {
		t.Fatal("expected runtime_installed to fail")
	}
}

func TestNodeStrategyHealthCheckNpmMissing(t *testing.T) {
	host := healthyFakeHost().WithLookPath("npm", runtimehost.ToolAvailability{Name: "npm", Status: runtimehost.AvailabilityNotFound})
	strategy := execution.NewNodeStrategy(host)

	health, _ := strategy.HealthCheck(context.Background(), healthyNodeApp())
	if findItem(t, health, execution.HealthCheckPackageManagerInstalled).Status != execution.HealthStatusFail {
		t.Fatal("expected package_manager_installed to fail")
	}
}

func TestNodeStrategyHealthCheckPathEmpty(t *testing.T) {
	strategy := execution.NewNodeStrategy(healthyFakeHost())
	app := healthyNodeApp()
	app.Source.LocalPath = ""

	health, _ := strategy.HealthCheck(context.Background(), app)
	if findItem(t, health, execution.HealthCheckPathValid).Status != execution.HealthStatusFail {
		t.Fatal("expected path_valid to fail for an empty path")
	}
	if findItem(t, health, execution.HealthCheckDirectoryAccessible).Status != execution.HealthStatusFail {
		t.Fatal("expected directory_accessible to also fail for an empty path")
	}
}

func TestNodeStrategyHealthCheckDirectoryMissing(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithLookPath("node", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithLookPath("npm", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound})
	strategy := execution.NewNodeStrategy(host)

	health, _ := strategy.HealthCheck(context.Background(), healthyNodeApp())
	if findItem(t, health, execution.HealthCheckDirectoryAccessible).Status != execution.HealthStatusFail {
		t.Fatal("expected directory_accessible to fail when the directory was never configured on the fake host")
	}
}

func TestNodeStrategyHealthCheckManifestMissing(t *testing.T) {
	// A fake host with node/npm/dir configured, but no package.json file.
	host := runtimehosttest.NewFakeHost().
		WithLookPath("node", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithLookPath("npm", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithDir("/srv/apps/my-api")
	strategy := execution.NewNodeStrategy(host)

	health, _ := strategy.HealthCheck(context.Background(), healthyNodeApp())
	if findItem(t, health, execution.HealthCheckManifestExists).Status != execution.HealthStatusFail {
		t.Fatal("expected manifest_exists to fail when package.json is missing")
	}
	if findItem(t, health, execution.HealthCheckScriptsAvailable).Status != execution.HealthStatusFail {
		t.Fatal("expected scripts_available to also fail when package.json cannot be read")
	}
}

func TestNodeStrategyHealthCheckNoScripts(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithLookPath("node", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithLookPath("npm", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithDir("/srv/apps/my-api").
		WithFile("/srv/apps/my-api/package.json").
		WithReadFile("/srv/apps/my-api/package.json", []byte(`{"name":"my-api"}`), nil)
	strategy := execution.NewNodeStrategy(host)

	health, _ := strategy.HealthCheck(context.Background(), healthyNodeApp())
	if findItem(t, health, execution.HealthCheckScriptsAvailable).Status != execution.HealthStatusFail {
		t.Fatal("expected scripts_available to fail when package.json declares no scripts")
	}
}

func TestNodeStrategyHealthCheckMalformedPackageJSON(t *testing.T) {
	host := runtimehosttest.NewFakeHost().
		WithLookPath("node", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithLookPath("npm", runtimehost.ToolAvailability{Status: runtimehost.AvailabilityFound}).
		WithDir("/srv/apps/my-api").
		WithFile("/srv/apps/my-api/package.json").
		WithReadFile("/srv/apps/my-api/package.json", []byte(`{not valid json`), nil)
	strategy := execution.NewNodeStrategy(host)

	health, _ := strategy.HealthCheck(context.Background(), healthyNodeApp())
	if findItem(t, health, execution.HealthCheckScriptsAvailable).Status != execution.HealthStatusFail {
		t.Fatal("expected scripts_available to fail for malformed package.json")
	}
}

func TestNodeStrategyHealthCheckNoStartCommand(t *testing.T) {
	strategy := execution.NewNodeStrategy(healthyFakeHost())
	app := healthyNodeApp()
	app.Config.StartCommand = ""

	health, _ := strategy.HealthCheck(context.Background(), app)
	if findItem(t, health, execution.HealthCheckCommandsConfigured).Status != execution.HealthStatusFail {
		t.Fatal("expected commands_configured to fail when no start command is set")
	}
}
