package deployment

import (
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

// DeploymentContext is shared, mutable state threaded through a single
// zero-downtime deployment's operations — each operation's closure reads
// what an earlier one decided and records what it itself produced, instead
// of every operation rediscovering information or BuildOperations passing
// an ever-growing parameter list between operation builders.
//
// It exists only for the lifetime of one BuildOperations call and the
// operations.Executor.Execute run that follows it; it is never persisted
// or shared across deployments.
type DeploymentContext struct {
	App  *models.Application
	Plan DeploymentPlan

	// ActivePort/CandidatePort are decided once, before the operation
	// list is built (see activePortFor/candidatePortFor), and read by
	// every operation that needs to know which port to use.
	ActivePort    int
	CandidatePort int

	// PreviousSession is the session that was active before this
	// deployment started, captured before ConfirmSwitch promotes the
	// candidate over it — so StopOldSession can still stop it afterward,
	// once nothing tracks it under any key anymore.
	PreviousSession services.RunSession

	// CandidateSession is filled in by StartCandidateOperation once it
	// actually starts the new session, for every operation after it
	// (health check, Nginx switch, confirm) to read.
	CandidateSession services.RunSession

	// Warnings accumulates problems that must never be hidden but also
	// must never fail — and roll back — an otherwise-successful
	// deployment (e.g. the previous session failing to stop cleanly
	// after traffic already switched to the new one).
	Warnings []string
}
