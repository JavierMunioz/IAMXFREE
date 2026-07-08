package deployment

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/operations"
)

// stopOperation stops the application's currently tracked session, if any.
// It is only applicable when one is running — there is nothing to stop
// otherwise, and calling Stop with no session would be meaningless. It
// compensates with Start: if a later operation fails, the application is
// put back in the running state it was in before this deployment touched
// it.
func (e *Engine) stopOperation(app *models.Application) operations.Operation {
	op := operations.Operation{Name: "Stop application", Component: string(ComponentExecution), Method: "Stop"}

	session, running := e.executionService.ActiveSession(app.ID)
	if !running {
		op.SkipReason = "application is not currently running"
		return op
	}

	op.Applicable = true
	op.Run = func(ctx context.Context) error {
		return e.executionService.Stop(ctx, app.ID, session)
	}
	op.Compensate = func(ctx context.Context) error {
		_, err := e.executionService.Start(ctx, app.ID)
		return err
	}
	return op
}

// startOperation starts the application. Unlike stopOperation, it is
// always applicable: a deployment's purpose is to end with the
// application running, whatever its state was before.
func (e *Engine) startOperation(app *models.Application) operations.Operation {
	return operations.Operation{
		Name: "Start application", Component: string(ComponentExecution), Method: "Start",
		Applicable: true,
		Run: func(ctx context.Context) error {
			_, err := e.executionService.Start(ctx, app.ID)
			return err
		},
	}
}
