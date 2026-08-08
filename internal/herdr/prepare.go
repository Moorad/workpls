package herdr

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Moorad/workpls/internal/config"
)

type PreparedTab struct {
	TabID   string
	PaneID  string
	Command string
}

type PreparedWorkspace struct {
	WorkspaceID string
	FirstTabID  string
	Tabs        []PreparedTab
}

func WorkspaceLabel(forestRoot, worktreeName string) string {
	return filepath.Base(forestRoot) + "/" + worktreeName
}

func (c *Client) PrepareWorkspace(ctx context.Context, targetPath, label string, tabs []config.Tab) (_ *PreparedWorkspace, err error) {
	if len(tabs) == 0 {
		return nil, fmt.Errorf("prepare Herdr workspace: no configured tabs")
	}

	out, err := c.command(ctx, "create target Herdr workspace", "workspace", "create", "--cwd", targetPath, "--label", label, "--no-focus")
	if err != nil {
		return nil, err
	}
	var created struct {
		Workspace Workspace `json:"workspace"`
		Tab       Tab       `json:"tab"`
		RootPane  Pane      `json:"root_pane"`
	}
	if err := decodeResult(out, &created); err != nil {
		return nil, err
	}
	if created.Workspace.WorkspaceID == "" {
		return nil, fmt.Errorf("create target Herdr workspace: response omitted workspace ID")
	}
	workspaceID := created.Workspace.WorkspaceID
	if created.Tab.TabID == "" || created.RootPane.PaneID == "" {
		return nil, fmt.Errorf("create target Herdr workspace: response omitted required tab or pane ID")
	}

	if _, err := c.command(ctx, "rename initial Herdr tab", "tab", "rename", created.Tab.TabID, tabs[0].Name); err != nil {
		return nil, err
	}
	prepared := &PreparedWorkspace{
		WorkspaceID: workspaceID,
		FirstTabID:  created.Tab.TabID,
		Tabs: []PreparedTab{{
			TabID: created.Tab.TabID, PaneID: created.RootPane.PaneID, Command: tabs[0].Command,
		}},
	}

	for _, configured := range tabs[1:] {
		out, err := c.command(ctx, "create Herdr tab "+configured.Name, "tab", "create", "--workspace", workspaceID, "--cwd", targetPath, "--label", configured.Name, "--no-focus")
		if err != nil {
			return nil, err
		}
		var createdTab struct {
			Tab      Tab  `json:"tab"`
			RootPane Pane `json:"root_pane"`
		}
		if err := decodeResult(out, &createdTab); err != nil {
			return nil, err
		}
		if createdTab.Tab.TabID == "" || createdTab.RootPane.PaneID == "" {
			return nil, fmt.Errorf("create Herdr tab %q: response omitted required IDs", configured.Name)
		}
		prepared.Tabs = append(prepared.Tabs, PreparedTab{TabID: createdTab.Tab.TabID, PaneID: createdTab.RootPane.PaneID, Command: configured.Command})
	}

	return prepared, nil
}
