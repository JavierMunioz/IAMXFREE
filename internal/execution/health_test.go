package execution_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
)

func TestHealthCheckHealthyWithNoFailures(t *testing.T) {
	health := execution.HealthCheck{Items: []execution.HealthCheckItem{
		{Name: execution.HealthCheckRuntimeInstalled, Status: execution.HealthStatusPass},
		{Name: execution.HealthCheckPathValid, Status: execution.HealthStatusPass},
	}}

	if !health.Healthy() {
		t.Fatal("expected Healthy() to be true when every item passes")
	}
	if len(health.Failed()) != 0 {
		t.Fatalf("Failed() = %v, want empty", health.Failed())
	}
}

func TestHealthCheckUnhealthyWithAnyFailure(t *testing.T) {
	health := execution.HealthCheck{Items: []execution.HealthCheckItem{
		{Name: execution.HealthCheckRuntimeInstalled, Status: execution.HealthStatusPass},
		{Name: execution.HealthCheckPathValid, Status: execution.HealthStatusFail, Detail: "path is empty"},
	}}

	if health.Healthy() {
		t.Fatal("expected Healthy() to be false when any item fails")
	}

	failed := health.Failed()
	if len(failed) != 1 || failed[0].Name != execution.HealthCheckPathValid {
		t.Fatalf("Failed() = %+v, want a single path_valid failure", failed)
	}
}

func TestHealthCheckEmptyIsHealthy(t *testing.T) {
	var health execution.HealthCheck
	if !health.Healthy() {
		t.Fatal("expected an empty HealthCheck to report healthy (vacuously true)")
	}
}
