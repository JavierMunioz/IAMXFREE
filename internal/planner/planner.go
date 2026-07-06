package planner

import "github.com/JavierMunioz/IAMXFREE/internal/inspection"

// Planner interprets one technology's inspection.Detection into a
// DeploymentPlan. Each supported technology implements its own Planner
// (node_planner.go, python_planner.go, ...); none of that logic lives here.
type Planner interface {
	// Name identifies this planner (e.g. "node", "docker").
	Name() string

	// CanPlan reports whether this planner knows how to interpret
	// detection. It is a pure decision — no I/O.
	CanPlan(detection inspection.Detection) bool

	// Plan builds a DeploymentPlan from detection. result is the full
	// inspection.Result detection came from, so a Planner can note
	// complementary detections (e.g. a Dockerfile found alongside a Node
	// project) without needing to re-run detection itself. Plan never
	// fails: anything it cannot determine is left blank with a matching
	// Warning explaining why.
	Plan(detection inspection.Detection, result inspection.Result) DeploymentPlan
}
