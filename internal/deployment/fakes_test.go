package deployment

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// fakeAppService is a minimal services.ApplicationService test double.
type fakeAppService struct {
	app    *models.Application
	getErr error

	updatedConfig     models.DeploymentConfig
	updateConfigErr   error
	updateConfigCalls int
}

func (f *fakeAppService) Register(context.Context, *models.Application) error { return nil }
func (f *fakeAppService) Get(_ context.Context, id string) (*models.Application, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.app, nil
}
func (f *fakeAppService) List(context.Context) ([]*models.Application, error) { return nil, nil }
func (f *fakeAppService) UpdateConfig(_ context.Context, _ string, cfg models.DeploymentConfig) (*models.Application, error) {
	f.updatedConfig = cfg
	f.updateConfigCalls++
	if f.updateConfigErr != nil {
		return nil, f.updateConfigErr
	}
	return f.app, nil
}
func (f *fakeAppService) ChangeStatus(context.Context, string, models.ApplicationStatus) (*models.Application, error) {
	return nil, nil
}
func (f *fakeAppService) Remove(context.Context, string) error { return nil }
func (f *fakeAppService) ResolveExecutionStrategy(context.Context, string) (execution.Metadata, error) {
	return execution.Metadata{}, execution.ErrNoStrategyFound
}
func (f *fakeAppService) CheckExecutionHealth(context.Context, string) (services.ExecutionHealth, error) {
	return services.ExecutionHealth{}, execution.ErrNoStrategyFound
}
func (f *fakeAppService) CheckGitStatus(context.Context, string) (services.GitStatus, error) {
	return services.GitStatus{}, nil
}

// fakeExecutionService is a minimal services.ExecutionService test double.
type fakeExecutionService struct {
	session services.RunSession
	running bool

	installFn  func(ctx context.Context, appID string) error
	installErr error
	buildFn    func(ctx context.Context, appID string) error
	buildErr   error

	startFn      func(ctx context.Context, appID string) (services.RunSession, error)
	startSession services.RunSession
	startErr     error
	stopErr      error
	stoppedWith  services.RunSession

	startCandidateFn      func(ctx context.Context, appID string, port int) (services.RunSession, error)
	startCandidateSession services.RunSession
	startCandidateErr     error

	candidateSession services.RunSession
	hasCandidate     bool

	stopCandidateErr     error
	stoppedCandidateWith services.RunSession

	promoteErr error
	promoted   bool

	checkStatusFn      func(ctx context.Context, appID string, session services.RunSession) (services.RunSession, error)
	checkStatusSession services.RunSession
	checkStatusErr     error

	stopSessionErr     error
	stoppedSessionWith services.RunSession
}

func (f *fakeExecutionService) Install(ctx context.Context, appID string) error {
	if f.installFn != nil {
		return f.installFn(ctx, appID)
	}
	return f.installErr
}
func (f *fakeExecutionService) Build(ctx context.Context, appID string) error {
	if f.buildFn != nil {
		return f.buildFn(ctx, appID)
	}
	return f.buildErr
}
func (f *fakeExecutionService) Start(ctx context.Context, appID string) (services.RunSession, error) {
	if f.startFn != nil {
		return f.startFn(ctx, appID)
	}
	return f.startSession, f.startErr
}
func (f *fakeExecutionService) Stop(_ context.Context, _ string, session services.RunSession) error {
	f.stoppedWith = session
	return f.stopErr
}
func (f *fakeExecutionService) RefreshSession(context.Context, string, services.RunSession) (services.RunSession, error) {
	return services.RunSession{}, nil
}
func (f *fakeExecutionService) OpenLogs(context.Context, string, services.RunSession) (services.LogStream, error) {
	return nil, nil
}
func (f *fakeExecutionService) Snapshot(context.Context, services.RunSession) (services.RuntimeSnapshot, error) {
	return services.RuntimeSnapshot{}, nil
}
func (f *fakeExecutionService) ActiveSession(string) (services.RunSession, bool) {
	return f.session, f.running
}
func (f *fakeExecutionService) StartCandidate(ctx context.Context, appID string, port int) (services.RunSession, error) {
	if f.startCandidateFn != nil {
		return f.startCandidateFn(ctx, appID, port)
	}
	return f.startCandidateSession, f.startCandidateErr
}
func (f *fakeExecutionService) CandidateSession(string) (services.RunSession, bool) {
	return f.candidateSession, f.hasCandidate
}
func (f *fakeExecutionService) StopCandidate(_ context.Context, _ string, session services.RunSession) error {
	f.stoppedCandidateWith = session
	return f.stopCandidateErr
}
func (f *fakeExecutionService) PromoteCandidate(context.Context, string) error {
	f.promoted = true
	return f.promoteErr
}
func (f *fakeExecutionService) CheckStatus(ctx context.Context, appID string, session services.RunSession) (services.RunSession, error) {
	if f.checkStatusFn != nil {
		return f.checkStatusFn(ctx, appID, session)
	}
	return f.checkStatusSession, f.checkStatusErr
}
func (f *fakeExecutionService) StopSession(_ context.Context, _ string, session services.RunSession) error {
	f.stoppedSessionWith = session
	return f.stopSessionErr
}
