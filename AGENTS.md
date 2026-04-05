# Agent Instructions

This is a Go web application (ProdCal) for book production management on exe.dev.

See README.md for build/deploy basics.

## Key Architecture Docs

- **`srv/EMAIL_SYSTEM.md`** — Complete reference for all email pathways (6 total: 4 manual, 1 automatic). Read before modifying any email code.
- **`DEPLOY.md`** — Deployment and hosting notes
- **`CHECKPOINTS.md`** — Checkpoint tags and rollback workflow

## Quick Reference

- Build: `make build && sudo systemctl restart srv`
- Logs: `journalctl -u srv -f`
- Port: 8000 (proxied via exe.dev HTTPS)
- DB: SQLite at `db.sqlite3`, migrations in `db/migrations/`
- Email: AgentMail API, see `srv/EMAIL_SYSTEM.md`
