package cmd

import (
	"github.com/Moorad/workpls/internal/app"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <worktree-name> [branch]",
	Short: "Create a worktree and replace the current workspace with it",
	Long: `Create a new worktree in the forest and replace the current Herdr workspace
with a freshly constructed workspace for it.

When [branch] is omitted the configured default_branch is used. If that branch
is already checked out in another worktree you are prompted for a branch name,
which is created from the default branch when it does not exist yet.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: withApp(func(cmd *cobra.Command, args []string, application *app.App) error {
		var branch *string
		if len(args) == 2 {
			branch = &args[1]
		}

		return application.Create(cmd.Context(), args[0], branch)
	}),
}

func init() {
	rootCmd.AddCommand(createCmd)
}
