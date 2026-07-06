package inspection

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestPythonDetectorNotFoundWithoutMarkers(t *testing.T) {
	dir := t.TempDir()
	if _, ok := NewPythonDetector().Detect(buildInput(t, dir)); ok {
		t.Fatal("expected no detection without any Python marker file")
	}
}

func TestPythonDetectorPackageManagerPriority(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"uv wins", map[string]string{"uv.lock": "", "poetry.lock": "", "pyproject.toml": ""}, "uv"},
		{"poetry over pipenv", map[string]string{"poetry.lock": "", "Pipfile": ""}, "poetry"},
		{"pipenv over requirements", map[string]string{"Pipfile": "", "requirements.txt": ""}, "pipenv"},
		{"requirements alone", map[string]string{"requirements.txt": ""}, "pip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tt.files {
				writeFile(t, dir, name, content)
			}

			detection, ok := NewPythonDetector().Detect(buildInput(t, dir))
			if !ok {
				t.Fatal("expected a detection")
			}
			if detection.PackageManager != tt.want {
				t.Errorf("PackageManager = %q, want %q", detection.PackageManager, tt.want)
			}
		})
	}
}

func TestPythonDetectorPyprojectAloneLeavesPackageManagerAmbiguous(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[project]\nname = \"my-api\"\n")

	detection, ok := NewPythonDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.PackageManager != "" {
		t.Errorf("PackageManager = %q, want empty", detection.PackageManager)
	}
	if !containsNote(detection.Notes, "package manager") {
		t.Errorf("expected a note about the ambiguous package manager, got %v", detection.Notes)
	}
}

func TestPythonDetectorFrameworkFromRequirements(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "requirements.txt", "Django==4.2\ngunicorn==21.2\n")

	detection, ok := NewPythonDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Framework != models.FrameworkDjango {
		t.Errorf("Framework = %q, want %q", detection.Framework, models.FrameworkDjango)
	}
}

func TestPythonDetectorStartCommandNeverSuggested(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "requirements.txt", "flask==3.0\n")

	detection, ok := NewPythonDetector().Detect(buildInput(t, dir))
	if !ok {
		t.Fatal("expected a detection")
	}
	if detection.Suggested.Start != "" {
		t.Errorf("Suggested.Start = %q, want empty", detection.Suggested.Start)
	}
	if !containsNote(detection.Notes, "start command") {
		t.Errorf("expected a note explaining why no start command was suggested, got %v", detection.Notes)
	}
}
