package forest

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var worktreeNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type CommandError struct {
	Operation string
	Stderr    string
	Err       error
}

func (e *CommandError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("%s: %s", e.Operation, e.Stderr)
	}
	return fmt.Sprintf("%s: %v", e.Operation, e.Err)
}

type Git struct {
	Executable string
}

func NewGit(executable string) *Git { return &Git{Executable: executable} }

func (g *Git) run(ctx context.Context, operation string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, g.Executable, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, &CommandError{Operation: operation, Stderr: strings.TrimSpace(stderr.String()), Err: err}
	}
	return stdout.Bytes(), nil
}

func (g *Git) ListWorktrees(ctx context.Context, commonDir string) ([]Worktree, error) {
	out, err := g.run(ctx, "list Git worktrees", "--git-dir", commonDir, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, err
	}
	return ParseWorktreePorcelain(out)
}

func (g *Git) ValidateBranchName(ctx context.Context, branch string) error {
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("branch name must not be empty")
	}
	if _, err := g.run(ctx, fmt.Sprintf("validate branch %q", branch), "check-ref-format", "--branch", branch); err != nil {
		return fmt.Errorf("invalid local branch name %q: %w", branch, err)
	}
	return nil
}

func (g *Git) BranchExists(ctx context.Context, commonDir, branch string) (bool, error) {
	cmd := exec.CommandContext(ctx, g.Executable, "--git-dir", commonDir, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
		return false, nil
	}
	return false, &CommandError{Operation: fmt.Sprintf("check local branch %q", branch), Stderr: strings.TrimSpace(stderr.String()), Err: err}
}

func (g *Git) AddWorktree(ctx context.Context, commonDir, target, branch, startPoint string, createBranch bool) error {
	args := []string{"--git-dir", commonDir, "worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch, target, startPoint)
	} else {
		args = append(args, target, branch)
	}
	_, err := g.run(ctx, fmt.Sprintf("create worktree %s", target), args...)
	return err
}

func (g *Git) Status(ctx context.Context, worktreePath string) (string, error) {
	out, err := g.run(ctx, "read worktree status", "-C", worktreePath, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (g *Git) RemoveWorktree(ctx context.Context, commonDir, worktreePath string) error {
	cmd := exec.CommandContext(ctx, g.Executable, "--git-dir", commonDir, "worktree", "remove", worktreePath)
	cmd.Dir = filepath.Dir(commonDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return &CommandError{Operation: fmt.Sprintf("remove worktree %s", worktreePath), Stderr: strings.TrimSpace(stderr.String()), Err: err}
	}
	return nil
}

func ValidateWorktreeName(name string) error {
	if name == "" {
		return fmt.Errorf("worktree name must not be empty")
	}
	if name == "." || name == ".." || !worktreeNamePattern.MatchString(name) {
		return fmt.Errorf("invalid worktree name %q: use only letters, numbers, '.', '_', and '-'", name)
	}
	return nil
}

func Canonical(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("make path absolute %s: %w", path, err)
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", fmt.Errorf("canonicalize path %s: %w", path, err)
	}
	return filepath.Clean(canonical), nil
}

func isDirectChild(parent, child string) bool {
	return filepath.Dir(child) == parent && child != parent
}

func PathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
