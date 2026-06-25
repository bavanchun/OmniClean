# detector testdata — package manager fixtures

These fixtures lock the manual/orphan parsing for various package managers against representative real-world output.

## Provenance

Captured under a pinned `LC_ALL=C` locale. The `*.txt` files reproduce the verbatim stdout sections of the respective package managers.

### APT (Debian/Ubuntu)
Captured inside a throwaway container (`debian:12` / `ubuntu:24.04`):
```sh
apt-get update -qq
apt-get install -y -qq moreutils
apt-get remove  -y -qq moreutils
LC_ALL=C apt-get autoremove --dry-run   # -> apt-autoremove-<distro>.txt
apt-mark showmanual                     # -> apt-showmanual-<distro>.txt
```

### DNF / YUM (Fedora)
Captured inside a `fedora:latest` container:
```sh
dnf install -y dnf-plugins-core
rpm -e --nodeps dnf-plugins-core
dnf repoquery --userinstalled --qf "%{name}\n"  # -> dnf-userinstalled-fedora44.txt
dnf repoquery --unneeded --qf "%{name}\n"       # -> dnf-unneeded-fedora44.txt
```

### Pacman (Arch Linux)
Captured inside an `archlinux:latest` container:
```sh
pacman -Sy --noconfirm --disable-sandbox-syscalls git
pacman -R --noconfirm --disable-sandbox-syscalls git
pacman -Qeq   # -> pacman-explicit-arch.txt
pacman -Qtdq  # -> pacman-orphans-arch.txt
```

### Zypper (openSUSE)
Captured inside an `opensuse/leap:latest` container:
```sh
zypper --non-interactive install -y git
rpm -e --nodeps git
zypper search -si              # -> zypper-installed-opensuse.txt
zypper packages --unneeded     # -> zypper-orphaned-opensuse.txt
```
