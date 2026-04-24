# Agent Instructions

JDBBS is a **unified monorepo** combining:

- **`djinna/prodcal`** — Go web app for production-calendar + book project management
- **`djinna/book-prod`** — Typst-based book production pipeline (templates, scripts, fonts)

Both upstream repos are now archival; all changes happen here.

See `README.md` for build/deploy basics and `MIGRATION_LOG.md` for the migration history and open reconciliation items.

## Layout

```
cmd/srv/main.go      Go entrypoint
srv/                 HTTP handlers, embedded static SPA, tests
db/                  SQLite open + migrations + sqlc-generated queries
typesetting/         Typst templates, fonts, conversion scripts, Pandoc filter
manuscripts/         Sample + reference book sources (.typ, .md, .docx)
reference/           Reference PDFs/EPUBs for design comparison
corrections/         Example corrections ledgers
docs/                Architecture, deployment, typography, workflow docs
scripts/             Operational scripts (e.g. backup-db.sh)
```

## Key Architecture Docs

- **`srv/EMAIL_SYSTEM.md`** — Reference for all email pathways (6 total). Read before modifying any email code.
- **`docs/DEPLOY.md`** — Deployment and hosting notes (exe.dev VM, systemd, DNS).
- **`docs/CHECKPOINTS.md`** — Checkpoint tags and rollback workflow.
- **`docs/TYPOGRAPHY.md`** / **`docs/WORKFLOW.md`** — Book production design + workflow.
- **`MIGRATION_LOG.md`** — What was migrated from prodcal/book-prod and what reconciliation work is still open.

## Quick Reference

- Build: `make build && sudo systemctl restart srv`
- Logs: `journalctl -u srv -f`
- Port: 8000 (proxied via exe.dev HTTPS — `https://jdbbs.exe.xyz`)
- DB: SQLite at `db.sqlite3` (auto-migrates on startup), migrations in `db/migrations/`
- Email: AgentMail API, see `srv/EMAIL_SYSTEM.md`

## Typesetting integration

The Go server shells out to scripts and templates under `typesetting/`:

- `typesetting/scripts/md-to-chapter.py` — DOCX → Markdown → Typst chapter pipeline (called from `srv/books.go`)
- `typesetting/scripts/generate-word-template.py` — Spec JSON → styled `.docx` (called from `srv/bookspecs.go`)
- `typesetting/templates/series-template.typ` — Master Typst layout

The path is resolved at runtime by `srv.typesettingRoot()`:

1. `JDBBS_TYPESETTING_DIR` env var, if set
2. `./typesetting` relative to CWD (production)
3. Walk up parents looking for `typesetting/` (handy for `go test`)

When changing book-production logic, **also re-run the affected `srv/...` tests** — many integration paths flow through Go handlers.

## Don't

- Don't reintroduce hardcoded `/home/exedev/...` paths. Use `typesettingRoot()` or relative paths.
- Don't commit to `djinna/prodcal` or `djinna/book-prod` — those repos are archival.
- Don't hand-edit `db/dbgen/*.sql.go` — regenerate with sqlc.
