package execution

// HealthCheckName identifies one prerequisite a Strategy verified. These
// names are shared architecture: every technology's HealthCheck uses the
// same vocabulary, so a caller (e.g. the dashboard) never needs to know
// which strategy produced it. What a name means in practice — which tool,
// which file — is specific to the Strategy and goes into the item's Detail.
type HealthCheckName string

const (
	HealthCheckRuntimeInstalled        HealthCheckName = "runtime_installed"
	HealthCheckPackageManagerInstalled HealthCheckName = "package_manager_installed"
	HealthCheckManifestExists          HealthCheckName = "manifest_exists"
	HealthCheckPathValid               HealthCheckName = "path_valid"
	HealthCheckScriptsAvailable        HealthCheckName = "scripts_available"
	HealthCheckCommandsConfigured      HealthCheckName = "commands_configured"
	HealthCheckDirectoryAccessible     HealthCheckName = "directory_accessible"
)

// HealthStatus is the structured pass/fail outcome of one HealthCheckItem —
// never free text.
type HealthStatus string

const (
	HealthStatusPass HealthStatus = "pass"
	HealthStatusFail HealthStatus = "fail"
)

// HealthCheckItem is one verified prerequisite. Detail is short,
// human-readable context (e.g. "node not found on PATH") — supplementary
// to Status, never a replacement for it.
type HealthCheckItem struct {
	Name   HealthCheckName
	Status HealthStatus
	Detail string
}

// HealthCheck is the full structured result of diagnosing whether a
// Strategy can manage an application.
type HealthCheck struct {
	Items []HealthCheckItem
}

// Healthy reports whether every item passed.
func (h HealthCheck) Healthy() bool {
	return len(h.Failed()) == 0
}

// Failed returns every item whose Status is HealthStatusFail.
func (h HealthCheck) Failed() []HealthCheckItem {
	var failed []HealthCheckItem
	for _, item := range h.Items {
		if item.Status == HealthStatusFail {
			failed = append(failed, item)
		}
	}
	return failed
}
