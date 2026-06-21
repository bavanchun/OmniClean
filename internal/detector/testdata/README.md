# detector testdata — apt classifier fixtures

These fixtures lock the apt manual/orphan parsing (`parseAptRemv`, `apt-mark
showmanual`) against representative real-world output.

## Provenance

Captured under a pinned `LC_ALL=C` locale. The `*.txt` files reproduce the
verbatim stdout sections of:

```sh
# inside a throwaway container (debian:12 / ubuntu:24.04):
apt-get update -qq
apt-get install -y -qq <pkg-with-autodeps>
apt-get remove  -y -qq <pkg>            # leaves its auto-deps orphaned
LC_ALL=C apt-get autoremove --dry-run   # -> apt-autoremove-<distro>.txt
apt-mark showmanual                      # -> apt-showmanual-<distro>.txt
```

The `Remv {pkg} [version]` line format is stable across Debian 11/12,
Ubuntu 22.04/24.04, and apt 2.x (verified by research report
`plans/reports/researcher-260621-1146-apt-orphan-detection-parsing-stability-report.md`).

`apt-autoremove-translated-de.txt` is a German-locale sample whose removal
lines are translated (`Entf` instead of `Remv`). It documents the exact failure
the `LC_ALL=C` pin (DefaultRunner) prevents: the parser must find **zero**
orphans and never panic, rather than mis-detecting.

> Note: these representative fixtures were authored from the research-verified
> literal format because a container runtime was unavailable at capture time.
> The live-apt guarantee is provided by the CI `classify-smoke` job, which runs
> the binary against real apt. Refresh these files from a real container when
> convenient using the commands above.
