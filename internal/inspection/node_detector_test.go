package inspection

import (
	"strings"
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestNodeDetectorNotFoundWithoutPackageJSON(t *testing.T) {
	dir := t.TempDir()
	input := buildInput(t, dir)

	if _, ok := NewNodeDetector().Detect(input); ok {
		t.Fatal("expected no detection without package.json")
	}
}

func TestNodeDetectorFullDetection(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{
		"name": "my-api",
		"version": "1.2.3",
		"scripts": {"build": "next build", "start": "next start"},
		"dependencies": {"next": "^14.0.0", "react": "^18.0.0"}
	}`)
	writeFile(t, dir, "pnpm-lock.yaml", "lockfileVersion: '6.0'")

	detection, ok := NewNodeDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}

	if detection.Type != ProjectTypeNode {
		t.Errorf("Type = %q, want %q", detection.Type, ProjectTypeNode)
	}
	if detection.Runtime != models.RuntimeNode {
		t.Errorf("Runtime = %q, want %q", detection.Runtime, models.RuntimeNode)
	}
	if detection.PackageManager != "pnpm" {
		t.Errorf("PackageManager = %q, want %q", detection.PackageManager, "pnpm")
	}
	if detection.Name != "my-api" || detection.Version != "1.2.3" {
		t.Errorf("Name/Version = %q/%q, want %q/%q", detection.Name, detection.Version, "my-api", "1.2.3")
	}
	if detection.Framework != models.FrameworkNextJS {
		t.Errorf("Framework = %q, want %q (next before react)", detection.Framework, models.FrameworkNextJS)
	}
	if detection.Confidence != ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q", detection.Confidence, ConfidenceHigh)
	}
	if detection.Suggested.Install != "pnpm install" {
		t.Errorf("Suggested.Install = %q, want %q", detection.Suggested.Install, "pnpm install")
	}
	if detection.Suggested.Build != "pnpm run build" {
		t.Errorf("Suggested.Build = %q, want %q", detection.Suggested.Build, "pnpm run build")
	}
	if detection.Suggested.Start != "pnpm start" {
		t.Errorf("Suggested.Start = %q, want %q", detection.Suggested.Start, "pnpm start")
	}
	if len(detection.Notes) != 0 {
		t.Errorf("Notes = %v, want empty", detection.Notes)
	}
}

func TestNodeDetectorPackageManagerFromEachLockfile(t *testing.T) {
	tests := []struct {
		lockfile string
		want     string
	}{
		{"bun.lockb", "bun"},
		{"bun.lock", "bun"},
		{"pnpm-lock.yaml", "pnpm"},
		{"yarn.lock", "yarn"},
		{"package-lock.json", "npm"},
	}

	for _, tt := range tests {
		t.Run(tt.lockfile, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "package.json", `{"name": "app"}`)
			writeFile(t, dir, tt.lockfile, "")

			detection, ok := NewNodeDetector().Detect(buildInput(t, dir))
			if !ok {
				t.Fatal("expected a detection")
			}
			if detection.PackageManager != tt.want {
				t.Errorf("PackageManager = %q, want %q", detection.PackageManager, tt.want)
			}
		})
	}
}

func TestNodeDetectorNoLockfileLeavesPackageManagerEmpty(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"name": "app"}`)

	detection, ok := NewNodeDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.PackageManager != "" {
		t.Errorf("PackageManager = %q, want empty", detection.PackageManager)
	}
	if !containsNote(detection.Notes, "package manager") {
		t.Errorf("expected a note about the unknown package manager, got %v", detection.Notes)
	}
	if detection.Suggested.Install != "" {
		t.Errorf("Suggested.Install = %q, want empty without a known package manager", detection.Suggested.Install)
	}
}

func TestNodeDetectorMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{not valid json`)

	detection, ok := NewNodeDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection even with malformed package.json")
	}
	if detection.Name != "" {
		t.Errorf("Name = %q, want empty for malformed JSON", detection.Name)
	}
	if detection.Confidence != ConfidenceMedium {
		t.Errorf("Confidence = %q, want %q", detection.Confidence, ConfidenceMedium)
	}
	if !containsNote(detection.Notes, "could not be parsed") {
		t.Errorf("expected a parse-failure note, got %v", detection.Notes)
	}
}

func TestNodeDetectorMissingNameAndVersionAreNoted(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{}`)

	detection, ok := NewNodeDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if !containsNote(detection.Notes, `"name"`) {
		t.Errorf("expected a note about the missing name field, got %v", detection.Notes)
	}
	if !containsNote(detection.Notes, `"version"`) {
		t.Errorf("expected a note about the missing version field, got %v", detection.Notes)
	}
}

func containsNote(notes []string, substr string) bool {
	for _, n := range notes {
		if strings.Contains(n, substr) {
			return true
		}
	}
	return false
}
