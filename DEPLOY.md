# Deployment Guide

Production Calendar + Manuscript Transmittal

## Current Setup (exe.dev)

- **URL**: https://jdbbs.exe.xyz/
- **Service**: systemd unit `srv` on port 8000
- **Database**: SQLite at `/home/exedev/prodcal/db.sqlite3` (WAL mode)
- **Backups**: Daily at 3 AM to `~/backups/`, 7-day retention
- **Binary**: `/home/exedev/prodcal/prodcal`

## Build & Deploy

```bash
cd /home/exedev/prodcal
make build
sudo systemctl restart srv
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
3. Rebuild and restart: `make build && sudo systemctl restart srv`

## Database: Seeding a New Project

1. Create the project via the admin UI or API:
   ```bash
   curl -X POST http://localhost:8000/api/projects \
     -H 'Content-Type: application/json' \
     -d '{"name": "My Book", "client_slug": "client", "project_slug": "book", "start_date": "2026-01-01"}'
   ```

2. Seed tasks from `seed_data.json`:
   ```bash
   curl -X POST http://localhost:8000/api/projects/1/seed \
     -H 'Content-Type: application/json' \
     -d @seed_data.json
   ```

3. Set the project password:
   ```bash
   curl -X POST http://localhost:8000/api/projects/1/auth \
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
sudo systemctl stop srv
gunzip ~/backups/prodcal-YYYYMMDD-HHMMSS.sqlite3.gz
cp ~/backups/prodcal-YYYYMMDD-HHMMSS.sqlite3 /home/exedev/prodcal/db.sqlite3
sudo systemctl start srv
```

## Health Check

```bash
curl http://localhost:8000/healthz
# {"status":"ok"}
```

## Admin Dashboard

Accessible at `/admin/` — requires exe.dev login (X-ExeDev-UserID header).
Shows all projects with task completion, auth status, transmittal status.

## Docker (Portable)

Not used in production currently, but available for portability:

```bash
docker build -t prodcal .
docker run -p 8000:8000 -v prodcal-data:/app/data prodcal
```

Note: When running in Docker, the DB path should be under `/app/data/`
for persistence. The working directory defaults to `/app`.

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
seed_data.json      ← template task data
```

## Monitoring

- **Logs**: `journalctl -u srv -f`
- **Health**: `curl localhost:8000/healthz`
- **Backups**: `ls -la ~/backups/`
- **Cron log**: `cat ~/backups/backup.log`
