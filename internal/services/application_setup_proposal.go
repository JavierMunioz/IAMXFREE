package services

import "github.com/JavierMunioz/IAMXFREE/internal/models"

// ApplicationSetupProposal is a pre-filled draft for the create-application
// wizard, built by inspecting a filesystem path and planning a
// configuration from what was found there. It is meant to be shown to the
// user and edited freely — nothing here is authoritative until the user
// confirms the wizard.
//
// It intentionally does not reuse any internal/inspection or
// internal/planner type: this is the only shape the wizard/TUI ever needs
// to know about, so those packages stay out of the presentation layer's
// dependency graph entirely.
type ApplicationSetupProposal struct {
	Path string

	// ProjectType is the detected technology (e.g. "node", "docker"), or
	// empty if nothing could be detected at Path.
	ProjectType string

	SuggestedName  string
	Type           models.ApplicationType
	Framework      models.Framework
	Runtime        models.Runtime
	PackageManager string
	SuggestedPort  int

	InstallCommand string
	BuildCommand   string
	StartCommand   string

	MatchedFiles []string
	Dependencies []string

	// Confidence is "high", "medium" or "low".
	Confidence string

	// Warnings explain what could not be determined and needs the user's
	// attention. Notes are additional context that isn't a gap.
	Warnings []string
	Notes    []string
}
