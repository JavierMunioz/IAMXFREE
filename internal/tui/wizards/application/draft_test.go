package application_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizards/application"
)

func TestDraftFromResult(t *testing.T) {
	result := wizard.Result{Values: map[string]string{
		application.KeyName:           "my-api",
		application.KeyType:           "api",
		application.KeyFramework:      "fastapi",
		application.KeyRuntime:        "python",
		application.KeyPath:           "/srv/apps/my-api",
		application.KeyPort:           "8000",
		application.KeyDomain:         "my-api.example.com",
		application.KeyRepoURL:        "https://github.com/user/my-api.git",
		application.KeyPackageManager: "uv",
		application.KeyInstallCommand: "uv sync",
		application.KeyBuildCommand:   "",
		application.KeyStartCommand:   "uvicorn main:app",
	}}

	draft, err := application.DraftFromResult(result)
	if err != nil {
		t.Fatalf("DraftFromResult() error = %v", err)
	}

	if draft.Name != "my-api" {
		t.Errorf("Name = %q, want %q", draft.Name, "my-api")
	}
	if draft.Port != 8000 {
		t.Errorf("Port = %d, want 8000", draft.Port)
	}
	if draft.Domain != "my-api.example.com" {
		t.Errorf("Domain = %q, want %q", draft.Domain, "my-api.example.com")
	}
	if draft.PackageManager != "uv" {
		t.Errorf("PackageManager = %q, want %q", draft.PackageManager, "uv")
	}
	if draft.InstallCommand != "uv sync" {
		t.Errorf("InstallCommand = %q, want %q", draft.InstallCommand, "uv sync")
	}
	if draft.StartCommand != "uvicorn main:app" {
		t.Errorf("StartCommand = %q, want %q", draft.StartCommand, "uvicorn main:app")
	}
}

func TestDraftFromResultInvalidPort(t *testing.T) {
	result := wizard.Result{Values: map[string]string{
		application.KeyPort: "not-a-port",
	}}

	if _, err := application.DraftFromResult(result); err == nil {
		t.Fatal("expected an error for an invalid port")
	}
}
