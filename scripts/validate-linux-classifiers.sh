#!/usr/bin/env bash
# validate-linux-classifiers.sh — exercise the OmniClean cleanup classifier
# against a REAL package manager inside a throwaway container.
#
# Usage:
#   scripts/validate-linux-classifiers.sh            # default: debian:12 (apt)
#   IMAGE=fedora:latest scripts/validate-linux-classifiers.sh
#   IMAGE=archlinux:latest scripts/validate-linux-classifiers.sh
#   IMAGE=opensuse/leap:latest scripts/validate-linux-classifiers.sh
#
# Requires: a running Docker daemon. Run from the repo root.
set -euo pipefail

IMAGE="${IMAGE:-debian:12}"
echo "==> validating classifier in ${IMAGE}"

# Detect manager based on image name
if [[ "${IMAGE}" =~ "debian" || "${IMAGE}" =~ "ubuntu" ]]; then
  MANAGER="apt"
  ORPHAN_PARENT="moreutils"
  INSTALL_CMD="apt-get update -qq && apt-get install -y -qq golang jq ${ORPHAN_PARENT} >/dev/null"
  REMOVE_CMD="apt-get remove -y -qq ${ORPHAN_PARENT} >/dev/null"
  ENV_ENV="export DEBIAN_FRONTEND=noninteractive LC_ALL=C"
elif [[ "${IMAGE}" =~ "fedora" ]]; then
  MANAGER="dnf"
  ORPHAN_PARENT="dnf-plugins-core"
  INSTALL_CMD="dnf install -y golang jq ${ORPHAN_PARENT} >/dev/null"
  REMOVE_CMD="rpm -e --nodeps ${ORPHAN_PARENT}"
  ENV_ENV="export LC_ALL=C"
elif [[ "${IMAGE}" =~ "archlinux" ]]; then
  MANAGER="pacman"
  ORPHAN_PARENT="git"
  INSTALL_CMD="pacman -Sy --noconfirm --disable-sandbox-syscalls go jq ${ORPHAN_PARENT} >/dev/null"
  REMOVE_CMD="pacman -R --noconfirm --disable-sandbox-syscalls ${ORPHAN_PARENT} >/dev/null"
  ENV_ENV="export LC_ALL=C"
elif [[ "${IMAGE}" =~ "opensuse" ]]; then
  MANAGER="zypper"
  ORPHAN_PARENT="git"
  INSTALL_CMD="zypper --non-interactive install -y go jq ${ORPHAN_PARENT} >/dev/null"
  REMOVE_CMD="rpm -e --nodeps ${ORPHAN_PARENT}"
  ENV_ENV="export LC_ALL=C"
else
  echo "Unsupported image: ${IMAGE}"
  exit 1
fi

docker run --security-opt seccomp=unconfined --rm -v "$PWD":/src -w /src "${IMAGE}" bash -euo pipefail -c "
  ${ENV_ENV}
  echo '--- Installing dependencies ---'
  ${INSTALL_CMD}

  # Build the binary inside the container.
  echo '--- Building OmniClean ---'
  go build -o /tmp/omniclean ./cmd/omniclean

  # Plant the orphan.
  echo '--- Planting orphan ---'
  ${REMOVE_CMD}

  echo '--- cleanup --json --manager ${MANAGER} ---'
  /tmp/omniclean cleanup --json --manager ${MANAGER} | tee /tmp/out.json
  echo

  # Assert: valid JSON array, and at least one orphan candidate present.
  jq -e 'type == \"array\"' /tmp/out.json >/dev/null
  if jq -e 'any(.[]; .role == \"orphan\")' /tmp/out.json >/dev/null; then
    echo 'PASS: at least one orphan detected by the classifier'
  else
    echo 'WARN: no orphan detected'
    exit 3
  fi
"
