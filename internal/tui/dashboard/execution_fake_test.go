package dashboard

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// fakeExecutionService is a minimal services.ExecutionService test double.
type fakeExecutionService struct {
	sessionByID map[string]services.RunSession
}

func (f *fakeExecutionService) Start(context.Context, string) (services.RunSession, error) {
	return services.RunSession{}, nil
}
func (f *fakeExecutionService) Stop(context.Context, string, services.RunSession) error { return nil }
func (f *fakeExecutionService) RefreshSession(context.Context, string, services.RunSession) (services.RunSession, error) {
	return services.RunSession{}, nil
}
func (f *fakeExecutionService) OpenLogs(context.Context, string, services.RunSession) (services.LogStream, error) {
	return nil, nil
}
func (f *fakeExecutionService) Snapshot(context.Context, services.RunSession) (services.RuntimeSnapshot, error) {
	return services.RuntimeSnapshot{}, nil
}
func (f *fakeExecutionService) ActiveSession(appID string) (services.RunSession, bool) {
	session, ok := f.sessionByID[appID]
	return session, ok
}
