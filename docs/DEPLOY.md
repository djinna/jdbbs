# Deployment Guide

JDBBS — Production Calendar + Book Production Pipeline.

## Current setup (exe.dev)

- **URL**: <https://jdbbs.exe.xyz/>
- **Service**: systemd unit `srv` on port 8000
- **Working directory**: `/home/exedev/jdbbs`
- **Binary**: `/home/exedev/jdbbs/jdbbs`
- **Database**: SQLite at `/home/exedev/jdbbs/db.sqlite3` (WAL mode, auto-migrated on startup)
- **Backups**: Daily at 3 AM to `~/backups/`, 7-day retention
- **Typesetting bundle**: `/home/exedev/jdbbs/typesetting/` (templates, scripts, fonts)

## Build & deploy

```bash
cd /home/exedev/jdbbs
git pull origin main
make build
sudo systemctl restart srv
```

## First-time migration from `prodcal`

If you're cutting over from the old `prodcal` deployment:

1. Back up the existing DB:
   ```bash
   sqlite3 /home/exedev/prodcal/db.sqlite3 \
     ".backup /home/exedev/backups/pre-jdbbs-$(date +%Y%m%d).sqlite3"
   ```

2. Clone jdbbs alongside prodcal:
   ```bash
   cd /home/exedev
   git clone https://github.com/djinna/jdbbs.git
   cd jdbbs
   make build
   ```

3. Copy the database into the new working directory:
   ```bash
   cp /home/exedev/prodcal/db.sqlite3 /home/exedev/jdbbs/db.sqlite3
   ```
   (Migrations run automatically on startup.)

4. Carry over the AgentMail env vars:
   ```bash
   cp /home/exedev/prodcal/.env /home/exedev/jdbbs/.env
   ```

5. Update the systemd unit:
   ```bash
   sudo cp srv.service /etc/systemd/system/srv.service
   sudo systemctl daemon-reload
   sudo systemctl restart srv
   ```

6. Verify:
   ```bash
   curl http://localhost:8000/healthz
   # → {"status":"ok"}
   ```

7. Once verified end-to-end, the old `/home/exedev/prodcal/` and
   `/home/exedev/book-production/` directories can be moved aside as backups.

## Runtime dependencies on the VM

| Dependency | Used for | Install |
|---|---|---|
| Go 1.26+ | Build the server binary | already installed |
| SQLite 3 | Database | already installed |
| Typst | PDF typesetting | `curl -fsSL https://typst.community/typst-install/install.sh \| sh` (or use the binary release) |
| Pandoc | DOCX ↔ Markdown / EPUB | `apt install pandoc` |
| Python 3 | Conversion + word-template scripts | already installed |
| `python-docx` | Word template generation | `pip3 install python-docx` |
| `pyyaml` | Corrections YAML parsing | `pip3 install pyyaml` |
| Libertinus Serif | Body font | `apt install fonts-libertinus` (Source Sans + JetBrains Mono are bundled in `typesetting/fonts/`) |

## Database: migrations

Migrations live in `db/migrations/` and follow `NNN-name.sql`. They run
automatically on startup via `db.RunMigrations()`. Each file ends with an
`INSERT OR IGNORE INTO migrations (...)` row.

To add a new migration:

1. Create `db/migrations/012-description.sql`.
2. Append the tracking row.
3. Rebuild + restart: `make build && sudo systemctl restart srv`.

## Database: seeding a new project

```bash
curl -X POST http://localhost:8000/api/projects \
  -H 'Content-Type: application/json' \
  -d '{"name":"My Book","client_slug":"client","project_slug":"book","start_date":"2026-01-01"}'

curl -X POST http://localhost:8000/api/projects/1/seed \
  -H 'Content-Type: application/json' -d @seed_data.json

curl -X POST http://localhost:8000/api/projects/1/auth \
  -H 'Content-Type: application/json' -d '{"password":"mypassword"}'
```

Or use the **Make New** button in the calendar UI to clone an existing project
with shifted dates.

## Database: backup / restore

```bash
# manual
sqlite3 db.sqlite3 ".backup /tmp/jdbbs-backup.sqlite3"

# scripted (the daily cron)
/home/exedev/jdbbs/scripts/backup-db.sh

# restore
sudo systemctl stop srv
gunzip ~/backups/jdbbs-YYYYMMDD-HHMMSS.sqlite3.gz
cp ~/backups/jdbbs-YYYYMMDD-HHMMSS.sqlite3 /home/exedev/jdbbs/db.sqlite3
sudo systemctl start srv
```

## Health check

```bash
curl http://localhost:8000/healthz
# → {"status":"ok"}
```

## Admin dashboard

Accessible at `/admin/` — requires exe.dev login (`X-ExeDev-UserID` header).
Shows all projects with task completion, auth status, transmittal status, and
the Books tab for the conversion pipeline.

## Docker (portable)

Not used in production currently, but supported:

```bash
docker build -t jdbbs .
docker run -p 8000:8000 -v jdbbs-data:/app/data jdbbs
```

The image bundles Typst, Pandoc, Python + python-docx + pyyaml, Libertinus
Serif (apt), and `typesetting/`. DB path defaults under `/app/data/`.

## Architecture

```
cmd/srv/main.go     ← entrypoint
srv/server.go       ← HTTP routes, handlers
srv/admin.go        ← admin dashboard
srv/transmittal.go  ← transmittal API
srv/books.go        ← DOCX → PDF pipeline (calls typesetting/scripts/*)
srv/bookspecs.go    ← book-spec CRUD + Word template generation
srv/epub.go         ← EPUB generation via Pandoc
srv/static/         ← embedded frontend (HTML/JS/CSS)
db/db.go            ← SQLite open, migrations
db/migrations/      ← sequential SQL migrations
db/queries/         ← sqlc query definitions
db/dbgen/           ← sqlc generated code
typesetting/        ← Typst templates, fonts, Python/shell scripts
scripts/backup-db.sh ← daily backup script
```

## Monitoring

- **Logs**: `journalctl -u srv -f`
- **Health**: `curl localhost:8000/healthz`
- **Backups**: `ls -la ~/backups/`
- **Cron log**: `cat ~/backups/backup.log`
