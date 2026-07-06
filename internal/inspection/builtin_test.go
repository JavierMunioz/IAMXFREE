package inspection

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestNewDefaultRegistryRegistersEveryBuiltinDetector(t *testing.T) {
	registry := NewDefaultRegistry()
	got := registry.Detectors()

	wantNames := []string{"node", "python", "go", "php", "rust", "java", "docker"}
	if len(got) != len(wantNames) {
		t.Fatalf("len(Detectors()) = %d, want %d", len(got), len(wantNames))
	}
	for i, want := range wantNames {
		if got[i].Name() != want {
			t.Errorf("Detectors()[%d].Name() = %q, want %q", i, got[i].Name(), want)
		}
	}
}

func TestDefaultRegistryDetectsNodeProjectEndToEnd(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"name": "my-api", "version": "1.0.0"}`)
	writeFile(t, dir, "package-lock.json", "{}")

	result, err := NewInspector(NewDefaultRegistry()).Inspect(dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	primary, ok := result.Primary()
	if !ok {
		t.Fatal("expected a primary detection")
	}
	if primary.Type != ProjectTypeNode {
		t.Fatalf("Primary().Type = %q, want %q", primary.Type, ProjectTypeNode)
	}
	if primary.Runtime != models.RuntimeNode {
		t.Fatalf("Primary().Runtime = %q, want %q", primary.Runtime, models.RuntimeNode)
	}
}

func TestDefaultRegistryDetectsNodeAndDockerTogether(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"name": "my-api", "version": "1.0.0"}`)
	writeFile(t, dir, "package-lock.json", "{}")
	writeFile(t, dir, "Dockerfile", "FROM node:20\n")

	result, err := NewInspector(NewDefaultRegistry()).Inspect(dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	if len(result.Detections) != 2 {
		t.Fatalf("len(Detections) = %d, want 2 (node + docker)", len(result.Detections))
	}

	var sawNode, sawDocker bool
	for _, d := range result.Detections {
		switch d.Type {
		case ProjectTypeNode:
			sawNode = true
		case ProjectTypeDocker:
			sawDocker = true
		}
	}
	if !sawNode || !sawDocker {
		t.Fatalf("expected both node and docker detections, got %+v", result.Detections)
	}
}

func TestDefaultRegistryFindsNothingInAnEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	result, err := NewInspector(NewDefaultRegistry()).Inspect(dir)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if len(result.Detections) != 0 {
		t.Fatalf("Detections = %+v, want empty", result.Detections)
	}
}
