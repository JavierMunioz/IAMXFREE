package deployment

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/nginx"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// TestZeroDowntimeDeploymentSucceedsEndToEnd wires BuildOperations straight
// into operations.Executor for a fully applicable zero-downtime deployment:
// candidate starts, health-checks clean, Nginx switches over, the previous
// session stops, and the new primary port is persisted.
func TestZeroDowntimeDeploymentSucceedsEndToEnd(t *testing.T) {
	app := zeroDowntimeApp()
	appSvc := &fakeAppService{app: app}
	execSvc := &fakeExecutionService{
		session:               services.RunSession{PID: 1000, Port: 3000, Status: "running"},
		running:               true,
		startCandidateSession: services.RunSession{PID: 2000, Port: 3001, Status: "running"},
		checkStatusSession:    services.RunSession{Status: "running"},
	}
	nginxHost := fakeNginxHost().
		WithFile("/etc/nginx/sites-available/example.com.conf").
		WithRunResult("nginx", []string{"-t"}, runtimehost.CommandResult{ExitCode: 0}, nil).
		WithRunResult("nginx", []string{"-s", "reload"}, runtimehost.CommandResult{ExitCode: 0}, nil)

	engine := &Engine{
		appService:       appSvc,
		executionService: execSvc,
		nginxManager:     nginx.NewManager(nginxHost),
	}

	ops, err := engine.BuildOperations(context.Background(), DeploymentPlan{ApplicationID: app.ID})
	if err != nil {
		t.Fatalf("BuildOperations() error = %v", err)
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if summary.Overall != operations.StateSuccess {
		t.Fatalf("Overall = %q, want %q; operations = %+v", summary.Overall, operations.StateSuccess, summary.Operations)
	}
	if summary.Failed != 0 || summary.Compensated != 0 || summary.CompensationFailed != 0 {
		t.Fatalf("expected a clean success, got %+v", summary)
	}
	if !execSvc.promoted {
		t.Error("expected the candidate to be promoted")
	}
	if execSvc.stoppedSessionWith.PID != 1000 {
		t.Errorf("stopped previous session PID = %d, want 1000", execSvc.stoppedSessionWith.PID)
	}
	if appSvc.updatedConfig.PrimaryPort != 3001 || appSvc.updatedConfig.SecondaryPort != 3000 {
		t.Errorf("persisted ports = primary %d/secondary %d, want 3001/3000",
			appSvc.updatedConfig.PrimaryPort, appSvc.updatedConfig.SecondaryPort)
	}

	content, ok := nginxHost.WrittenFile("/etc/nginx/sites-available/example.com.conf")
	if !ok || !containsPort(string(content), 3001) {
		t.Errorf("expected the rewritten Nginx config to reference port 3001, got: %s", content)
	}
}

// TestZeroDowntimeDeploymentHealthCheckFailureNeverTouchesNginx proves that
// a candidate that fails its health check gets stopped and compensated
// without Nginx ever being touched — the previously active session keeps
// serving traffic throughout.
func TestZeroDowntimeDeploymentHealthCheckFailureNeverTouchesNginx(t *testing.T) {
	app := zeroDowntimeApp()
	appSvc := &fakeAppService{app: app}
	execSvc := &fakeExecutionService{
		session:               services.RunSession{PID: 1000, Port: 3000, Status: "running"},
		running:               true,
		startCandidateSession: services.RunSession{PID: 2000, Port: 3001, Status: "running"},
		checkStatusSession:    services.RunSession{Status: "stopped"},
	}
	nginxHost := fakeNginxHost().WithFile("/etc/nginx/sites-available/example.com.conf")

	engine := &Engine{
		appService:       appSvc,
		executionService: execSvc,
		nginxManager:     nginx.NewManager(nginxHost),
	}

	ops, err := engine.BuildOperations(context.Background(), DeploymentPlan{ApplicationID: app.ID})
	if err != nil {
		t.Fatalf("BuildOperations() error = %v", err)
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if summary.Overall != operations.StateFailed {
		t.Fatalf("Overall = %q, want %q", summary.Overall, operations.StateFailed)
	}
	if summary.Compensated != 1 {
		t.Fatalf("Compensated = %d, want 1 (StartCandidate)", summary.Compensated)
	}
	if execSvc.stoppedCandidateWith.PID != 2000 {
		t.Errorf("stopped candidate PID = %d, want 2000", execSvc.stoppedCandidateWith.PID)
	}
	if execSvc.promoted {
		t.Error("expected the candidate to never be promoted")
	}
	if execSvc.stoppedSessionWith.PID != 0 {
		t.Error("expected the previous session to never be stopped")
	}
	if appSvc.updateConfigCalls != 0 {
		t.Error("expected the port config to never be persisted")
	}
	if _, ok := nginxHost.WrittenFile("/etc/nginx/sites-available/example.com.conf"); ok {
		t.Error("expected Nginx to never be touched when the health check fails")
	}
}

// TestZeroDowntimeDeploymentReloadFailureRestoresUpstream proves that if
// Nginx's reload fails after the upstream file was already rewritten, the
// upstream gets restored to the previously active port.
func TestZeroDowntimeDeploymentReloadFailureRestoresUpstream(t *testing.T) {
	app := zeroDowntimeApp()
	appSvc := &fakeAppService{app: app}
	execSvc := &fakeExecutionService{
		session:               services.RunSession{PID: 1000, Port: 3000, Status: "running"},
		running:               true,
		startCandidateSession: services.RunSession{PID: 2000, Port: 3001, Status: "running"},
		checkStatusSession:    services.RunSession{Status: "running"},
	}
	nginxHost := fakeNginxHost().
		WithFile("/etc/nginx/sites-available/example.com.conf").
		WithRunResult("nginx", []string{"-t"}, runtimehost.CommandResult{ExitCode: 0}, nil).
		WithRunResult("nginx", []string{"-s", "reload"},
			runtimehost.CommandResult{ExitCode: 1, Stderr: "nginx: [error] reload failed"},
			&runtimehost.ExecutionError{Command: "nginx", ExitCode: 1, Err: errors.New("exit status 1")},
		)

	engine := &Engine{
		appService:       appSvc,
		executionService: execSvc,
		nginxManager:     nginx.NewManager(nginxHost),
	}

	ops, err := engine.BuildOperations(context.Background(), DeploymentPlan{ApplicationID: app.ID})
	if err != nil {
		t.Fatalf("BuildOperations() error = %v", err)
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if summary.Overall != operations.StateFailed {
		t.Fatalf("Overall = %q, want %q", summary.Overall, operations.StateFailed)
	}
	// StartCandidate and UpdateUpstream both succeeded before Reload failed,
	// so both get compensated (in reverse: upstream restored, then
	// candidate stopped).
	if summary.Compensated != 2 {
		t.Fatalf("Compensated = %d, want 2 (UpdateUpstream + StartCandidate); operations = %+v", summary.Compensated, summary.Operations)
	}
	if execSvc.stoppedCandidateWith.PID != 2000 {
		t.Errorf("stopped candidate PID = %d, want 2000", execSvc.stoppedCandidateWith.PID)
	}

	content, ok := nginxHost.WrittenFile("/etc/nginx/sites-available/example.com.conf")
	if !ok {
		t.Fatal("expected the site config to have been written at least once")
	}
	if !containsPort(string(content), 3000) {
		t.Errorf("expected the restored config to reference the original port 3000, got: %s", content)
	}
	if appSvc.updateConfigCalls != 0 {
		t.Error("expected the port config to never be persisted after a failed deployment")
	}
}

// TestZeroDowntimeDeploymentUpstreamUpdateFailureCompensatesCandidate
// proves that if writing the new upstream itself fails (e.g. the site
// doesn't exist on disk), the candidate session started earlier still
// gets compensated (stopped).
func TestZeroDowntimeDeploymentUpstreamUpdateFailureCompensatesCandidate(t *testing.T) {
	app := zeroDowntimeApp()
	appSvc := &fakeAppService{app: app}
	execSvc := &fakeExecutionService{
		startCandidateSession: services.RunSession{PID: 2000, Port: 3001, Status: "running"},
		checkStatusSession:    services.RunSession{Status: "running"},
	}
	// No site file written to sites-available: UpdateVirtualHost returns
	// ErrSiteNotFound.
	nginxHost := fakeNginxHost()

	engine := &Engine{
		appService:       appSvc,
		executionService: execSvc,
		nginxManager:     nginx.NewManager(nginxHost),
	}

	ops, err := engine.BuildOperations(context.Background(), DeploymentPlan{ApplicationID: app.ID})
	if err != nil {
		t.Fatalf("BuildOperations() error = %v", err)
	}

	summary := operations.NewExecutor().Execute(context.Background(), ops, nil)

	if summary.Overall != operations.StateFailed {
		t.Fatalf("Overall = %q, want %q", summary.Overall, operations.StateFailed)
	}
	if summary.Compensated != 1 {
		t.Fatalf("Compensated = %d, want 1 (StartCandidate)", summary.Compensated)
	}
	if execSvc.stoppedCandidateWith.PID != 2000 {
		t.Errorf("stopped candidate PID = %d, want 2000", execSvc.stoppedCandidateWith.PID)
	}
}

func containsPort(config string, port int) bool {
	return strings.Contains(config, ":"+strconv.Itoa(port))
}
