package inspection

import (
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

type goDetector struct{}

// NewGoDetector returns a Detector for Go projects (go.mod).
func NewGoDetector() Detector {
	return &goDetector{}
}

func (d *goDetector) Name() string { return "go" }

func (d *goDetector) Detect(in DetectionInput) (Detection, bool) {
	if !in.Has("go.mod") {
		return Detection{}, false
	}

	detection := Detection{
		Type:           ProjectTypeGo,
		Runtime:        models.RuntimeGo,
		PackageManager: "go modules",
		MatchedFiles:   []string{"go.mod"},
		Confidence:     ConfidenceHigh,
		Suggested: SuggestedCommands{
			Install: "go mod download",
			Build:   "go build ./...",
			Start:   "go run .",
		},
	}

	data, err := in.ReadFile("go.mod")
	if err != nil {
		detection.Notes = append(detection.Notes, "go.mod could not be read")
		return detection, true
	}

	module := parseGoModuleName(string(data))
	detection.Name = module
	if module == "" {
		detection.Notes = append(detection.Notes, "module name could not be determined from go.mod")
	}
	// go.mod's "go X.Y" directive is the minimum Go language version, not a
	// project version, so Version is deliberately left blank here.
	detection.Notes = append(detection.Notes, "go.mod does not declare a project version")

	return detection, true
}

func parseGoModuleName(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
