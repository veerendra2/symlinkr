# Symlinkr

<p align="center">
  <img src="logo.png" alt="Symlinkr Logo" width="300"/>
  <br>
</p>
<p align="center">Declarative symlink manager</p>

<p align="center">
  <a href="https://github.com/veerendra2/symlinkr/actions"><img src="https://github.com/veerendra2/symlinkr/workflows/CI/badge.svg" alt="Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/veerendra2/symlinkr"><img src="https://goreportcard.com/badge/github.com/veerendra2/symlinkr" alt="Go Report Card"></a>
  <a href="https://github.com/veerendra2/symlinkr/releases"><img src="https://img.shields.io/github/v/release/veerendra2/symlinkr" alt="Release"></a>
  <a href="https://github.com/veerendra2/symlinkr/blob/main/LICENSE"><img src="https://img.shields.io/github/license/veerendra2/symlinkr" alt="License"></a>
  <a href="https://github.com/veerendra2/symlinkr/stargazers"><img src="https://img.shields.io/github/stars/veerendra2/symlinkr" alt="Stars"></a>
  <a href="https://github.com/veerendra2/symlinkr/network/members"><img src="https://img.shields.io/github/forks/veerendra2/symlinkr" alt="Forks"></a>
  <a href="https://github.com/veerendra2/homebrew-tap"><img src="https://img.shields.io/badge/homebrew-tap-blue?style=flat&logo=homebrew&logoColor=white" alt="Homebrew"></a>
  <a href="https://github.com/veerendra2/symlinkr/releases"><img src="https://img.shields.io/badge/archs-amd64%20%7C%20arm64-blue?style=flat" alt="Architectures"></a>
</p>

> _Inspired by [mise dotfiles](https://mise.jdx.dev/dotfiles.html)_

## Installation

**Via Homebrew:**

```bash
brew tap veerendra2/homebrew-tap
brew install symlinkr
```

**Or download from [GitHub Releases](https://github.com/veerendra2/symlinkr/releases)**

## Usage

```bash
symlinkr - Declarative symlink manager

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

### Configuration

Default configuration file is `symlinkr.yaml` in the current working directory.

```yaml
root_dir: "~/projects/dotfiles"
force_overwrite: false

symlinks:
  - ~/.bashrc: bashrc
  - ~/.gitconfig: gitconfig
  - ~/.config/nvim: config/nvim
    recursive: true
```

| Field                  | Type    | Required | Description                                                                                   |
| :--------------------- | :------ | :------- | :-------------------------------------------------------------------------------------------- |
| `root_dir`             | string  | Yes      | Base directory containing source files. Supports `~` and `$VAR` / `${VAR}`.                   |
| `force_overwrite`      | boolean | No       | Overwrite non-symlink targets with `.bak` backup. Overridden by `-f` flag (default: `false`). |
| `symlinks`             | list    | Yes      | List of `destination: source` mappings. Supports `~` and env vars.                            |
| `symlinks[].recursive` | boolean | No       | When `true`, mirrors directory tree with individual file symlinks (default: `false`).         |

#### Recursive Behavior

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
