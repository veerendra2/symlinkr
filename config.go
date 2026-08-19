package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RootDir        string    `yaml:"root_dir"`
	ForceOverwrite bool      `yaml:"force_overwrite"`
	Symlinks       []Symlink `yaml:"symlinks"`
}

type Symlink struct {
	Source    string `yaml:"-"`
	Dest      string `yaml:"-"`
	Recursive bool   `yaml:"recursive"`
}

func (s *Symlink) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("symlink entry must be a mapping")
	}

	for i := 0; i < len(node.Content); i += 2 {
		if i+1 >= len(node.Content) {
			return fmt.Errorf("malformed symlink entry: odd number of mapping elements")
		}
		key := node.Content[i].Value
		value := node.Content[i+1].Value

		if key == "recursive" {
			if err := node.Content[i+1].Decode(&s.Recursive); err != nil {
				return fmt.Errorf("invalid recursive value: %w", err)
			}
		} else if s.Dest == "" {
			s.Dest = key
			s.Source = value
		} else {
			return fmt.Errorf("symlink entry has multiple mappings: %q and %q", s.Dest, key)
		}
	}

	if s.Source == "" || s.Dest == "" {
		return fmt.Errorf("symlink entry must have source and destination")
	}

	return nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if cfg.RootDir == "" {
		return nil, fmt.Errorf("root_dir is required")
	}

	if len(cfg.Symlinks) == 0 {
		return nil, fmt.Errorf("symlinks list cannot be empty")
	}

	if err := cfg.expandPaths(); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) expandPaths() error {
	var err error
	c.RootDir, err = expandPath(c.RootDir)
	if err != nil {
		return fmt.Errorf("failed to expand root_dir: %w", err)
	}

	cleanRoot := filepath.Clean(c.RootDir)

	for i := range c.Symlinks {
		absSource := filepath.Clean(filepath.Join(cleanRoot, c.Symlinks[i].Source))
		rel, err := filepath.Rel(cleanRoot, absSource)
		if err != nil || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
			return fmt.Errorf("source path escapes root_dir: %s", c.Symlinks[i].Source)
		}
		c.Symlinks[i].Source = absSource

		c.Symlinks[i].Dest, err = expandPath(c.Symlinks[i].Dest)
		if err != nil {
			return fmt.Errorf("failed to expand destination path: %w", err)
		}
	}

	return nil
}

func (c *Config) validate() error {
	if _, err := os.Stat(c.RootDir); err != nil {
		return fmt.Errorf("cannot access root_dir %s: %w", c.RootDir, err)
	}

	return nil
}

func expandPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}

	path = os.ExpandEnv(path)

	return filepath.Clean(path), nil
}
