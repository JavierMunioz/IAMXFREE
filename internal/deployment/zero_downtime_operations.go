package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

// healthCheckAttempts/healthCheckInterval bound how long
// healthCheckCandidateOperation is willing to wait for a freshly started
// candidate session to prove it didn't crash on startup. This is a process
// liveness check only (via ExecutionService.CheckStatus) — it does not
// probe the candidate's port over the network; that is a natural next
// step this leaves room for, not something invented here.
const (
	healthCheckAttempts = 5
	healthCheckInterval = 500 * time.Millisecond
)

// zeroDowntimeOperations builds the operation sequence for
// DeploymentStrategyZeroDowntime: start a second ("candidate") session on
// the port not currently serving traffic, health-check it, and only then
// switch Nginx's upstream over to it — keeping the previous session
// running throughout until the switch is confirmed. app and plan are
// otherwise treated the same as buildOperations's standard sequence
// (pre/post-deploy hooks, Fetch, Install, Build reuse the exact same
// operations, gated by the exact same plan analysis).
func (e *Engine) zeroDowntimeOperations(app *models.Application, plan DeploymentPlan) []operations.Operation {
	previousSession, hasActive := e.executionService.ActiveSession(app.ID)

	dctx := &DeploymentContext{
		App:             app,
		Plan:            plan,
		PreviousSession: previousSession,
	}
	dctx.ActivePort = activePortFor(app.Config, previousSession, hasActive)
	dctx.CandidatePort = candidatePortFor(app.Config, dctx.ActivePort)

	return []operations.Operation{
		e.preDeployHookOperation(app, plan),
		e.fetchOperation(app, plan),
		e.installOperation(app, plan),
		e.buildOperation(app, plan),
		e.startCandidateOperation(dctx),
		e.healthCheckCandidateOperation(dctx),
		e.updateNginxUpstreamOperation(dctx),
		e.validateNginxConfigOperation(dctx),
		e.zeroDowntimeReloadNginxOperation(dctx),
		e.confirmSwitchOperation(dctx),
		e.stopOldSessionOperation(dctx),
		e.persistNewPrimaryPortOperation(dctx),
		e.postDeployHookOperation(app, plan),
	}
}

// startCandidateOperation starts a second session of app on
// dctx.CandidatePort, alongside whatever is already active. It is always
// applicable — that is the point of a zero-downtime deployment. Its
// Compensate stops that candidate session again, which is what undoes it
// if a later operation (the health check, the Nginx switch) fails.
func (e *Engine) startCandidateOperation(dctx *DeploymentContext) operations.Operation {
	return operations.Operation{
		Name: "Start candidate session", Component: string(ComponentExecution), Method: "StartCandidate",
		Applicable: true,
		Run: func(ctx context.Context) error {
			session, err := e.executionService.StartCandidate(ctx, dctx.App.ID, dctx.CandidatePort)
			if err != nil {
				return err
			}
			dctx.CandidateSession = session
			return nil
		},
		Compensate: func(ctx context.Context) error {
			return e.executionService.StopCandidate(ctx, dctx.App.ID, dctx.CandidateSession)
		},
	}
}

// healthCheckCandidateOperation polls the candidate session's process
// liveness a few times, failing if it ever reports anything other than
// running — catching a candidate that crashed immediately on startup. It
// has no Compensate of its own: a failure here is undone by
// startCandidateOperation's Compensate, which the Executor runs
// automatically once this operation fails, per the existing reverse
// compensation walk.
func (e *Engine) healthCheckCandidateOperation(dctx *DeploymentContext) operations.Operation {
	return operations.Operation{
		Name: "Health check candidate session", Component: string(ComponentExecution), Method: "HealthCheck",
		Applicable: true,
		Run: func(ctx context.Context) error {
			for attempt := 0; attempt < healthCheckAttempts; attempt++ {
				if attempt > 0 {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(healthCheckInterval):
					}
				}

				updated, err := e.executionService.CheckStatus(ctx, dctx.App.ID, dctx.CandidateSession)
				if err != nil {
					return fmt.Errorf("health check: %w", err)
				}
				if updated.Status != "running" {
					return fmt.Errorf("candidate session is not running (status: %s)", updated.Status)
				}
			}
			return nil
		},
	}
}

// updateNginxUpstreamOperation switches the application's Nginx site to
// reverse-proxy to the candidate port. Its Compensate restores the
// previous upstream — used if this succeeds but a later step (config
// validation, reload) fails.
func (e *Engine) updateNginxUpstreamOperation(dctx *DeploymentContext) operations.Operation {
	op := operations.Operation{Name: "Update Nginx upstream", Component: string(ComponentNginx), Method: "UpdateUpstream"}

	if dctx.App.Config.Domain == "" {
		op.SkipReason = "no domain configured — reverse proxy not required"
		return op
	}

	op.Applicable = true
	op.Run = func(ctx context.Context) error {
		return e.nginxManager.UpdateVirtualHost(ctx, virtualHostFor(dctx.App, dctx.CandidatePort))
	}
	op.Compensate = func(ctx context.Context) error {
		return e.nginxManager.UpdateVirtualHost(ctx, virtualHostFor(dctx.App, dctx.ActivePort))
	}
	return op
}

// validateNginxConfigOperation runs `nginx -t` against the config already
// written by updateNginxUpstreamOperation. It has no Compensate of its
// own — a failure here is undone by updateNginxUpstreamOperation's
// Compensate, restoring the previous upstream immediately.
func (e *Engine) validateNginxConfigOperation(dctx *DeploymentContext) operations.Operation {
	op := operations.Operation{Name: "Validate Nginx configuration", Component: string(ComponentNginx), Method: "ValidateConfig"}

	if dctx.App.Config.Domain == "" {
		op.SkipReason = "no domain configured — reverse proxy not required"
		return op
	}

	op.Applicable = true
	op.Run = func(ctx context.Context) error {
		result, err := e.nginxManager.ValidateConfig(ctx)
		if err != nil {
			return err
		}
		if !result.Valid {
			return fmt.Errorf("nginx configuration is invalid: %s", result.Output)
		}
		return nil
	}
	return op
}

// zeroDowntimeReloadNginxOperation reloads Nginx to pick up the switched
// upstream. Unlike the standard sequence's reloadNginxOperation (gated by
// whether the site was already enabled), a zero-downtime deployment always
// needs a reload once it has changed the upstream — so applicability here
// only depends on there being a site at all. It has no Compensate of its
// own, for the same reason validateNginxConfigOperation doesn't: a failure
// here is undone by updateNginxUpstreamOperation's Compensate.
func (e *Engine) zeroDowntimeReloadNginxOperation(dctx *DeploymentContext) operations.Operation {
	op := operations.Operation{Name: "Reload Nginx", Component: string(ComponentNginx), Method: "Reload"}

	if dctx.App.Config.Domain == "" {
		op.SkipReason = "no domain configured — reverse proxy not required"
		return op
	}

	op.Applicable = true
	op.Run = func(ctx context.Context) error {
		return e.nginxManager.Reload(ctx)
	}
	return op
}

// confirmSwitchOperation makes the candidate session the new active one in
// ExecutionService's own bookkeeping, now that Nginx is actually routing
// traffic to it.
func (e *Engine) confirmSwitchOperation(dctx *DeploymentContext) operations.Operation {
	return operations.Operation{
		Name: "Confirm switch to candidate session", Component: string(ComponentExecution), Method: "PromoteCandidate",
		Applicable: true,
		Run: func(ctx context.Context) error {
			return e.executionService.PromoteCandidate(ctx, dctx.App.ID)
		},
	}
}

// stopOldSessionOperation stops the session that was active before this
// deployment, now that traffic has already switched away from it. A
// failure here is deliberately never treated as a deployment failure —
// traffic already switched successfully, so rolling everything back would
// be wrong — it is instead recorded on dctx.Warnings, so it is surfaced,
// never hidden, without failing (or compensating) an otherwise-successful
// deployment.
func (e *Engine) stopOldSessionOperation(dctx *DeploymentContext) operations.Operation {
	op := operations.Operation{Name: "Stop previous session", Component: string(ComponentExecution), Method: "StopSession"}

	if dctx.PreviousSession.PID == 0 {
		op.SkipReason = "no previous session was running"
		return op
	}

	op.Applicable = true
	op.Run = func(ctx context.Context) error {
		if err := e.executionService.StopSession(ctx, dctx.App.ID, dctx.PreviousSession); err != nil {
			dctx.Warnings = append(dctx.Warnings, fmt.Sprintf(
				"previous session (PID %d) did not stop cleanly: %v", dctx.PreviousSession.PID, err))
		}
		return nil
	}
	return op
}

// persistNewPrimaryPortOperation durably records the outcome of this
// deployment's port swap: the candidate port (now serving traffic)
// becomes PrimaryPort, and the port that used to be active becomes
// SecondaryPort — free for the next deployment's candidate.
func (e *Engine) persistNewPrimaryPortOperation(dctx *DeploymentContext) operations.Operation {
	return operations.Operation{
		Name: "Persist new primary port", Component: string(ComponentExecution), Method: "UpdateConfig",
		Applicable: true,
		Run: func(ctx context.Context) error {
			cfg := dctx.App.Config
			cfg.PrimaryPort = dctx.CandidatePort
			cfg.SecondaryPort = dctx.ActivePort
			_, err := e.appService.UpdateConfig(ctx, dctx.App.ID, cfg)
			return err
		},
	}
}
