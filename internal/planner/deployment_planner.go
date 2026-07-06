package planner

import (
	"fmt"

	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
)

// DeploymentPlanner turns an inspection.Result into a single DeploymentPlan.
// It never fails: when nothing was detected, or no registered Planner
// recognizes the primary detection, it still returns a DeploymentPlan —
// just one whose Warnings explain why it is mostly empty.
type DeploymentPlanner struct {
	registry *Registry
}

// NewDeploymentPlanner builds a DeploymentPlanner backed by registry.
func NewDeploymentPlanner(registry *Registry) *DeploymentPlanner {
	return &DeploymentPlanner{registry: registry}
}

// Plan builds a DeploymentPlan from result, using result.Primary() to pick
// which detection drives the proposal. Complementary detections (e.g. a
// Dockerfile found alongside a Node project) are still available to the
// chosen Planner via result.
func (p *DeploymentPlanner) Plan(result inspection.Result) DeploymentPlan {
	primary, ok := result.Primary()
	if !ok {
		return DeploymentPlan{
			Confidence: inspection.ConfidenceLow,
			Warnings:   []string{"no technology could be detected in this directory"},
		}
	}

	for _, pl := range p.registry.Planners() {
		if pl.CanPlan(primary) {
			return pl.Plan(primary, result)
		}
	}

	return DeploymentPlan{
		Confidence: inspection.ConfidenceLow,
		Warnings:   []string{fmt.Sprintf("no planner is registered for %q projects yet", primary.Type)},
	}
}
