# WhiteDNS Server Management 0.10.1 — fork integration

Merges upstream through `b26b402` into rjalilocsdg/Soar-Server.

- Adds Telegram management, native Linux amd64/arm64 bundles, outbound anti-flood
  controls and quarantine, the updated UI and sidebar folding, and managed
  CottenMax alongside Normal mode.
- Uses the updated Soar/Cotten sources and patched MasterDNS session handling.
- Preserves StormDNS, three-character account labels (configurable 3–26), and
  optional XOR for Cotten Normal, StormDNS and MasterDNS. AES-256-GCM is the
  default; CottenMax retains its native encryption.
- Preserves beta3 account domains, key derivations, subscription credentials,
  exported profiles, usage and selected engines. Existing `all` accounts retain
  all four engines. MasterDNS also accepts upstream's `mdns-1` alias.
- Adds upstream's separate MasterDNS session limit (default 255); omitted fields
  preserve existing choices in older API and Telegram account updates.
- Saving server settings restarts active account engines so changes take effect;
  clients reconnect with their existing credentials.
- The installer and panel updater stay on this fork. The one-command installer
  discovers the latest stable fork release and chooses amd64 or arm64.

Install or upgrade on supported Ubuntu/Debian:

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/rjalilocsdg/Soar-Server/main/panel-setup.sh)
```

Refresh the browser after upgrading. Refresh client subscriptions when changing
account engines or encryption. See [validation](UPSTREAM-MERGE-VALIDATION.md).
