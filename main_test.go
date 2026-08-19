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
  - ~/.bashrc: bashrc
  - ~/.config/nvim: config/nvim
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

	t.Run("reject path traversal", func(t *testing.T) {
		traversalConfig := `root_dir: ` + rootDir + `
symlinks:
  - ~/.evil: ../outside
`
		configPath := filepath.Join(tmpDir, "traversal.yaml")
		if err := os.WriteFile(configPath, []byte(traversalConfig), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("LoadConfig() expected error for path traversal, got nil")
		}
	})
}

func TestSymlinkOperations(t *testing.T) {
	tmpDir := t.TempDir()

	sourceFile := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(sourceFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("create symlink", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "dest.txt")
		stats := &Stats{}

		err := CreateSymlink(sourceFile, destFile, false, false, stats)
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

		if stats.Created != 1 {
			t.Errorf("stats.Created = %d, want 1", stats.Created)
		}
	})

	t.Run("skip existing correct symlink", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "dest2.txt")
		stats := &Stats{}

		if err := os.Symlink(sourceFile, destFile); err != nil {
			t.Fatal(err)
		}

		err := CreateSymlink(sourceFile, destFile, false, false, stats)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}
	})

	t.Run("skip on existing non-symlink without force", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "dest3.txt")
		if err := os.WriteFile(destFile, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}
		stats := &Stats{}

		err := CreateSymlink(sourceFile, destFile, false, false, stats)
		if err == nil {
			t.Error("CreateSymlink() expected skip error for existing non-symlink, got nil")
		}
		if stats.Skipped != 1 || stats.Exists != 1 {
			t.Errorf("stats.Skipped = %d, stats.Exists = %d, want 1, 1", stats.Skipped, stats.Exists)
		}
	})

	t.Run("overwrite existing non-symlink with force", func(t *testing.T) {
		destFile := filepath.Join(tmpDir, "dest4.txt")
		if err := os.WriteFile(destFile, []byte("existing"), 0644); err != nil {
			t.Fatal(err)
		}
		stats := &Stats{}

		err := CreateSymlink(sourceFile, destFile, true, false, stats)
		if err != nil {
			t.Fatalf("CreateSymlink() error = %v", err)
		}

		info, err := os.Lstat(destFile)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected %s to be symlink after force overwrite", destFile)
		}
		if stats.Created != 1 {
			t.Errorf("stats.Created = %d, want 1", stats.Created)
		}
	})
}

func TestRecursiveSymlinkOperations(t *testing.T) {
	tmpDir := t.TempDir()

	sourceDir := filepath.Join(tmpDir, "source")
	subDir := filepath.Join(sourceDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	file1 := filepath.Join(sourceDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(tmpDir, "dest")

	t.Run("create recursive", func(t *testing.T) {
		stats := &Stats{}
		err := CreateRecursive(sourceDir, destDir, false, false, stats)
		if err != nil {
			t.Fatalf("CreateRecursive() error = %v", err)
		}

		destFile1 := filepath.Join(destDir, "file1.txt")
		destFile2 := filepath.Join(destDir, "sub", "file2.txt")

		info1, err := os.Lstat(destFile1)
		if err != nil || info1.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected %s to be symlink", destFile1)
		}

		info2, err := os.Lstat(destFile2)
		if err != nil || info2.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected %s to be symlink", destFile2)
		}

		if stats.Created != 2 {
			t.Errorf("stats.Created = %d, want 2", stats.Created)
		}
	})

	t.Run("remove recursive", func(t *testing.T) {
		stats := &Stats{}
		err := RemoveRecursive(destDir, false, stats)
		if err != nil {
			t.Fatalf("RemoveRecursive() error = %v", err)
		}

		destFile1 := filepath.Join(destDir, "file1.txt")
		if _, err := os.Lstat(destFile1); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed", destFile1)
		}

		if stats.Removed != 2 {
			t.Errorf("stats.Removed = %d, want 2", stats.Removed)
		}
	})
}
