# Book-Production Project Summary

**Generated:** 2026-04-07

---

## 1. Git Log — Development Trajectory (last 30 commits)

**Recent focus areas (newest first):**
1. **EPUB metadata fixes** — fallback title metadata for docx conversions
2. **Pipeline snapshot** — pre-test-run asset capture
3. **CI removal** — builds happen on VM, not GitHub Actions
4. **Hosting resolved** — exe.dev VM is permanent ($20/mo)
5. **Reference materials & session docs** — EPUBs, PDFs, extracted sources
6. **Test plan creation** — systematic spec→config→template pipeline testing
7. **Typst template full configurability** — series-template.typ driven by config dict
8. **Word template generation** — python-docx based
9. **Typesetting spec refinement** — running heads, elements layout, pt sizes
10. **Admin SPA** — Typst Frontend (typesetting tab), paper sizes, trim/dimension sync
11. **Bowker ISBN integration research**
12. **Admin dashboard polish** — font comparison, status tooltips, dark mode, capsule UI

**Trajectory:** Project evolved from a dashboard mockup → full production pipeline: Word→Typst→PDF/EPUB with an admin SPA for spec management, config generation, and compilation.

---

## 2. Planning Documents

### TEST_PLAN_2026-03-09.md
- **Covers:** 8-phase systematic test plan for the spec→config→template pipeline
- **Status:** PLANNED — written as a step-by-step QA guide, not yet executed
- **Key decisions:**
  - Test UI binding (dropdown sync, value display)
  - Config generation correctness (em math, quoting, missing fields)
  - Save/load round-trip fidelity
  - Backward compat with old `_em` field names
- **Phases:** (1) Admin SPA basics, (2) Spec editing/saving, (3) Config generation, (4) Font picker, (5) Running heads & elements, (6) Pull from transmittal, (7) Cover upload, (8) E2E Typst compilation
- **Outstanding:** Entire plan is untested — waiting for manual walkthrough

### BOWKER_ISBN_INTEGRATION_PLAN.md
- **Covers:** Research on Bowker/MyIdentifiers.com for ISBN management + integration plan
- **Status:** PLANNED — research complete, implementation not started
- **Key decisions:**
  - **No public API exists** — Bowker is Drupal 7 + AngularJS with internal JSON endpoints
  - **Chose Option B** (Manual workflow with smart prefill) over browser automation
  - New `isbns` DB table, client dashboard ISBN tab, Bowker cheat sheet generation
  - ISBN-13 validation with check digit algorithm
- **Outstanding:** No code written yet. Depends on Phase 1 data model, Phase 2 client UI, Phase 3 pipeline integration

### TYPST_FRONTEND_PLAN.md
- **Covers:** Full architecture for a "Typesetting" tab in the Admin Dashboard
- **Status:** IN PROGRESS — significant implementation done per session prompt 2026-03-08d
- **Key decisions:**
  - New `book_specs` DB table (one spec per project, JSON `data` column)
  - Spec JSON structure: metadata, page, typography, headings, elements, running_heads, front/back matter, custom_styles
  - Two-column admin form layout
  - Config override mechanism: `series-template.typ` accepts external config via `merge-config()`
  - Pull from Transmittal maps transmittal fields → spec fields
- **Outstanding:** Custom styles editor untested, EPUB generation pipeline stubs need completion, multi-chapter support deferred

### TYPOGRAPHY_REFINEMENT_PROMPT.md
- **Covers:** Exact CSS/typography specs extracted from InDesign-produced reference books
- **Status:** REFERENCE DOCUMENT — used to guide template refinement
- **Key decisions:**
  - Body text is LEFT-ALIGNED (ragged right), not justified
  - Font sizes use em units relative to base
  - Code blocks much smaller (0.667em ≈ 8pt) with generous line-height (1.5)
  - Section breaks use breve character (˘)
  - Small-caps treatment for technical acronyms
  - Aggressive orphan/widow control
- **Outstanding:** Template still described as "pretty far off" from InDesign originals

### NEXT_SESSION_PROMPT_2026-03-08d.md
- **Covers:** Session recap — completed integration of admin SPA ↔ Go backend ↔ Typst template
- **Status:** DONE — documents completed work
- **Key accomplishments:**
  - 14 Typst paper sizes in grouped optgroups (replaced 4 hardcoded)
  - `specToTypstConfig()` generates complete Typst config (headings, elements, running heads)
  - 15 new config keys in `series-template.typ` (all hardcoded values → config references)
  - Backward compat: old `_em` fields, `running_headers` key, legacy trim formats
- **Open threads:** Print planning dashboard integration, EPUB pipeline, E2E compile testing, custom styles in spec→Typst untested, push prodcal to GitHub

---

## 3. Test Files Analysis

### Test Infrastructure

| File | Purpose | Quality |
|------|---------|--------|
| `edge-case-generator.py` | Generates Word docs with 10 edge case categories | Good — comprehensive fixture generator using python-docx |
| `test-edge-cases.sh` | Shell-based integration test runner | Good — tests conversion + decision hooks with assertions |
| `test_preflight_detector.py` | Python unittest suite for preflight detector | Good — 14 test methods, covers auto-preserve, HTML output, image inventory, scripts |
| `preflight-detector-declared-styles.json` | Test fixture for declared style matching | Simple — 2 entries (tweet-p-ascii, emoji-c) |
| `EDGE_CASE_HANDLING_GUIDE.md` | Documentation of all edge case categories + solutions | Thorough — covers 8 categories, priority fixes, implementation phases |

### Test Coverage

**What IS tested:**
- Word→Typst conversion for 10 edge case types (bold/italic, color, fonts, spacing, section breaks, lists, Unicode, tables, custom styles, nested lists)
- Decision hook system (strip/keep/convert) for manual lists, colored text, highlighted text
- Grouped list block normalization (numbered, bullet, nested)
- Preflight detector: manual formatting detection, ASCII art candidates, image inventory, non-Latin scripts (CJK, Thai), declared style matching, HTML report generation
- Regression assertions in shell script (grep-based)

**What is NOT tested:**
- EPUB output validation (epubcheck mentioned but not automated)
- PDF compilation end-to-end
- Config generation correctness (the TEST_PLAN covers this but is unexecuted)
- Font picker population
- Save/load round-trip
- Transmittal pull
- Cover upload
- Custom styles → Typst output
- Typography fidelity vs InDesign reference
- Running heads output
- Front/back matter generation

### Test Output Files
- `output/clean-reference.typ` — Clean Typst from well-formatted Word doc ✓
- `output/edge-cases-test.typ` — Conversion with editorial review comments (no decisions applied)
- `output/edge-cases-enhanced.typ` — Enhanced conversion (lists partially handled but not grouped)
- `output/edge-cases-decision-hook.typ` — Conversion WITH decisions applied (stripped items removed, converts mapped to Typst list syntax)
- `output/test-edge-decisions.json` — 14 editorial decisions (strip/convert/keep)
- `preflight-detector-test-report.json` — JSON findings from detector (observed styles, manual formatting, images, scripts)
- `preflight-detector-test-report.html` — Styled HTML report with IBM Plex fonts, grid layout

---

## 4. Context Files (Hermes Agent Configuration)

### .context/hermes-investigation.md
- **Covers:** Fixed `base_url` pointing to OpenRouter despite `provider: anthropic`
- **Status:** DONE — config fixed

### .context/hermes-openai-fix.md
- **Covers:** ChatGPT Codex backend returning empty output; fixed by routing to standard OpenAI API
- **Status:** DONE — `active_provider` cleared, `base_url` set to api.openai.com
- **Remaining:** Anthropic key over limits until 2026-05-01; compression summary model needs switching

### .context/hermes-cost-control-options.md
- **Covers:** 14 cost control levers for Hermes 0.8.0 (reasoning_effort, smart routing, compression, etc.)
- **Status:** REFERENCE — options documented, recommended "frugal" config provided

### .shared-context/hermes-openai-key-findings.md
- **Covers:** Explains two separate OpenAI env vars (`OPENAI_API_KEY` for LLM, `VOICE_TOOLS_OPENAI_KEY` for STT/TTS)
- **Status:** DONE — mystery solved, config is correct

---

## 5. Key Outstanding Issues & Blockers

1. **No automated E2E tests** — The test plan exists but hasn't been executed; no CI pipeline
2. **Typography fidelity gap** — Templates still don't match InDesign reference output
3. **EPUB pipeline incomplete** — Backend stubs may need completion
4. **ISBN integration not started** — Plan exists, no code
5. **Custom styles untested** — Plumbing exists but unverified
6. **prodcal not on GitHub** — Only book-production is pushed; prodcal is VM-only
7. **Anthropic API key over limits** — Compression summary model needs alternative until 2026-05-01
8. **Manual list nesting** — Flattened; true hierarchy/indent-level reconstruction deferred
9. **Multi-chapter books** — Pipeline handles single-chapter only; anthologies need chapter-level metadata
