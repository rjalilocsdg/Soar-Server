# Panel 0.14.2

- Add/Edit Account offers Soar server 0.2.4 (default) or pinned 0.2.3. Existing accounts retain the default. Changing versions restarts that account; links, keys and usage remain unchanged. This does not change the client app engine.
- Older versions button beside update checking lists official stable releases from 0.10.1 onward. Downgrading requires the administrator password and confirmation.
- Downloads and checksums are verified before stopping the panel. A private backup is taken and the selected binary must read the account database before activation. Startup failure rolls back the panel and data. Current maintenance tools are retained for future updates.
- Releases older than 0.14.1 are blocked while engine proxy routes are configured, preventing unintended direct outbound connections. Older versions may lack newer settings; avoid editing accounts there if preserving newer preferences is required.
