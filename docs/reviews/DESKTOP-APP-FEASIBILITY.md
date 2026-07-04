# Feasibility: ProdCal as a personal macOS app (Vellum-style)

Date: 2026-07-04 · Status: **assessment only, nothing built** · Scope decisions locked in with JD:
- **Personal use only** — single machine (JD's Mac), **no App Store, no distribution to others.**
- Goal: Word doc in → **EPUB + print PDF out**, with the same affordances as the web app.
- Local persistence model **A** (separate sandbox, own data, not synced to prod).

## Verdict
**Feasible and low-effort.** ProdCal is already a local-capable web app whose engine *is* Vellum's value prop (DOCX → EPUB + print PDF). The distance to a macOS app is **a native window around the existing local server**, not a rewrite. Because it's personal + undistributed, the two hard parts (notarization, font-license redistribution) **evaporate**. The persistent-local sandbox (model A) is ~80% of the finished product.

## Why it's close
- Backend is a self-contained Go binary: `-listen` flag, auto-migrates SQLite on boot, email off unless `AGENTMAIL_*` set, static SPA embedded via `//go:embed`.
- The book engine already exists end-to-end: upload → book spec → preflight → EPUB/PDF, driven by `pandoc` + `typst` + `python3`.

## What changes because it's personal / undistributed
| Concern | If distributed | Personal-only (our case) |
|---|---|---|
| Apple notarization + Developer ID + hardened runtime | required, real overhead | **not needed** — ad-hoc/unsigned runs locally |
| Licensed fonts (`plantinMTpro`, `proximanova`) embedded | redistribution license question (possible blocker) | **non-issue** — nothing is redistributed; fonts you already license, on your own machine |
| Bundling `pandoc`/`typst`/`python3` into the `.app` | required for portability | **optional** — the app can call the tools already installed via Homebrew |

## The three gaps (in effort order)
1. **Native shell — small (days).** Use **Wails** (Go backend + webview → native `.app`) — purpose-built for this stack; reuse the *entire* existing Go server + embedded SPA nearly untouched, in a real macOS window. Electron/Tauri also work; Wails is the natural fit because the backend is already Go. "Same affordances as the web app" is satisfied by construction — it *is* the web UI, natively windowed.
2. **Dependency handling — now optional.** Call system-installed `typst`/`pandoc`/`python3`. Bundling only mattered for other Macs, which are out of scope. *(The Python→Go port below stays a nice-to-have, not a requirement.)*
3. **Product scoping — a design decision, not code.** ProdCal is two products fused: **(a)** a book-production engine (Vellum's space) and **(b)** a client/project-management SaaS (clients, transmittals, email digests, calendar, the `X-ExeDev-UserID` admin gate). On a single-user Mac, most of (b) is meaningless. The `.app` should run a **"desktop mode"** that hides the SaaS surface and drops the proxy-based auth (a local single user is trusted). This also sidesteps the whole B1/header-trust security class from the launch triage.

## Recommended path
`A (persistent local sandbox)` → thin **Wails** shell that launches the server + points a webview at `localhost` → desktop-mode flag hiding the SaaS/auth surface → (optional, for robustness) port the 3 Python scripts to Go, which also fixes the one test that's red on every laptop.

## Nice-to-have that pays off regardless of the app
**Eliminate the Python dependency** by porting the 3 server-invoked scripts to Go:
- `generate-word-template.py` (`bookspecs.go:849`), `apply-corrections-docx.py` (`corrections_apply.go:94`), `detect-edge-cases.py` (`preflight.go:106`).
- All three are DOCX/OOXML zip manipulation — doable in Go with a docx/zip library or raw XML.
- Payoff: removes the hardest thing to bundle *and* makes `go test ./...` green locally (kills `TestWordTemplateGeneration*` `ModuleNotFoundError`). Worth doing even if the desktop app never happens.

## Prerequisites / open items
- Confirm `typst` + `python-docx` install locally (`brew install typst`; `pip install python-docx`) per `docs/MANUSCRIPT-PIPELINE.md` — needed for full local EPUB/PDF today regardless of the app.
- Decide the desktop-mode feature cut (which SaaS surfaces to hide).
- The launch-triage fixes (esp. the silent-failure pattern) benefit the desktop app too — shared codebase.

## Effort summary
- Wails wrapper reusing SPA + server: **days**
- Call system tools (no bundling): **trivial**
- Python→Go port (optional): **moderate — highest-leverage single refactor**
- Signing/fonts/notarization: **none required for personal use**

---

## Update — working prototype built this session

A runnable prototype now exists in-repo. **Toolkit choice: `webview/webview_go`, not Wails.** Rationale: ProdCal's frontend is already an HTTP-served embedded SPA; Wails wants to own a JS-frontend build we don't have. `webview_go` is a thin CGO wrapper over the system WebKit — a single Go binary opens a native window pointed at the local server. No Node, no JS scaffolding, no `wails` CLI.

### What was added (all additive)
- **`internal/localrun/`** — shared launcher core: builds `srv.New` on a data-dir DB, serves the real handler on an ephemeral `127.0.0.1` port, and fronts it with a loopback-only reverse proxy that injects the `X-ExeDev-UserID: local-admin` header. Refuses any non-loopback address. `Start(Options) (*Instance, error)` + `Shutdown(ctx)`.
- **`cmd/prodcal-local/`** — headless persistent service (model A), now a thin CLI over `internal/localrun` (+ launchd agent, `scripts/run-local.sh`, `docs/LOCAL-USAGE.md` from the harness step).
- **`cmd/prodcal-app/`** (`//go:build darwin`) — the desktop shell: starts the launcher on an ephemeral port, opens a WebKit window (`webview.New` → `Navigate`) at it. ~40 lines; reuses the entire server + SPA unchanged.

### Verified
- `go build ./cmd/prodcal-app` (CGO + WebKit) links; `go build ./...`, `gofmt`, `go vet` all clean.
- Launcher core smoke-tested: admin UI → HTTP 200 through the injected header, `/healthz` ok, DB created in the data dir, loopback guard fatal on `0.0.0.0`.
- **Not verified here:** the actual window render — a GUI can't be exercised headlessly. Run `go build -o prodcal-app ./cmd/prodcal-app && ./prodcal-app` on your Mac to see it.

### Remaining to make it a "real" `.app` (all optional for personal use)
1. **Bundle as `.app`** — wrap the binary in `ProdCal.app/Contents/{MacOS,Info.plist}` so it's double-clickable + gets a Dock icon. Trivial script; no signing needed for personal use.
2. **Desktop-mode scoping** — hide the SaaS surface (clients/transmittals/email/calendar admin) the single-user app doesn't need; expose the book engine (upload → spec → preflight → EPUB/PDF). This is a frontend/routing decision, not plumbing.
3. **Install the pipeline tools** locally (`brew install typst`, `pip install python-docx`) so EPUB/PDF generation runs in-app; or later bundle them.
4. **(Optional) port the 3 Python scripts to Go** to drop the Python dependency entirely — still the highest-leverage cleanup, but no longer required since it's your machine.

The prototype confirms the core feasibility claim: the distance from web app to personal desktop app was a native window + a loopback launcher, both of which now exist and reuse the unmodified server.
