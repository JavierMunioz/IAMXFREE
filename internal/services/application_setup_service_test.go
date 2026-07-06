package services_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/planner"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
)

func newSetupService() services.ApplicationSetupService {
	inspector := inspection.NewInspector(inspection.NewDefaultRegistry())
	deploymentPlanner := planner.NewDeploymentPlanner(planner.NewDefaultRegistry())
	return services.NewApplicationSetupService(inspector, deploymentPlanner)
}

func writeSetupFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", name, err)
	}
}

func TestApplicationSetupServiceInspectNodeProject(t *testing.T) {
	dir := t.TempDir()
	writeSetupFile(t, dir, "package.json", `{
		"name": "my-api",
		"scripts": {"start": "node server.js --port 4000"},
		"dependencies": {"express": "^4.0.0"}
	}`)
	writeSetupFile(t, dir, "package-lock.json", "{}")

	proposal, err := newSetupService().Inspect(context.Background(), dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	if proposal.Path != dir {
		t.Errorf("Path = %q, want %q", proposal.Path, dir)
	}
	if proposal.ProjectType != "node" {
		t.Errorf("ProjectType = %q, want %q", proposal.ProjectType, "node")
	}
	if proposal.SuggestedName != "my-api" {
		t.Errorf("SuggestedName = %q, want %q", proposal.SuggestedName, "my-api")
	}
	if proposal.Type != models.ApplicationTypeBackend {
		t.Errorf("Type = %q, want %q", proposal.Type, models.ApplicationTypeBackend)
	}
	if proposal.PackageManager != "npm" {
		t.Errorf("PackageManager = %q, want %q", proposal.PackageManager, "npm")
	}
	if proposal.SuggestedPort != 4000 {
		t.Errorf("SuggestedPort = %d, want 4000", proposal.SuggestedPort)
	}
	if proposal.StartCommand != "npm start" {
		t.Errorf("StartCommand = %q, want %q", proposal.StartCommand, "npm start")
	}
	if proposal.Confidence != "high" {
		t.Errorf("Confidence = %q, want %q", proposal.Confidence, "high")
	}
}

func TestApplicationSetupServiceInspectEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	proposal, err := newSetupService().Inspect(context.Background(), dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if proposal.ProjectType != "" {
		t.Errorf("ProjectType = %q, want empty", proposal.ProjectType)
	}
	if len(proposal.Warnings) == 0 {
		t.Error("expected a warning explaining nothing was detected")
	}
}

func TestApplicationSetupServiceInspectNonexistentPath(t *testing.T) {
	if _, err := newSetupService().Inspect(context.Background(), "/does/not/exist/at/all"); err == nil {
		t.Fatal("expected an error for a nonexistent path")
	}
}
