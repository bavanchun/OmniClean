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
- **Cleanup suggestions** — `omniclean cleanup` classifies each package by the manager's own dependency bookkeeping into removable roles — **orphan** (auto-installed dependency no longer required) and **leaf** (top-level manual install) — and never surfaces a still-required dependency. Reuses the confirm/delete flow; candidates are listed orphans-first, oldest-install first.
- **Pipeable** — `omniclean analyze --json` and `omniclean cleanup --json` (or any non-TTY stdout) emit stable JSON for scripting.

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

### Cleanup capability tiers

`omniclean cleanup` only classifies what a manager can prove from its own bookkeeping — no usage-time or last-used heuristics:

| Tier        | Managers                                         | Suggests                          |
|-------------|--------------------------------------------------|-----------------------------------|
| Full        | brew, apt, dnf, pacman, zypper                   | orphan **and** leaf candidates    |
| Leaf-only   | flatpak, cargo, pip, npm, snap, scoop            | leaf (top-level) candidates only  |
| Unknown     | winget, choco                                    | nothing (no dependency graph)     |

flatpak is leaf-only because it has no documented read-only/dry-run way to list unused runtimes; OmniClean never runs a mutating command to detect orphans.

`installedAt` is **best-effort** and derived from filesystem mtimes (e.g. the Homebrew Cellar dir or the dpkg `.list` file). It tracks the last file-list write / relink, **not** a true install date, and is shown only as a sort hint — never as a removal decision driver.

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

# Cleanup suggestions (removable orphan/leaf packages)
omniclean cleanup                  # interactive review TUI
omniclean cleanup --dry-run        # preview removals only
omniclean cleanup --manager brew   # restrict to specific manager(s)
omniclean cleanup --json | jq      # machine-readable candidate list

# Animated main menu (spinner star + rotating gradient borders)
omniclean --fancy
```

## Root Flags

| Flag           | Description                                                    |
|----------------|----------------------------------------------------------------|
| `--dry-run`    | Simulate uninstallation without making changes                 |
| `--manager`    | Filter to specific manager(s) (e.g. `apt,pip`)                 |
| `--no-confirm` | Skip confirmation prompt before uninstalling                   |
| `--verbose`    | Enable verbose debug logging                                   |
| `--fancy`      | Animated UI effects in the main menu (slight idle CPU cost)    |

## Key Bindings

### Main Menu

| Key            | Action                        |
|----------------|-------------------------------|
| `↑/↓` `j/k`    | Navigate cards                |
| `1`…`4` (`5`)  | Jump to card and select       |
| `enter`        | Select highlighted card       |
| `?`            | Toggle full help              |
| `q` / `esc`    | Quit                          |

### Package List

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

### Validating the Linux classifiers

The cleanup classifiers (apt/flatpak) are unit-tested against captured fixtures, but you can exercise them against a **real** apt inside a throwaway container:

```bash
scripts/validate-linux-classifiers.sh             # debian:12 (default)
IMAGE=ubuntu:24.04 scripts/validate-linux-classifiers.sh
```

It plants an orphaned dependency and asserts `omniclean cleanup --json --manager apt` reports it. CI runs the same check as the soft-gated `classify-smoke` job. flatpak is leaf-only (no orphan signal), so verify it manually: every entry from `flatpak list --app` classifies as a `leaf`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to add a new package manager or submit a bug fix.

## License

[MIT](LICENSE)
