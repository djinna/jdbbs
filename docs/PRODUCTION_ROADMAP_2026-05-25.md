# Production Roadmap — 2026-05-25

Synthesis of three earlier planning docs (`TEST_PLAN_2026-03-09`, `TYPOGRAPHY_REFINEMENT_PROMPT`, `TYPST_FRONTEND_PLAN` — all from the archived book-prod tree) plus the new `docs/translation layer 2026-05-25.md`, cross-referenced against what's actually shipped in jdbbs as of 2026-05-25.

This doc is the **single source of truth for the path to "v1 workflow complete"** — replaces the three older plans, which are kept in `book-prod-archived-2026-05-25/` for history.

## Status snapshot

| Track | Ships in admin SPA | Backend wiring | End-to-end | Critical gap |
|---|---|---|---|---|
| Typesetting tab (spec editor) | ✅ Done | ✅ Done | — | none |
| Spec → Typst config generator | ✅ Done | ✅ Done | — | not consumed by compile pipeline |
| Spec → compiled PDF | UI button exists | ❌ Stub | ❌ | **wire spec into `runConversion()`** |
| Spec → compiled EPUB | UI button exists | ❌ Stub | ❌ | **wire spec into `docx2epub.sh`/`md2epub.sh`** |
| Typography fidelity (EPUB) | ~80% | ✅ | — | body-text alignment may need toggle |
| Typography fidelity (Typst) | ~80% | ✅ Done | — | Ghosts parity matrix never built |
| Corrections workflow | UI exists | ❌ Endpoints unexposed | ❌ | expose corrections API |
| Word template generation | ✅ Done | ✅ Done | — | none |
| Translation layer | — | — | — | **new v2 work — separate roadmap** |

## v1 critical path (in priority order)

The minimum to call the workflow "complete": a user picks a project, edits the spec, hits **Compile PDF** or **Compile EPUB**, gets a downloadable artifact that reflects their spec edits. Plus corrections round-trip. Plus a Ghosts parity sanity check.

### CP-1. Spec → compile pipeline (Typst) — **the keystone**
**TRK-DEV-002** (new). Wire `books.go::runConversion()` to load the project's `book_spec`, render a `config.typ` from the spec, and inject it into the generated `main.typ` before invoking `typst compile`. Add `POST /api/projects/{id}/book-spec/compile` endpoint that orchestrates: load spec → load source manuscript → run conversion → return artifact + log.

**Effort:** ~1 dedicated session (3-4 hours). **Blockers:** none — all pieces exist independently.

### CP-2. Spec → compile pipeline (EPUB)
**TRK-DEV-003** (new). Same shape as CP-1 but for EPUB. `docx2epub.sh` and `md2epub.sh` need to consume the spec (or a rendered CSS override from the spec). Resolves **MIG-005** ("EPUB strategy: Go handler vs shell script") — recommendation: keep shell scripts, have Go pass `--metadata` and a rendered CSS overlay.

**Effort:** ~1 dedicated session (2-3 hours). **Blockers:** none.

### CP-3. Corrections API + round-trip
**TRK-MIG-006 already open** for this — keep the ticket, fold this in. Migration `010-corrections.sql` exists; `corrections/example-ghosts.yaml` exists; `scripts/apply-corrections.py` + `apply-corrections-docx.py` exist; admin UI has a corrections table. Missing: REST endpoints in `server.go` to CRUD corrections, plus an "apply now" trigger that runs the appropriate patcher.

**Effort:** ~1 dedicated session (3 hours). **Blockers:** none.

### CP-4. Libertinus on VM
**TRK-MIG-007 already open.** Verify Libertinus Serif is available to Typst on the VM (or bundle it into `typesetting/fonts/`). Currently `typesetting/fonts/` has Source Sans 3 and JetBrains Mono only.

**Effort:** ~20 min on VM + commit. **Blockers:** none. Can be folded into CP-1's session.

### CP-5. Ghosts parity matrix
**TRK-DESIGN-001 already open.** With CP-1 done, compile `manuscripts/ghosts/` end-to-end and diff against `reference/GHOSTS.pdf`. Build a parity matrix (trim, margins, font, leading, baseline grid, drop caps, chapter numbering, running heads, page numbers, widow/orphan, end matter). For each row: ✅ matches / ❌ visible difference / ⚠️ close-enough-to-defer. File child tickets (TRK-DESIGN-002..N) for ❌.

**Effort:** ~1 dedicated session (2-3 hours, partly visual inspection). **Blockers:** CP-1 (compile pipeline) + CP-4 (Libertinus).

### CP-6. EPUB typography refinement (drift audit)
**TRK-DESIGN-003** (new). Audit `epub-styles.css` against `TYPOGRAPHY_REFINEMENT_PROMPT` (archived in book-prod). Confirmed gap: body text is `justify` in current CSS, prompt specifies left-aligned. Possible others. Decide per-style: change CSS, expose in spec as a toggle, or leave as-is (and update the spec).

**Effort:** ~1 hour. **Blockers:** none. Can be folded into CP-2.

## v1 stretch (worth doing in v1.x)

### CP-7. Test suite for the compile pipeline
**TRK-TEST-001 already open** (end-to-end fixture pipeline DOCX/MD → PDF + EPUB). After CP-1/CP-2 exist, fixture a single "hello world" manuscript with all the common Word styles. Run on every commit (CI) so regressions surface.

**TRK-TEST-002** (visual regression for Ghosts golden) — after CP-5, set up an image-diff check against `reference/GHOSTS.pdf` pages.

### CP-8. Multi-chapter / anthology support
Currently single-chapter focused. For Ghosts (an anthology), need chapter-level metadata (per-chapter title/author/section breaks). Probably emerges naturally from CP-5 (Ghosts is the forcing function). File as **TRK-DEV-004** when concrete needs surface.

### CP-9. Backup alarm channel (open question from TRACKER)
TRK-OPS-007 closed with `check-backups.sh` writing `.HEALTH-FAIL` but no actual notification. Discord webhook or ntfy.sh. ~30 min. Worth doing once the workflow has live traffic.

## v2: Translation layer

This is its own roadmap, well outside the v1 critical path. Cleanly separable: the v1 workflow ships per source-language; translation is a downstream pipeline that consumes `ms:final` artifacts.

**Recommended approach:** Variant E (orchestrator) + Variant F (cross-lingual consistency). Skip Variant G (MT as primary, human rubber stamp) — quality + rights-and-reputation cost too high.

**v2 phases (each ~1 session unless noted):**

### v2.A. Per-title manifest schema
**TRK-TRANS-001** (new). Design and document the per-title `manifest.yaml` translation section: `translations.{lang}.{status, translator, reviewer, glossary, style_guide, mt_engine, llm_pass, target_pub_date, locked_at}`. Treat `zh-Hans`/`zh-Hant` as separate. Decide es-ES vs es-419 policy and document. Add to DB schema (`translations` table linked to `book_id`). Probably 1 dedicated session.

### v2.B. Bilingual side-by-side output format
**TRK-TRANS-002** (new). The doc identifies this as the highest-leverage missing piece for translator adoption ("translators hate switching windows"). Markdown with table layouts, or XLIFF if translators want CAT-tool integration. Prototype + ship to one real translator for feedback. ~1 session for prototype.

### v2.C. MT + LLM draft pipeline (one language at a time)
**TRK-TRANS-003** (new). Pick `fr` first (deepL + LLM well-trodden, native reviewers easier to find than zh). End-to-end: chunker (chapter-aware, not token-aware), glossary loader, deepL call, claude pass with glossary + style guide + prior-chunk context, bilingual output writer. ~2 sessions per language.

### v2.D. Per-language Typst templates
**TRK-TRANS-004** (new). `series-template-fr.typ` with French typography (nbsp before `;`, `:`, `?`, `!`; guillemets; em-dash dialogue; hyphenation patterns; accented capitals). Then `-es.typ`, then `-zh-Hans.typ` (CJK fonts, fullwidth punctuation, line-breaking rules — much bigger jump). Each ~1 session.

### v2.E. Per-language EPUB stylesheets
**TRK-TRANS-005** (new). `epub-styles-fr.css` (font stack), `-es.css`, `-zh-Hans.css` (Source Han Serif / Noto Serif CJK). `<html lang="…">` tagging. ~1 session per language.

### v2.F. Per-language automated validators
**TRK-TRANS-006** (new). Regex-based checks (this is not an LLM job): French bare `:`/`;`/`?`/`!` without nbsp; Spanish `?` without leading `¿`; Chinese halfwidth punctuation in CJK runs. Run as part of compile. ~1 session.

### v2.G. ISBN/ONIX/cover registry per language
**TRK-TRANS-007** (new). At 20-50 books × 3 langs × 2 formats = 60-200 ISBNs/year. Schema in DB, UI in admin, ONIX record generator per distributor. ~2-3 sessions; meaningful design surface area.

### v2.H. Cross-lingual glossary system
**TRK-TRANS-008** (new). Per-series glossary in git (`glossaries/{series}-{lang}.yaml`). Hermes/orchestrator commits updates with translator attribution. Backlist re-translation flagging. Terminology drift detection across volumes. ~2 sessions.

### v2.I. Orchestrator agent (hermes-style)
**TRK-TRANS-009** (new). Spawns translation jobs on `ms:final`. Posts to Slack on draft-ready ("87k words, ~40-60hr review, side-by-side at translations/fr/draft-v1.md"). Diffs MTPE vs LLM draft, flags hotspots for native reviewer. Weekly digest cron. Significant agent-design work — separate planning session first.

## Suggested session-by-session sequencing

Pick from the top of v1 critical path first. Each row is a discrete session.

| # | Session focus | Hours | Unblocks |
|---|---|---|---|
| 1 | CP-1 + CP-4 (Typst spec→compile + Libertinus on VM) | 3-4 | CP-5, CP-7 |
| 2 | CP-2 + CP-6 (EPUB spec→compile + CSS drift audit) | 3 | CP-7 EPUB |
| 3 | CP-3 (corrections API + round-trip) | 3 | full v1 workflow |
| 4 | CP-5 (Ghosts parity matrix) | 2-3 | release readiness |
| 5 | CP-7 (test suite for compile pipeline) | 2-3 | regression safety |
| 6+ | v2 sessions per the v2 phases above, in user-prioritized order | 1-2 each | foreign editions |

Sessions 1-3 are the **minimum to ship a real book through the v1 pipeline**. Session 4 is the **release-confidence check**. Session 5 is the **regression-safety net**. Beyond that, v2 work can interleave with normal feature/ops work.

## Open questions to settle before starting

- **Body text alignment in EPUB**: justified (current) or left-aligned (per refinement prompt)? Confirm with a side-by-side of a sample chapter in both renderings before changing.
- **EPUB strategy decision (MIG-005)**: confirm "shell script consuming spec via env + rendered CSS overlay" approach before CP-2.
- **Anthology vs single-chapter**: is Ghosts the next book through the pipeline, or is the first real book a single-chapter (like Twitter Years)? Affects whether CP-8 (anthology) becomes urgent.
- **Translation languages priority**: if v2 starts, which language first? Doc recommends `fr` (easiest MT pair, easiest reviewer recruitment); user has not explicitly chosen.
- **Translation rights**: per the doc — verify contracts with source authors permit MT-then-edit workflow before building the v2 pipeline.
