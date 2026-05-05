# Handoff Log

A running, reverse-chronological log of session summaries. Newest entries on top.
Each session prepends a brief summary of what was done.

---

## 2026-05-XX — Landing page redesign

**What changed**
- Redesigned `srv/static/landing.html` to reflect the app as it has evolved (Projects / Typesetting / Transmittals / Calendars).
- New hero with eyebrow + serif headline (`Book production, end-to-end.`), framed client-portal card (input + Open button), and a 4-up feature grid summarizing what the studio does.
- Reuses existing CSS variables / theme system (light + dark) and the floating font/dark-mode toolbar — no changes to JS behavior.
- Footer simplified and bumped to © 2026 JDBB; Admin link preserved.
- Verified light + dark visually at http://localhost:8000/.

**Build / deploy**
- Static assets are embedded via `//go:embed static/*` in `srv/server.go`, so changes require a rebuild.
- Rebuilt with `make build` (→ `go build -o prodcal ./cmd/srv`) and restarted the systemd unit: `sudo systemctl restart srv`.

**Process notes**
- Going forward, prepend a new dated entry at the top of this file at the end of each session, summarizing what we did. The user uses this to hand off context between sessions.
