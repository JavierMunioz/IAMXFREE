package deployment

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func TestResolvePortsBootstrapsFromInternalPortWhenUnset(t *testing.T) {
	primary, secondary := resolvePorts(models.DeploymentConfig{InternalPort: 3000})
	if primary != 3000 || secondary != 3001 {
		t.Errorf("resolvePorts() = (%d, %d), want (3000, 3001)", primary, secondary)
	}
}

func TestResolvePortsRespectsAlreadyConfiguredPorts(t *testing.T) {
	primary, secondary := resolvePorts(models.DeploymentConfig{
		InternalPort: 3000, PrimaryPort: 8080, SecondaryPort: 8081,
	})
	if primary != 8080 || secondary != 8081 {
		t.Errorf("resolvePorts() = (%d, %d), want (8080, 8081)", primary, secondary)
	}
}

func TestResolvePortsRespectsPrimaryAloneBeingSet(t *testing.T) {
	// Only one of the pair being non-zero still counts as "configured" —
	// resolvePorts never second-guesses a deliberately zero SecondaryPort.
	primary, secondary := resolvePorts(models.DeploymentConfig{PrimaryPort: 8080})
	if primary != 8080 || secondary != 0 {
		t.Errorf("resolvePorts() = (%d, %d), want (8080, 0)", primary, secondary)
	}
}

func TestActivePortForUsesTrackedSessionPort(t *testing.T) {
	cfg := models.DeploymentConfig{PrimaryPort: 8080, SecondaryPort: 8081}
	got := activePortFor(cfg, services.RunSession{Port: 8081}, true)
	if got != 8081 {
		t.Errorf("activePortFor() = %d, want 8081 (session's own port)", got)
	}
}

func TestActivePortForFallsBackToPrimaryWhenNoSession(t *testing.T) {
	cfg := models.DeploymentConfig{PrimaryPort: 8080, SecondaryPort: 8081}
	got := activePortFor(cfg, services.RunSession{}, false)
	if got != 8080 {
		t.Errorf("activePortFor() = %d, want 8080 (primary, nothing running)", got)
	}
}

func TestActivePortForFallsBackToPrimaryForLegacySessionWithoutPort(t *testing.T) {
	cfg := models.DeploymentConfig{PrimaryPort: 8080, SecondaryPort: 8081}
	got := activePortFor(cfg, services.RunSession{PID: 4242}, true)
	if got != 8080 {
		t.Errorf("activePortFor() = %d, want 8080 (session predates Port tracking)", got)
	}
}

func TestCandidatePortForPicksTheOtherPort(t *testing.T) {
	cfg := models.DeploymentConfig{PrimaryPort: 8080, SecondaryPort: 8081}

	if got := candidatePortFor(cfg, 8080); got != 8081 {
		t.Errorf("candidatePortFor(active=8080) = %d, want 8081", got)
	}
	if got := candidatePortFor(cfg, 8081); got != 8080 {
		t.Errorf("candidatePortFor(active=8081) = %d, want 8080", got)
	}
}

func TestCandidatePortForBootstrapCase(t *testing.T) {
	cfg := models.DeploymentConfig{InternalPort: 3000}
	primary, _ := resolvePorts(cfg)

	got := candidatePortFor(cfg, primary)
	if got != 3001 {
		t.Errorf("candidatePortFor() = %d, want 3001 (bootstrapped secondary)", got)
	}
}
