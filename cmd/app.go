package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/Moorad/workpls/internal/app"
	"github.com/Moorad/workpls/internal/ui"
	"github.com/spf13/cobra"
)

// withApp adapts an application action into a cobra RunE. Arguments are already
// validated by the time it runs, so anything that fails from here on is a
// runtime problem rather than a usage mistake and should not reprint the usage.
func withApp(run func(cmd *cobra.Command, args []string, application *app.App) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		application, err := newApp()
		if err != nil {
			return err
		}

		return run(cmd, args, application)
	}
}

// newApp resolves the external dependencies shared by every command. It fails
// before any work happens if the environment cannot support an interactive
// worktree switch.
func newApp() (*app.App, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("MVP supports macOS only")
	}

	if err := ui.RequireTerminal(os.Stdin, os.Stdout); err != nil {
		return nil, err
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("find git in PATH: %w", err)
	}

	herdrPath, err := exec.LookPath("herdr")
	if err != nil {
		return nil, fmt.Errorf("find herdr in PATH: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("read current directory: %w", err)
	}

	return &app.App{
		GitPath:   gitPath,
		HerdrPath: herdrPath,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		CWD:       cwd,
		Attach: func(path string) error {
			return syscall.Exec(path, []string{path}, os.Environ())
		},
	}, nil
}
