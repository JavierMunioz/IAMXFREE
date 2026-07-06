package execution

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// Start runs app's configured start command as a background process via
// runtimehost.Host — never systemd, PM2 or Docker; IAMXFREE manages it
// directly. It refuses to start when Readiness reports the application is
// not ready, so a caller never has to duplicate that check.
func (s *nodeStrategy) Start(ctx context.Context, app *models.Application) (Session, error) {
	readiness, err := s.Readiness(ctx, app)
	if err != nil {
		return Session{}, err
	}
	if !readiness.Ready {
		return Session{}, fmt.Errorf("execution: node: application is not ready to start: %v", readiness.BlockingErrors)
	}

	name, args := splitCommand(app.Config.StartCommand)

	pid, err := s.host.StartProcess(ctx, runtimehost.Command{
		Name: name,
		Args: args,
		Dir:  app.Source.LocalPath,
	})
	if err != nil {
		return Session{}, err
	}

	return Session{
		PID:        pid,
		StartedAt:  time.Now().UTC(),
		Command:    name,
		Args:       args,
		WorkingDir: app.Source.LocalPath,
		Status:     StatusRunning,
		Runtime:    app.Runtime,
	}, nil
}

// Stop terminates the process described by session via runtimehost.Host.
// Automatic restart-on-stop is explicitly out of scope for this iteration.
func (s *nodeStrategy) Stop(_ context.Context, _ *models.Application, session Session) error {
	return s.host.StopProcess(session.PID)
}

// splitCommand breaks a configured command string (e.g. "npm start") into
// the executable name and its arguments. It is a plain whitespace split, not
// a full shell parser — a configured command containing quoted arguments
// with embedded spaces is not supported yet.
func splitCommand(command string) (name string, args []string) {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "", nil
	}
	return fields[0], fields[1:]
}
