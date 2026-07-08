package models

// DeploymentConfig holds everything needed to install, build and run an
// application: its networking and the shell command for each lifecycle
// stage. Commands are stored as plain strings; how they get executed is a
// concern for the future process manager, not for this model.
type DeploymentConfig struct {
	InternalPort int    `json:"internal_port,omitempty"`
	PublicPort   int    `json:"public_port,omitempty"`
	Domain       string `json:"domain,omitempty"`
	Subdomain    string `json:"subdomain,omitempty"`
	EnvFile      string `json:"env_file,omitempty"`

	// PackageManager names the tool used to install/run dependencies (e.g.
	// "npm", "pnpm", "poetry"). Future execution.Strategy implementations
	// for a given runtime will need this to pick the right command.
	PackageManager string `json:"package_manager,omitempty"`

	InstallCommand string `json:"install_command,omitempty"`
	BuildCommand   string `json:"build_command,omitempty"`
	StartCommand   string `json:"start_command,omitempty"`
	StopCommand    string `json:"stop_command,omitempty"`
	RestartCommand string `json:"restart_command,omitempty"`

	// PreDeployHook and PostDeployHook are optional shell commands run
	// around a deployment: PreDeployHook right before it (e.g. notify a
	// channel, put up a maintenance page), PostDeployHook right after it
	// finishes successfully (e.g. warm a cache, notify a channel). Blank
	// means no hook — never invented, only run when configured.
	PreDeployHook  string `json:"pre_deploy_hook,omitempty"`
	PostDeployHook string `json:"post_deploy_hook,omitempty"`
}
