# Multi-engine panel 0.6.1-beta3

Adds an account encryption selector for Cotten, StormDNS and MasterDNS.

- In **New account** or **Edit account**, choose **AES-256-GCM** (default) or
  **XOR — weak**. The choice applies to all three DNS engines in that account.
- XOR is weak encryption and has no AES-GCM authentication tag. Choose it only
  when its tradeoff is acceptable; no speed or DPI-evasion gain is promised.
- Soar keeps its existing encryption. Existing accounts continue using AES.
- Changing the selection closes the account's sessions. Refresh the client
  subscription or import the updated link before reconnecting.
- Saved settings, domains, keys, quotas and subscription URLs are preserved.
- Older API callers may omit `encryption_method`; the current choice is kept.
  Explicit supported values are 1 (XOR) and 5 (AES-256-GCM).
- Retains the short domain settings from beta2 and the four-engine integration.

Install or upgrade on supported Ubuntu/Debian amd64:

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/rjalilocsdg/Soar-Server/main/panel-setup.sh)
```

The installer verifies archive and bundled checksums. See
[validation details](MULTI-ENGINE-VALIDATION.md) for local test results and limits.
