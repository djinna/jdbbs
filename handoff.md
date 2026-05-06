# Handoff Log

A running, reverse-chronological log of session summaries. Newest entries on top.
Each session prepends a brief summary of what was done.

---

## 2026-05-XX (cont.) — Landing forks; password reset affordance; public summary API

**Landing forks** (for design review at `/static/`):
- `landing-alt.html` — left-aligned 2-col with rotated specimen card stack.
- `landing-editorial.html` — single column, drop cap, masthead, small-caps roman-numeral feature list, colophon. "Small press" feel.
- `landing-tease.html` — left col + sticky **live tile** on the right pulling from `/api/public/summary` (active projects, clients, status dots). Shows real numbers.
- `landing-navbar.html` — sticky topbar with brand + portal input inline + admin link, quiet hero + 4-col feature strip below.
- Main `/` still uses the centered redesign from prior session.

**New backend endpoint**
- Added `GET /api/public/summary` in `srv/server.go` returning aggregate counts (`projects_active`, `projects_total`, `clients`). 60s public cache. Used only by the tease fork right now — safe to keep regardless.

**Client portal: "Forgot password?" affordance** (`srv/static/client.html`)
- The client-level password gate now shows a clear link: "Forgot or need a reset?"
- Clicking it opens a `mailto:j@djinna.com` with subject `Portal access — {slug}` and a prefilled body containing the client code and the portal URL, so the request is easy to triage.
- This is the cheap path; if we want self-serve later, a one-time admin-issued token + reset page would be the next step.

**Misc**
- Dropped the `e.g. vgr` placeholder in the main landing portal input — now reads `Your client code`.
- Footer year bumped to 2026.

**Build/deploy**
- Static assets are embedded; rebuild + restart needed: `go build -o prodcal ./cmd/srv && sudo systemctl restart srv`.

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
