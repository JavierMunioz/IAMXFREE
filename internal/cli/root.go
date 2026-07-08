// Package cli defines IAMXFREE's command surface using Cobra. It is
// responsible for parsing flags/subcommands and delegating to core; running
// with no subcommand launches the interactive TUI.
package cli

import (
	"github.com/JavierMunioz/IAMXFREE/internal/config"
	"github.com/JavierMunioz/IAMXFREE/internal/execution"
	"github.com/JavierMunioz/IAMXFREE/internal/git"
	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
	"github.com/JavierMunioz/IAMXFREE/internal/monitor"
	"github.com/JavierMunioz/IAMXFREE/internal/planner"
	"github.com/JavierMunioz/IAMXFREE/internal/repositories/jsonstore"
	"github.com/JavierMunioz/IAMXFREE/internal/runtimehost"
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

			sessionsDir, err := config.DefaultSessionsDir()
			if err != nil {
				return err
			}

			sessionRepo, err := jsonstore.NewSessionRepository(sessionsDir)
			if err != nil {
				return err
			}

			host := runtimehost.NewLinuxHost()

			executionRegistry := execution.NewRegistry()
			executionRegistry.Register(execution.NewNodeStrategy(host))
			resolver := execution.NewResolver(executionRegistry)

			gitManager := git.NewManager(host)
			service := services.NewApplicationService(repo, resolver, gitManager)
			runtimeMonitor := monitor.New(host)
			executionService := services.NewExecutionService(repo, resolver, runtimeMonitor, sessionRepo)

			inspector := inspection.NewInspector(inspection.NewDefaultRegistry())
			deploymentPlanner := planner.NewDeploymentPlanner(planner.NewDefaultRegistry())
			setup := services.NewApplicationSetupService(inspector, deploymentPlanner)

			return tui.Run(service, executionService, setup)
		},
	}

	return root
}

// Execute runs the CLI, parsing os.Args.
func Execute() error {
	return NewRootCommand().Execute()
}
