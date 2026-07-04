# ProdCal — Local Single-User Usage (Model A)

This document explains how to run ProdCal as a persistent **local** service on
your own Mac, using its **own data directory** that is completely separate from
the production/review database and from the repo's throwaway `db.sqlite3`.

This is **model A**: one binary, one private data dir, running in the background
for a single trusted user (you). It is also the **seed of the planned Wails
desktop app** — the launcher (`cmd/prodcal-local`) is a clean, reusable core
(data dir → `srv.New` → internal server → loopback proxy) that a desktop shell
will later embed.

## Why a launcher (and not just `./prodcal`)

The production server treats any non-empty `X-ExeDev-UserID` request header as an
authenticated admin. In production that header is injected by the exe.dev proxy
that sits in front of the app. **Locally there is no such proxy**, so the admin
UI is unreachable — hitting `/admin/` just 302-redirects to an exe.dev login flow
that does not exist on your machine.

`prodcal-local` solves this by running a small **loopback-only reverse proxy**:

```
you ─▶ 127.0.0.1:8000 (front proxy, injects admin header)
           └─▶ 127.0.0.1:<ephemeral> (real ProdCal server)
```

The front proxy injects `X-ExeDev-UserID: local-admin` (and
`X-ExeDev-Email: local@localhost`) on every request, then forwards to the real
server running on an ephemeral loopback port. That makes the admin UI work
locally exactly as it does behind the prod proxy.

## Data directory

- Default: `$HOME/Library/Application Support/ProdCal/`
  - `db.sqlite3` — the local database (created + migrated automatically on first run)
  - `.prodcal-secret` — the persisted HMAC secret, so client tokens survive restarts
- Override with `-data <dir>` or the `PRODCAL_LOCAL_DATA` env var.

This is intentionally **not** the repo's `db.sqlite3` (which is a throwaway dev
DB) and **not** any prod/review database. Model A keeps your local data isolated.

## Running it

Fastest path — build and run in one step:

```sh
./scripts/run-local.sh
# or with overrides:
./scripts/run-local.sh -addr 127.0.0.1:8055 -data ./scratch/local-data
```

Or via make:

```sh
make local
```

On startup the launcher:

- creates the data dir (`0755`) and opens/migrates `db.sqlite3` inside it,
- sets `PRODCAL_BASE_URL=http://<addr>` (unless you already set it) so generated
  links point at your local instance instead of prod `exe.xyz`,
- serves the real handler on an ephemeral `127.0.0.1` port,
- serves the header-injecting proxy on `-addr` (default `127.0.0.1:8000`),
- best-effort `open`s the URL in your browser,
- shuts down cleanly on `Ctrl-C` (SIGINT) or SIGTERM.

### Flags and environment

| Flag    | Env                 | Default                                        |
|---------|---------------------|------------------------------------------------|
| `-data` | `PRODCAL_LOCAL_DATA`| `$HOME/Library/Application Support/ProdCal`     |
| `-addr` | `PRODCAL_LOCAL_ADDR`| `127.0.0.1:8000`                                |

Env vars set the defaults; an explicit flag on the command line wins.

## Persistence with launchd (auto-start / keep-alive)

To keep ProdCal running in the background across logins, install the LaunchAgent
in `deploy/local/com.jdbb.prodcal-local.plist`:

```sh
# 1. Build and install the binary where the plist expects it
go build -o prodcal-local ./cmd/prodcal-local
cp prodcal-local /usr/local/bin/prodcal-local          # or edit ProgramArguments

# 2. launchd does NOT expand "~" — substitute your real home into the log paths
sed "s|__HOME__|$HOME|g" deploy/local/com.jdbb.prodcal-local.plist \
    > ~/Library/LaunchAgents/com.jdbb.prodcal-local.plist

# 3. Load it (RunAtLoad + KeepAlive start it now and restart it if it dies)
launchctl load ~/Library/LaunchAgents/com.jdbb.prodcal-local.plist
```

Logs go to `~/Library/Logs/prodcal-local.log`.

To stop / uninstall:

```sh
launchctl unload ~/Library/LaunchAgents/com.jdbb.prodcal-local.plist
rm ~/Library/LaunchAgents/com.jdbb.prodcal-local.plist
```

To pass custom flags (e.g. a different data dir or port), add them as extra
`<string>` entries in the plist's `ProgramArguments` array, or set
`PRODCAL_LOCAL_DATA` / `PRODCAL_LOCAL_ADDR` via an `EnvironmentVariables` dict.

## Security note — read this

- **Loopback only.** The launcher binds only to `127.0.0.1`/`localhost` and
  **refuses to start** if `-addr` resolves to anything else (it explicitly rejects
  `0.0.0.0` and bare `:port`).
- **Unauthenticated by design.** The front proxy injects an admin identity header
  on *every* request. Anything that can reach the port is admin. That is
  acceptable **only** because it is bound to loopback on your single-user machine.
- **NEVER expose this port.** Do not port-forward it, do not put it behind a
  tunnel, do not bind it to a LAN interface. There is no login in front of it.
- This launcher does **not** modify the production auth model — the real server's
  `X-ExeDev-UserID` handling is untouched; the launcher simply supplies the header
  that the prod proxy would otherwise supply.

## Email is intentionally off

The launcher does **not** set `AGENTMAIL_API_KEY` / `AGENTMAIL_INBOX_ID`, so the
server logs `email not configured` and all email pathways are inert. This is
deliberate: a local single-user instance should never send real email. Leave the
`AGENTMAIL_*` variables unset.
