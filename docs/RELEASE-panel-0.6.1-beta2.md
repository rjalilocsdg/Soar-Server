# Multi-engine panel 0.6.1-beta2

This preview adds configurable short account domains to the four-engine panel.

- Choose **Account domain label length** in **Server settings → Account domains**.
- New accounts default to three characters; choose 3–26 in the panel.
- Duplicate labels are rejected and regenerated inside the database transaction.
- Full internal IDs and encryption keys retain their original lengths.
- Existing account domains, keys, quotas and subscription URLs remain unchanged.
- Explicitly using **Rotate link and key** applies the current length to an old
  account and revokes its old domain/key. Refresh its clients' subscriptions.

Install or upgrade on a supported Ubuntu/Debian amd64 server:

```sh
bash <(curl -fsSL https://raw.githubusercontent.com/rjalilocsdg/Soar-Server/main/panel-setup.sh)
```

The installer pins this version and verifies the archive and bundled checksums.
This is a prerelease; see [multi-engine documentation](MULTI-ENGINE.md) and
[the validation report](MULTI-ENGINE-VALIDATION.md).
