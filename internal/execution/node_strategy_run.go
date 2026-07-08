package execution

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
)

// Start runs app's configured start command as a background process via
// runtimehost.Host — never systemd, PM2 or Docker; IAMXFREE manages it
// directly. It refuses to start when Readiness reports the application is
// not ready, so a caller never has to duplicate that check. port is passed
// to the process as a PORT environment variable — the convention most Node
// servers (Express, Next.js, etc.) read to decide where to listen — which
// is what lets two sessions of the same application coexist on different
// ports during a zero-downtime deployment.
func (s *nodeStrategy) Start(ctx context.Context, app *models.Application, port int) (Session, error) {
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
		Env:  []string{"PORT=" + strconv.Itoa(port)},
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
		Port:       port,
	}, nil
}

// Stop terminates the process described by session via runtimehost.Host.
// Automatic restart-on-stop is explicitly out of scope for this iteration.
func (s *nodeStrategy) Stop(_ context.Context, _ *models.Application, session Session) error {
	return s.host.StopProcess(session.PID)
}

// Status re-checks session's process via runtimehost.Host and returns an
// updated Session reflecting whether it is still running.
func (s *nodeStrategy) Status(_ context.Context, _ *models.Application, session Session) (Session, error) {
	running, err := s.host.IsProcessRunning(session.PID)
	if err != nil {
		return Session{}, err
	}

	updated := session
	if running {
		updated.Status = StatusRunning
	} else {
		updated.Status = StatusStopped
	}
	return updated, nil
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
