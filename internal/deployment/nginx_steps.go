package deployment

import (
	"context"
	"fmt"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
)

// nginxSteps answers "does an Nginx site exist for this application" and
// describes whether Nginx would need reloading, delegating entirely to the
// Nginx Manager's read-only ListVirtualHosts.
func (e *Engine) nginxSteps(ctx context.Context, app *models.Application) []DeploymentStep {
	verify := DeploymentStep{
		Name:      "Verify Nginx configuration",
		Component: ComponentNginx,
		Operation: OperationVerifyNginxConfig,
		Required:  app.Config.Domain != "",
	}

	if app.Config.Domain == "" {
		verify.Status = StepStatusSkipped
		verify.Expected = DeploymentResult{Description: "no domain configured — reverse proxy not required", WouldSucceed: true}
		return []DeploymentStep{verify}
	}

	sites, err := e.nginxManager.ListVirtualHosts(ctx)
	if err != nil {
		verify.Status = StepStatusBlocked
		verify.Risks = append(verify.Risks, fmt.Sprintf("could not list Nginx sites: %v", err))
		verify.Expected = DeploymentResult{Description: "confirm a site exists for the configured domain", WouldSucceed: false}
		return []DeploymentStep{verify}
	}

	var site *nginx.SiteSummary
	for i := range sites {
		if sites[i].ServerName == app.Config.Domain {
			site = &sites[i]
			break
		}
	}

	reload := DeploymentStep{
		Name:      "Reload Nginx",
		Component: ComponentNginx,
		Operation: OperationReloadNginx,
	}

	switch {
	case site == nil:
		verify.Status = StepStatusWarning
		verify.Warnings = append(verify.Warnings, fmt.Sprintf("no Nginx site configured for %q yet", app.Config.Domain))
		verify.Expected = DeploymentResult{Description: "site would need to be created", WouldSucceed: false}
		reload.Required = true
		reload.Status = StepStatusReady
		reload.Expected = DeploymentResult{Description: "reload Nginx once the site is created", WouldSucceed: true}

	case !site.Enabled:
		verify.Status = StepStatusWarning
		verify.Warnings = append(verify.Warnings, fmt.Sprintf("site %q exists but is not enabled", site.ServerName))
		verify.Expected = DeploymentResult{Description: fmt.Sprintf("site %q found, disabled", site.ServerName), WouldSucceed: true}
		reload.Required = true
		reload.Status = StepStatusReady
		reload.Expected = DeploymentResult{Description: "enable the site and reload Nginx", WouldSucceed: true}

	default:
		verify.Status = StepStatusReady
		verify.Expected = DeploymentResult{Description: fmt.Sprintf("site %q already configured and enabled", site.ServerName), WouldSucceed: true}
		reload.Required = false
		reload.Status = StepStatusSkipped
		reload.Expected = DeploymentResult{Description: "site already enabled — no reload necessary", WouldSucceed: true}
	}

	return []DeploymentStep{verify, reload}
}
