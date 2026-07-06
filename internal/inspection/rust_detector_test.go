package inspection

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestRustDetectorNotFoundWithoutCargoToml(t *testing.T) {
	dir := t.TempDir()
	if _, ok := NewRustDetector().Detect(buildInput(t, dir)); ok {
		t.Fatal("expected no detection without Cargo.toml")
	}
}

func TestRustDetectorParsesPackageTable(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cargo.toml", `[package]
name = "my-api"
version = "0.1.0"
edition = "2021"

[dependencies]
tokio = "1"
`)

	detection, ok := NewRustDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Runtime != models.RuntimeRust {
		t.Errorf("Runtime = %q, want %q", detection.Runtime, models.RuntimeRust)
	}
	if detection.PackageManager != "cargo" {
		t.Errorf("PackageManager = %q, want %q", detection.PackageManager, "cargo")
	}
	if detection.Name != "my-api" || detection.Version != "0.1.0" {
		t.Errorf("Name/Version = %q/%q, want %q/%q", detection.Name, detection.Version, "my-api", "0.1.0")
	}
	if detection.Confidence != ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q", detection.Confidence, ConfidenceHigh)
	}
	if detection.Suggested.Build != "cargo build" || detection.Suggested.Start != "cargo run" {
		t.Errorf("Suggested = %+v", detection.Suggested)
	}
}

func TestRustDetectorIgnoresFieldsOutsidePackageTable(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cargo.toml", `[dependencies]
name = "not-the-package-name"

[package]
name = "my-api"
`)

	detection, ok := NewRustDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Name != "my-api" {
		t.Errorf("Name = %q, want %q", detection.Name, "my-api")
	}
}

func TestRustDetectorMissingFieldsAreNoted(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cargo.toml", "[workspace]\nmembers = []\n")

	detection, ok := NewRustDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Name != "" || detection.Version != "" {
		t.Errorf("Name/Version = %q/%q, want empty/empty", detection.Name, detection.Version)
	}
	if !containsNote(detection.Notes, "package name") {
		t.Errorf("expected a note about the missing package name, got %v", detection.Notes)
	}
}
