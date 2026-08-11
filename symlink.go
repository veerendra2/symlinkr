package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Stats struct {
	Created int
	Skipped int
	Errors  int
}

func CreateSymlink(source, dest string, force, dryRun bool) error {
	if _, err := os.Lstat(source); os.IsNotExist(err) {
		return fmt.Errorf("⚠ Skipped: source not found for %q", filepath.Base(source))
	}

	destInfo, err := os.Lstat(dest)
	if err == nil {
		if destInfo.Mode()&os.ModeSymlink != 0 {
			link, _ := os.Readlink(dest)
			if link == source {
				return nil
			}
			if dryRun {
				fmt.Printf("[DRY-RUN] Would remove old symlink: %s\n", dest)
				fmt.Printf("[DRY-RUN] Would create symlink: %s -> %s\n", dest, source)
				return nil
			}
			if err := os.Remove(dest); err != nil {
				return fmt.Errorf("✗ Error: failed to remove old symlink: %w", err)
			}
		} else {
			if !force {
				return fmt.Errorf("✗ Error: %s exists and is not a symlink (use -f to force)", dest)
			}
			if dryRun {
				fmt.Printf("[DRY-RUN] Would backup: %s to %s.bak\n", dest, dest)
				fmt.Printf("[DRY-RUN] Would create symlink: %s -> %s\n", dest, source)
				return nil
			}
			backupPath := dest + ".bak"
			if err := os.Rename(dest, backupPath); err != nil {
				return fmt.Errorf("✗ Error: failed to backup file: %w", err)
			}
		}
	}

	if dryRun {
		fmt.Printf("[DRY-RUN] Would create symlink: %s -> %s\n", dest, source)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("✗ Error: failed to create parent directory: %w", err)
	}

	if err := os.Symlink(source, dest); err != nil {
		return fmt.Errorf("✗ Error: failed to create symlink: %w", err)
	}

	fmt.Printf("✓ Created: %s -> %s\n", dest, source)
	return nil
}

func CreateRecursive(sourceDir, destDir string, force, dryRun bool) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			if dryRun {
				fmt.Printf("[DRY-RUN] Would create directory: %s\n", destPath)
				return nil
			}
			return os.MkdirAll(destPath, 0755)
		}

		return CreateSymlink(path, destPath, force, dryRun)
	})
}

func RemoveSymlink(path string, dryRun bool) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}

	if dryRun {
		fmt.Printf("[DRY-RUN] Would remove symlink: %s\n", path)
		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("✗ Error: failed to remove symlink: %w", err)
	}

	fmt.Printf("✓ Removed: %s\n", path)
	return nil
}

func RemoveRecursive(destDir string, dryRun bool) error {
	var symlinks []string
	var dirs []string

	err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			symlinks = append(symlinks, path)
		} else if info.IsDir() && path != destDir {
			dirs = append(dirs, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, link := range symlinks {
		if err := RemoveSymlink(link, dryRun); err != nil {
			return err
		}
	}

	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		if len(entries) == 0 {
			if dryRun {
				fmt.Printf("[DRY-RUN] Would remove empty directory: %s\n", dir)
				continue
			}
			if err := os.Remove(dir); err != nil {
				// Log but don't fail - cleanup is best-effort
				fmt.Printf("⚠ Warning: could not remove directory %s: %v\n", dir, err)
			}
		}
	}

	return nil
}
