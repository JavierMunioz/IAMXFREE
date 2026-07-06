package application_test

import (
	"testing"

	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizards/application"
)

func TestStepsOrderAndKeys(t *testing.T) {
	steps := application.Steps(&fakeSetupService{})

	wantKeys := []string{
		application.KeyPath,
		application.KeyAnalysis,
		application.KeyName,
		application.KeyType,
		application.KeyFramework,
		application.KeyRuntime,
		application.KeyPackageManager,
		application.KeyInstallCommand,
		application.KeyBuildCommand,
		application.KeyStartCommand,
		application.KeyPort,
		application.KeyDomain,
		application.KeyRepoURL,
		"confirm",
	}

	if len(steps) != len(wantKeys) {
		t.Fatalf("len(steps) = %d, want %d", len(steps), len(wantKeys))
	}
	for i, want := range wantKeys {
		if steps[i].Key != want {
			t.Errorf("steps[%d].Key = %q, want %q", i, steps[i].Key, want)
		}
	}
}
