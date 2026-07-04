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
