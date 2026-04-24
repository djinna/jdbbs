# JDBBS Kickoff Prompt

> **Purpose**: This document is the founding prompt for the `djinna/jdbbs` repository. Hand it to an AI coding assistant or developer to scaffold, populate, and ship the unified monorepo. It is self-contained — no prior conversation context is required.

---

## 1. Project Identity & Context

| Key | Value |
|---|---|
| **New repo** | `djinna/jdbbs` |
| **Production URL** | `https://jdbbs.exe.xyz` |
| **Hosting** | exe.dev VM at `/home/exedev/` |
| **Archival repo 1** | `djinna/prodcal` — Go web app (production calendar + book management) |
| **Archival repo 2** | `djinna/book-prod` — Typst-based book production pipeline |

This repo **unifies** the two archival repos into a single monorepo. Both archival repos are frozen after this migration — no further commits to either. `djinna/jdbbs` becomes the single source of truth going forward.

### Connecting the GitHub Repo to the exe.dev VM

The app is served from the exe.dev VM, not from GitHub Pages or a container registry. To connect the `djinna/jdbbs` repo to the VM:

1. **Clone the repo on the VM**:
   ```bash
   cd /home/exedev
   git clone https://github.com/djinna/jdbbs.git
   ```

2. **Build and install**:
   ```bash
   cd /home/exedev/jdbbs
   make build
   # Binary goes to /home/exedev/jdbbs/jdbbs (or prodcal, adjust Makefile accordingly)
   ```

3. **Update the systemd unit** (`/etc/systemd/system/srv.service` or equivalent):
   - Change `ExecStart` to point to the new binary path: `/home/exedev/jdbbs/jdbbs`
   - Change `WorkingDirectory` to `/home/exedev/jdbbs`
   - Ensure environment variables (`AGENTMAIL_API_KEY`, `AGENTMAIL_INBOX_ID`, etc.) carry over
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl restart srv
   ```

4. **DNS**: `jdbbs.exe.xyz` should already resolve via the exe.dev proxy. The server listens on port 8000; the exe.dev HTTPS proxy terminates TLS and forwards to it. Verify: `curl http://localhost:8000/healthz`.

5. **Deploy workflow** (ongoing):
   ```bash
   cd /home/exedev/jdbbs
   git pull origin main
   make build
   sudo systemctl restart srv
   ```
   Optionally set up a webhook or cron-based pull for automated deploys.

6. **Database migration**: The existing SQLite database at `/home/exedev/prodcal/db.sqlite3` should be copied (or symlinked) to the new working directory. The server auto-runs migrations on startup.
   ```bash
   cp /home/exedev/prodcal/db.sqlite3 /home/exedev/jdbbs/db.sqlite3
   ```

7. **Book-production assets**: Previously at `/home/exedev/book-production/` — now absorbed into the repo at `typesetting/`. Remove (or keep as backup) the old standalone directory after verifying the monorepo works.

---

## 2. What Each Archival Repo Contains

### 2a. `djinna/prodcal` (last commit: `dbb3bf6`, 2026-04-12)

**Go HTTP server** (`srv/`) with handlers for:

| Handler file | Responsibility |
|---|---|
| `server.go` | Core routing, auth (exe.dev proxy `X-ExeDev-UserID` header), embedded static files |
| `admin.go` | Admin dashboard, project CRUD, client management |
| `books.go` | Book upload, DOCX → Typst → PDF conversion pipeline |
| `bookspecs.go` | Typesetting spec CRUD, transmittal-to-spec import, config generation, Word template generation, cover upload |
| `corrections.go` | Corrections ledger CRUD + YAML export |
| `epub.go` | EPUB generation via Pandoc (docx → epub3) with spec-driven CSS |
| `transmittal.go` | Manuscript transmittal form (metadata capture) |
| `client.go` | Client portal (password-protected per-project views) |
| `email.go`, `*_email.go` | AgentMail integration (6 email pathways, see `srv/EMAIL_SYSTEM.md`) |
| `journal.go` | Activity journal (Calls, Decisions, Approvals) |
| `filelog.go` | File log (inbound/outbound manuscript asset tracking) |

**Database**:
- SQLite with auto-migrations on startup
- Migrations in `db/migrations/` (001 through 011, including `007-books.sql`, `008-book-specs.sql`, `010-corrections.sql`, `011-project-archive.sql`)
- sqlc-generated Go code in `db/dbgen/`, query definitions in `db/queries/`
- Schema covers: projects, tasks, transmittals (with versions), books, book_specs, corrections, clients, visitors, file_log, journal

**Frontend** (embedded via `//go:embed static/*`):
- `srv/static/admin.html` — Admin SPA with tabbed UI (Projects | Books)
- `srv/static/app.js` — Gantt chart / timeline scheduler
- `srv/static/transmittal.html` / `transmittal.js` — Transmittal form
- `srv/static/client.html` — Client portal
- `srv/static/landing.html` — Landing page
- `srv/static/style.css`, `transmittal.css` — Stylesheets

**Auth**: exe.dev proxy injects `X-ExeDev-UserID` header. Admin endpoints check for this. Client portal uses per-project passwords (SHA-256 hashed).

**Email**: AgentMail API (`AGENTMAIL_API_KEY` + `AGENTMAIL_INBOX_ID` env vars). See `srv/EMAIL_SYSTEM.md` for the 6 pathways.

**Infrastructure**:
- `Dockerfile` (two-stage: Go build → Debian slim runtime; currently not used in prod)
- `Makefile` (assumed — `make build` referenced in AGENTS.md)
- systemd unit `srv` on port 8000, proxied via exe.dev HTTPS
- Daily SQLite backups at 3 AM (`scripts/backup-db.sh`)
- Checkpoint tags for rollback (`CHECKPOINTS.md`)

**Critical hardcoded dependency on book-prod** (`srv/books.go`, lines 21–26):
```go
const (
    bookProdRoot    = "/home/exedev/book-production"
    convertScript   = bookProdRoot + "/scripts/md-to-chapter.py"
    seriesTemplate  = bookProdRoot + "/templates/series-template.typ"
    fontsDir        = bookProdRoot + "/fonts"
)
```
These absolute paths must be replaced with paths relative to the monorepo root.

### 2b. `djinna/book-prod` (last commit: `a595088`, 2026-03-18)

**Typst templates**:
- `templates/series-template.typ` — Master layout engine (535 lines). Defines `default-config` dict with page geometry, fonts, margins, headings, running heads, section breaks. Provides `merge-config()` for overrides, plus functions for front matter, paragraphs, section breaks, epigraphs, etc.
- `templates/styles.typ` — Shared style definitions (small-caps, monospace, etc.)
- `templates/images.typ` — Image handling utilities
- `templates/pandoc-typst.typ` — Pandoc-specific Typst template
- `templates/epub/epub-styles.css` — EPUB stylesheet
- `templates/word/` — Word templates: `author-template.docx`, `default-reference.docx`, `protocolized-style-guide.docx`

**Conversion scripts**:
- `scripts/md-to-chapter.py` — Markdown → Typst chapter conversion (called by prodcal's `books.go`)
- `scripts/apply-corrections.py` — EPUB patcher (YAML-driven find/replace in XHTML content, 329 lines)
- `scripts/apply-corrections-docx.py` — Word document patcher
- `scripts/generate-word-template.py` — Book spec JSON → styled `.docx` template (495 lines)
- `scripts/docx2epub.sh` — DOCX → EPUB3 via Pandoc (with font embedding)
- `scripts/md2epub.sh` — Markdown → EPUB3 via Pandoc
- `scripts/docx2pdf.sh` — DOCX → PDF
- `scripts/md2pdf.sh` — Markdown → PDF
- `scripts/build.sh`, `scripts/build-ghosts.sh` — Build scripts

**Pandoc Lua filter**:
- `scripts/docx-to-typst.lua` — Custom DOCX → Typst conversion filter

**Fonts** (`fonts/`):
- `fonts/sourcesans/` — Source Sans 3 (including WOFF2 for EPUB embedding)
- `fonts/jetbrainsmono/` — JetBrains Mono (including WOFF2 webfonts)
- Libertinus Serif is expected to be system-installed (not bundled as files)
- `fonts/README.md` — Font documentation

**Source material** (`src/`):
- `src/ghosts/` — Reference book source (`main.typ` and related files)
- `src/sample-book.typ`, `src/sample-story.md`, `src/sample-anthology.md`, `src/sample-chapter.docx` — Sample/test content

**Reference materials** (`reference/`):
- Reference PDFs and EPUBs (GHOSTS, TT, LIBRARIANS)
- Extracted EPUB internals for analysis
- `SERIES_DESIGN_SPEC.md` — Series visual identity specification
- PDF page screenshots for comparison

**Documentation** (`docs/`):
- `TYPOGRAPHY.md` — Typography specification
- `WORKFLOW.md` — Production workflow documentation
- `typst-paper-sizes.md` — Paper size reference
- `word-styles.md` — Word style mapping documentation

**Corrections** (`corrections/`):
- `example-ghosts.yaml` — Example corrections ledger

**Other**:
- `print-planning.html`, `font-compare.html` — Standalone HTML tools
- Various `NEXT_SESSION_PROMPT_*.md` files (session planning docs)

---

## 3. Proposed Monorepo Structure

```
jdbbs/
├── cmd/srv/
│   └── main.go                   # Go entrypoint (from prodcal)
├── srv/                          # Go handlers (from prodcal)
│   ├── books.go                  # UPDATE: replace hardcoded paths → use typesetting/ relative paths
│   ├── bookspecs.go              # Spec CRUD, config generation, Word template generation
│   ├── corrections.go            # Corrections ledger + YAML export
│   ├── epub.go                   # EPUB generation via Pandoc
│   ├── server.go                 # Core routing, auth, static file embedding
│   ├── admin.go                  # Admin dashboard
│   ├── transmittal.go            # Transmittal API
│   ├── client.go                 # Client portal
│   ├── email.go                  # AgentMail integration
│   ├── *_email.go                # Email pathway handlers
│   ├── journal.go                # Activity journal
│   ├── filelog.go                # File log
│   ├── EMAIL_SYSTEM.md           # Email architecture doc
│   ├── static/                   # Frontend assets (HTML, JS, CSS)
│   │   ├── admin.html
│   │   ├── app.js
│   │   ├── transmittal.html
│   │   ├── transmittal.js
│   │   ├── client.html
│   │   ├── landing.html
│   │   ├── style.css
│   │   └── transmittal.css
│   └── *_test.go                 # All existing tests
├── db/
│   ├── db.go                     # SQLite open, migration runner
│   ├── migrations/               # All SQL migrations 001–011 (from prodcal)
│   ├── queries/                  # sqlc query files (from prodcal)
│   ├── dbgen/                    # sqlc-generated Go code (from prodcal)
│   └── sqlc.yaml                 # sqlc config
├── typesetting/                  # ← absorbed from book-prod
│   ├── templates/
│   │   ├── series-template.typ   # Master Typst layout engine
│   │   ├── styles.typ            # Shared style definitions
│   │   ├── images.typ            # Image utilities
│   │   ├── pandoc-typst.typ      # Pandoc Typst template
│   │   ├── epub/
│   │   │   └── epub-styles.css   # EPUB stylesheet
│   │   └── word/
│   │       ├── author-template.docx
│   │       ├── default-reference.docx
│   │       └── protocolized-style-guide.docx
│   ├── scripts/
│   │   ├── md-to-chapter.py      # Markdown → Typst chapter
│   │   ├── apply-corrections.py  # EPUB patcher
│   │   ├── apply-corrections-docx.py  # Word doc patcher
│   │   ├── generate-word-template.py  # Spec → styled .docx
│   │   ├── docx2epub.sh          # DOCX → EPUB3
│   │   ├── md2epub.sh            # Markdown → EPUB3
│   │   ├── docx2pdf.sh           # DOCX → PDF
│   │   ├── md2pdf.sh             # Markdown → PDF
│   │   ├── build.sh              # Generic build script
│   │   └── build-ghosts.sh       # Book-specific build
│   ├── filters/
│   │   └── docx-to-typst.lua     # Pandoc Lua filter
│   └── fonts/
│       ├── README.md
│       ├── sourcesans/           # Source Sans 3 (OTF + WOFF2)
│       └── jetbrainsmono/        # JetBrains Mono (OTF + WOFF2)
├── manuscripts/                  # ← from book-prod src/
│   ├── ghosts/                   # Reference book (main.typ, etc.)
│   ├── sample-book.typ
│   ├── sample-story.md
│   ├── sample-anthology.md
│   └── sample-chapter.docx
├── reference/                    # ← from book-prod reference/
│   ├── SERIES_DESIGN_SPEC.md
│   ├── *.pdf, *.epub             # Reference output files
│   └── pdf_pages/                # Page comparison screenshots
├── corrections/                  # ← from book-prod corrections/
│   └── example-ghosts.yaml       # Example corrections ledger
├── docs/
│   ├── TYPOGRAPHY.md             # ← from book-prod docs/
│   ├── WORKFLOW.md               # ← from book-prod docs/
│   ├── typst-paper-sizes.md      # ← from book-prod docs/
│   ├── word-styles.md            # ← from book-prod docs/
│   ├── DEPLOY.md                 # ← from prodcal (updated for monorepo)
│   ├── EMAIL_SYSTEM.md           # ← from prodcal srv/ (or keep in srv/)
│   └── CHECKPOINTS.md            # ← from prodcal
├── scripts/
│   └── backup-db.sh              # ← from prodcal scripts/
├── Dockerfile                    # Updated for unified deps (Typst, Pandoc, Python)
├── Makefile                      # Build targets
├── go.mod / go.sum               # ← from prodcal
├── seed_data.json                # ← from prodcal
├── .gitignore
├── AGENTS.md                     # Updated agent instructions for monorepo
└── README.md                     # New unified README
```

---

## 4. Critical Reconciliation Tasks

These must be done carefully during migration. Check off each one and document decisions in a migration log.

- [ ] **Spec/config parity**: Compare `defaultSpecData()` in `srv/bookspecs.go` (prodcal) with `default-config` dict in `templates/series-template.typ` (book-prod). Key mappings to verify:
  - `page.width_in` / `page.height_in` (Go, inches) ↔ `page-width` / `page-height` (Typst, points): `5.5in = 396pt ≠ 353.811pt` — **drift detected**, reconcile which is canonical
  - `typography.base_size_pt: 10` ↔ `base-size: 10pt` — matches
  - `typography.leading_pt: 2` ↔ `leading: 2pt` — matches
  - `headings.h1_size_pt: 16.67` ↔ `h1-size: 1.667em` — matches (1.667 × 10pt = 16.67pt)
  - `elements.poem_size_pt: 7.5` ↔ `poem-size: 0.75em` — matches (0.75 × 10pt = 7.5pt)
  - Note: Go spec uses absolute `pt` values; Typst uses relative `em`. The Go→Typst config generator in `buildTypstConfig()` must translate correctly
  - Front matter / back matter flags exist in Go spec but not in Typst `default-config` — verify how these are used

- [ ] **Wire corrections end-to-end**: The corrections pipeline currently has a gap:
  - prodcal's `corrections.go` stores corrections in SQLite and exports to YAML via `handleExportCorrections`
  - book-prod's `scripts/apply-corrections.py` reads YAML and patches EPUBs
  - These are currently manual steps. Wire them together: after EPUB generation in `epub.go`, auto-apply pending corrections from the DB by generating the YAML in-memory and running `apply-corrections.py`
  - Also integrate `apply-corrections-docx.py` for Word document patching

- [ ] **Replace hardcoded paths** in `srv/books.go` (lines 21–26):
  ```go
  // BEFORE (hardcoded):
  bookProdRoot    = "/home/exedev/book-production"

  // AFTER (relative to binary or working directory):
  bookProdRoot    = "typesetting"  // or use os.Executable() to find repo root
  convertScript   = filepath.Join(bookProdRoot, "scripts", "md-to-chapter.py")
  seriesTemplate  = filepath.Join(bookProdRoot, "templates", "series-template.typ")
  fontsDir        = filepath.Join(bookProdRoot, "fonts")
  ```
  Also update symlink creation for `styles.typ` and `images.typ` (lines 228–233) to use relative paths.
  Also update `typst compile --root` flag (line 238) — currently uses `"/"` for absolute paths; adjust for relative template paths.

- [ ] **Check for server-side drift**: Verify whether book-prod scripts were modified on the VM (`/home/exedev/book-production/`) after the last commit (2026-03-18) without being pushed. Run on the VM:
  ```bash
  cd /home/exedev/book-production
  git status
  git diff
  ```
  If there are uncommitted changes, incorporate them into `jdbbs` rather than losing them.

- [ ] **Reconcile EPUB generation**: Two EPUB paths exist:
  1. `srv/epub.go` (Go) — Pandoc-based, uses book spec settings for CSS generation, cover images, metadata. Invoked via API (`POST /api/books/{id}/generate-epub`)
  2. `scripts/docx2epub.sh` / `scripts/md2epub.sh` (shell) — Standalone Pandoc wrappers with font embedding, fixed CSS
  - Decision needed: Keep both? The Go handler is the primary path for the web app. The shell scripts are useful for CLI/batch usage. If keeping both, ensure they produce consistent output (same CSS, same font embedding logic).

- [ ] **Reconcile Word template generation**: Two implementations exist:
  1. `srv/bookspecs.go` — `handleGenerateWordTemplate()` (Go, invoked via API `POST /api/projects/{id}/book-spec/word-template`). Uses the `python3 generate-word-template.py` script internally
  2. `scripts/generate-word-template.py` (Python, 495 lines) — Standalone script, reads spec JSON
  - These are actually the same pipeline (Go calls Python). Verify the Go handler passes the correct path to the Python script after the path changes.

- [ ] **Update Dockerfile** for unified dependencies:
  ```dockerfile
  # Current (prodcal): Go binary only, Debian slim
  # Needed: Add Typst, Pandoc, Python3 + python-docx + pyyaml, fonts
  RUN apt-get install -y pandoc python3 python3-pip
  RUN pip3 install python-docx pyyaml
  # Install Typst (binary release)
  # Copy fonts into image
  COPY typesetting/fonts/ /usr/share/fonts/custom/
  ```

- [ ] **Update deployment docs** (`DEPLOY.md`) for the new unified structure — paths, binary name, working directory, systemd unit changes

- [ ] **Verify fonts**: Ensure all three font families are properly included and referenced:
  - Source Sans 3: bundled in `typesetting/fonts/sourcesans/`
  - JetBrains Mono: bundled in `typesetting/fonts/jetbrainsmono/`
  - Libertinus Serif: **NOT bundled** in book-prod — assumed system-installed. Either bundle it or document the system dependency. Check if it's installed on the exe.dev VM.

- [ ] **Set up CI (GitHub Actions)**: Neither archival repo has active CI. Add the workflow below to `.github/workflows/ci.yml` in the new repo (see Section 5, Phase 1, step 7 for the ready-to-use file).

---

## 5. Migration Strategy

### Phase 1: Scaffold

1. Create the `djinna/jdbbs` repo on GitHub (empty, with README)
2. Set up the directory structure from Section 3
3. Copy Go code from `prodcal` as-is:
   - Option A (preserving history): `git subtree add --prefix=. https://github.com/djinna/prodcal.git main`
   - Option B (clean start): Copy files with commit messages noting origin, e.g. `"Import Go server from djinna/prodcal@dbb3bf6"`
4. Copy typesetting assets from `book-prod` into `typesetting/`:
   - `book-prod/templates/` → `typesetting/templates/`
   - `book-prod/scripts/` → `typesetting/scripts/`
   - `book-prod/fonts/` → `typesetting/fonts/`
   - `book-prod/scripts/docx-to-typst.lua` → `typesetting/filters/docx-to-typst.lua`
5. Copy remaining book-prod content:
   - `book-prod/src/` → `manuscripts/`
   - `book-prod/reference/` → `reference/`
   - `book-prod/corrections/` → `corrections/`
   - `book-prod/docs/` → `docs/`
6. Verify: `go build ./cmd/srv` succeeds and `go test ./srv/...` passes
7. Add CI — create `.github/workflows/ci.yml` with this workflow:

   ```yaml
   name: CI
   on:
     push:
       branches: [main]
     pull_request:
       branches: [main]

   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4

         - uses: actions/setup-go@v5
           with:
             go-version: "1.26"

         - name: Build
           run: go build ./cmd/srv

         - name: Vet
           run: go vet ./...

         - name: Test
           run: go test ./... -v -count=1
   ```

   This gives you automated build + lint + test on every push and PR. You'll see a green check or red X on each PR before merging. Expand later with Typst compile checks or Python linting as needed.

### Phase 2: Integration

1. Fix hardcoded paths in `srv/books.go` (see Section 4)
2. Update any shell scripts that use `PROJECT_ROOT` to work within the monorepo structure
3. Run `go test ./...` — all existing tests must pass
4. Manually test the conversion pipeline: upload a DOCX → convert → download PDF
5. Wire corrections pipeline end-to-end (DB → YAML → EPUB patcher)

### Phase 3: Reconciliation

1. Audit and reconcile the spec/config data models (Go `defaultSpecData()` vs Typst `default-config`)
2. Decide on EPUB generation strategy (Go handler vs shell scripts)
3. Deduplicate any overlapping logic
4. Update Dockerfile to include all dependencies
5. Update systemd unit and deployment workflow
6. Update DNS if needed (the current DEPLOY.md already shows `jdbbs.exe.xyz` as the URL)

### Phase 4: Archive

1. Add to `djinna/prodcal` README.md:
   ```markdown
   > **ARCHIVED** — This repository has been merged into [djinna/jdbbs](https://github.com/djinna/jdbbs).
   > No further development will happen here. This repo is preserved as a read-only archive.
   ```
2. Add the same notice to `djinna/book-prod` README.md
3. Set both repos to archived on GitHub (Settings → Danger Zone → Archive this repository)

---

## 6. Safety Rules

1. **Do NOT delete or modify anything in `djinna/prodcal` or `djinna/book-prod`** — they are archival references. The only exception is adding the archive notice to their READMEs in Phase 4.
2. **Every file copied into `jdbbs`** should have a commit message noting its origin repo and commit hash (e.g., `"Import typesetting templates from djinna/book-prod@a595088"`).
3. **Run all existing tests after each phase** before proceeding (`go test ./...`).
4. **Keep a migration log** (`MIGRATION_LOG.md` in the new repo) documenting every decision made during reconciliation — especially around the spec/config drift and EPUB generation strategy.
5. **Back up the production database** before any deployment changes:
   ```bash
   sqlite3 /home/exedev/prodcal/db.sqlite3 ".backup /home/exedev/backups/pre-migration-$(date +%Y%m%d).sqlite3"
   ```
6. **Test on the VM** with the production database copy before switching the systemd unit to the new binary.
7. **Do not modify tests** to make them pass — if a test fails after migration, the code has a bug that must be fixed.

---

## 7. Environment & Dependencies

### Runtime dependencies on the exe.dev VM:

| Dependency | Purpose | Install |
|---|---|---|
| Go 1.26+ | Build the server binary | Already installed |
| SQLite 3 | Database | Already installed |
| Typst | PDF typesetting engine | `curl -fsSL https://typst.community/typst-install/install.sh \| sh` or check if already at `/usr/local/bin/typst` |
| Pandoc | DOCX → Markdown, DOCX → EPUB conversions | `apt install pandoc` or check existing |
| Python 3 | Conversion scripts, Word template generation | Already installed |
| `python-docx` | Word template generation | `pip3 install python-docx` |
| `pyyaml` | Corrections YAML parsing | `pip3 install pyyaml` |
| Libertinus Serif font | Body text | Check: `fc-list \| grep -i libertinus` |
| Source Sans 3 font | Headings | Bundled in `typesetting/fonts/sourcesans/` |
| JetBrains Mono font | Code blocks | Bundled in `typesetting/fonts/jetbrainsmono/` |

### Environment variables:

| Variable | Purpose |
|---|---|
| `AGENTMAIL_API_KEY` | AgentMail API authentication |
| `AGENTMAIL_INBOX_ID` | AgentMail inbox for sending emails |
| `PRODCAL_BASE_URL` | Override base URL (default: `https://{hostname}.exe.xyz`) |

---

## 8. Key API Routes (for reference)

Carried over from prodcal — these should all work unchanged after migration:

```
GET  /healthz                              # Health check
GET  /admin/                               # Admin dashboard
GET  /api/projects                         # List projects
POST /api/projects                         # Create project
GET  /api/projects/{id}                    # Get project
PUT  /api/projects/{id}                    # Update project
POST /api/projects/{id}/archive            # Archive project
POST /api/projects/{id}/restore            # Restore archived project
GET  /api/projects/{id}/tasks              # List tasks
POST /api/projects/{id}/tasks              # Create task
GET  /api/projects/{id}/transmittal        # Get transmittal
PUT  /api/projects/{id}/transmittal        # Update transmittal
GET  /api/books                            # List books
POST /api/books/upload                     # Upload DOCX
POST /api/books/{id}/convert               # DOCX → PDF pipeline
POST /api/books/{id}/generate-epub         # DOCX → EPUB pipeline
GET  /api/books/{id}/download/{format}     # Download PDF or EPUB
GET  /api/projects/{id}/book-spec          # Get book spec
PUT  /api/projects/{id}/book-spec          # Update book spec
POST /api/projects/{id}/book-spec/pull-transmittal   # Import transmittal → spec
POST /api/projects/{id}/book-spec/generate-config    # Generate Typst config
POST /api/projects/{id}/book-spec/word-template      # Generate Word template
POST /api/projects/{id}/book-spec/cover    # Upload cover image
GET  /api/projects/{id}/corrections        # List corrections
POST /api/projects/{id}/corrections        # Create correction
POST /api/projects/{id}/corrections/export # Export corrections YAML
GET  /api/projects/{id}/file-log           # File log
GET  /api/projects/{id}/journal            # Activity journal
POST /api/projects/{id}/transmittal/email  # Send transmittal email
POST /api/projects/{id}/snapshot/email     # Send project snapshot
POST /api/projects/{id}/activity/email     # Send activity email
GET  /api/clients/{client}                 # Client info
GET  /api/clients/{client}/projects        # Client's projects
```

---

## 9. Quick Smoke Test (post-migration)

After Phase 2, run these to verify nothing is broken:

```bash
# 1. Build
cd /home/exedev/jdbbs
go build ./cmd/srv

# 2. Tests
go test ./...

# 3. Start server
./jdbbs -listen :8000 &    # or: go run ./cmd/srv -listen :8000

# 4. Health check
curl http://localhost:8000/healthz
# → {"status":"ok"}

# 5. Admin dashboard loads
curl -s http://localhost:8000/admin/ | head -5
# → HTML with <title>ProdCal Admin</title>

# 6. Book conversion pipeline (requires Typst + Pandoc + Python)
# Upload a test DOCX via the admin UI or API, then trigger convert
# Verify PDF download works

# 7. EPUB generation
# Trigger EPUB generation via API, verify download works
```
