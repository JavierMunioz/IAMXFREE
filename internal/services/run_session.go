package services

import (
	"time"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// RunSession is what callers (the detail view, most notably) see about a
// currently-tracked execution session — a plain projection of
// execution.Session, so internal/tui never needs to import
// internal/execution to hold onto one between calls.
type RunSession struct {
	PID        int
	StartedAt  time.Time
	Command    string
	Args       []string
	WorkingDir string
	Status     string
	Runtime    models.Runtime
}
