package planner

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
)

func TestDeploymentPlannerNoDetectionsWarns(t *testing.T) {
	dp := NewDeploymentPlanner(NewRegistry())
	plan := dp.Plan(inspection.Result{})

	if len(plan.Warnings) == 0 {
		t.Fatal("expected a warning when nothing was detected")
	}
	if plan.Confidence != inspection.ConfidenceLow {
		t.Errorf("Confidence = %q, want %q", plan.Confidence, inspection.ConfidenceLow)
	}
}

func TestDeploymentPlannerNoMatchingPlannerWarns(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&fakePlanner{name: "python", projectType: inspection.ProjectTypePython})

	dp := NewDeploymentPlanner(registry)
	result := inspection.Result{Detections: []inspection.Detection{
		{Type: inspection.ProjectTypeNode, Confidence: inspection.ConfidenceHigh},
	}}

	plan := dp.Plan(result)
	if len(plan.Warnings) == 0 {
		t.Fatal("expected a warning when no registered planner matches")
	}
}

func TestDeploymentPlannerDelegatesToMatchingPlanner(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&fakePlanner{name: "node", projectType: inspection.ProjectTypeNode})

	dp := NewDeploymentPlanner(registry)
	result := inspection.Result{Detections: []inspection.Detection{
		{Type: inspection.ProjectTypeNode, Name: "my-api", Confidence: inspection.ConfidenceHigh},
	}}

	plan := dp.Plan(result)
	if plan.SuggestedName != "my-api" {
		t.Errorf("SuggestedName = %q, want %q", plan.SuggestedName, "my-api")
	}
	if len(plan.Warnings) != 0 {
		t.Errorf("Warnings = %v, want empty", plan.Warnings)
	}
}

func TestDeploymentPlannerUsesPrimaryAmongMultipleDetections(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&fakePlanner{name: "node", projectType: inspection.ProjectTypeNode})

	dp := NewDeploymentPlanner(registry)
	result := inspection.Result{Detections: []inspection.Detection{
		{Type: inspection.ProjectTypeDocker, Confidence: inspection.ConfidenceMedium},
		{Type: inspection.ProjectTypeNode, Name: "my-api", Confidence: inspection.ConfidenceHigh},
	}}

	plan := dp.Plan(result)
	if plan.SuggestedName != "my-api" {
		t.Fatalf("expected the node detection to be used as primary, got plan %+v", plan)
	}
}
