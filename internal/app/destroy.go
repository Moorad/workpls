package app

import (
	"context"
	"fmt"

	"github.com/Moorad/workpls/internal/ui"
)

func (a *App) Destroy(ctx context.Context) error {
	env, err := a.load(ctx)
	if err != nil {
		return err
	}

	confirmed, err := a.validate(ctx, env)
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	currWorkspace, err := env.herdr.GetCurrentWorkspace(ctx)
	if err != nil {
		return err
	}

	if err := env.git.RemoveWorktree(ctx, env.forest.CommonDir, env.forest.Current.Path); err != nil {
		return err
	}

	if err := env.herdr.CloseWorkspace(ctx, currWorkspace); err != nil {
		return err
	}

	return nil
}

// validate reports whether the user confirmed destroying the current worktree.
func (a *App) validate(ctx context.Context, env *environment) (bool, error) {
	statusText, err := env.git.Status(ctx, env.forest.Current.Path)
	if err != nil {
		return false, err
	}
	if statusText != "" {
		fmt.Fprint(a.Stderr, statusText)
		return false, fmt.Errorf("worktree is dirty; commit, stash, or remove the changes before destroying it")
	}
	branch := env.forest.Current.Branch
	if branch == "" {
		branch = "detached HEAD"
	}

	return ui.Confirm(fmt.Sprintf("Destroy worktree %s (branch %s) and close its workspace?", env.forest.Current.Name(), branch))
}
