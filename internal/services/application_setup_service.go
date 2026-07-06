package services

import (
	"context"

	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
	"github.com/JavierMunioz/IAMXFREE/internal/planner"
)

// ApplicationSetupService coordinates project inspection and deployment
// planning into a single proposal for the create-application wizard. The
// wizard depends only on this interface — never directly on
// internal/inspection or internal/planner — so that separation of
// responsibilities holds even as the wizard grows.
type ApplicationSetupService interface {
	// Inspect analyzes path and returns a pre-filled proposal. It is safe
	// (and cheap) to call again for the same path — inspection itself is a
	// handful of local file reads — but callers that want to guarantee it
	// only runs once per project (e.g. across wizard step navigation)
	// should cache the result themselves.
	Inspect(ctx context.Context, path string) (ApplicationSetupProposal, error)
}

type applicationSetupService struct {
	inspector *inspection.Inspector
	planner   *planner.DeploymentPlanner
}

// NewApplicationSetupService builds the default ApplicationSetupService,
// backed by inspector and deploymentPlanner.
func NewApplicationSetupService(inspector *inspection.Inspector, deploymentPlanner *planner.DeploymentPlanner) ApplicationSetupService {
	return &applicationSetupService{inspector: inspector, planner: deploymentPlanner}
}

func (s *applicationSetupService) Inspect(_ context.Context, path string) (ApplicationSetupProposal, error) {
	result, err := s.inspector.Inspect(path)
	if err != nil {
		return ApplicationSetupProposal{}, err
	}

	plan := s.planner.Plan(result)

	return ApplicationSetupProposal{
		Path:           path,
		ProjectType:    string(plan.ProjectType),
		SuggestedName:  plan.SuggestedName,
		Type:           plan.Type,
		Framework:      plan.Framework,
		Runtime:        plan.Runtime,
		PackageManager: plan.PackageManager,
		SuggestedPort:  plan.SuggestedPort,
		InstallCommand: plan.InstallCommand,
		BuildCommand:   plan.BuildCommand,
		StartCommand:   plan.StartCommand,
		MatchedFiles:   plan.MatchedFiles,
		Dependencies:   plan.Dependencies,
		Confidence:     string(plan.Confidence),
		Warnings:       plan.Warnings,
		Notes:          plan.Notes,
	}, nil
}
