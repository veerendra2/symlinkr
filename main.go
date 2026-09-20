package main

import (
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
	var configPath string
	flag.StringVar(&configPath, "config", "", "Config file path")
	flag.StringVar(&configPath, "c", "", "Config file path (shorthand)")
	remove := flag.Bool("r", false, "Remove mode (uninstall)")
	force := flag.Bool("f", false, "Force overwrite existing files")
	dryRun := flag.Bool("dry-run", false, "Preview changes without executing")
	showVersion := flag.Bool("v", false, "Show version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "symlinkr - Declarative symlink manager\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  symlinkr [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -c, --config <path>  Config file path (default: symlinkr.yaml or symlinkr.yml)\n")
		fmt.Fprintf(os.Stderr, "  -r                   Uninstall mode (remove all symlinks)\n")
		fmt.Fprintf(os.Stderr, "  -f                   Force overwrite existing files\n")
		fmt.Fprintf(os.Stderr, "  --dry-run            Preview changes without executing\n")
		fmt.Fprintf(os.Stderr, "  -v                   Show version information\n\n")
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

	cfg, err := LoadConfig(resolveConfigPath(configPath))
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

// resolveConfigPath returns flagPath if set, otherwise symlinkr.yaml if it
// exists, then symlinkr.yml. When neither exists it returns symlinkr.yaml,
// so the missing-config error stays the same.
func resolveConfigPath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}

	if info, err := os.Stat("symlinkr.yaml"); err == nil && info.Mode().IsRegular() {
		return "symlinkr.yaml"
	}

	if info, err := os.Stat("symlinkr.yml"); err == nil && info.Mode().IsRegular() {
		return "symlinkr.yml"
	}

	return "symlinkr.yaml"
}

func runApplyMode(cfg *Config, force, dryRun bool, stats *Stats) error {
	for _, symlink := range cfg.Symlinks {
		var err error
		if symlink.Recursive {
			err = CreateRecursive(symlink.Source, symlink.Dest, force, dryRun, stats)
		} else {
			err = CreateSymlink(symlink.Source, symlink.Dest, force, dryRun, stats)
		}

		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return nil
}

func runRemoveMode(cfg *Config, dryRun bool, stats *Stats) error {
	for _, symlink := range cfg.Symlinks {
		var err error
		if symlink.Recursive {
			err = RemoveRecursive(symlink.Source, symlink.Dest, dryRun, stats)
		} else {
			err = RemoveSymlink(symlink.Dest, dryRun, stats)
		}

		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return nil
}

func printSummary(stats Stats, dryRun bool) {
	if dryRun {
		fmt.Printf("\nSummary: %d would create, %d would remove, %d would skip, %d would error\n",
			stats.Created, stats.Removed, stats.Skipped, stats.Errors)
	} else {
		fmt.Printf("\nSummary: %d created, %d removed, %d skipped, %d error\n",
			stats.Created, stats.Removed, stats.Skipped, stats.Errors)
	}

	if stats.Exists > 0 {
		fmt.Println("Tip: Use -f to overwrite existing files")
	}
}
