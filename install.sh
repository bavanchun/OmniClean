#!/usr/bin/env bash
# OmniClean installer — downloads the requested release tarball, checks
# its SHA-256 against the published checksums file, and installs the
# binary to /usr/local/bin (override with INSTALL_DIR).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/bavanchun/OmniClean/main/install.sh | bash
#   ./install.sh --version v0.3.0
#   INSTALL_DIR=$HOME/.local/bin ./install.sh

set -euo pipefail

REPO="bavanchun/OmniClean"
VERSION=""
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

while [ $# -gt 0 ]; do
  case "$1" in
    --version)
      VERSION="$2"; shift 2 ;;
    --version=*)
      VERSION="${1#*=}"; shift ;;
    --install-dir)
      INSTALL_DIR="$2"; shift 2 ;;
    --install-dir=*)
      INSTALL_DIR="${1#*=}"; shift ;;
    -h|--help)
      sed -n '2,12p' "$0"; exit 0 ;;
    *)
      echo "unknown option: $1" >&2; exit 2 ;;
  esac
done

# Resolve version (default: latest release tag).
if [ -z "$VERSION" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -m1 '"tag_name"' | cut -d '"' -f4)"
fi
if [ -z "$VERSION" ]; then
  echo "could not determine latest version; pass --version vX.Y.Z" >&2
  exit 1
fi

# Detect OS / arch and pick the matching asset name.
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac
asset="omniclean_${os}_${arch}.tar.gz"

base="https://github.com/${REPO}/releases/download/${VERSION}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading ${asset} (${VERSION})…"
curl -fsSL -o "${tmp}/${asset}" "${base}/${asset}"
curl -fsSL -o "${tmp}/checksums.txt" "${base}/checksums.txt"

echo "Verifying SHA-256…"
expected="$(grep " ${asset}\$" "${tmp}/checksums.txt" | awk '{print $1}')"
if [ -z "$expected" ]; then
  echo "no checksum entry for ${asset} in checksums.txt" >&2
  exit 1
fi
if command -v sha256sum >/dev/null; then
  actual="$(sha256sum "${tmp}/${asset}" | awk '{print $1}')"
else
  actual="$(shasum -a 256 "${tmp}/${asset}" | awk '{print $1}')"
fi
if [ "$expected" != "$actual" ]; then
  echo "checksum mismatch! expected=$expected actual=$actual" >&2
  exit 1
fi

tar -C "$tmp" -xzf "${tmp}/${asset}"
mkdir -p "$INSTALL_DIR"
install -m 0755 "${tmp}/omniclean" "${INSTALL_DIR}/omniclean"
echo "Installed omniclean ${VERSION} to ${INSTALL_DIR}/omniclean"
