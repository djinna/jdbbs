# ProdCal

Book production management for a small press: production calendars, manuscript
transmittals, a per-project file log + journal, a Word → EPUB / print-PDF
typesetting pipeline, and a password-gated client portal. Go + SQLite with a
vanilla-JS frontend embedded in the binary.

## Build

```bash
go build -o prodcal ./cmd/srv    # or: make build
```

The `-o prodcal` is required — without it the binary is named `srv`, which
collides with the `srv/` source directory.

## Test

```bash
go test ./...
```

The Word-template tests (`srv/word_template_test.go`) shell out to `python3` +
`python-docx`, so they only pass where the doc pipeline is installed (the VM).
Locally, expect that one failure unless you `pip install python-docx`.

## Run locally

```bash
./scripts/run-local.sh    # or: make local
```

This starts a loopback-only launcher that injects the admin header in front of
the unmodified server, with its own data directory. See `docs/LOCAL-USAGE.md`.

## Deploy

Hub-and-spoke git flow (details in `AGENTS.md`): commit locally → push to
GitHub → then on the VM:

```bash
git pull --ff-only origin main
make build
sudo systemctl restart prodcal
```

Production runs as the systemd unit `prodcal.service` on port 8000 behind the
exe.dev HTTPS proxy, which injects `X-ExeDev-UserID` / `X-ExeDev-Email` for
admin auth. Logs: `journalctl -u prodcal -f`. See `DEPLOY.md` for the full
runbook and `CHECKPOINTS.md` for checkpoint tags + rollback.

## Data

- SQLite at `db.sqlite3` (WAL mode); migrations in `db/migrations/` run
  automatically at startup.
- Backups: daily `scripts/backup-db.sh` to `~/backups/`, synced offsite to R2
  (`scripts/sync-to-r2.sh`); status surfaces in the admin dashboard.

## Email

Outbound email goes through AgentMail. Read `srv/EMAIL_SYSTEM.md` before
modifying any email pathway.

## Code layout

- `cmd/srv` — main package (server binary entrypoint)
- `cmd/prodcal-local`, `cmd/prodcal-app` — Mac-only local launcher / desktop
  app over `internal/localrun` (see `docs/LOCAL-USAGE.md`)
- `srv` — HTTP handlers, email, typesetting pipeline; `srv/static` — embedded
  frontend (HTML/JS/CSS)
- `db` — SQLite open + migrations; `db/queries` + `db/dbgen` — sqlc
- `scripts` — backups, R2 sync, local run, dev setup
- `docs` — pipeline, local usage, API notes; `docs/reviews` — launch triage
