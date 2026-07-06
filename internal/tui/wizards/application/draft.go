package application

import (
	"fmt"
	"strconv"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
)

// DraftFromResult converts a completed Wizard's generic Result into a typed
// models.ApplicationDraft. This is the only place that knows the step keys
// defined in Steps line up with ApplicationDraft's fields.
func DraftFromResult(result wizard.Result) (models.ApplicationDraft, error) {
	port, err := strconv.Atoi(result.Get(KeyPort))
	if err != nil {
		return models.ApplicationDraft{}, fmt.Errorf("invalid port %q: %w", result.Get(KeyPort), err)
	}

	return models.ApplicationDraft{
		Name:      result.Get(KeyName),
		Type:      models.ApplicationType(result.Get(KeyType)),
		Framework: models.Framework(result.Get(KeyFramework)),
		Runtime:   models.Runtime(result.Get(KeyRuntime)),
		Path:      result.Get(KeyPath),
		Port:      port,
		Domain:    result.Get(KeyDomain),
		RepoURL:   result.Get(KeyRepoURL),
	}, nil
}
