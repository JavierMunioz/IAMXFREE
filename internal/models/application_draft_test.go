package models_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

func TestApplicationDraftToApplication(t *testing.T) {
	draft := models.ApplicationDraft{
		Name:           "my-api",
		Type:           models.ApplicationTypeAPI,
		Framework:      models.FrameworkFastAPI,
		Runtime:        models.RuntimePython,
		Path:           "/srv/apps/my-api",
		Port:           8000,
		Domain:         "my-api.example.com",
		RepoURL:        "https://github.com/user/my-api.git",
		PackageManager: "uv",
		InstallCommand: "uv sync",
		BuildCommand:   "",
		StartCommand:   "uvicorn main:app",
	}

	app := draft.ToApplication()

	if app.ID == "" {
		t.Fatal("expected ToApplication to generate an ID")
	}
	if app.Name != draft.Name {
		t.Errorf("Name = %q, want %q", app.Name, draft.Name)
	}
	if app.Type != draft.Type {
		t.Errorf("Type = %q, want %q", app.Type, draft.Type)
	}
	if app.Framework != draft.Framework {
		t.Errorf("Framework = %q, want %q", app.Framework, draft.Framework)
	}
	if app.Runtime != draft.Runtime {
		t.Errorf("Runtime = %q, want %q", app.Runtime, draft.Runtime)
	}
	if app.Source.LocalPath != draft.Path {
		t.Errorf("Source.LocalPath = %q, want %q", app.Source.LocalPath, draft.Path)
	}
	if app.Source.RepositoryURL != draft.RepoURL {
		t.Errorf("Source.RepositoryURL = %q, want %q", app.Source.RepositoryURL, draft.RepoURL)
	}
	if app.Config.InternalPort != draft.Port {
		t.Errorf("Config.InternalPort = %d, want %d", app.Config.InternalPort, draft.Port)
	}
	if app.Config.Domain != draft.Domain {
		t.Errorf("Config.Domain = %q, want %q", app.Config.Domain, draft.Domain)
	}
	if app.Config.PackageManager != draft.PackageManager {
		t.Errorf("Config.PackageManager = %q, want %q", app.Config.PackageManager, draft.PackageManager)
	}
	if app.Config.InstallCommand != draft.InstallCommand {
		t.Errorf("Config.InstallCommand = %q, want %q", app.Config.InstallCommand, draft.InstallCommand)
	}
	if app.Config.StartCommand != draft.StartCommand {
		t.Errorf("Config.StartCommand = %q, want %q", app.Config.StartCommand, draft.StartCommand)
	}
	if err := app.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}
