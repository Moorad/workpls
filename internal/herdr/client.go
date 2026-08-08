package herdr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct {
	Executable string
}

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
func (e *CommandError) Unwrap() error { return e.Err }

func (c *Client) command(ctx context.Context, operation string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.Executable, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, &CommandError{Operation: operation, Stderr: strings.TrimSpace(stderr.String()), Err: err}
	}
	return stdout.Bytes(), nil
}

func decodeResult(data []byte, destination any) error {
	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("decode Herdr response: %w", err)
	}
	if len(envelope.Error) > 0 && string(envelope.Error) != "null" {
		return fmt.Errorf("Herdr API error: %s", envelope.Error)
	}
	if len(envelope.Result) == 0 {
		return fmt.Errorf("decode Herdr response: missing result")
	}
	if err := json.Unmarshal(envelope.Result, destination); err != nil {
		return fmt.Errorf("decode Herdr result: %w", err)
	}
	return nil
}

func (c *Client) CheckServer(ctx context.Context) error {
	out, err := c.command(ctx, "check Herdr server", "status", "server")
	if err != nil {
		return err
	}
	status := string(out)
	if !strings.Contains(status, "status: running") {
		return fmt.Errorf("Herdr server is not running")
	}
	if !strings.Contains(status, "compatible: yes") {
		return fmt.Errorf("running Herdr server is not compatible")
	}
	return nil
}

func IsInside() bool {
	return os.Getenv("HERDR_ENV") == "1"
}

type Pane struct {
	PaneID      string `json:"pane_id"`
	TabID       string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
	Focused     bool   `json:"focused"`
	CWD         string `json:"cwd"`
}

type Workspace struct {
	WorkspaceID string `json:"workspace_id"`
	ActiveTabID string `json:"active_tab_id"`
	Label       string `json:"label"`
}

type Tab struct {
	TabID       string `json:"tab_id"`
	WorkspaceID string `json:"workspace_id"`
	Label       string `json:"label"`
}

func (c *Client) ListPanes(ctx context.Context) ([]Pane, error) {
	out, err := c.command(ctx, "list Herdr panes", "pane", "list")
	if err != nil {
		return nil, err
	}
	var result struct {
		Panes []Pane `json:"panes"`
	}
	if err := decodeResult(out, &result); err != nil {
		return nil, err
	}
	for i, pane := range result.Panes {
		if pane.PaneID == "" || pane.WorkspaceID == "" || pane.CWD == "" {
			return nil, fmt.Errorf("decode Herdr pane %d: missing required ID or cwd", i)
		}
	}
	return result.Panes, nil
}

func (c *Client) GetCurrentWorkspace(ctx context.Context) (string, error) {
	currWorkspace := os.Getenv("HERDR_WORKSPACE_ID")

	if currWorkspace == "" {
		panes, err := c.ListPanes(ctx)
		if err != nil {
			return "", err
		}

		for _, pane := range panes {
			if pane.Focused {
				currWorkspace = pane.WorkspaceID
				break
			}
		}

		if currWorkspace == "" {
			return "", fmt.Errorf("no Herdr panes are focused")
		}
	}

	return currWorkspace, nil
}

func (c *Client) CloseWorkspace(ctx context.Context, workspaceID string) error {
	_, err := c.command(ctx, "close Herdr workspace "+workspaceID, "workspace", "close", workspaceID)
	return err
}

func (c *Client) RunPane(ctx context.Context, paneID, command string) error {
	_, err := c.command(ctx, "run command in Herdr pane "+paneID, "pane", "run", paneID, command)
	return err
}

func (c *Client) FocusWorkspace(ctx context.Context, workspaceID string) error {
	_, err := c.command(ctx, "focus Herdr workspace "+workspaceID, "workspace", "focus", workspaceID)
	return err
}
