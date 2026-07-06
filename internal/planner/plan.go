package planner

import (
	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// DeploymentPlan is a proposed IAMXFREE configuration for a project, built
// from what an inspection.Result found. It is only a proposal: nothing in
// this package runs a command, writes a file, or persists anything.
type DeploymentPlan struct {
	// ProjectType is the detected technology that produced this plan (e.g.
	// inspection.ProjectTypeNode), or empty if nothing was detected.
	ProjectType inspection.ProjectType

	// General information
	SuggestedName  string
	Type           models.ApplicationType
	Framework      models.Framework
	Runtime        models.Runtime
	PackageManager string
	SuggestedPort  int
	// Domain is always left blank — a domain is never inferred from a
	// project's files.
	Domain string

	// Configuration
	InstallCommand string
	BuildCommand   string
	StartCommand   string

	MatchedFiles []string
	Dependencies []string

	Confidence inspection.Confidence

	// Warnings explain what could not be determined and needs the user's
	// attention or a decision — every blank field above that a Planner
	// could not confidently fill in has a matching Warning explaining why.
	Warnings []string

	// Notes are additional context worth telling the user that isn't a gap
	// (e.g. "a Dockerfile was also found").
	Notes []string
}
