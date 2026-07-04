# Agent Instructions

This is a Go web application (ProdCal) for book production management on exe.dev.

See README.md for build/deploy basics.

## Key Architecture Docs

- **`srv/EMAIL_SYSTEM.md`** — Complete reference for all email pathways (5 total: 4 manual, 1 automatic). Read before modifying any email code.
- **`DEPLOY.md`** — Deployment and hosting notes
- **`CHECKPOINTS.md`** — Checkpoint tags and rollback workflow

## Quick Reference

- Build + restart (VM): `make build && sudo systemctl restart prodcal`
- Logs (VM): `journalctl -u prodcal -f`  — the systemd unit is `prodcal.service` (older docs say `srv`; that unit does not exist)
- Port: 8000 (proxied via exe.dev HTTPS)
- DB: SQLite at `db.sqlite3`, migrations in `db/migrations/`
- Email: AgentMail API, see `srv/EMAIL_SYSTEM.md`

## Environments

ProdCal runs in two places:

- **VM `jdbbs.exe.xyz` (canonical / deploy).** `prodcal.service` on port 8000 behind the exe.dev HTTPS proxy, which injects `X-ExeDev-UserID` (= admin) and `X-ExeDev-Email`. Has the full doc pipeline (pandoc/typst/python-docx) + licensed fonts. Most dev happens here (Shelley sessions); source of truth for print output.
- **Local Mac (dogfood / offline).** Same Go server, its own SQLite data dir, email off. There is no exe.dev proxy locally, so admin is reached via the loopback launcher below (it injects the admin header) — never by weakening the server's auth.

### Git flow — hub-and-spoke (avoids VM↔GitHub divergence)

Commit locally → push to GitHub (`origin`) → on the VM `git pull --ff-only origin main` → `make build` → `sudo systemctl restart prodcal`. The VM only ever fast-forwards and never pushes divergent commits. (If the VM must publish, it can over SSH: `git push git@github.com:djinna/jdbbs.git main`.) The go.mod pins `go 1.26.4`; the VM has `GOTOOLCHAIN=auto`, so `make build` auto-fetches the toolchain if the installed Go is older.

### Local usage & desktop app — see `docs/LOCAL-USAGE.md`

- `cmd/prodcal-local` — headless persistent local service: a loopback-only reverse proxy that injects the admin header in front of the **unmodified** server; own data dir under `~/Library/Application Support/ProdCal`. Run with `./scripts/run-local.sh` or `make local`.
- `cmd/prodcal-app` (`//go:build darwin`) — a native WebKit window over the same launcher core (`internal/localrun`); the personal macOS desktop-app prototype (Word → EPUB/print-PDF, same UI as the web app).
- These are Mac-only conveniences. The deploy build (`make build` → `./cmd/srv`) is unaffected, and the darwin-tagged desktop app never compiles on the Linux VM.

## Reviews

- **`docs/reviews/LAUNCH-TRIAGE.md`** — pre-launch code + UX review; blockers fixed, HIGH tier tracked.
- **`docs/reviews/DESKTOP-APP-FEASIBILITY.md`** — macOS app plan + prototype notes.
