package forest

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type Worktree struct {
	Path           string
	HEAD           string
	Branch         string
	Detached       bool
	Bare           bool
	Prunable       bool
	PrunableReason string
}

func (w Worktree) Name() string { return filepath.Base(w.Path) }

// ParseWorktreePorcelain parses both Git's newline-delimited porcelain format
// and its -z variant. Unknown attributes are deliberately ignored.
func ParseWorktreePorcelain(data []byte) ([]Worktree, error) {
	if bytes.IndexByte(data, 0) >= 0 {
		return parseNULPorcelain(data)
	}
	return parseLinePorcelain(data)
}

func parseNULPorcelain(data []byte) ([]Worktree, error) {
	fields := bytes.Split(data, []byte{0})
	var result []Worktree
	var current *Worktree
	for _, raw := range fields {
		if len(raw) == 0 {
			continue
		}
		line := string(raw)
		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				result = append(result, *current)
			}
			current = &Worktree{Path: strings.TrimPrefix(line, "worktree ")}
			continue
		}
		if current == nil {
			return nil, fmt.Errorf("parse git worktree list: attribute before worktree record")
		}
		parseAttribute(current, line)
	}
	if current != nil {
		result = append(result, *current)
	}
	return validateRecords(result)
}

func parseLinePorcelain(data []byte) ([]Worktree, error) {
	var result []Worktree
	var current *Worktree
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if current != nil {
				result = append(result, *current)
				current = nil
			}
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				result = append(result, *current)
			}
			path := strings.TrimPrefix(line, "worktree ")
			if strings.HasPrefix(path, "\"") {
				unquoted, err := strconv.Unquote(path)
				if err != nil {
					return nil, fmt.Errorf("parse git worktree path %s: %w", path, err)
				}
				path = unquoted
			}
			current = &Worktree{Path: path}
			continue
		}
		if current == nil {
			return nil, fmt.Errorf("parse git worktree list: attribute before worktree record")
		}
		parseAttribute(current, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse git worktree list: %w", err)
	}
	if current != nil {
		result = append(result, *current)
	}
	return validateRecords(result)
}

func parseAttribute(w *Worktree, line string) {
	key, value, found := strings.Cut(line, " ")
	if !found {
		key = line
	}
	switch key {
	case "HEAD":
		w.HEAD = value
	case "branch":
		w.Branch = strings.TrimPrefix(value, "refs/heads/")
	case "detached":
		w.Detached = true
	case "bare":
		w.Bare = true
	case "prunable":
		w.Prunable = true
		w.PrunableReason = value
	}
}

func validateRecords(records []Worktree) ([]Worktree, error) {
	for i := range records {
		if records[i].Path == "" {
			return nil, fmt.Errorf("parse git worktree list: record has no path")
		}
	}
	return records, nil
}
