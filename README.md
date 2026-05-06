# OmniClean

[![CI](https://github.com/bavanchun/OmniClean/actions/workflows/ci.yml/badge.svg)](https://github.com/bavanchun/OmniClean/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bavanchun/OmniClean)](https://goreportcard.com/report/github.com/bavanchun/OmniClean)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/bavanchun/OmniClean)](https://github.com/bavanchun/OmniClean/releases/latest)

A cross-platform Go TUI that turns ten different package managers, project artifact graveyards, and disk-usage hot spots into one keyboard-driven workflow.

<!-- TODO: add a screenshot or GIF of the TUI here -->

## Features

- **Smart uninstall** — single searchable list across apt, brew, snap, flatpak, pip, npm, cargo, winget, choco, and scoop, with package-detail view and dry-run.
- **Leftover detection v2** — per-manager scanners locate residual caches, configs, and per-app data with byte-accurate sizes and a user-editable whitelist (`~/.config/omniclean/whitelist`).
- **Project artifact purge** — `omniclean purge` walks configured project roots and lists `node_modules`, `target`, `.venv`, `__pycache__`, `.gradle`, `bin/obj`, `dist`, and friends, with per-stack badges, recent-build protection, and a confirm-before-delete flow.
- **Interactive disk explorer** — `omniclean analyze` is a Bubbletea/Lip Gloss disk explorer with bar-rendered usage, breadcrumb navigation, a large-files overlay, and OS-native trash (Finder, gio/trash-cli, Recycle Bin).
- **Pipeable** — `omniclean analyze --json` (or any non-TTY stdout) emits stable JSON for scripting.

## Supported Package Managers

| Manager       | Platform                  |
|---------------|---------------------------|
| apt/dpkg      | Linux (Debian/Ubuntu)     |
| brew          | macOS / Linux             |
| snap          | Linux                     |
| flatpak       | Linux                     |
| pip           | Cross-platform            |
| npm           | Cross-platform            |
| cargo         | Cross-platform            |
| winget        | Windows                   |
| choco         | Windows (Chocolatey)      |
| scoop         | Windows (Scoop)           |

## Installation

### From GitHub Releases

Download the latest binary for your platform from [Releases](https://github.com/bavanchun/OmniClean/releases).

```bash
# Linux amd64
curl -Lo omniclean.tar.gz https://github.com/bavanchun/OmniClean/releases/latest/download/omniclean_linux_amd64.tar.gz
tar xzf omniclean.tar.gz
sudo mv omniclean /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/bavanchun/OmniClean/cmd/omniclean@latest
```

### From Source

```bash
git clone https://github.com/bavanchun/OmniClean.git
cd OmniClean
make build
sudo mv bin/omniclean /usr/local/bin/
```

## Usage

```bash
# Launch the package uninstall TUI
omniclean

# Preview without making changes
omniclean --dry-run

# Restrict to specific managers
omniclean --manager apt,pip

# Project artifact purge
omniclean purge                    # interactive review
omniclean purge --dry-run          # preview only
omniclean purge --paths            # edit configured scan roots
omniclean purge --stack node,rust  # restrict by ecosystem

# Disk explorer
omniclean analyze ~/Code           # TUI explorer
omniclean analyze --json | jq      # machine-readable
omniclean analyze --large-min=1G   # raise the large-files threshold
```

## Key Bindings

| Key            | Action                        |
|----------------|-------------------------------|
| `↑/↓`          | Navigate list                 |
| `space`        | Toggle package selection      |
| `/`            | Search/filter packages        |
| `tab`          | Cycle filter by manager       |
| `enter`        | Confirm selected packages     |
| `d`            | View package details          |
| `esc`          | Go back                       |
| `q` / `ctrl+c` | Quit                          |

## Development

```bash
make run           # run directly
make run-dry       # run with --dry-run
make test          # run all tests with race detector
make lint          # golangci-lint
make build         # build to bin/omniclean
make fmt           # format with gofmt + goimports
make install-tools # install golangci-lint, goreleaser, goimports
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to add a new package manager or submit a bug fix.

## License

[MIT](LICENSE)
