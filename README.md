# Symlinkr

[![Go Report Card](https://goreportcard.com/badge/github.com/veerendra2/symlinkr)](https://goreportcard.com/report/github.com/veerendra2/symlinkr)
[![Release](https://img.shields.io/github/v/release/veerendra2/symlinkr)](https://github.com/veerendra2/symlinkr/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/veerendra2/symlinkr)](go.mod)
[![Build](https://github.com/veerendra2/symlinkr/actions/workflows/release.yml/badge.svg)](https://github.com/veerendra2/symlinkr/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/veerendra2/symlinkr.svg)](https://pkg.go.dev/github.com/veerendra2/symlinkr)
[![GitHub stars](https://img.shields.io/github/stars/veerendra2/symlinkr)](https://github.com/veerendra2/symlinkr/stargazers)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/veerendra2/symlinkr/graphs/commit-activity)

<div align="center">
  <img src="logo.png" alt="Symlinkr Logo" width="400"/>
</div>

A lightweight CLI tool to manage symlinks from a YAML configuration file.

> Declarative symlink management for dotfiles. - _"Inspired by mise dotfile"_

## Installation

**Via Homebrew:**

```bash
brew tap veerendra2/homebrew-tap
brew install symlinkr
```

**Or download from [GitHub Releases](https://github.com/veerendra2/symlinkr/releases)**

## Usage

```bash
symlinkr - Manage symlinks from a YAML configuration

Usage:
  symlinkr [flags]

Flags:
  --config <path>    Config file path (default: symlinkr.yaml)
  -r                 Uninstall mode (remove all symlinks)
  -f                 Force overwrite existing files
  --dry-run          Preview changes without executing
  -v                 Show version information

Examples:
  symlinkr                              # Apply config
  symlinkr --dry-run                    # Preview changes
  symlinkr --config ~/dotfiles.yaml     # Custom config
  symlinkr -f                           # Force overwrite
  symlinkr -r                           # Uninstall
  symlinkr -r --dry-run                 # Preview uninstall
```

Create a `symlinkr.yaml` config file:

```yaml
root_dir: "~/projects/dotfiles"
force_overwrite: false

symlinks:
  - bashrc: ~/.bashrc
  - gitconfig: ~/.gitconfig
  - config/nvim: ~/.config/nvim
    recursive: true
```

### Configuration

**`root_dir`** (required) - Base directory containing source files

**`force_overwrite`** (optional) - Overwrite existing non-symlink files (creates `.bak` backup)

**`symlinks`** (required) - List of symlink mappings:

- Simple mapping: `source: destination`
- Recursive: adds `recursive: true` to mirror directory structure

### Recursive Behavior

**Without `recursive`** - Creates single directory symlink:

```
~/.config/nvim -> ~/dotfiles/config/nvim
```

**With `recursive: true`** - Creates directory structure with individual file symlinks:

```
~/.config/nvim/init.lua -> ~/dotfiles/config/nvim/init.lua
~/.config/nvim/lua/plugins.lua -> ~/dotfiles/config/nvim/lua/plugins.lua
```

## Development

Build and test using [Taskfile](https://taskfile.dev/):

```bash
task build    # Build binary
task test     # Run tests
task lint     # Run linter
```
