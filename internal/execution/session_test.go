package execution_test

import (
	"testing"
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestSessionFields(t *testing.T) {
	startedAt := time.Now().UTC()
	session := execution.Session{
		PID:        4242,
		StartedAt:  startedAt,
		Command:    "npm",
		Args:       []string{"start"},
		WorkingDir: "/srv/apps/my-api",
		Status:     execution.StatusRunning,
		Runtime:    models.RuntimeNode,
	}

	if session.PID != 4242 {
		t.Errorf("PID = %d, want 4242", session.PID)
	}
	if !session.StartedAt.Equal(startedAt) {
		t.Errorf("StartedAt = %v, want %v", session.StartedAt, startedAt)
	}
	if session.Command != "npm" || len(session.Args) != 1 || session.Args[0] != "start" {
		t.Errorf("Command/Args = %q/%v, want npm/[start]", session.Command, session.Args)
	}
	if session.Status != execution.StatusRunning {
		t.Errorf("Status = %q, want %q", session.Status, execution.StatusRunning)
	}
	if session.Runtime != models.RuntimeNode {
		t.Errorf("Runtime = %q, want %q", session.Runtime, models.RuntimeNode)
	}
}

func TestSessionZeroValueIsUsable(t *testing.T) {
	var session execution.Session
	if session.PID != 0 || session.Status != "" {
		t.Fatalf("expected a zero-value Session to be empty, got %+v", session)
	}
}
