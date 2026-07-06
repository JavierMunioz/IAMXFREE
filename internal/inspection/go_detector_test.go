package inspection

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestGoDetectorNotFoundWithoutGoMod(t *testing.T) {
	dir := t.TempDir()
	if _, ok := NewGoDetector().Detect(buildInput(t, dir)); ok {
		t.Fatal("expected no detection without go.mod")
	}
}

func TestGoDetectorParsesModuleName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module github.com/example/my-api\n\ngo 1.24\n")

	detection, ok := NewGoDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Runtime != models.RuntimeGo {
		t.Errorf("Runtime = %q, want %q", detection.Runtime, models.RuntimeGo)
	}
	if detection.Name != "github.com/example/my-api" {
		t.Errorf("Name = %q, want %q", detection.Name, "github.com/example/my-api")
	}
	if detection.PackageManager != "go modules" {
		t.Errorf("PackageManager = %q, want %q", detection.PackageManager, "go modules")
	}
	if detection.Version != "" {
		t.Errorf("Version = %q, want empty (go.mod has no project version)", detection.Version)
	}
	if detection.Suggested.Build != "go build ./..." {
		t.Errorf("Suggested.Build = %q, want %q", detection.Suggested.Build, "go build ./...")
	}
}

func TestGoDetectorMissingModuleLineIsNoted(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "go 1.24\n")

	detection, ok := NewGoDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Name != "" {
		t.Errorf("Name = %q, want empty", detection.Name)
	}
	if !containsNote(detection.Notes, "module name") {
		t.Errorf("expected a note about the missing module name, got %v", detection.Notes)
	}
}
