// Package cli defines IAMXFREE's command surface using Cobra. It is
// responsible for parsing flags/subcommands and delegating to core; running
// with no subcommand launches the interactive TUI.
package cli

import (
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
			return tui.Run()
		},
	}

	return root
}

// Execute runs the CLI, parsing os.Args.
func Execute() error {
	return NewRootCommand().Execute()
}
