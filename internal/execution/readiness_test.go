package execution_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
)

func TestDeriveReadinessAllPassing(t *testing.T) {
	health := execution.HealthCheck{Items: []execution.HealthCheckItem{
		{Name: execution.HealthCheckRuntimeInstalled, Status: execution.HealthStatusPass},
		{Name: execution.HealthCheckPathValid, Status: execution.HealthStatusPass},
	}}

	readiness := execution.DeriveReadiness(health)
	if !readiness.Ready {
		t.Fatal("expected Ready to be true when everything passes")
	}
	if len(readiness.MissingDependencies) != 0 || len(readiness.BlockingErrors) != 0 || len(readiness.Warnings) != 0 {
		t.Fatalf("expected no dependencies/errors/warnings, got %+v", readiness)
	}
}

func TestDeriveReadinessMissingDependency(t *testing.T) {
	health := execution.HealthCheck{Items: []execution.HealthCheckItem{
		{Name: execution.HealthCheckRuntimeInstalled, Status: execution.HealthStatusFail, Detail: "node not found"},
	}}

	readiness := execution.DeriveReadiness(health)
	if readiness.Ready {
		t.Fatal("expected Ready to be false when a dependency is missing")
	}
	if len(readiness.MissingDependencies) != 1 || readiness.MissingDependencies[0] != string(execution.HealthCheckRuntimeInstalled) {
		t.Fatalf("MissingDependencies = %v, want [runtime_installed]", readiness.MissingDependencies)
	}
}

func TestDeriveReadinessBlockingError(t *testing.T) {
	health := execution.HealthCheck{Items: []execution.HealthCheckItem{
		{Name: execution.HealthCheckManifestExists, Status: execution.HealthStatusFail, Detail: "package.json not found"},
	}}

	readiness := execution.DeriveReadiness(health)
	if readiness.Ready {
		t.Fatal("expected Ready to be false on a blocking error")
	}
	if len(readiness.BlockingErrors) != 1 || readiness.BlockingErrors[0] != "package.json not found" {
		t.Fatalf("BlockingErrors = %v, want [package.json not found]", readiness.BlockingErrors)
	}
}

func TestDeriveReadinessScriptsAvailableIsOnlyAWarning(t *testing.T) {
	health := execution.HealthCheck{Items: []execution.HealthCheckItem{
		{Name: execution.HealthCheckScriptsAvailable, Status: execution.HealthStatusFail, Detail: "package.json declares no scripts"},
	}}

	readiness := execution.DeriveReadiness(health)
	if !readiness.Ready {
		t.Fatal("expected Ready to still be true when only scripts_available failed")
	}
	if len(readiness.Warnings) != 1 || readiness.Warnings[0] != "package.json declares no scripts" {
		t.Fatalf("Warnings = %v, want [package.json declares no scripts]", readiness.Warnings)
	}
	if len(readiness.BlockingErrors) != 0 {
		t.Fatalf("BlockingErrors = %v, want empty", readiness.BlockingErrors)
	}
}
