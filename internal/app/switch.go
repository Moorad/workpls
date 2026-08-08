package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/Moorad/workpls/internal/forest"
	"github.com/Moorad/workpls/internal/ui"
)

func (a *App) Switch(ctx context.Context) error {
	env, err := a.load(ctx)
	if err != nil {
		return err
	}

	confirmed, target, err := prompt(ctx, env)
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	return a.replace(ctx, env, target)
}

func prompt(ctx context.Context, env *environment) (bool, forest.Worktree, error) {
	var choices []ui.Choice
	byPath := make(map[string]forest.Worktree)
	for _, wt := range env.forest.ManagedWorktrees() {
		if wt.Path == env.forest.Current.Path {
			continue
		}

		detail := wt.Branch
		if wt.Detached || detail == "" {
			commit := wt.HEAD
			if len(commit) > 7 {
				commit = commit[:7]
			}
			detail = "detached @ " + commit
		}
		choices = append(choices, ui.Choice{Label: fmt.Sprintf("%-20s %s", wt.Name(), detail), Value: wt.Path})
		byPath[wt.Path] = wt
	}

	if len(choices) == 0 {
		fmt.Println("No alternative worktrees.")
		return false, forest.Worktree{}, nil
	}

	sort.SliceStable(choices, func(i, j int) bool { return choices[i].Label < choices[j].Label })
	path, selected, err := ui.Select("Switch worktree", choices)
	if err != nil {
		return false, forest.Worktree{}, err
	}

	if !selected {
		return false, forest.Worktree{}, nil
	}

	return true, byPath[path], nil
}
