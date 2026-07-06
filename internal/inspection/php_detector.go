package inspection

import (
	"encoding/json"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

type composerJSON struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Require map[string]string `json:"require"`
}

type phpDetector struct{}

// NewPHPDetector returns a Detector for PHP projects (composer.json).
func NewPHPDetector() Detector {
	return &phpDetector{}
}

func (d *phpDetector) Name() string { return "php" }

func (d *phpDetector) Detect(in DetectionInput) (Detection, bool) {
	if !in.Has("composer.json") {
		return Detection{}, false
	}

	detection := Detection{
		Type:           ProjectTypePHP,
		Runtime:        models.RuntimePHP,
		PackageManager: "composer",
		MatchedFiles:   []string{"composer.json"},
		Suggested:      SuggestedCommands{Install: "composer install"},
	}

	data, err := in.ReadFile("composer.json")
	if err != nil {
		detection.Confidence = ConfidenceMedium
		detection.Notes = append(detection.Notes, "composer.json could not be read")
		return detection, true
	}

	var pkg composerJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		detection.Confidence = ConfidenceMedium
		detection.Notes = append(detection.Notes, "composer.json could not be parsed as JSON")
		return detection, true
	}

	detection.Confidence = ConfidenceHigh
	detection.Name = pkg.Name
	detection.Version = pkg.Version
	if pkg.Version == "" {
		detection.Notes = append(detection.Notes, `composer.json has no "version" field`)
	}

	if _, ok := pkg.Require["laravel/framework"]; ok {
		detection.Framework = models.FrameworkLaravel
		detection.Suggested.Start = "php artisan serve"
	} else {
		detection.Notes = append(detection.Notes, "framework could not be determined")
	}

	return detection, true
}
