package inspection

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestPHPDetectorNotFoundWithoutComposerJSON(t *testing.T) {
	dir := t.TempDir()
	if _, ok := NewPHPDetector().Detect(buildInput(t, dir)); ok {
		t.Fatal("expected no detection without composer.json")
	}
}

func TestPHPDetectorLaravel(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{
		"name": "acme/my-api",
		"version": "1.0.0",
		"require": {"laravel/framework": "^10.0"}
	}`)

	detection, ok := NewPHPDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Runtime != models.RuntimePHP {
		t.Errorf("Runtime = %q, want %q", detection.Runtime, models.RuntimePHP)
	}
	if detection.PackageManager != "composer" {
		t.Errorf("PackageManager = %q, want %q", detection.PackageManager, "composer")
	}
	if detection.Name != "acme/my-api" || detection.Version != "1.0.0" {
		t.Errorf("Name/Version = %q/%q, want %q/%q", detection.Name, detection.Version, "acme/my-api", "1.0.0")
	}
	if detection.Framework != models.FrameworkLaravel {
		t.Errorf("Framework = %q, want %q", detection.Framework, models.FrameworkLaravel)
	}
	if detection.Suggested.Start != "php artisan serve" {
		t.Errorf("Suggested.Start = %q, want %q", detection.Suggested.Start, "php artisan serve")
	}
}

func TestPHPDetectorUnknownFramework(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{"name": "acme/my-api", "version": "1.0.0"}`)

	detection, ok := NewPHPDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Framework != "" {
		t.Errorf("Framework = %q, want empty", detection.Framework)
	}
	if !containsNote(detection.Notes, "framework") {
		t.Errorf("expected a note about the undetermined framework, got %v", detection.Notes)
	}
}
