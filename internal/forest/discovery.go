package forest

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type Forest struct {
	Root       string
	CommonDir  string
	Current    Worktree
	Worktrees  []Worktree
	ConfigPath string
}

func (g *Git) Discover(ctx context.Context, cwd string) (*Forest, error) {
	topRaw, err := g.run(ctx, "find current Git checkout", "-C", cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("Workplease must run inside a linked worktree: %w", err)
	}
	commonRaw, err := g.run(ctx, "find common Git directory", "-C", cwd, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, err
	}
	top, err := Canonical(strings.TrimSpace(string(topRaw)))
	if err != nil {
		return nil, err
	}
	common, err := Canonical(strings.TrimSpace(string(commonRaw)))
	if err != nil {
		return nil, err
	}
	root := filepath.Dir(common)
	expectedCommon := filepath.Join(root, ".git")
	if common != expectedCommon {
		return nil, fmt.Errorf("unsupported repository layout: common Git directory must be <forest>/.git (found %s)", common)
	}
	bareRaw, err := g.run(ctx, "verify bare common Git repository", "--git-dir", common, "rev-parse", "--is-bare-repository")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(bareRaw)) != "true" {
		return nil, fmt.Errorf("unsupported repository layout: %s is not a bare repository", common)
	}
	if !isDirectChild(root, top) {
		return nil, fmt.Errorf("unsupported worktree location: %s must be a direct child of %s", top, root)
	}

	configPath := filepath.Join(root, ".workpls.toml")
	exists, err := PathExists(configPath)
	if err != nil {
		return nil, fmt.Errorf("check configuration: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("missing configuration %s", configPath)
	}

	worktrees, err := g.ListWorktrees(ctx, common)
	if err != nil {
		return nil, err
	}
	currentIndex := -1
	for i := range worktrees {
		canonical, canonicalErr := Canonical(worktrees[i].Path)
		if canonicalErr != nil {
			if worktrees[i].Prunable {
				continue
			}
			return nil, canonicalErr
		}
		worktrees[i].Path = canonical
		if canonical == top && !worktrees[i].Bare {
			currentIndex = i
		}
	}
	if currentIndex < 0 {
		return nil, fmt.Errorf("current checkout %s is not a non-bare Git worktree record", top)
	}

	return &Forest{
		Root:       root,
		CommonDir:  common,
		Current:    worktrees[currentIndex],
		Worktrees:  worktrees,
		ConfigPath: configPath,
	}, nil
}

func (f *Forest) ManagedWorktrees() []Worktree {
	managed := make([]Worktree, 0, len(f.Worktrees))
	for _, wt := range f.Worktrees {
		if !wt.Bare && !wt.Prunable && isDirectChild(f.Root, wt.Path) {
			managed = append(managed, wt)
		}
	}
	return managed
}

func (f *Forest) FindWorktreeWithBranch(branch string) *Worktree {
	for i := range f.Worktrees {
		if f.Worktrees[i].Branch == branch && !f.Worktrees[i].Prunable {
			return &f.Worktrees[i]
		}
	}
	return nil
}
