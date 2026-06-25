# Expand Linux support — dnf/yum, pacman, zypper classifiers + snap Classify

**Date**: 2026-06-25 13:42
**Severity**: Low
**Component**: internal/detector (dnf, pacman, zypper, snap) + scripts
**Status**: Resolved

## What Happened

Completed 6-phase implementation:
1. Researched and verified no-sudo, read-only manual vs. orphan commands for DNF, Pacman, and Zypper via containerized environments.
2. Captured real-distro output fixtures under pinned `LC_ALL=C` locale.
3. Implemented DNF detector and classifier utilizing `dnf repoquery` and `rpm -qa` commands.
4. Implemented Pacman detector and classifier with special handling for `pacman -Qtdq` non-zero exit status on clean environments.
5. Implemented Zypper detector and classifier with a generic header-aware tabular parser.
6. Extended existing Snap list-only detector to support leaf-only classification (RoleManual).
7. Parameterized and generalized the containerized validator script (`scripts/validate-linux-classifiers.sh`).
8. Registered all three new package managers in registry and package configurations.
9. Verified all unit tests pass with the race detector and successfully validated the DNF classifier against a live Fedora image.

## Technical Details

- **DNF**:
  - Manual: `dnf repoquery --userinstalled --qf "%{name}\n"`
  - Orphan: `dnf repoquery --unneeded --qf "%{name}\n"`
  - Info: `rpm -qa --qf "%{NAME}\t%{VERSION}-%{RELEASE}\t%{SIZE}\t%{INSTALLTIME}\t%{SUMMARY}\n"`
- **Pacman**:
  - Manual: `pacman -Qeq`
  - Orphan: `pacman -Qtdq`
  - Clean environment: `pacman -Qtdq` exits with code 1. Safely intercepted and mapped to an empty map (no orphans) to avoid degrading manual classification.
- **Zypper**:
  - Header-aware parser: Dynamically detects column offsets (e.g. `Name` and status `S`) from tabular outputs (`zypper search -si` and `zypper packages --unneeded`) to handle any column permutation.
- **Snap**:
  - Classified as `RoleManual` (leaf-only) to avoid systemd-dependent container execution issues and maintain Flatpak parity.

## Lessons Learned

1. **Pacman sandbox / QEMU emulation conflicts**: Under Apple Silicon (arm64) emulation, QEMU struggles with pacman's seccomp restrictions during package operations. Disabling sandboxing via `--disable-sandbox-syscalls` is necessary inside container environments.
2. **Tabular parser header awareness**: Table headers of package manager tools (e.g. Zypper) can change column order based on flags/subcommands. Finding headers dynamically via index lookups is safer than hardcoding column indexes.
3. **RPM installtime field**: Stating Unix timestamps directly from the rpm database avoids heavy disk traversal and provides O(1) file-lookup performance for package listing.

## Next Steps

- Merge `feature/expand-linux-package-managers` to `main`.
- Monitor Fedora / RHEL DNF5 rollout issues (verified DNF5 backward compatibility).
