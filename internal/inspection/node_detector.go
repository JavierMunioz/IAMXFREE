package inspection

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

type nodePackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// nodeFrameworkMarkers maps a package.json dependency name to the framework
// it implies. Order matters: meta-frameworks are checked before the base
// library they build on (Next.js before React, Nuxt before Vue, ...).
var nodeFrameworkMarkers = []struct {
	pkg       string
	framework models.Framework
}{
	{"next", models.FrameworkNextJS},
	{"nuxt", models.FrameworkNuxt},
	{"astro", models.FrameworkAstro},
	{"@sveltejs/kit", models.FrameworkSvelte},
	{"svelte", models.FrameworkSvelte},
	{"@angular/core", models.FrameworkAngular},
	{"vue", models.FrameworkVue},
	{"react", models.FrameworkReact},
	{"@nestjs/core", models.FrameworkNestJS},
	{"express", models.FrameworkExpress},
}

type nodeDetector struct{}

// NewNodeDetector returns a Detector for Node.js projects (package.json).
func NewNodeDetector() Detector {
	return &nodeDetector{}
}

func (d *nodeDetector) Name() string { return "node" }

func (d *nodeDetector) Detect(in DetectionInput) (Detection, bool) {
	if !in.Has("package.json") {
		return Detection{}, false
	}

	detection := Detection{
		Type:         ProjectTypeNode,
		Runtime:      models.RuntimeNode,
		MatchedFiles: []string{"package.json"},
	}

	switch {
	case in.Has("bun.lockb"):
		detection.PackageManager = "bun"
		detection.MatchedFiles = append(detection.MatchedFiles, "bun.lockb")
	case in.Has("bun.lock"):
		detection.PackageManager = "bun"
		detection.MatchedFiles = append(detection.MatchedFiles, "bun.lock")
	case in.Has("pnpm-lock.yaml"):
		detection.PackageManager = "pnpm"
		detection.MatchedFiles = append(detection.MatchedFiles, "pnpm-lock.yaml")
	case in.Has("yarn.lock"):
		detection.PackageManager = "yarn"
		detection.MatchedFiles = append(detection.MatchedFiles, "yarn.lock")
	case in.Has("package-lock.json"):
		detection.PackageManager = "npm"
		detection.MatchedFiles = append(detection.MatchedFiles, "package-lock.json")
	default:
		detection.Notes = append(detection.Notes, "no lockfile found; package manager could not be determined")
	}

	data, err := in.ReadFile("package.json")
	if err != nil {
		detection.Confidence = ConfidenceMedium
		detection.Notes = append(detection.Notes, "package.json could not be read")
		return detection, true
	}

	var pkg nodePackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		detection.Confidence = ConfidenceMedium
		detection.Notes = append(detection.Notes, "package.json could not be parsed as JSON")
		return detection, true
	}

	detection.Confidence = ConfidenceHigh
	detection.Name = pkg.Name
	detection.Version = pkg.Version
	detection.Scripts = pkg.Scripts

	if pkg.Name == "" {
		detection.Notes = append(detection.Notes, `package.json has no "name" field`)
	}
	if pkg.Version == "" {
		detection.Notes = append(detection.Notes, `package.json has no "version" field`)
	}

	deps := make(map[string]string, len(pkg.Dependencies)+len(pkg.DevDependencies))
	for name, version := range pkg.Dependencies {
		deps[name] = version
	}
	for name, version := range pkg.DevDependencies {
		deps[name] = version
	}

	if len(deps) > 0 {
		names := make([]string, 0, len(deps))
		for name := range deps {
			names = append(names, name)
		}
		sort.Strings(names)
		detection.Dependencies = names
	}

	for _, marker := range nodeFrameworkMarkers {
		if _, ok := deps[marker.pkg]; ok {
			detection.Framework = marker.framework
			break
		}
	}
	if detection.Framework == "" {
		detection.Notes = append(detection.Notes, "framework could not be determined")
	}

	if detection.PackageManager != "" {
		detection.Suggested.Install = detection.PackageManager + " install"
		if _, ok := pkg.Scripts["build"]; ok {
			detection.Suggested.Build = fmt.Sprintf("%s run build", detection.PackageManager)
		}
		if _, ok := pkg.Scripts["start"]; ok {
			detection.Suggested.Start = detection.PackageManager + " start"
		}
	} else {
		detection.Notes = append(detection.Notes, "commands not suggested because the package manager is unknown")
	}

	return detection, true
}
