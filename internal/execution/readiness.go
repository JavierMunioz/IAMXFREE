package execution

// Readiness is a go/no-go assessment of whether an application can be
// started right now, derived from a HealthCheck.
type Readiness struct {
	Ready bool

	// MissingDependencies lists failed runtime/package-manager checks.
	MissingDependencies []string

	// BlockingErrors lists failures that prevent starting (missing
	// manifest, invalid path, no start command, ...).
	BlockingErrors []string

	// Warnings lists failures that are unusual but not fatal on their own
	// (e.g. no scripts declared) — Ready can still be true alongside them.
	Warnings []string
}

// DeriveReadiness interprets health into a Readiness. This policy is shared
// architecture: every Strategy derives Readiness from its HealthCheck the
// same way, so the rule for what merely warns versus what blocks is
// defined once, not per technology.
func DeriveReadiness(health HealthCheck) Readiness {
	readiness := Readiness{Ready: true}

	for _, item := range health.Items {
		if item.Status != HealthStatusFail {
			continue
		}

		switch item.Name {
		case HealthCheckRuntimeInstalled, HealthCheckPackageManagerInstalled:
			readiness.MissingDependencies = append(readiness.MissingDependencies, string(item.Name))
			readiness.Ready = false
		case HealthCheckScriptsAvailable:
			readiness.Warnings = append(readiness.Warnings, item.Detail)
		default:
			readiness.BlockingErrors = append(readiness.BlockingErrors, item.Detail)
			readiness.Ready = false
		}
	}

	return readiness
}
