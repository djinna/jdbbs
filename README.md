# JDBBS

**Production Calendar + Book Production** in one Go binary.

JDBBS unifies two former repos:

| Former repo | Role |
|---|---|
| [`djinna/prodcal`](https://github.com/djinna/prodcal) | Go web app: project scheduling, transmittals, book specs, client portals, AgentMail integration |
| [`djinna/book-prod`](https://github.com/djinna/book-prod) | Typst-based book production pipeline: templates, fonts, DOCX/MD → PDF/EPUB scripts |

Live at **<https://jdbbs.exe.xyz>**, hosted on the exe.dev VM at `/home/exedev/jdbbs/`.

## Quick start (development)

```bash
# 1. Build
make build                     # → ./jdbbs

# 2. Run (defaults to port 8000, sqlite at ./db.sqlite3)
./jdbbs -listen :8000

# 3. Health check
curl http://localhost:8000/healthz
```

`make test` runs the Go test suite. The book-production tests need Python +
`python-docx` + `pyyaml`, and the PDF/EPUB pipelines additionally need
[Typst](https://typst.app/), [Pandoc](https://pandoc.org/), and a system
Libertinus Serif font. Use `make typeset-deps` to install the Python pieces.

## Deployment

Production runs as a systemd unit (`srv.service`) on the exe.dev VM:

```bash
cd /home/exedev/jdbbs
git pull origin main
make build
sudo systemctl restart srv
```

DNS (`jdbbs.exe.xyz`) is handled by the exe.dev HTTPS proxy, which terminates
TLS and forwards to localhost:8000. See `docs/DEPLOY.md` for the full setup.

## Repo layout

```
cmd/srv/main.go      Binary entrypoint
srv/                 HTTP handlers + embedded static SPA + tests
  static/            admin.html, app.js, transmittal.*, client.html, ...
  EMAIL_SYSTEM.md    Email pathways (6 total)
db/                  SQLite open, migrations, sqlc queries + dbgen
typesetting/
  templates/         Typst master template + Pandoc + Word + EPUB
  scripts/           md-to-chapter, generate-word-template, apply-corrections, ...
  filters/           Pandoc Lua filters
  fonts/             Source Sans 3, JetBrains Mono (Libertinus is system-installed)
manuscripts/         Sample + ghosts source files (.typ, .md, .docx)
reference/           Reference PDFs/EPUBs + extracted internals
corrections/         Example corrections ledger (YAML)
docs/                Architecture, deploy, typography, workflow
scripts/backup-db.sh Daily backup helper
```

## Configuration

Environment variables (all optional except AgentMail):

| Variable | Purpose |
|---|---|
| `AGENTMAIL_API_KEY` | AgentMail authentication |
| `AGENTMAIL_INBOX_ID` | AgentMail sending inbox |
| `PRODCAL_BASE_URL` | Override base URL (default: `https://{hostname}.exe.xyz`) |
| `JDBBS_TYPESETTING_DIR` | Override path to `typesetting/` (default: `./typesetting`) |

## See also

- `AGENTS.md` — guidance for AI coding assistants
- `MIGRATION_LOG.md` — what was merged from prodcal + book-prod, open reconciliation items
- `srv/EMAIL_SYSTEM.md` — email architecture
- `docs/DEPLOY.md` — deployment and operational runbook
- `docs/CHECKPOINTS.md` — checkpoint tags and rollback
