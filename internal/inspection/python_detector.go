package inspection

import (
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

var pythonMarkerFiles = []string{"requirements.txt", "pyproject.toml", "Pipfile", "poetry.lock", "uv.lock"}

// pythonFrameworkMarkers are matched as plain substrings against the raw
// text of requirements.txt/pyproject.toml — a best-effort heuristic, not a
// dependency-graph resolution.
var pythonFrameworkMarkers = []struct {
	needle    string
	framework models.Framework
}{
	{"fastapi", models.FrameworkFastAPI},
	{"django", models.FrameworkDjango},
	{"flask", models.FrameworkFlask},
}

type pythonDetector struct{}

// NewPythonDetector returns a Detector for Python projects.
func NewPythonDetector() Detector {
	return &pythonDetector{}
}

func (d *pythonDetector) Name() string { return "python" }

func (d *pythonDetector) Detect(in DetectionInput) (Detection, bool) {
	var matched []string
	for _, marker := range pythonMarkerFiles {
		if in.Has(marker) {
			matched = append(matched, marker)
		}
	}
	if len(matched) == 0 {
		return Detection{}, false
	}

	detection := Detection{
		Type:         ProjectTypePython,
		Runtime:      models.RuntimePython,
		MatchedFiles: matched,
	}

	switch {
	case in.Has("uv.lock"):
		detection.PackageManager = "uv"
	case in.Has("poetry.lock"):
		detection.PackageManager = "poetry"
	case in.Has("Pipfile"):
		detection.PackageManager = "pipenv"
	case in.Has("requirements.txt"):
		detection.PackageManager = "pip"
	default:
		detection.Notes = append(detection.Notes, "pyproject.toml present but no lock file found; package manager could not be determined")
	}

	if in.Has("requirements.txt") || in.Has("pyproject.toml") || in.Has("Pipfile") {
		detection.Confidence = ConfidenceHigh
	} else {
		detection.Confidence = ConfidenceMedium
	}

	detection.Framework = detectPythonFramework(in)
	if detection.Framework == "" {
		detection.Notes = append(detection.Notes, "framework could not be determined")
	}

	switch detection.PackageManager {
	case "uv":
		detection.Suggested.Install = "uv sync"
	case "poetry":
		detection.Suggested.Install = "poetry install"
	case "pipenv":
		detection.Suggested.Install = "pipenv install"
	case "pip":
		detection.Suggested.Install = "pip install -r requirements.txt"
	default:
		detection.Notes = append(detection.Notes, "install command not suggested because the package manager is unknown")
	}
	detection.Notes = append(detection.Notes, "start command depends on the framework and entry point; not inferred")

	return detection, true
}

func detectPythonFramework(in DetectionInput) models.Framework {
	for _, name := range []string{"requirements.txt", "pyproject.toml"} {
		if !in.Has(name) {
			continue
		}
		data, err := in.ReadFile(name)
		if err != nil {
			continue
		}
		lower := strings.ToLower(string(data))
		for _, marker := range pythonFrameworkMarkers {
			if strings.Contains(lower, marker.needle) {
				return marker.framework
			}
		}
	}
	return ""
}
