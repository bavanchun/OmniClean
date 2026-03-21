# OmniClean

A cross-platform TUI tool that aggregates packages from multiple package managers into a single searchable interface for clean uninstallation.

## Supported Package Managers

| Manager  | Platform              |
|----------|-----------------------|
| apt/dpkg | Linux (Debian/Ubuntu) |
| brew     | macOS / Linux         |
| snap     | Linux                 |
| flatpak  | Linux                 |
| pip      | Cross-platform        |
| npm      | Cross-platform        |
| cargo    | Cross-platform        |

## Installation

### From GitHub Releases

Download the latest binary for your platform from [Releases](https://github.com/bavanchun/OmniClean/releases).

```bash
# Linux amd64
curl -Lo omniclean.tar.gz https://github.com/bavanchun/OmniClean/releases/latest/download/omniclean_linux_amd64.tar.gz
tar xzf omniclean.tar.gz
sudo mv omniclean /usr/local/bin/
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
```

## Key Bindings

| Key            | Action                    |
|----------------|---------------------------|
| `↑/↓`          | Navigate list             |
| `space`        | Toggle package selection  |
| `/`            | Search/filter packages    |
| `enter`        | Confirm selected packages |
| `d`            | View package details      |
| `esc`          | Go back                   |
| `q` / `ctrl+c` | Quit                      |

## Development

```bash
make run        # run directly
make test       # run all tests with race detector
make lint       # golangci-lint
make build      # build to bin/omniclean
make install-tools  # install golangci-lint, goreleaser, goimports
```

## License

MIT
