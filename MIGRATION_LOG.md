# Migration Log

This file documents the merge of `djinna/prodcal` and `djinna/book-prod` into
`djinna/jdbbs`. It is the source of truth for what was migrated, what was
changed, and what reconciliation work is still open.

## Sources

| Repo | Last commit imported | Date | Role |
|---|---|---|---|
| [`djinna/prodcal`](https://github.com/djinna/prodcal) | `3a8df8a` | 2026-04-24 | Go web app — production calendar, transmittals, book specs, client portals, AgentMail |
| [`djinna/book-prod`](https://github.com/djinna/book-prod) | `a595088` | 2026-03-18 | Typst-based book production pipeline — templates, fonts, conversion scripts |

Both upstream repos are now archival.

## Strategy

**Clean import** (Section 5 Phase 1, Option B in `docs/JDBBS_KICKOFF_PROMPT.md`).
File contents are copied verbatim and the upstream commit SHAs are recorded
here. History is preserved by the original repos remaining read-only on GitHub.

The initial three commits on `main` mirror the kickoff phases:

1. **Import Go server, DB, and embedded SPA from `djinna/prodcal@3a8df8a`** — verbatim copy, no modifications.
2. **Import typesetting templates, fonts, scripts, manuscripts, reference, corrections from `djinna/book-prod@a595088`** — verbatim copy, no modifications.
3. **Phase 2: monorepo path fixes, Dockerfile/Makefile/CI/docs** — see "Phase 2 changes" below.

## Layout mapping

### From `prodcal@3a8df8a`

| Source | Destination | Notes |
|---|---|---|
| `cmd/srv/` | `cmd/srv/` | unchanged |
| `srv/` | `srv/` | hardcoded paths replaced — see Phase 2 below |
| `db/` | `db/` | unchanged |
| `go.mod` / `go.sum` | same | module name kept as `srv.exe.dev` |
| `.gitignore` / `.dockerignore` | same | merged with monorepo additions |
| `Dockerfile` | `Dockerfile` | rewritten for unified deps (Typst + Pandoc + python-docx + pyyaml + Libertinus) |
| `Makefile` | `Makefile` | binary renamed `prodcal` → `jdbbs` |
| `srv.service` | `srv.service` | `WorkingDirectory` → `/home/exedev/jdbbs`, `ExecStart` → `/home/exedev/jdbbs/jdbbs` |
| `scripts/backup-db.sh` | `scripts/backup-db.sh` | (path inside script still references the old `/home/exedev/prodcal/db.sqlite3` — see open items) |
| `CHECKPOINTS.md`, `DEPLOY.md` | `docs/CHECKPOINTS.md`, `docs/DEPLOY.md` | DEPLOY.md was rewritten for the monorepo |
| `AGENTS.md` | `AGENTS.md` | rewritten for monorepo |
| `README.md` | `README.md` | rewritten as unified README |
| `JDBBS_KICKOFF_PROMPT.md` | `docs/JDBBS_KICKOFF_PROMPT.md` | preserved as historical context |
| `docs/` (notes / plans / checklists / book-production-deepdive.html) | `docs/` | preserved |
| ad-hoc session notes (`NEXT-SESSION-*.md`, `SESSION-SUMMARY.txt`, `TEAM-UPDATE.txt`) | dropped | session-specific scratch; not relevant going forward |

### From `book-prod@a595088`

| Source | Destination | Notes |
|---|---|---|
| `templates/` | `typesetting/templates/` | series-template.typ, styles.typ, images.typ, pandoc-typst.typ, epub/, word/ |
| `fonts/` | `typesetting/fonts/` | Source Sans 3 + JetBrains Mono. Libertinus Serif still expected to be system-installed. |
| `scripts/*.py`, `*.sh` | `typesetting/scripts/` | conversion, corrections, build helpers |
| `scripts/docx-to-typst.lua` | `typesetting/filters/docx-to-typst.lua` | Pandoc filter promoted to its own dir |
| `src/` | `manuscripts/` | sample books, ghosts source |
| `reference/` | `reference/` | reference PDFs/EPUBs + extracted internals |
| `corrections/` | `corrections/` | example corrections ledger |
| `docs/{TYPOGRAPHY,WORKFLOW,typst-paper-sizes,word-styles}.md` | `docs/` | merged into the prodcal docs/ tree |
| `BOWKER_ISBN_INTEGRATION_PLAN.md`, `TYPST_FRONTEND_PLAN.md`, `TYPOGRAPHY_REFINEMENT_PROMPT.md`, `TEST_PLAN_2026-03-09.md`, `SESSION_SUMMARY_*.md`, `NEXT_SESSION_PROMPT_*.md` | dropped | session-specific scratch |
| `print-planning.html`, `font-compare.html` | dropped | standalone HTML tools, can be re-imported on demand |

## Phase 2 changes (path & deployment fixes)

### `srv/books.go`

Replaced the four hardcoded `bookProdRoot = "/home/exedev/book-production"`
constants with a runtime-resolved `typesettingRoot()` helper. Resolution order:

1. `JDBBS_TYPESETTING_DIR` env var
2. `./typesetting` relative to CWD (production layout)
3. Walk up parents looking for a sibling `typesetting/` (so `go test ./srv/...`
   works when CWD is the package dir)
4. Fall back to `./typesetting` (absolute)

Helpers `convertScriptPath()`, `seriesTemplatePath()`, `fontsDirPath()` derive
their paths from `typesettingRoot()`. The `typst compile --root /` flag is
unchanged because the helpers return absolute paths.

### `srv/bookspecs.go`

`generate-word-template.py` and `typst fonts --font-path` paths now resolve
through the same helpers.

### `typesetting/scripts/build-ghosts.sh`

`cd ~/book-production` replaced with a `REPO_ROOT` derivation from the
script's own location. Output paths updated for the monorepo layout
(`src/ghosts/` → `manuscripts/ghosts/`, `templates/series-template.typ` →
`typesetting/templates/series-template.typ`).

The other shell scripts (`md2pdf.sh`, `md2epub.sh`, `docx2pdf.sh`,
`docx2epub.sh`) already used `PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"`, so
they continue to work unchanged when run from inside `typesetting/`.

### Dockerfile

Single-stage build → multi-stage with full typesetting toolchain:
Typst (binary release), Pandoc, python3 + python-docx + pyyaml,
fonts-libertinus, fontconfig, plus the bundled fonts copied into
`/usr/share/fonts/`.

### CI

`.github/workflows/ci.yml` added — Go 1.26, `go vet`, `go test`. Python
deps (`python-docx`, `pyyaml`) are installed so the word-template
integration test passes; Pandoc is installed for the eventual
PDF/EPUB pipeline tests.

### `srv.service`

Now points at `/home/exedev/jdbbs/jdbbs`. The old `prodcal` working
directory and binary path are gone.

## Open Phase 3 reconciliation items

These need a deliberate decision before being resolved.

### 1. Page-width drift (`spec.page.width_in` vs Typst `page-width`)

`srv/bookspecs.go::defaultSpecData()` declares `page.width_in: 5.5` (= 396pt).
`typesetting/templates/series-template.typ::default-config` declares
`page-width: 353.811pt` (≈ 4.91in, the trade-paperback trim).

These do not match. The Go → Typst config generator in `buildTypstConfig()`
will overwrite the Typst default with whatever the spec says, so in practice
the spec value wins — but the spec default is wrong. Decide which is canonical
and align both. Likely the 4.91" trim is correct (it's the Protocolized
Anthology design).

### 2. EPUB generation strategy

Two paths exist:

- `srv/epub.go` — Go handler invoked via `POST /api/books/{id}/generate-epub`.
  Uses spec-driven CSS, cover, metadata.
- `typesetting/scripts/{docx2epub,md2epub}.sh` — Standalone Pandoc wrappers
  with embedded WOFF2 fonts and a fixed CSS.

Both produce EPUB but with different CSS / font handling. Decision needed:

- Keep both (Go = web app, scripts = CLI/batch)? Then make sure they produce
  consistent output.
- Or have the Go handler shell out to the script, so there's a single source
  of truth?

### 3. Corrections wiring

`srv/corrections.go` stores corrections in SQLite and exports YAML.
`typesetting/scripts/apply-corrections.py` consumes that YAML and patches
EPUBs. The two halves are not yet wired: the YAML is a manual export step.

To close: after EPUB generation in `srv/epub.go`, materialize the YAML
in-memory and invoke `apply-corrections.py` to apply pending corrections.
Same for `apply-corrections-docx.py` for Word.

### 4. Libertinus Serif bundling

book-prod assumes Libertinus is system-installed. The Dockerfile installs
`fonts-libertinus` from apt to handle the container case. The exe.dev VM
should be checked: `fc-list | grep -i libertinus`. If absent, install
`fonts-libertinus` (or bundle the fonts in `typesetting/fonts/libertinus/`
for full self-containment).

### 5. `scripts/backup-db.sh` path

The daily cron script still references `/home/exedev/prodcal/db.sqlite3`.
Update to `/home/exedev/jdbbs/db.sqlite3` as part of the VM cutover, and the
in-repo copy here too.

### 6. Server-side drift in `book-production` on the VM

Per the kickoff: check whether the live `/home/exedev/book-production/`
contains uncommitted changes after `book-prod@a595088`. If so, port them in
before declaring this migration complete.

### 7. Phase 4 archive notices

Once the migration is verified live on the VM:

- Add an "ARCHIVED — merged into djinna/jdbbs" banner to
  `djinna/prodcal/README.md` and `djinna/book-prod/README.md`.
- Set both repos to **Archived** in GitHub Settings → Danger Zone.

(Deferred until cutover confirms jdbbs is fully working.)

## Sanity checks performed during scaffold

- `go build -o jdbbs ./cmd/srv` succeeds.
- `go vet ./...` clean.
- `go test ./...` passes (with `python-docx` + `pyyaml` installed locally —
  same condition as in prodcal previously).
- All `srv/static/*` assets load via `//go:embed static/*`.
- All `db/migrations/001`–`011` apply cleanly on a fresh DB.
