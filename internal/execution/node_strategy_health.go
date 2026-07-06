package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// nodePackageJSON is the minimal shape node_strategy needs from package.json
// to verify the application is runnable. It is intentionally separate from
// (and simpler than) inspection's own package.json struct: a Strategy never
// depends on internal/inspection — detection and execution stay decoupled.
type nodePackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func (s *nodeStrategy) HealthCheck(ctx context.Context, app *models.Application) (HealthCheck, error) {
	items := []HealthCheckItem{
		s.checkToolInstalled("node", HealthCheckRuntimeInstalled),
		s.checkToolInstalled("npm", HealthCheckPackageManagerInstalled),
		s.checkPathValid(app),
		s.checkDirectoryAccessible(app),
		s.checkManifestExists(app),
		s.checkScriptsAvailable(app),
		s.checkCommandsConfigured(app),
	}

	return HealthCheck{Items: items}, nil
}

func (s *nodeStrategy) checkToolInstalled(name string, checkName HealthCheckName) HealthCheckItem {
	availability, err := s.host.LookPath(name)
	if err != nil {
		return HealthCheckItem{Name: checkName, Status: HealthStatusFail, Detail: err.Error()}
	}
	if !availability.Found() {
		return HealthCheckItem{Name: checkName, Status: HealthStatusFail, Detail: fmt.Sprintf("%s not found on PATH", name)}
	}
	return HealthCheckItem{Name: checkName, Status: HealthStatusPass, Detail: availability.Path}
}

func (s *nodeStrategy) checkPathValid(app *models.Application) HealthCheckItem {
	if strings.TrimSpace(app.Source.LocalPath) == "" {
		return HealthCheckItem{Name: HealthCheckPathValid, Status: HealthStatusFail, Detail: "no local path is configured"}
	}
	return HealthCheckItem{Name: HealthCheckPathValid, Status: HealthStatusPass, Detail: app.Source.LocalPath}
}

func (s *nodeStrategy) checkDirectoryAccessible(app *models.Application) HealthCheckItem {
	if strings.TrimSpace(app.Source.LocalPath) == "" {
		return HealthCheckItem{Name: HealthCheckDirectoryAccessible, Status: HealthStatusFail, Detail: "no local path is configured"}
	}

	exists, err := s.host.DirExists(app.Source.LocalPath)
	if err != nil {
		return HealthCheckItem{Name: HealthCheckDirectoryAccessible, Status: HealthStatusFail, Detail: err.Error()}
	}
	if !exists {
		return HealthCheckItem{Name: HealthCheckDirectoryAccessible, Status: HealthStatusFail, Detail: fmt.Sprintf("%s does not exist", app.Source.LocalPath)}
	}
	return HealthCheckItem{Name: HealthCheckDirectoryAccessible, Status: HealthStatusPass}
}

func (s *nodeStrategy) packageJSONPath(app *models.Application) string {
	return filepath.Join(app.Source.LocalPath, "package.json")
}

func (s *nodeStrategy) checkManifestExists(app *models.Application) HealthCheckItem {
	exists, err := s.host.FileExists(s.packageJSONPath(app))
	if err != nil {
		return HealthCheckItem{Name: HealthCheckManifestExists, Status: HealthStatusFail, Detail: err.Error()}
	}
	if !exists {
		return HealthCheckItem{Name: HealthCheckManifestExists, Status: HealthStatusFail, Detail: "package.json not found"}
	}
	return HealthCheckItem{Name: HealthCheckManifestExists, Status: HealthStatusPass}
}

func (s *nodeStrategy) readPackageJSON(app *models.Application) (nodePackageJSON, error) {
	data, err := s.host.ReadFile(s.packageJSONPath(app))
	if err != nil {
		return nodePackageJSON{}, err
	}
	var pkg nodePackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nodePackageJSON{}, err
	}
	return pkg, nil
}

func (s *nodeStrategy) checkScriptsAvailable(app *models.Application) HealthCheckItem {
	pkg, err := s.readPackageJSON(app)
	if err != nil {
		return HealthCheckItem{Name: HealthCheckScriptsAvailable, Status: HealthStatusFail, Detail: "package.json scripts could not be read"}
	}
	if len(pkg.Scripts) == 0 {
		return HealthCheckItem{Name: HealthCheckScriptsAvailable, Status: HealthStatusFail, Detail: "package.json declares no scripts"}
	}
	return HealthCheckItem{Name: HealthCheckScriptsAvailable, Status: HealthStatusPass}
}

func (s *nodeStrategy) checkCommandsConfigured(app *models.Application) HealthCheckItem {
	if strings.TrimSpace(app.Config.StartCommand) == "" {
		return HealthCheckItem{Name: HealthCheckCommandsConfigured, Status: HealthStatusFail, Detail: "no start command is configured"}
	}
	return HealthCheckItem{Name: HealthCheckCommandsConfigured, Status: HealthStatusPass, Detail: app.Config.StartCommand}
}
