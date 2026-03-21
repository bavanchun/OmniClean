# OmniClean

[![CI](https://github.com/bavanchun/OmniClean/actions/workflows/ci.yml/badge.svg)](https://github.com/bavanchun/OmniClean/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bavanchun/OmniClean)](https://goreportcard.com/report/github.com/bavanchun/OmniClean)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/bavanchun/OmniClean)](https://github.com/bavanchun/OmniClean/releases/latest)

A cross-platform TUI tool that aggregates packages from multiple package managers into a single searchable interface for clean uninstallation.

<!-- TODO: add a screenshot or GIF of the TUI here -->

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
# Launch interactive TUI
omniclean

# Preview what would be removed (no changes made)
omniclean --dry-run

# Only show packages from specific managers
omniclean --manager apt,pip

# Skip the confirmation prompt
omniclean --no-confirm

# Show version
omniclean --version
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
