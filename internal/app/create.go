package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Moorad/workpls/internal/forest"
	"github.com/Moorad/workpls/internal/herdr"
	"github.com/Moorad/workpls/internal/ui"
)

func (a *App) Create(ctx context.Context, name string, suppliedBranch *string) error {
	env, err := a.load(ctx)
	if err != nil {
		return err
	}

	println("Validating...")
	target, branch, err := validate(ctx, *env, name, suppliedBranch)
	if err != nil {
		return err
	}

	println("Creating worktree...")
	newWorktree, err := createWorktree(ctx, *env, target, branch)
	if err != nil {
		return err
	}

	println("Preparing workspace...")
	return a.replace(ctx, env, newWorktree)
}

func validate(ctx context.Context, env environment, name string, suppliedBranch *string) (string, string, error) {
	if err := forest.ValidateWorktreeName(name); err != nil {
		return "", "", err
	}

	target := filepath.Join(env.forest.Root, name)
	if exists, err := forest.PathExists(target); err != nil {
		return "", "", err
	} else if exists {
		return "", "", fmt.Errorf("worktree path already exists: %s", target)
	}

	for _, wt := range env.forest.Worktrees {
		if filepath.Clean(wt.Path) == target || (!wt.Bare && filepath.Base(wt.Path) == name) {
			return "", "", fmt.Errorf("git already records worktree %q at %s", name, wt.Path)
		}
	}

	defaultExists, err := env.git.BranchExists(ctx, env.forest.CommonDir, env.config.DefaultBranch)
	if err != nil {
		return "", "", err
	}
	if !defaultExists {
		return "", "", fmt.Errorf("configured default branch %q does not exist locally", env.config.DefaultBranch)
	}

	branch := env.config.DefaultBranch
	if suppliedBranch != nil {
		branch = *suppliedBranch
		if err := env.git.ValidateBranchName(ctx, branch); err != nil {
			return "", "", err
		}
	}

	worktreeWithBranch := env.forest.FindWorktreeWithBranch(env.config.DefaultBranch)
	if worktreeWithBranch != nil {
		branch, err = ui.Input(
			fmt.Sprintf(
				"The worktree %s is already checked out to %s\nWhat branch do you want the new worktree to checkout (created from %s if missing)?",
				worktreeWithBranch.Name(),
				branch,
				env.config.DefaultBranch,
			), func(value string) error {
				return env.git.ValidateBranchName(ctx, value)
			},
		)
		if err != nil {
			return "", "", fmt.Errorf("read branch name: %w", err)
		}
	}

	return target, branch, nil
}

func createWorktree(ctx context.Context, env environment, target, branch string) (forest.Worktree, error) {
	var newWorktree forest.Worktree
	branchExists, err := env.git.BranchExists(ctx, env.forest.CommonDir, branch)
	if err != nil {
		return newWorktree, err
	}

	if err := env.git.AddWorktree(ctx, env.forest.CommonDir, target, branch, env.config.DefaultBranch, !branchExists); err != nil {
		return newWorktree, err
	}

	newWorktree = forest.Worktree{Path: target, Branch: branch}
	env.forest.Worktrees = append(env.forest.Worktrees, newWorktree)

	return newWorktree, nil
}

func (a *App) replace(ctx context.Context, env *environment, target forest.Worktree) error {
	prepared, err := env.herdr.PrepareWorkspace(ctx, target.Path, herdr.WorkspaceLabel(env.forest.Root, target.Name()), env.config.Tabs)
	if err != nil {
		return err
	}

	for _, pane := range prepared.Tabs {
		if strings.TrimSpace(pane.Command) == "" {
			continue
		}

		if err := env.herdr.RunPane(ctx, pane.PaneID, pane.Command); err != nil {
			return err
		}
	}

	if herdr.IsInside() {
		if err = env.herdr.FocusWorkspace(ctx, prepared.WorkspaceID); err != nil {
			return err
		}
	}

	currWorkspace, err := env.herdr.GetCurrentWorkspace(ctx)
	if err != nil {
		return err
	}

	if err := env.herdr.CloseWorkspace(ctx, currWorkspace); err != nil {
		return err
	}

	if !herdr.IsInside() {
		if err = a.Attach(env.herdr.Executable); err != nil {
			return err
		}
	}
	return nil
}
