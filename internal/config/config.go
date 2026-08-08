package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Version       int    `toml:"version"`
	DefaultBranch string `toml:"default_branch"`
	Setup         Setup  `toml:"setup"`
	Tabs          []Tab  `toml:"tabs"`
}

type Setup struct {
	Copy     []CopyMapping `toml:"copy"`
	Commands []string      `toml:"commands"`
}

type CopyMapping struct {
	From string `toml:"from"`
	To   string `toml:"to"`
}

type Tab struct {
	Name    string `toml:"name"`
	Command string `toml:"command"`
}

// Load decodes a closed-schema configuration and validates all rules which do
// not require consulting Git. The caller must separately verify DefaultBranch.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read configuration %s: %w", path, err)
	}
	return Parse(data)
}

func Parse(data []byte) (*Config, error) {
	var cfg Config
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		var unknown *toml.StrictMissingError
		if errors.As(err, &unknown) {
			fields := make([]string, 0, len(unknown.Errors))
			for _, detail := range unknown.Errors {
				fields = append(fields, strings.Join([]string(detail.Key()), "."))
			}
			return nil, fmt.Errorf("parse .workpls.toml: unknown field(s): %s", strings.Join(fields, ", "))
		}
		return nil, fmt.Errorf("parse .workpls.toml: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("validate .workpls.toml: version must equal 1")
	}
	if strings.TrimSpace(c.DefaultBranch) == "" {
		return fmt.Errorf("validate .workpls.toml: default_branch is required")
	}
	if len(c.Tabs) == 0 {
		return fmt.Errorf("validate .workpls.toml: at least one [[tabs]] entry is required")
	}

	names := make(map[string]struct{}, len(c.Tabs))
	for i, tab := range c.Tabs {
		if strings.TrimSpace(tab.Name) == "" {
			return fmt.Errorf("validate .workpls.toml: tabs[%d].name is required", i)
		}
		if _, exists := names[tab.Name]; exists {
			return fmt.Errorf("validate .workpls.toml: duplicate tab name %q", tab.Name)
		}
		names[tab.Name] = struct{}{}
	}

	for i, mapping := range c.Setup.Copy {
		if err := validateRelativePath(mapping.From); err != nil {
			return fmt.Errorf("validate .workpls.toml: setup.copy[%d].from: %w", i, err)
		}
		if err := validateRelativePath(mapping.To); err != nil {
			return fmt.Errorf("validate .workpls.toml: setup.copy[%d].to: %w", i, err)
		}
	}
	return nil
}

func validateRelativePath(path string) error {
	if path == "" {
		return fmt.Errorf("path must not be empty")
	}
	if filepath.IsAbs(path) {
		return fmt.Errorf("path must be relative")
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path must stay within its root")
	}
	return nil
}
