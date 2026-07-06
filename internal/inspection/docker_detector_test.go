package inspection

import "testing"

func TestDockerDetectorNotFoundWithoutMarkers(t *testing.T) {
	dir := t.TempDir()
	if _, ok := NewDockerDetector().Detect(buildInput(t, dir)); ok {
		t.Fatal("expected no detection without Dockerfile or a compose file")
	}
}

func TestDockerDetectorDockerfileOnly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Dockerfile", "FROM node:20\n")

	detection, ok := NewDockerDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.PackageManager != "docker" {
		t.Errorf("PackageManager = %q, want %q", detection.PackageManager, "docker")
	}
	if detection.Suggested.Build != "docker build ." {
		t.Errorf("Suggested.Build = %q, want %q", detection.Suggested.Build, "docker build .")
	}
	if detection.Suggested.Start != "" {
		t.Errorf("Suggested.Start = %q, want empty without a compose file", detection.Suggested.Start)
	}
	if !containsNote(detection.Notes, "compose file") {
		t.Errorf("expected a note about the missing compose file, got %v", detection.Notes)
	}
}

func TestDockerDetectorComposeFile(t *testing.T) {
	for _, name := range dockerComposeFiles {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, name, "services: {}\n")

			detection, ok := NewDockerDetector().Detect(buildInput(t, dir))
			if !ok {
				t.Fatal("expected a detection")
			}
			if detection.PackageManager != "docker compose" {
				t.Errorf("PackageManager = %q, want %q", detection.PackageManager, "docker compose")
			}
			if detection.Suggested.Start != "docker compose up" {
				t.Errorf("Suggested.Start = %q, want %q", detection.Suggested.Start, "docker compose up")
			}
		})
	}
}

func TestDockerDetectorBothDockerfileAndCompose(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Dockerfile", "FROM node:20\n")
	writeFile(t, dir, "docker-compose.yml", "services: {}\n")

	detection, ok := NewDockerDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if len(detection.MatchedFiles) != 2 {
		t.Errorf("MatchedFiles = %v, want both Dockerfile and docker-compose.yml", detection.MatchedFiles)
	}
	if detection.PackageManager != "docker compose" {
		t.Errorf("PackageManager = %q, want %q (compose takes priority)", detection.PackageManager, "docker compose")
	}
}
