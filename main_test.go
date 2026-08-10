package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expand tilde",
			input:    "~/test",
			expected: filepath.Join(home, "test"),
		},
		{
			name:     "expand env var",
			input:    "$HOME/test",
			expected: filepath.Join(home, "test"),
		},
		{
			name:     "literal path",
			input:    "/usr/local/bin",
			expected: "/usr/local/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandPath(tt.input)
			if err != nil {
				t.Fatalf("expandPath() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("expandPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigParsing(t *testing.T) {
	tmpDir := t.TempDir()
	rootDir := filepath.Join(tmpDir, "dotfiles")
	if err := os.Mkdir(rootDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `root_dir: ` + rootDir + `
force_overwrite: false

symlinks:
  - bashrc: ~/.bashrc
  - config/nvim: ~/.config/nvim
    recursive: true
`

	configPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.RootDir != rootDir {
		t.Errorf("RootDir = %v, want %v", cfg.RootDir, rootDir)
	}

	if cfg.ForceOverwrite != false {
		t.Errorf("ForceOverwrite = %v, want false", cfg.ForceOverwrite)
	}

	if len(cfg.Symlinks) != 2 {
		t.Fatalf("len(Symlinks) = %v, want 2", len(cfg.Symlinks))
	}

	if cfg.Symlinks[1].Recursive != true {
		t.Errorf("Symlinks[1].Recursive = %v, want true", cfg.Symlinks[1].Recursive)
	}
}

func TestSymlinkOperations(t *testing.T) {
	tmpDir := t.TempDir()

	sourceFile := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(sourceFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("create symlink", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "dest.txt")

		err := CreateSymlink(sourceFile, destFile, false, false)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}

		info, err := os.Lstat(destFile)
		if err != nil {
			t.Fatal(err)
		}

		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("destination is not a symlink")
		}
	})

	t.Run("skip existing correct symlink", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "dest2.txt")

		os.Symlink(sourceFile, destFile)

		err := CreateSymlink(sourceFile, destFile, false, false)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}
	})

	t.Run("error on existing non-symlink without force", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "dest3.txt")
		os.WriteFile(destFile, []byte("existing"), 0644)

		err := CreateSymlink(sourceFile, destFile, false, false)
		if err == nil {
			t.Error("CreateSymlink() expected error for existing non-symlink, got nil")
		}
	})
}
