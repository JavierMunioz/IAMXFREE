package deployment

import (
	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// Engine coordinates ApplicationService, ExecutionService, the Git Manager
// and the Nginx Manager to build a DeploymentPlan. It has no deployment
// logic of its own — every fact it reports comes from delegating to the
// component that actually knows it.
type Engine struct {
	appService       services.ApplicationService
	executionService services.ExecutionService
	gitManager       *git.Manager
	nginxManager     *nginx.Manager
}

// NewEngine returns an Engine backed by appService, executionService,
// gitManager and nginxManager.
func NewEngine(appService services.ApplicationService, executionService services.ExecutionService, gitManager *git.Manager, nginxManager *nginx.Manager) *Engine {
	return &Engine{
		appService:       appService,
		executionService: executionService,
		gitManager:       gitManager,
		nginxManager:     nginxManager,
	}
}
