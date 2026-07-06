package inspection

import (
	"os"
	"path/filepath"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// ProjectType names the technology stack a Detector recognizes.
type ProjectType string

const (
	ProjectTypeNode   ProjectType = "node"
	ProjectTypePython ProjectType = "python"
	ProjectTypeGo     ProjectType = "go"
	ProjectTypePHP    ProjectType = "php"
	ProjectTypeRust   ProjectType = "rust"
	ProjectTypeJava   ProjectType = "java"
	ProjectTypeDocker ProjectType = "docker"
)

// Confidence states how sure a Detection is, so a caller never has to guess
// whether a field reflects solid evidence or a best-effort guess.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// SuggestedCommands are commands a Detector believes would install, build
// and start the project. A field is left empty whenever the detector isn't
// confident enough to suggest one — an empty string is never a guess.
type SuggestedCommands struct {
	Install string
	Build   string
	Start   string
}

// Detection is everything one Detector found in a directory.
type Detection struct {
	Type           ProjectType
	Runtime        models.Runtime
	Framework      models.Framework
	PackageManager string
	Name           string
	Version        string
	Scripts        map[string]string
	MatchedFiles   []string
	Suggested      SuggestedCommands
	Confidence     Confidence

	// Dependencies lists every dependency name this Detector could enumerate
	// (e.g. package.json's combined dependencies + devDependencies keys),
	// regardless of dependency "kind". Detectors that cannot enumerate
	// dependencies (Docker, Java via Gradle, ...) leave this nil.
	Dependencies []string

	// Notes explicitly states what could not be determined, instead of a
	// Detection silently omitting it.
	Notes []string
}

// DetectionInput is what a Detector receives: the project root and the set
// of entries the Inspector already found directly inside it. Detectors read
// file contents on demand via ReadFile; they never list directories or
// touch anything outside root themselves.
type DetectionInput struct {
	Root    string
	Entries map[string]bool
}

// Has reports whether name exists directly inside the inspected root.
func (in DetectionInput) Has(name string) bool {
	return in.Entries[name]
}

// ReadFile reads name from directly inside the inspected root.
func (in DetectionInput) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(in.Root, name))
}

// Detector recognizes one technology from a directory's contents. It only
// reads files under the root it is given — it never executes a command,
// installs anything, or writes to the filesystem.
type Detector interface {
	// Name identifies this detector (e.g. "node", "docker").
	Name() string

	// Detect inspects input and reports what it found. ok is false when
	// this detector found no evidence of its technology at all.
	Detect(input DetectionInput) (Detection, bool)
}
