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

	InstallCommand string `json:"install_command,omitempty"`
	BuildCommand   string `json:"build_command,omitempty"`
	StartCommand   string `json:"start_command,omitempty"`
	StopCommand    string `json:"stop_command,omitempty"`
	RestartCommand string `json:"restart_command,omitempty"`
}
