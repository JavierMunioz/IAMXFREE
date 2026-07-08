package deployment

import (
	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// Engine coordinates ApplicationService, ExecutionService, the Git Manager
// and the Nginx Manager to build a DeploymentPlan. It has no deployment
// logic of its own — every fact it reports comes from delegating to the
// component that actually knows it.
//
// host runs pre/post-deploy hook commands directly (they are plain shell
// commands, not tied to any application's runtime, so they don't belong to
// an execution.Strategy).
type Engine struct {
	appService       services.ApplicationService
	executionService services.ExecutionService
	gitManager       *git.Manager
	nginxManager     *nginx.Manager
	host             runtimehost.Host
}

// NewEngine returns an Engine backed by appService, executionService,
// gitManager, nginxManager and host.
func NewEngine(appService services.ApplicationService, executionService services.ExecutionService, gitManager *git.Manager, nginxManager *nginx.Manager, host runtimehost.Host) *Engine {
	return &Engine{
		appService:       appService,
		executionService: executionService,
		gitManager:       gitManager,
		nginxManager:     nginxManager,
		host:             host,
	}
}
