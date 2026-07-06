package planner

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
)

// fakePlanner is used only to exercise the Planner contract in isolation
// from any real technology.
type fakePlanner struct {
	projectType inspection.ProjectType
	name        string
}

func (p *fakePlanner) Name() string { return p.name }

func (p *fakePlanner) CanPlan(detection inspection.Detection) bool {
	return detection.Type == p.projectType
}

func (p *fakePlanner) Plan(detection inspection.Detection, _ inspection.Result) DeploymentPlan {
	return DeploymentPlan{
		SuggestedName: detection.Name,
		Confidence:    detection.Confidence,
	}
}

var _ Planner = (*fakePlanner)(nil)

func TestFakePlannerCanPlan(t *testing.T) {
	p := &fakePlanner{projectType: inspection.ProjectTypeNode, name: "fake"}

	if !p.CanPlan(inspection.Detection{Type: inspection.ProjectTypeNode}) {
		t.Error("expected CanPlan to be true for a matching project type")
	}
	if p.CanPlan(inspection.Detection{Type: inspection.ProjectTypePython}) {
		t.Error("expected CanPlan to be false for a non-matching project type")
	}
}

func TestFakePlannerPlan(t *testing.T) {
	p := &fakePlanner{projectType: inspection.ProjectTypeNode, name: "fake"}
	detection := inspection.Detection{
		Type:       inspection.ProjectTypeNode,
		Name:       "my-api",
		Confidence: inspection.ConfidenceHigh,
	}

	plan := p.Plan(detection, inspection.Result{})
	if plan.SuggestedName != "my-api" {
		t.Errorf("SuggestedName = %q, want %q", plan.SuggestedName, "my-api")
	}
	if plan.Confidence != inspection.ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q", plan.Confidence, inspection.ConfidenceHigh)
	}
}

func TestDeploymentPlanZeroValueIsUsable(t *testing.T) {
	var plan DeploymentPlan
	if plan.SuggestedName != "" || plan.Domain != "" || len(plan.Warnings) != 0 {
		t.Fatalf("expected a zero-value DeploymentPlan to be empty, got %+v", plan)
	}
}
