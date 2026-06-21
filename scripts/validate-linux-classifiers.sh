#!/usr/bin/env bash
# validate-linux-classifiers.sh — exercise the OmniClean cleanup classifier
# against a REAL apt inside a throwaway container, asserting a planted orphan is
# detected. This is the live-apt guarantee unit fixtures cannot give.
#
# Usage:
#   scripts/validate-linux-classifiers.sh            # default: debian:12
#   IMAGE=ubuntu:24.04 scripts/validate-linux-classifiers.sh
#
# Requires: a running Docker daemon. Run from the repo root.
#
# The cleanup classify path is read-only (apt-get autoremove --dry-run); the
# only state this script mutates is the explicit install/remove it performs to
# manufacture an orphan inside the disposable container.
set -euo pipefail

IMAGE="${IMAGE:-debian:12}"
# moreutils reliably pulls a private dependency (libipc-run-perl) that becomes
# orphaned once moreutils itself is removed — a deterministic orphan generator.
ORPHAN_PARENT="moreutils"

echo "==> validating apt classifier in ${IMAGE}"

docker run --rm -v "$PWD":/src -w /src "${IMAGE}" bash -euo pipefail -c '
  export DEBIAN_FRONTEND=noninteractive LC_ALL=C
  apt-get update -qq
  apt-get install -y -qq golang jq '"${ORPHAN_PARENT}"' >/dev/null

  # Build the binary inside the container.
  go build -o /tmp/omniclean ./cmd/omniclean

  # Plant the orphan: remove the parent, leaving its private dep auto-installed
  # and no longer required.
  apt-get remove -y -qq '"${ORPHAN_PARENT}"' >/dev/null

  echo "--- cleanup --json --manager apt ---"
  /tmp/omniclean cleanup --json --manager apt | tee /tmp/out.json
  echo

  # Assert: valid JSON array, and at least one orphan candidate present.
  jq -e "type == \"array\"" /tmp/out.json >/dev/null
  if jq -e "any(.[]; .role == \"orphan\")" /tmp/out.json >/dev/null; then
    echo "PASS: at least one orphan detected by the classifier"
  else
    echo "WARN: no orphan detected — apt may not have produced one on this image"
    exit 3
  fi
'

echo "==> note: flatpak is leaf-only (no orphan signal); verify manually with:"
echo "    flatpak list --app --columns=application   # every entry classifies as Manual"
