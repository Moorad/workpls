package cmd

import (
	"github.com/Moorad/workpls/internal/app"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch to another worktree in the forest",
	Long: `Pick another worktree in the forest and replace the current Herdr workspace
with a freshly constructed workspace for it.`,
	Args: cobra.NoArgs,
	RunE: withApp(func(cmd *cobra.Command, args []string, application *app.App) error {
		return application.Switch(cmd.Context())
	}),
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
