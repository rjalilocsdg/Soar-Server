# Multi-engine panel 0.6.1-beta1

Preview release of the Soar-Server fork maintained at
https://github.com/rjalilocsdg/Soar-Server.

- Run Soar, CottenDNS, StormDNS and MasterDnsVPN together on the same server.
- Share account quota, expiry, speed and stream limits across selected engines.
- Export per-engine links and QR codes, or all four in one subscription.
- Preserve existing accounts and the original Soar/Cotten `both` selection.
- Use this fork for both one-command installation and in-panel update checks.

Install or upgrade from an interactive SSH terminal on a supported Ubuntu or
Debian amd64 server:

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/rjalilocsdg/Soar-Server/main/panel-setup.sh)
```

The installer downloads this exact version and checks both the archive checksum
and bundled file checksums before changing services. Existing panel data is
preserved. After upgrading, enable additional engines for an account and refresh
its subscription in the client.

Real upstream clients, browser workflows and focused race tests passed locally.
Linux binaries were cross-compiled on macOS with Go 1.26.0. Two existing Cotten
tests require oversized UDP datagrams unavailable on macOS; public VPS/systemd
and physical Android validation remain outstanding. See
[the validation report](MULTI-ENGINE-VALIDATION.md). This is a prerelease.
