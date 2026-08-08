package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	VERSION string
	COMMIT  string
	DATE    string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version of workpls and its dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s (%s - %s)\n", VERSION, COMMIT, DATE)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
