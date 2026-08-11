package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var (
	version   = "dev"
	revision  = "none"
	date      = "unknown"
	buildUser = "unknown"
)

func main() {
	configPath := flag.String("config", "symlinkr.yaml", "Config file path")
	remove := flag.Bool("r", false, "Remove mode (uninstall)")
	force := flag.Bool("f", false, "Force overwrite existing files")
	dryRun := flag.Bool("dry-run", false, "Preview changes without executing")
	showVersion := flag.Bool("v", false, "Show version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "symlinkr - Manage symlinks from a YAML configuration\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  symlinkr [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  --config <path>    Config file path (default: symlinkr.yaml)\n")
		fmt.Fprintf(os.Stderr, "  -r                 Uninstall mode (remove all symlinks)\n")
		fmt.Fprintf(os.Stderr, "  -f                 Force overwrite existing files\n")
		fmt.Fprintf(os.Stderr, "  --dry-run          Preview changes without executing\n")
		fmt.Fprintf(os.Stderr, "  -v                 Show version information\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  symlinkr                              # Apply config\n")
		fmt.Fprintf(os.Stderr, "  symlinkr --dry-run                    # Preview changes\n")
		fmt.Fprintf(os.Stderr, "  symlinkr --config ~/dotfiles.yaml     # Custom config\n")
		fmt.Fprintf(os.Stderr, "  symlinkr -f                           # Force overwrite\n")
		fmt.Fprintf(os.Stderr, "  symlinkr -r                           # Uninstall\n")
		fmt.Fprintf(os.Stderr, "  symlinkr -r --dry-run                 # Preview uninstall\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("symlinkr %s (commit: %s, built at: %s by %s)\n", version, revision, date, buildUser)
		os.Exit(0)
	}

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	forceMode := *force
	if cfg.ForceOverwrite && !forceMode {
		forceMode = true
	}

	stats := Stats{}

	if *remove {
		if err := runRemoveMode(cfg, *dryRun, &stats); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := runApplyMode(cfg, forceMode, *dryRun, &stats); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	printSummary(stats, *dryRun)

	if stats.Errors > 0 {
		os.Exit(1)
	}
}

func runApplyMode(cfg *Config, force, dryRun bool, stats *Stats) error {
	for _, symlink := range cfg.Symlinks {
		var err error
		if symlink.Recursive {
			err = CreateRecursive(symlink.Source, symlink.Dest, force, dryRun)
		} else {
			err = CreateSymlink(symlink.Source, symlink.Dest, force, dryRun)
		}

		if err != nil {
			var skipErr *SkipError
			if errors.As(err, &skipErr) {
				fmt.Println(err.Error())
				stats.Skipped++
			} else {
				fmt.Println(err.Error())
				stats.Errors++
			}
		} else {
			stats.Created++
		}
	}

	return nil
}

func runRemoveMode(cfg *Config, dryRun bool, stats *Stats) error {
	for _, symlink := range cfg.Symlinks {
		var err error
		if symlink.Recursive {
			err = RemoveRecursive(symlink.Dest, dryRun)
		} else {
			err = RemoveSymlink(symlink.Dest, dryRun)
		}

		if err != nil {
			fmt.Println(err.Error())
			stats.Errors++
		} else {
			stats.Removed++
		}
	}

	return nil
}

func printSummary(stats Stats, dryRun bool) {
	if dryRun {
		fmt.Printf("\n[DRY-RUN] Summary: %d would create, %d would remove, %d would skip, %d would error\n",
			stats.Created, stats.Removed, stats.Skipped, stats.Errors)
	} else {
		fmt.Printf("\nSummary: %d created, %d removed, %d skipped, %d error\n",
			stats.Created, stats.Removed, stats.Skipped, stats.Errors)
	}
}
