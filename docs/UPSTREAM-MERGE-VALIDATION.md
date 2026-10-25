# Upstream merge validation — 0.10.1

Upstream: WhiteDNS/Soar-Server `b26b402`. Fork base: `9659e18` (beta3).
Local host: macOS arm64, Go 1.26.0, installed Chrome and Playwright 1.58.2.
Native Linux amd64/arm64 CI is required before publication.

## Compatibility

`TestBeta3UpgradePreservesAccountsProfilesAndSubscriptions` loads accounts and
profile links exported by the actual beta3 code. Test credentials belong to
throwaway fixtures. The test verifies unchanged profile strings, account data,
key seeds and subscription credentials, in-memory defaults for new fields,
omitted-field preservation in older API requests and valid backup inspection.
The fixture covers AES/XOR, four-engine, legacy Soar+Cotten and Soar-only
accounts, and short and long account labels.

MasterDNS uses the upstream pinned engine plus its synchronized client response
state and embedding patch, retaining the fork's original domains and key
context. The upstream domain alias is accepted on the same runtime. StormDNS
uses the original fork source. Cotten uses the new source with shared
Normal/Max session accounting; the fork's explicit Normal-mode encryption
selection is retained in the checksummed embedding patch.

## Local platform limits

The two pre-existing Cotten large-UDP tests are excluded from local ordinary
runs because macOS rejects their oversized datagrams. They remain enabled in
native Linux CI. The host's system Bash is 3.2; modern-Bash installer checks
run in Linux CI. No production VPS or physical-device deployment is performed
by this merge task.

## Verification scope

Local checks include beta3 fixture compatibility, native AES/XOR client traffic,
short-label allocation, settings restarts and anti-flood behavior under Go's
race detector. The embedded Cotten and MasterDNS session tests pass under the
race detector. Both release patches also apply successfully to clean checkouts
of their pinned source revisions and match their complete file manifests.
Python release/bootstrap checks and Go vet pass. Browser checks cover English
and Persian, account editing, encryption, labels, sharing, subscriptions and
responsive layouts.

The publication gate is the `Verify management panel` workflow on both native
Linux architectures, including the full Go/race suite, engine suites, installer
checks, offline bundle verification and both browser acceptance suites. Release
assets are downloaded directly from that successful workflow run.
