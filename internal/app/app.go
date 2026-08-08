package app

import (
	"context"
	"fmt"
	"os"

	"github.com/Moorad/workpls/internal/config"
	"github.com/Moorad/workpls/internal/forest"
	"github.com/Moorad/workpls/internal/herdr"
)

type App struct {
	GitPath   string
	HerdrPath string
	Stdin     *os.File
	Stdout    *os.File
	Stderr    *os.File
	CWD       string
	Attach    func(string) error
}

type environment struct {
	forest *forest.Forest
	config *config.Config
	git    *forest.Git
	herdr  *herdr.Client
}

func (a *App) load(ctx context.Context) (*environment, error) {
	git := forest.NewGit(a.GitPath)
	discovered, err := git.Discover(ctx, a.CWD)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(discovered.ConfigPath)
	if err != nil {
		return nil, err
	}
	if err := git.ValidateBranchName(ctx, cfg.DefaultBranch); err != nil {
		return nil, fmt.Errorf("invalid configured default_branch: %w", err)
	}
	client := &herdr.Client{Executable: a.HerdrPath}
	if err := client.CheckServer(ctx); err != nil {
		return nil, err
	}
	return &environment{forest: discovered, config: cfg, git: git, herdr: client}, nil
}
