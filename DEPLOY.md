# Deployment Guide

Production Calendar + Manuscript Transmittal

## Current Setup (exe.dev)

- **URL**: https://jdbbs.exe.xyz/
- **Service**: systemd unit `prodcal.service` on port 8000 (unit file `prodcal.service` in the repo root)
- **Database**: SQLite at `/home/exedev/prodcal/db.sqlite3` (WAL mode)
- **Backups**: Daily at 3 AM to `~/backups/`, 7-day retention
- **Binary**: `/home/exedev/prodcal/prodcal`

## Build & Deploy

```bash
cd /home/exedev/prodcal
make build
sudo systemctl restart prodcal
```

## Pipeline tools (VM)

The build pipeline shells out to tools that are **not** managed by `make build`:

| Tool | Required | Installed how | Check |
|---|---|---|---|
| pandoc | **≥ 3.2** (`-t typst+smart` needs the smart extension for the typst writer; 3.1.3 from Ubuntu apt fails with `exit 23`) | GitHub .deb: `curl -sLO https://github.com/jgm/pandoc/releases/download/3.11/pandoc-3.11-1-amd64.deb && sudo dpkg -i pandoc-3.11-1-amd64.deb` | `pandoc --version` |
| typst | 0.13.1 (≥ 0.13 required: series template uses `text(costs:)`; 0.12.0 kept as `/usr/local/bin/typst-0.12` for rollback) | `/usr/local/bin/typst` | `typst --version` |
| python3 + python-docx | any recent | apt / pip | `python3 -c 'import docx'` |
| fonttools (python) | optional, ≥ 4.x | pip (`pip install fonttools`) | `python3 -c 'import fontTools'` — subsets the Noto CJK/Thai fallback fonts in EPUBs (~1 MB instead of 16 MB); without it the full fonts are embedded |

No service restart is needed after upgrading these — they are exec'd per build.
Smoke: upload a small .docx from admin and confirm status reaches `ready`.
(2026-09-03: found the VM at pandoc 3.1.3 with every build failing since May.)

## Mail: Resend from mail.jdbb.studio (since 2026-09-18)

`.env` has `PRODCAL_MAIL_FROM=factory@mail.jdbb.studio`; the app posts to the
exe.dev Resend proxy (`https://resend.int.exe.xyz`, sending-only key injected
at the edge — nothing on the VM). DNS for `mail.jdbb.studio` lives at
Porkbun: CNAMEs `send.mail` + `rsend.mail` → `*.forge.rmta.net`, TXT
`resend._domainkey.mail` (DKIM), TXT `_dmarc` (p=none, reports to Jenna).
Roll back to AgentMail: delete the `PRODCAL_MAIL_FROM` line, restart.
Details in `srv/EMAIL_SYSTEM.md`.

## Store: sandbox → live Stripe (scheduled flip)

The Factory Pass store talks to Stripe through the exe.dev proxy. Which
account it hits is one line in `.env`:

- `PRODCAL_STRIPE_URL=https://stripe-test.int.exe.xyz` → test key (sandbox, `cs_test_…`)
- line absent → `https://stripe.int.exe.xyz`, the live key (`rk_live_…`, attached 2026-09-17)

`scripts/store-go-live.sh` deletes that line, restarts `prodcal` (which
creates the live catalog + WORKSHOP49/PROTOCOL50 coupons at boot via
`store.ensureCatalog`), checks `/api/public/store/config`, and emails Jenna.
Idempotent; `--dry-run` edits a copy under `scratch/` and touches nothing.

Scheduled once by `prodcal-store-live.timer` (units in `deploy/`) for
**Tue 22 Sep 2026 16:00 UTC = Wed 23 Sep 00:00 Hong Kong** — workshop
attendees build free through Tuesday HKT, billing starts Wednesday.

```
systemctl list-timers prodcal-store-live.timer      # when it fires
sudo systemctl start prodcal-store-live.service     # fire it now instead
sudo systemctl disable --now prodcal-store-live.timer   # cancel
journalctl -u prodcal-store-live.service            # what it did
```

Roll back: put the `PRODCAL_STRIPE_URL` line back (copy kept as
`.env.pre-live.<stamp>`) and `sudo systemctl restart prodcal`.

### Index add-on (`index`, $100) — same switch

The back-of-book index is a catalog item like `builds-3` (`srv/store.go`,
lookup key `index`, fulfilled by `passes.index_included = 1`). It follows the
same Stripe account switch: **while the test key is in, the workshop room
"buys" it with a `4242 4242 4242 4242` card and pays nothing** — that is the
free-for-the-workshop path Jenna okayed (2026-09-20). When `store-go-live.sh`
flips to the live key, `ensureCatalog` creates the live `index` price and the
$100 becomes real. To give it away after that, grant it instead of selling it:
`POST /api/admin/passes/{id}/index` (admin header), or the **+ index** button on
the pass row at `/admin/store/`. Drafting costs the studio ≈ $0.45 per 100 pages at the LLM gateway
(`INDEXER_LLM_MODEL`, default `claude-sonnet-4-5`) — a pass can re-draft as
often as it likes, so watch `factory_events` kind `index.drafted` if that
ever matters.

## Database: Migrations

Migrations live in `db/migrations/` and follow the naming pattern `NNN-name.sql`.
They run automatically on startup — the server calls `db.RunMigrations()` at boot.

Each migration file must end with an `INSERT OR IGNORE INTO migrations` to record itself.

To add a new migration:

1. Create `db/migrations/005-description.sql`
2. Include the tracking insert:
   ```sql
   INSERT OR IGNORE INTO migrations (migration_number, migration_name)
   VALUES (005, '005-description');
   ```
3. Rebuild and restart: `make build && sudo systemctl restart prodcal`

## Database: Seeding a New Project

Admin-gated API calls need the exe.dev admin header. The proxy injects it for
logged-in browser sessions, but curl from localhost on the VM must pass it
explicitly — without `-H 'X-ExeDev-UserID: admin'` these calls return 401.
(Easier alternative: create and seed projects through the admin UI at `/admin/`.)

1. Create the project via the admin UI or API:
   ```bash
   curl -X POST http://localhost:8000/api/projects \
     -H 'X-ExeDev-UserID: admin' \
     -H 'Content-Type: application/json' \
     -d '{"name": "My Book", "client_slug": "client", "project_slug": "book", "start_date": "2026-01-01"}'
   ```

2. Seed tasks from a JSON task list (`{"tasks": [...], "start_date": "..."}`):
   ```bash
   curl -X POST http://localhost:8000/api/projects/1/seed \
     -H 'X-ExeDev-UserID: admin' \
     -H 'Content-Type: application/json' \
     -d @tasks.json
   ```

3. Set the project password:
   ```bash
   curl -X POST http://localhost:8000/api/projects/1/auth \
     -H 'X-ExeDev-UserID: admin' \
     -H 'Content-Type: application/json' \
     -d '{"password": "mypassword"}'
   ```

Or use the "Make New" button in the calendar UI to duplicate an existing project
with shifted dates.

## Database: Manual Backup

```bash
/home/exedev/prodcal/scripts/backup-db.sh
```

Or manually:
```bash
sqlite3 db.sqlite3 ".backup /tmp/prodcal-backup.sqlite3"
```

## Database: Restore from Backup

```bash
sudo systemctl stop prodcal.service
# decompress WITHOUT consuming the backup (-k keeps the .gz), or use zcat
zcat ~/backups/prodcal-YYYYMMDD-HHMMSS.sqlite3.gz > /home/exedev/prodcal/db.sqlite3
# drop stale WAL/shm from the stopped server so SQLite doesn't merge them into the restore
rm -f /home/exedev/prodcal/db.sqlite3-wal /home/exedev/prodcal/db.sqlite3-shm
sudo systemctl start prodcal.service
curl -s http://localhost:8000/healthz   # expect {"status":"ok"}
```

## Database: R2 Restore Drill

`scripts/r2-restore-drill.sh` exercises the disaster-recovery path end to end:
it downloads the latest backup from R2 and integrity-checks it. Run it monthly
via cron (suggested: `30 4 1 * *`). A failure writes a `.LAST-R2-DRILL-FAILURE`
sentinel in the backup dir, which surfaces
as a problem in the admin backup-status endpoint (`/api/admin/backup-status`).

## Health Check

```bash
curl http://localhost:8000/healthz
# {"status":"ok"}
```

## Admin Dashboard

Accessible at `/admin/` — requires exe.dev login (X-ExeDev-UserID header).
Shows all projects with task completion, auth status, transmittal status.

## Architecture

```
cmd/srv/main.go     ← entrypoint
srv/server.go       ← HTTP routes, handlers
srv/admin.go        ← admin dashboard
srv/transmittal.go  ← transmittal API
srv/static/         ← embedded frontend (HTML/JS/CSS)
db/db.go            ← SQLite open, migrations
db/migrations/      ← sequential SQL migrations
db/queries/         ← sqlc query definitions
db/dbgen/           ← sqlc generated code
scripts/backup-db.sh ← daily backup script
scripts/sync-to-r2.sh ← offsite R2 sync
```

## Monitoring

- **Logs**: `journalctl -u prodcal -f`
- **Health**: `curl localhost:8000/healthz`
- **Backups**: `ls -la ~/backups/`
- **Cron log**: `cat ~/backups/backup.log`
