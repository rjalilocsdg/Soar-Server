# WhiteDNS Server Management 0.14.1 — fork integration

Updates rjalilocsdg/Soar-Server from 0.10.1 through upstream `1079be8` (0.14.0).

- Adds upstream diagnostics, optional AI assistance, Telegram emergency recovery,
  SSH management, listener fixes and UI performance improvements.
- Adds independent outbound proxy routes and optional traffic reports for all
  four engines, including StormDNS. Proxy failures do not fall back to direct
  access. Destination logging remains disabled by default.
- Updates Soar from v0.2.3 to v0.2.4, including the unused client stream fix.
- Updates MasterDnsVPN to release `v2026.10.26.100415-5f83ee9`; preserves the
  embedding, response synchronization and terminal stream cleanup patches.
- Preserves StormDNS, Cotten Normal/Max, short account IDs, optional XOR,
  existing account domains, keys, subscriptions, quotas and usage.
- Installer and panel updater continue using this fork's stable releases.

Upgrade from **Panel updates**, or run the existing one-command installer:

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/rjalilocsdg/Soar-Server/main/panel-setup.sh)
```

Refresh your browser after upgrading. No account recreation is required.
