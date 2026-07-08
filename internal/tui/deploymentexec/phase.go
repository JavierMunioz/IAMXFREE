package deploymentexec

import "github.com/JavierMunioz/IAMXFREE/internal/operations"

// runPhase is the coarse, always-visible stage a deployment run is in —
// what the user needs to understand at a glance, distinct from any single
// operation's state.
type runPhase string

const (
	phaseExecuting    runPhase = "Executing"
	phaseFailed       runPhase = "Failed"
	phaseCompensating runPhase = "Compensating"
	phaseFinished     runPhase = "Finished"
)

// currentPhase derives the run's phase from its accumulated results: once
// finished it's always Finished; otherwise it's Compensating as soon as
// any compensation has started, Failed as soon as any operation failed
// (before compensation begins), or Executing.
func currentPhase(results []operations.OperationResult, finished bool) runPhase {
	if finished {
		return phaseFinished
	}

	for _, r := range results {
		if r.Compensation != nil {
			return phaseCompensating
		}
	}

	for _, r := range results {
		if r.State == operations.StateFailed {
			return phaseFailed
		}
	}

	return phaseExecuting
}
