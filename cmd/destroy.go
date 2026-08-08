package cmd

import (
	"github.com/Moorad/workpls/internal/app"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Remove the current worktree and close its workspace",
	Long: `Remove the worktree you are currently in and close its Herdr workspace.

The worktree must be clean; commit, stash, or remove any changes first.`,
	Args: cobra.NoArgs,
	RunE: withApp(func(cmd *cobra.Command, args []string, application *app.App) error {
		return application.Destroy(cmd.Context())
	}),
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
