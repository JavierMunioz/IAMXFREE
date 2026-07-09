package deployment

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost/runtimehosttest"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func zeroDowntimeApp() *models.Application {
	return &models.Application{
		ID:   "app-1",
		Name: "my-api",
		Source: models.SourceInfo{
			LocalPath: "/srv/apps/my-api",
		},
		Config: models.DeploymentConfig{
			Domain:       "example.com",
			InternalPort: 3000,
			Strategy:     models.DeploymentStrategyZeroDowntime,
		},
	}
}

func TestZeroDowntimeOperationsOrderAndCount(t *testing.T) {
	app := zeroDowntimeApp()
	engine := &Engine{
		appService:       &fakeAppService{app: app},
		executionService: &fakeExecutionService{},
		gitManager:       git.NewManager(runtimehosttest.NewFakeHost()),
		nginxManager:     nginx.NewManager(runtimehosttest.NewFakeHost()),
	}

	ops := engine.zeroDowntimeOperations(app, DeploymentPlan{ApplicationID: app.ID})

	wantMethods := []string{
		"Run", "Fetch", "Install", "Build",
		"StartCandidate", "HealthCheck", "UpdateUpstream", "ValidateConfig", "Reload",
		"PromoteCandidate", "StopSession", "UpdateConfig", "Run",
	}
	if len(ops) != len(wantMethods) {
		t.Fatalf("len(ops) = %d, want %d", len(ops), len(wantMethods))
	}
	for i, want := range wantMethods {
		if ops[i].Method != want {
			t.Errorf("ops[%d].Method = %q, want %q", i, ops[i].Method, want)
		}
	}
}

func TestStartCandidateOperationStartsOnCandidatePortAndUpdatesContext(t *testing.T) {
	app := zeroDowntimeApp()
	execSvc := &fakeExecutionService{
		startCandidateSession: services.RunSession{PID: 2000, Port: 3001, Status: "running"},
	}
	engine := &Engine{executionService: execSvc}
	dctx := &DeploymentContext{App: app, ActivePort: 3000, CandidatePort: 3001}

	op := engine.startCandidateOperation(dctx)
	if !op.Applicable {
		t.Fatal("expected StartCandidate to always be applicable")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if dctx.CandidateSession.PID != 2000 {
		t.Errorf("dctx.CandidateSession.PID = %d, want 2000", dctx.CandidateSession.PID)
	}
}

func TestStartCandidateOperationCompensateStopsCandidate(t *testing.T) {
	app := zeroDowntimeApp()
	execSvc := &fakeExecutionService{}
	engine := &Engine{executionService: execSvc}
	dctx := &DeploymentContext{App: app, CandidateSession: services.RunSession{PID: 2000}}

	op := engine.startCandidateOperation(dctx)
	if err := op.Compensate(context.Background()); err != nil {
		t.Fatalf("Compensate() error = %v", err)
	}
	if execSvc.stoppedCandidateWith.PID != 2000 {
		t.Errorf("stopped candidate PID = %d, want 2000", execSvc.stoppedCandidateWith.PID)
	}
}

func TestHealthCheckCandidateOperationPassesWhenRunning(t *testing.T) {
	app := zeroDowntimeApp()
	execSvc := &fakeExecutionService{
		checkStatusSession: services.RunSession{Status: "running"},
	}
	engine := &Engine{executionService: execSvc}
	dctx := &DeploymentContext{App: app, CandidateSession: services.RunSession{PID: 2000}}

	op := engine.healthCheckCandidateOperation(dctx)
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestHealthCheckCandidateOperationFailsWhenNotRunning(t *testing.T) {
	app := zeroDowntimeApp()
	execSvc := &fakeExecutionService{
		checkStatusSession: services.RunSession{Status: "stopped"},
	}
	engine := &Engine{executionService: execSvc}
	dctx := &DeploymentContext{App: app, CandidateSession: services.RunSession{PID: 2000}}

	op := engine.healthCheckCandidateOperation(dctx)
	if err := op.Run(context.Background()); err == nil {
		t.Fatal("expected an error when the candidate is not running")
	}
}

func TestHealthCheckCandidateOperationPropagatesStrategyError(t *testing.T) {
	app := zeroDowntimeApp()
	wantErr := errors.New("boom")
	execSvc := &fakeExecutionService{checkStatusErr: wantErr}
	engine := &Engine{executionService: execSvc}
	dctx := &DeploymentContext{App: app, CandidateSession: services.RunSession{PID: 2000}}

	op := engine.healthCheckCandidateOperation(dctx)
	if err := op.Run(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}

func TestUpdateNginxUpstreamOperationSkippedWithoutDomain(t *testing.T) {
	app := zeroDowntimeApp()
	app.Config.Domain = ""
	engine := &Engine{nginxManager: nginx.NewManager(fakeNginxHost())}
	dctx := &DeploymentContext{App: app}

	op := engine.updateNginxUpstreamOperation(dctx)
	if op.Applicable {
		t.Fatal("expected UpdateUpstream to not be applicable without a domain")
	}
}

func TestUpdateNginxUpstreamOperationSwitchesUpstream(t *testing.T) {
	app := zeroDowntimeApp()
	host := fakeNginxHost().WithFile("/etc/nginx/sites-available/example.com.conf")
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	dctx := &DeploymentContext{App: app, ActivePort: 3000, CandidatePort: 3001}

	op := engine.updateNginxUpstreamOperation(dctx)
	if !op.Applicable {
		t.Fatal("expected UpdateUpstream to be applicable with a domain configured")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	content, ok := host.WrittenFile("/etc/nginx/sites-available/example.com.conf")
	if !ok {
		t.Fatal("expected the site config to be rewritten")
	}
	if !strings.Contains(string(content), "3001") {
		t.Errorf("rendered config = %s, want it to reference port 3001", content)
	}
}

func TestUpdateNginxUpstreamOperationCompensateRestoresOldUpstream(t *testing.T) {
	app := zeroDowntimeApp()
	host := fakeNginxHost().WithFile("/etc/nginx/sites-available/example.com.conf")
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	dctx := &DeploymentContext{App: app, ActivePort: 3000, CandidatePort: 3001}

	op := engine.updateNginxUpstreamOperation(dctx)
	if err := op.Compensate(context.Background()); err != nil {
		t.Fatalf("Compensate() error = %v", err)
	}

	content, ok := host.WrittenFile("/etc/nginx/sites-available/example.com.conf")
	if !ok {
		t.Fatal("expected the site config to be rewritten")
	}
	if !strings.Contains(string(content), "3000") {
		t.Errorf("rendered config = %s, want it to reference the restored port 3000", content)
	}
}

func TestValidateNginxConfigOperationFailsWhenInvalid(t *testing.T) {
	app := zeroDowntimeApp()
	host := fakeNginxHost().WithRunResult("nginx", []string{"-t"},
		runtimehost.CommandResult{ExitCode: 1, Stderr: "nginx: [emerg] invalid"},
		&runtimehost.ExecutionError{Command: "nginx", ExitCode: 1, Err: errors.New("exit status 1")},
	)
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	dctx := &DeploymentContext{App: app}

	op := engine.validateNginxConfigOperation(dctx)
	if err := op.Run(context.Background()); err == nil {
		t.Fatal("expected Run to fail for an invalid nginx configuration")
	}
}

func TestValidateNginxConfigOperationSucceedsWhenValid(t *testing.T) {
	app := zeroDowntimeApp()
	host := fakeNginxHost().WithRunResult("nginx", []string{"-t"}, runtimehost.CommandResult{ExitCode: 0}, nil)
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	dctx := &DeploymentContext{App: app}

	op := engine.validateNginxConfigOperation(dctx)
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestZeroDowntimeReloadNginxOperationAlwaysApplicableWithDomain(t *testing.T) {
	app := zeroDowntimeApp()
	host := fakeNginxHost().WithRunResult("nginx", []string{"-s", "reload"}, runtimehost.CommandResult{ExitCode: 0}, nil)
	engine := &Engine{nginxManager: nginx.NewManager(host)}
	dctx := &DeploymentContext{App: app}

	op := engine.zeroDowntimeReloadNginxOperation(dctx)
	if !op.Applicable {
		t.Fatal("expected Reload to always be applicable for a zero-downtime deploy with a domain")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestConfirmSwitchOperationPromotesCandidate(t *testing.T) {
	app := zeroDowntimeApp()
	execSvc := &fakeExecutionService{}
	engine := &Engine{executionService: execSvc}
	dctx := &DeploymentContext{App: app}

	op := engine.confirmSwitchOperation(dctx)
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !execSvc.promoted {
		t.Fatal("expected PromoteCandidate to have been called")
	}
}

func TestStopOldSessionOperationSkippedWithoutPreviousSession(t *testing.T) {
	app := zeroDowntimeApp()
	engine := &Engine{executionService: &fakeExecutionService{}}
	dctx := &DeploymentContext{App: app}

	op := engine.stopOldSessionOperation(dctx)
	if op.Applicable {
		t.Fatal("expected StopSession to not be applicable without a previous session")
	}
}

func TestStopOldSessionOperationSwallowsFailureIntoWarnings(t *testing.T) {
	app := zeroDowntimeApp()
	wantErr := errors.New("process would not die")
	execSvc := &fakeExecutionService{stopSessionErr: wantErr}
	engine := &Engine{executionService: execSvc}
	dctx := &DeploymentContext{App: app, PreviousSession: services.RunSession{PID: 1000}}

	op := engine.stopOldSessionOperation(dctx)
	if !op.Applicable {
		t.Fatal("expected StopSession to be applicable with a previous session")
	}
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v, want nil (failure must be swallowed into warnings, never fail the deployment)", err)
	}
	if len(dctx.Warnings) != 1 {
		t.Fatalf("len(dctx.Warnings) = %d, want 1", len(dctx.Warnings))
	}
	if !strings.Contains(dctx.Warnings[0], "1000") {
		t.Errorf("warning = %q, want it to mention the PID", dctx.Warnings[0])
	}
}

func TestPersistNewPrimaryPortOperationSwapsPorts(t *testing.T) {
	app := zeroDowntimeApp()
	appSvc := &fakeAppService{app: app}
	engine := &Engine{appService: appSvc}
	dctx := &DeploymentContext{App: app, ActivePort: 3000, CandidatePort: 3001}

	op := engine.persistNewPrimaryPortOperation(dctx)
	if err := op.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if appSvc.updatedConfig.PrimaryPort != 3001 {
		t.Errorf("PrimaryPort = %d, want 3001 (the candidate that just took over)", appSvc.updatedConfig.PrimaryPort)
	}
	if appSvc.updatedConfig.SecondaryPort != 3000 {
		t.Errorf("SecondaryPort = %d, want 3000 (the port that used to be active)", appSvc.updatedConfig.SecondaryPort)
	}
}
