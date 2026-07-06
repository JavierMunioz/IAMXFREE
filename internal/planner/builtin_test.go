package planner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestNewDefaultRegistryRegistersNodePlanner(t *testing.T) {
	registry := NewDefaultRegistry()
	got := registry.Planners()

	if len(got) != 1 {
		t.Fatalf("len(Planners()) = %d, want 1 (only node is implemented so far)", len(got))
	}
	if got[0].Name() != "node" {
		t.Fatalf("Planners()[0].Name() = %q, want %q", got[0].Name(), "node")
	}
}

// TestEndToEndInspectionToPlan proves the full pipeline works together:
// inspection.Inspector finds a real Node project on disk, and this
// package's DeploymentPlanner turns that into a DeploymentPlan, without
// planner ever touching the filesystem itself.
func TestEndToEndInspectionToPlan(t *testing.T) {
	dir := t.TempDir()
	writeProjectFile(t, dir, "package.json", `{
		"name": "my-api",
		"scripts": {"start": "node server.js --port 3000"},
		"dependencies": {"express": "^4.0.0"}
	}`)
	writeProjectFile(t, dir, "package-lock.json", "{}")

	result, err := inspection.NewInspector(inspection.NewDefaultRegistry()).Inspect(dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	plan := NewDeploymentPlanner(NewDefaultRegistry()).Plan(result)

	if plan.SuggestedName != "my-api" {
		t.Errorf("SuggestedName = %q, want %q", plan.SuggestedName, "my-api")
	}
	if plan.PackageManager != "npm" {
		t.Errorf("PackageManager = %q, want %q", plan.PackageManager, "npm")
	}
	if plan.Type != models.ApplicationTypeBackend {
		t.Errorf("Type = %q, want %q", plan.Type, models.ApplicationTypeBackend)
	}
	if plan.SuggestedPort != 3000 {
		t.Errorf("SuggestedPort = %d, want 3000", plan.SuggestedPort)
	}
	if plan.StartCommand != "npm start" {
		t.Errorf("StartCommand = %q, want %q", plan.StartCommand, "npm start")
	}
}

func writeProjectFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", name, err)
	}
}
