package deployment

// DeploymentOperation names the kind of action a DeploymentStep represents.
// This iteration only ever plans these operations — VerifyRepository and
// CheckLocalChanges are the only ones it can actually carry out (both are
// pure reads); Pull, InstallDependencies, Build, RestartApplication,
// VerifyNginxConfig and ReloadNginx are described in the plan but not yet
// executable.
type DeploymentOperation string

const (
	OperationVerifyRepository    DeploymentOperation = "verify_repository"
	OperationCheckLocalChanges   DeploymentOperation = "check_local_changes"
	OperationPull                DeploymentOperation = "pull"
	OperationInstallDependencies DeploymentOperation = "install_dependencies"
	OperationBuild               DeploymentOperation = "build"
	OperationCheckRunningStatus  DeploymentOperation = "check_running_status"
	OperationRestartApplication  DeploymentOperation = "restart_application"
	OperationVerifyNginxConfig   DeploymentOperation = "verify_nginx_config"
	OperationReloadNginx         DeploymentOperation = "reload_nginx"
)
