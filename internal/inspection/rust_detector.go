package inspection

import (
	"regexp"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

var (
	cargoNameRe    = regexp.MustCompile(`^name\s*=\s*"([^"]*)"`)
	cargoVersionRe = regexp.MustCompile(`^version\s*=\s*"([^"]*)"`)
)

type rustDetector struct{}

// NewRustDetector returns a Detector for Rust projects (Cargo.toml).
func NewRustDetector() Detector {
	return &rustDetector{}
}

func (d *rustDetector) Name() string { return "rust" }

func (d *rustDetector) Detect(in DetectionInput) (Detection, bool) {
	if !in.Has("Cargo.toml") {
		return Detection{}, false
	}

	detection := Detection{
		Type:           ProjectTypeRust,
		Runtime:        models.RuntimeRust,
		PackageManager: "cargo",
		MatchedFiles:   []string{"Cargo.toml"},
		Suggested:      SuggestedCommands{Build: "cargo build", Start: "cargo run"},
	}

	data, err := in.ReadFile("Cargo.toml")
	if err != nil {
		detection.Confidence = ConfidenceMedium
		detection.Notes = append(detection.Notes, "Cargo.toml could not be read")
		return detection, true
	}

	name, version := parseCargoPackageTable(string(data))
	detection.Name = name
	detection.Version = version

	if name != "" {
		detection.Confidence = ConfidenceHigh
	} else {
		detection.Confidence = ConfidenceMedium
		detection.Notes = append(detection.Notes, "package name could not be determined from Cargo.toml")
	}
	if version == "" {
		detection.Notes = append(detection.Notes, "package version could not be determined from Cargo.toml")
	}

	return detection, true
}

// parseCargoPackageTable is a minimal, best-effort scan for name/version
// inside a Cargo.toml's [package] table. It is not a general TOML parser —
// it only recognizes simple `key = "value"` lines, which is how these two
// fields are conventionally written.
func parseCargoPackageTable(content string) (name string, version string) {
	inPackageTable := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inPackageTable = trimmed == "[package]"
			continue
		}
		if !inPackageTable {
			continue
		}
		if name == "" {
			if m := cargoNameRe.FindStringSubmatch(trimmed); m != nil {
				name = m[1]
			}
		}
		if version == "" {
			if m := cargoVersionRe.FindStringSubmatch(trimmed); m != nil {
				version = m[1]
			}
		}
	}
	return name, version
}
