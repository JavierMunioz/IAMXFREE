// Package cli defines IAMXFREE's command surface using Cobra. It is
// responsible for parsing flags/subcommands and delegating to core; running
// with no subcommand launches the interactive TUI.
package cli

import (
	"github.com/JavierMunioz/IAMXFREE/internal/config"
	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
	"github.com/JavierMunioz/IAMXFREE/internal/planner"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories/jsonstore"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
	"github.com/JavierMunioz/IAMXFREE/internal/tui"
	"github.com/spf13/cobra"
)

// NewRootCommand builds the root Cobra command for IAMXFREE.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "iamxfree",
		Short: "IAMXFREE manages applications deployed on a VPS",
		Long: "IAMXFREE is a terminal UI for administering applications deployed " +
			"on a VPS: processes, reverse proxy, environment variables and more.",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := config.DefaultApplicationsDir()
			if err != nil {
				return err
			}

			repo, err := jsonstore.NewApplicationRepository(dir)
			if err != nil {
				return err
			}

			// No execution.Strategy is registered yet — that arrives with
			// the first concrete strategy (Node+npm, Docker, ...) in a
			// later iteration. Until then, ResolveExecutionStrategy always
			// reports execution.ErrNoStrategyFound.
			resolver := execution.NewResolver(execution.NewRegistry())

			service := services.NewApplicationService(repo, resolver)

			inspector := inspection.NewInspector(inspection.NewDefaultRegistry())
			deploymentPlanner := planner.NewDeploymentPlanner(planner.NewDefaultRegistry())
			setup := services.NewApplicationSetupService(inspector, deploymentPlanner)

			return tui.Run(service, setup)
		},
	}

	return root
}

// Execute runs the CLI, parsing os.Args.
func Execute() error {
	return NewRootCommand().Execute()
}
