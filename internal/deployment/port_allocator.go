package deployment

import (
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// resolvePorts returns the pair of ports a DeploymentStrategyZeroDowntime
// deployment alternates between. If cfg.PrimaryPort and cfg.SecondaryPort
// are both zero (never configured yet), it bootstraps them from
// cfg.InternalPort — the port the application already runs on today — and
// InternalPort+1, so zero-downtime deploys work without requiring a
// separate setup step first. Once either field is non-zero, both are
// taken exactly as configured.
func resolvePorts(cfg models.DeploymentConfig) (primary, secondary int) {
	if cfg.PrimaryPort != 0 || cfg.SecondaryPort != 0 {
		return cfg.PrimaryPort, cfg.SecondaryPort
	}
	return cfg.InternalPort, cfg.InternalPort + 1
}

// activePortFor determines which of cfg's two ports is currently serving
// traffic. The currently tracked active session is the ground truth when
// it has a recorded port; otherwise (no session running yet, or a session
// started before Port existed) it falls back to the resolved primary port
// — the sane default for a first-ever zero-downtime deployment.
func activePortFor(cfg models.DeploymentConfig, active services.RunSession, hasActive bool) int {
	if hasActive && active.Port != 0 {
		return active.Port
	}
	primary, _ := resolvePorts(cfg)
	return primary
}

// candidatePortFor picks whichever of cfg's two ports is not activePort —
// the port the next deployment's candidate session should start on.
func candidatePortFor(cfg models.DeploymentConfig, activePort int) int {
	primary, secondary := resolvePorts(cfg)
	if activePort == secondary {
		return primary
	}
	return secondary
}
