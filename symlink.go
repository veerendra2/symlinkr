package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Stats struct {
	Created int
	Removed int
	Skipped int
	Errors  int
	Exists  int
}

type SkipError struct {
	Msg string
}

func (e *SkipError) Error() string {
	return e.Msg
}

func CreateSymlink(source, dest string, force, dryRun bool, stats *Stats) error {
	if _, err := os.Lstat(source); err != nil {
		if os.IsNotExist(err) {
			stats.Skipped++
			return &SkipError{Msg: fmt.Sprintf("[skip]    %s (source not found: %q)", dest, filepath.Base(source))}
		}
		stats.Errors++
		return fmt.Errorf("[error]   %s (failed to access source: %w)", source, err)
	}

	destInfo, err := os.Lstat(dest)
	if err == nil {
		if destInfo.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(dest)
			if err != nil {
				stats.Errors++
				return fmt.Errorf("[error]   %s (failed to read symlink target: %w)", dest, err)
			}
			// Resolve relative symlink targets before comparing
			linkAbs := link
			if !filepath.IsAbs(link) {
				linkAbs = filepath.Clean(filepath.Join(filepath.Dir(dest), link))
			}
			if linkAbs == source {
				return nil
			}
			if dryRun {
				fmt.Printf("[dry-run] Would remove old symlink: %s\n", dest)
				fmt.Printf("[dry-run] Would create symlink: %s -> %s\n", dest, source)
				stats.Created++
				return nil
			}
			if err := os.Remove(dest); err != nil {
				stats.Errors++
				return fmt.Errorf("[error]   %s (failed to remove old symlink: %w)", dest, err)
			}
		} else {
			if !force {
				stats.Skipped++
				stats.Exists++
				return &SkipError{Msg: fmt.Sprintf("[skip]    %s (already exists)", dest)}
			}
			if dryRun {
				fmt.Printf("[dry-run] Would remove existing file: %s\n", dest)
				fmt.Printf("[dry-run] Would create symlink: %s -> %s\n", dest, source)
				stats.Created++
				return nil
			}
			if err := os.RemoveAll(dest); err != nil {
				stats.Errors++
				return fmt.Errorf("[error]   %s (failed to remove existing file: %w)", dest, err)
			}
		}
	}

	if dryRun {
		fmt.Printf("[dry-run] Would create symlink: %s -> %s\n", dest, source)
		stats.Created++
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		stats.Errors++
		return fmt.Errorf("[error]   %s (failed to create parent directory: %w)", dest, err)
	}

	if err := os.Symlink(source, dest); err != nil {
		stats.Errors++
		return fmt.Errorf("[error]   %s (failed to create symlink: %w)", dest, err)
	}

	fmt.Printf("[created] %s -> %s\n", dest, source)
	stats.Created++
	return nil
}

func CreateRecursive(sourceDir, destDir string, force, dryRun bool, stats *Stats) error {
	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			stats.Errors++
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			stats.Errors++
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			if dryRun {
				fmt.Printf("[dry-run] Would create directory: %s\n", destPath)
				return nil
			}
			if err := os.MkdirAll(destPath, 0755); err != nil {
				stats.Errors++
				return err
			}
			return nil
		}

		err = CreateSymlink(path, destPath, force, dryRun, stats)
		if err != nil {
			fmt.Println(err.Error())
			// Continue walking — stats.Errors already incremented inside CreateSymlink
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to process recursive symlinks: %w", err)
	}
	return nil
}

func RemoveSymlink(path string, dryRun bool, stats *Stats) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		stats.Errors++
		return err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}

	if dryRun {
		fmt.Printf("[dry-run] Would remove symlink: %s\n", path)
		stats.Removed++
		return nil
	}

	if err := os.Remove(path); err != nil {
		stats.Errors++
		return fmt.Errorf("[error]   %s (failed to remove symlink: %w)", path, err)
	}

	fmt.Printf("[removed] %s\n", path)
	stats.Removed++
	return nil
}

func RemoveRecursive(destDir string, dryRun bool, stats *Stats) error {
	var symlinks []string
	var dirs []string

	err := filepath.WalkDir(destDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			symlinks = append(symlinks, path)
		} else if d.IsDir() && path != destDir {
			dirs = append(dirs, path)
		}

		return nil
	})

	if err != nil {
		stats.Errors++
		return fmt.Errorf("failed to scan directory for removal: %w", err)
	}

	for _, link := range symlinks {
		if err := RemoveSymlink(link, dryRun, stats); err != nil {
			fmt.Println(err.Error())
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
				fmt.Printf("[dry-run] Would remove empty directory: %s\n", dir)
				continue
			}
			if err := os.Remove(dir); err != nil {
				stats.Errors++
				fmt.Printf("[error]   Could not remove directory %s: %v\n", dir, err)
			}
		}
	}

	return nil
}
