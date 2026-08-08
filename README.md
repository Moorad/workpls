# Workpls

Workpls is CLI for ephemeral Git worktrees and Herdr workspaces.

## Commands

```text
workpls create <worktree-name> [branch]
workpls switch
workpls destroy
```

All commands must be run from a linked, direct-child worktree in a workpls forest. They require an interactive terminal, `git`, and an already-running compatible Herdr server.

## Build

```bash
mise run dev
```

## Manual initialization

Initialization is intentionally outside workpls. Create the strict bare-repository forest yourself:

```bash
mkdir project
git clone --bare <remote> project/.git
git --git-dir=project/.git worktree add project/main main
$EDITOR project/.workpls.toml
cd project/main
```

A `.workpls.toml` file example:

```toml
version = 1
# When creating a new worktree, use this branch as the default base
default_branch = "main"

[setup]
# Copy these files when creating a new worktree
copy = [
  { from = "main/.env", to = ".env" },
]
# Run this commands when creating a new worktree
commands = [
  "pnpm install",
]

# Create these tabs when creating or switching to a worktree
[[tabs]]
name = "my-awesome-tab"
command = "echo 'Hi I am Tab 1!'"

[[tabs]]
name = "my-other-awesome-tab"
command = "echo 'Hi I am Tab 2!'"

```
