# Four-engine server integration

This development build embeds Soar, CottenDNS, StormDNS and MasterDnsVPN in
the panel process. One public UDP/53 listener dispatches requests by account
and engine domain. The existing TCP/53 listener can frame DNS messages for
the embedded engines too; this does not add TCP resolver support to upstream
clients that only use UDP.

## Accounts and compatibility

The `engines` API field accepts `soar`, `cotten`, `stormdns`, `masterdns`,
`both`, `all`, or a comma-separated combination of engine names. `both`
continues to mean Soar and Cotten, and omitted fields preserve existing
accounts. The UI defaults new accounts to all four. Existing accounts must
be edited explicitly to enable additional engines.

For account zone `ACCOUNT.t.example.org`, the routes are:

| Engine | Domain | Link scheme |
| --- | --- | --- |
| Soar | `ACCOUNT.t.example.org` | `soar://` |
| CottenDNS | `cot.ACCOUNT.t.example.org` | `cottendns://` |
| StormDNS | `storm.ACCOUNT.t.example.org` | `stormdns://` |
| MasterDnsVPN | `master.ACCOUNT.t.example.org` | `masterdns://` |

All routes use the existing NS delegation. No additional IP address,
nameserver record or public port is required. Alternate tunnel domains
produce corresponding profiles for each enabled engine.

The existing `whitedns.profile` version 1 envelope is used for new links,
with `server.domain`, `server.encryption_key` and `server.encryption_method`.
The account create/edit form offers 5 (AES-256-GCM, default) or 1 (XOR, weak)
for Cotten, StormDNS and MasterDNS. Soar keeps its existing encryption.
The Cotten encryption selector applies to Normal mode; Max retains its native
encryption. MasterDNS also accepts the upstream `mdns-1.` domain alias while
exporting the original `master.` domains and key derivation for fork compatibility.
XOR does not provide AES-GCM confidentiality or authentication; it is optional
and is not a guarantee of higher speed or DPI resistance. Changing the method
closes that account’s sessions; refresh its clients’ subscriptions before reconnecting.
Older accounts keep AES and API updates that omit the setting preserve it.
Engine keys use separate HMAC derivation
contexts from the account seed. Rotating the account key revokes all four;
rotating the subscription URL alone does not change tunnel keys.

Quota, expiry, enable/disable, speed and account/server stream limits are
enforced through the existing shared egress and metering path. Tunneled DNS
queries use the same metered, allowlisted resolver path. Engines cannot
open independent, unmetered egress through the embedded API.

StormDNS and MasterDnsVPN retain upstream session table limits (255 sessions
per engine per account). Soar/Cotten session settings remain specific to
those engines. A zero shared stream limit still means unlimited at the
panel level; upstream per-session safeguards remain in effect.

NS tunnel queries beneath the StormDNS route are dispatched to the engine.
Authoritative NS/SOA answers are retained at delegation, account and engine
apexes for QNAME minimisation. TXT/CNAME/SRV packet routing is also preserved.

## Short account domains

In **Server settings → Account domains → Account domain label length**, choose
3–26 characters. The default is 3, including on upgraded databases that have
not set a preference. The existing maximum account-zone length still applies
when using long base domains.

New accounts use the selected length. Labels are allocated with duplicate
checking inside the account write transaction, including alternate-domain
truncation. Internal account IDs remain 128-bit and encryption seeds remain
256-bit. Existing domains, keys, usage and subscriptions are preserved when
changing this setting or upgrading.

To shorten an existing account deliberately, use its **Rotate link and key**
action after choosing the length. This revokes its old key and domain; refresh
the subscription on its clients. This is never done automatically.

## Source and build

To install the custom Linux amd64 bundle on a supported server, upload the
archive and run these commands in an interactive SSH terminal:

```sh
tar -xzf soar-panel-0.14.1-linux-amd64.tar.gz
cd soar-panel-0.14.1-linux-amd64
sudo bash panel-install.sh
```

The existing installer preserves panel accounts, keys and usage. After
upgrading, edit an account's engine selection to enable the new routes,
then refresh its subscription in the client. This fork’s one-command installer
and in-panel updater both use releases from `rjalilocsdg/Soar-Server`. The
original WhiteDNS installer continues to install the original upstream project.

StormDNS remains vendored in `third_party/StormDNS`, with its original MIT
license and pinned commit in `UPSTREAM.json`. MasterDnsVPN uses its pinned
upstream source plus `patches/masterdns-engine.json`; the patch includes native
session response-state synchronization and the embedding API. Cotten uses the
new pinned source and patch in `patches/cotten-engine.json`, including Normal
and Max modes. All are compiled into the panel; no extra daemon is installed.

Prepare the verified workspace with `scripts/prepare_panel_engines.py`,
then build the real native clients for regression tests:

```sh
python3 scripts/prepare_panel_engines.py --workspace .engine-source/patched.go.work
export GOWORK="$PWD/.engine-source/patched.go.work"
go build -o .engine-source/storm-client stormdns-go/cmd/client
go build -o .engine-source/master-client masterdnsvpn-go/cmd/client
go build -o .engine-source/cotten-client ./.engine-source/CottenDNS/cmd/client
export PANEL_STORM_CLIENT="$PWD/.engine-source/storm-client"
export PANEL_MASTER_CLIENT="$PWD/.engine-source/master-client"
export PANEL_COTTEN_CLIENT="$PWD/.engine-source/cotten-client"
go test ./... -count=1 -timeout 240s
go test -race ./internal/panel ./internal/panelops -count=1 -timeout 900s
go vet ./...
python3 scripts/build_panel_bundle.py
```

Native tests cover encrypted TCP transfer, DNS record routing, shared
Soar/new-engine stream admission, usage accounting and key revocation.
`additional_engines` in runtime diagnostics reports per-engine starts,
running instances, startup errors, queue drops and stream counters.
