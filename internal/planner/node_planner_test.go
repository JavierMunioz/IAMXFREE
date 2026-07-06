package planner

import (
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func containsNote(notes []string, substr string) bool {
	for _, n := range notes {
		if strings.Contains(n, substr) {
			return true
		}
	}
	return false
}

func TestNodePlannerCanPlanOnlyNode(t *testing.T) {
	p := NewNodePlanner()
	if !p.CanPlan(inspection.Detection{Type: inspection.ProjectTypeNode}) {
		t.Error("expected CanPlan to be true for a node detection")
	}
	if p.CanPlan(inspection.Detection{Type: inspection.ProjectTypePython}) {
		t.Error("expected CanPlan to be false for a non-node detection")
	}
}

func fullNodeDetection() inspection.Detection {
	return inspection.Detection{
		Type:           inspection.ProjectTypeNode,
		Runtime:        models.RuntimeNode,
		Framework:      models.FrameworkReact,
		PackageManager: "pnpm",
		Name:           "my-api",
		Scripts:        map[string]string{"build": "vite build", "start": "vite preview --port 4173"},
		MatchedFiles:   []string{"package.json", "pnpm-lock.yaml"},
		Dependencies:   []string{"react", "vite"},
		Suggested: inspection.SuggestedCommands{
			Install: "pnpm install",
			Build:   "pnpm run build",
			Start:   "pnpm start",
		},
		Confidence: inspection.ConfidenceHigh,
	}
}

func TestNodePlannerHappyPath(t *testing.T) {
	plan := NewNodePlanner().Plan(fullNodeDetection(), inspection.Result{})

	if plan.SuggestedName != "my-api" {
		t.Errorf("SuggestedName = %q, want %q", plan.SuggestedName, "my-api")
	}
	if plan.Framework != models.FrameworkReact || plan.Runtime != models.RuntimeNode {
		t.Errorf("Framework/Runtime = %q/%q", plan.Framework, plan.Runtime)
	}
	if plan.PackageManager != "pnpm" {
		t.Errorf("PackageManager = %q, want %q", plan.PackageManager, "pnpm")
	}
	if plan.InstallCommand != "pnpm install" || plan.BuildCommand != "pnpm run build" || plan.StartCommand != "pnpm start" {
		t.Errorf("commands = %+v", plan)
	}
	if plan.SuggestedPort != 4173 {
		t.Errorf("SuggestedPort = %d, want 4173", plan.SuggestedPort)
	}
	if plan.Type != models.ApplicationTypeFrontend {
		t.Errorf("Type = %q, want %q", plan.Type, models.ApplicationTypeFrontend)
	}
	if plan.Domain != "" {
		t.Errorf("Domain = %q, want empty", plan.Domain)
	}
	if len(plan.Warnings) != 0 {
		t.Errorf("Warnings = %v, want empty", plan.Warnings)
	}
	if plan.Confidence != inspection.ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q (no warnings, no downgrade)", plan.Confidence, inspection.ConfidenceHigh)
	}
	if len(plan.Dependencies) != 2 {
		t.Errorf("Dependencies = %v", plan.Dependencies)
	}
}

func TestNodePlannerMissingNameWarns(t *testing.T) {
	detection := fullNodeDetection()
	detection.Name = ""

	plan := NewNodePlanner().Plan(detection, inspection.Result{})
	if !containsNote(plan.Warnings, "application name") {
		t.Errorf("expected a warning about the missing name, got %v", plan.Warnings)
	}
	if plan.Confidence != inspection.ConfidenceMedium {
		t.Errorf("Confidence = %q, want %q (downgraded)", plan.Confidence, inspection.ConfidenceMedium)
	}
}

func TestNodePlannerMissingPackageManagerWarns(t *testing.T) {
	detection := fullNodeDetection()
	detection.PackageManager = ""
	detection.Suggested.Install = ""

	plan := NewNodePlanner().Plan(detection, inspection.Result{})
	if !containsNote(plan.Warnings, "package manager") {
		t.Errorf("expected a warning about the missing package manager, got %v", plan.Warnings)
	}
	if plan.InstallCommand != "" {
		t.Errorf("InstallCommand = %q, want empty", plan.InstallCommand)
	}
}

func TestNodePlannerMissingBuildScriptIsNoteNotWarning(t *testing.T) {
	detection := fullNodeDetection()
	delete(detection.Scripts, "build")
	detection.Suggested.Build = ""

	plan := NewNodePlanner().Plan(detection, inspection.Result{})
	if plan.BuildCommand != "" {
		t.Errorf("BuildCommand = %q, want empty", plan.BuildCommand)
	}
	if !containsNote(plan.Notes, "build") {
		t.Errorf("expected a note about the missing build script, got %v", plan.Notes)
	}
	if containsNote(plan.Warnings, "build") {
		t.Errorf("missing build script should be a Note, not a Warning: %v", plan.Warnings)
	}
}

func TestNodePlannerMissingStartScriptWarns(t *testing.T) {
	detection := fullNodeDetection()
	delete(detection.Scripts, "start")
	detection.Suggested.Start = ""

	plan := NewNodePlanner().Plan(detection, inspection.Result{})
	if plan.StartCommand != "" {
		t.Errorf("StartCommand = %q, want empty", plan.StartCommand)
	}
	if plan.SuggestedPort != 0 {
		t.Errorf("SuggestedPort = %d, want 0", plan.SuggestedPort)
	}
	if !containsNote(plan.Warnings, `"start" script`) {
		t.Errorf("expected a warning about the missing start script, got %v", plan.Warnings)
	}
}

func TestNodePlannerStartScriptWithoutPortWarns(t *testing.T) {
	detection := fullNodeDetection()
	detection.Scripts["start"] = "node server.js"

	plan := NewNodePlanner().Plan(detection, inspection.Result{})
	if plan.SuggestedPort != 0 {
		t.Errorf("SuggestedPort = %d, want 0", plan.SuggestedPort)
	}
	if !containsNote(plan.Warnings, "port could not be inferred") {
		t.Errorf("expected a warning about the unknown port, got %v", plan.Warnings)
	}
}

func TestNodePlannerBackendFramework(t *testing.T) {
	detection := fullNodeDetection()
	detection.Framework = models.FrameworkExpress

	plan := NewNodePlanner().Plan(detection, inspection.Result{})
	if plan.Type != models.ApplicationTypeBackend {
		t.Errorf("Type = %q, want %q", plan.Type, models.ApplicationTypeBackend)
	}
}

func TestNodePlannerHybridFrameworkLeavesTypeUnclassified(t *testing.T) {
	detection := fullNodeDetection()
	detection.Framework = models.FrameworkNextJS

	plan := NewNodePlanner().Plan(detection, inspection.Result{})
	if plan.Type != "" {
		t.Errorf("Type = %q, want empty for a hybrid meta-framework", plan.Type)
	}
	if !containsNote(plan.Notes, "application type") {
		t.Errorf("expected a note explaining the unclassified type, got %v", plan.Notes)
	}
}

func TestNodePlannerNotesSiblingDockerDetection(t *testing.T) {
	result := inspection.Result{Detections: []inspection.Detection{
		{Type: inspection.ProjectTypeDocker, Confidence: inspection.ConfidenceHigh},
	}}

	plan := NewNodePlanner().Plan(fullNodeDetection(), result)
	if !containsNote(plan.Notes, "Docker") {
		t.Errorf("expected a note about the sibling Docker detection, got %v", plan.Notes)
	}
}

func TestNodePlannerNeverInventsAScript(t *testing.T) {
	detection := fullNodeDetection()
	detection.Scripts = map[string]string{"test": "jest"} // neither build nor start
	detection.Suggested = inspection.SuggestedCommands{Install: "pnpm install"}

	plan := NewNodePlanner().Plan(detection, inspection.Result{})
	if plan.BuildCommand != "" || plan.StartCommand != "" {
		t.Errorf("expected no build/start command to be invented, got %+v", plan)
	}
}
