package deployment

// DeploymentComponent names which existing manager is responsible for a
// DeploymentStep — the Engine never acts on its own behalf.
type DeploymentComponent string

const (
	ComponentGit       DeploymentComponent = "git"
	ComponentExecution DeploymentComponent = "execution"
	ComponentNginx     DeploymentComponent = "nginx"
)
