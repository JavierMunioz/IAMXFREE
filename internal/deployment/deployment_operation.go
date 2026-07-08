package deployment

// DeploymentOperation names the kind of action a DeploymentStep represents.
type DeploymentOperation string

const (
	OperationVerifyRepository    DeploymentOperation = "verify_repository"
	OperationCheckLocalChanges   DeploymentOperation = "check_local_changes"
	OperationPull                DeploymentOperation = "pull"
	OperationPreDeployHook       DeploymentOperation = "pre_deploy_hook"
	OperationInstallDependencies DeploymentOperation = "install_dependencies"
	OperationBuild               DeploymentOperation = "build"
	OperationCheckRunningStatus  DeploymentOperation = "check_running_status"
	OperationRestartApplication  DeploymentOperation = "restart_application"
	OperationVerifyNginxConfig   DeploymentOperation = "verify_nginx_config"
	OperationReloadNginx         DeploymentOperation = "reload_nginx"
	OperationPostDeployHook      DeploymentOperation = "post_deploy_hook"
)
