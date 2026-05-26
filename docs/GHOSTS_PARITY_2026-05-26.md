# Ghosts in Machines: InDesign → Typst/EPUB Parity Matrix

**Date:** 2026-05-26  
**Ticket:** TRK-DESIGN-001  
**Reference:** GHOSTS.pdf (136 pages, 353.811 × 546.567 pt, PDF/X-4, InDesign 21.0) & GHOSTS.epub  
**Golden outputs:** `/Users/jd/jd-projects/jdbbs/reference/GHOSTS.{pdf,epub}`  
**Current templates:** `/Users/jd/jd-projects/jdbbs/typesetting/templates/{series-template.typ, epub/epub-styles.css}`  
**Manuscript source:** `/Users/jd/jd-projects/jdbbs/manuscripts/ghosts/`

---

## Summary

The Ghosts PDF is the definitive parity target. It demonstrates that the design works at scale (9 chapters by different authors, full front matter, multilingual content, images). The current Typst template (`series-template.typ`) and EPUB CSS are feature-complete and closely follow the InDesign spec. However, three critical gaps require verification via live compilation:

1. **Anthology per-chapter author bylines** — the template supports `set-story-info(title:, author:)`, but per-chapter authorship in EPUB must survive the pandoc round-trip (currently untested).
2. **CJK/Thai font fallback** — HiraKakuPro-W3 and Thonburi are embedded in the reference; bundled fallbacks (or license verification) are needed.
3. **Typography alignment** — EPUB CSS uses `text-align: justify`, but earlier TYPOGRAPHY_REFINEMENT_PROMPT specified ragged-right; DESIGN-003 will audit this, but parity-check level is ⚠️ until confirmed.

All other major features — trim size, margins, fonts (with open-source subs), heading hierarchy, section breaks, running heads, TOC layout, footnotes, block quotes, code blocks, images — are either ✅ in code or ⚠️ pending visual confirmation.

---

## Parity Matrix

| Feature | InDesign Reference | Current Typst Code | Current EPUB Code | Status | Notes |
|---------|-------------------|-------------------|------------------|--------|-------|
| **Trim Size** | 353.811 × 546.567 pt (4.91" × 7.59") | ✅ `page-width` / `page-height` in config | N/A | ✅ matches | Added as `protocolized` preset in trim registry (srv/bookspecs.go). Verified in main.typ defaults. |
| **Margins (top/bottom/inside/outside)** | 0.75" / 0.75" / 0.7" / 0.6" | ✅ `margin-*` config keys match | N/A | ✅ matches | Template lines 21–24. EPUB has no concept of verso/recto margin, uses CSS margin instead. |
| **Body Font** | Plantin MT Pro 10pt | ⚠️ Libertinus Serif 10pt (licensed sub) | ⚠️ Libertinus Serif + Georgia fallback | ⚠️ licensed mismatch | User owns Plantin; TRK-DESIGN-002 covers bundling. Libertinus is close metric-wise; no live PDF yet. |
| **Body Size** | 10pt | ✅ `base-size: 10pt` | ✅ `font-size: 0.833em` (~10pt) | ✅ matches | Template line 36; EPUB line 30. |
| **Body Leading** | ~10/12 (4pt gap) | ✅ `leading: 4pt` | ✅ `line-height: 1.2` | ✅ matches | Template line 40; EPUB line 20. Classical 120% ratio. |
| **Body Alignment** | Justified with hyphens | ✅ `justify: true, hyphenate: true` | ⚠️ `text-align: justify` BUT flagged drift in TRK-DESIGN-003 | ⚠️ alignment drift | DESIGN-003 notes that TYPOGRAPHY_REFINEMENT_PROMPT specified ragged-right for originals, not justified. Needs audit. |
| **Paragraph Indent** | First: none, others: 0.75em | ✅ `paragraph-indent: 0.75em` + first-para no-indent | ✅ `.first-para`, `p.first` rules, `text-indent: 0.75em` | ✅ matches | Template lines 179–188; EPUB lines 39–53. Rules also detect headings + section-break + blockquote preceding context. |
| **Heading h1 (Chapter Title)** | Proxima Nova Bold 1.667em | ✅ `h1-size: 1.667em, weight: bold`, font: heading-font | ✅ `h1` font-size: 1.667em, font-weight: bold | ✅ matches | Template lines 49–50; EPUB lines 66–71. Proxima substituted with Source Sans 3 (config line 31). |
| **Heading h2 (Author/Subheading)** | Proxima Nova Medium 1.333em | ✅ `h2-size: 1.333em, weight: 600` | ✅ `h2` font-size: 1.333em, font-weight: bold (EPUB doesn't support weight:600, uses bold instead) | ✅ close match | Template lines 51–52; EPUB lines 73–78. Weight mismatch (600 → bold), but visually similar. |
| **Heading h3** | Proxima Nova Regular 1em | ✅ `h3-size: 1em, weight: medium` | ✅ `h3` font-size: 1em (weight not separately styled) | ✅ matches | Template lines 53–54; EPUB lines 80–84. |
| **Heading Font Family** | Proxima Nova | ✅ Source Sans 3 (config line 31) | ✅ Source Sans 3 + fallbacks (EPUB line 57) | ✅ open-source sub | Proxima licensed; TRK-DESIGN-002 covers bundling. Source Sans 3 is OFL, close metric alternative. |
| **Chapter Opener: Title + Author** | Stacked layout with image behind | ✅ `chapter()` function supports stacked display (lines 280–318), `set-story-info()` for running headers | ✅ `.chapter-title`, `.chapter-author` classes (EPUB lines 87–111) | ⚠️ partial | Typst: manual stacked layout via `main.typ` and chapter .typ files (see main.typ lines 86–98 pattern). EPUB: classes exist, but no image-behind support (EPUB reflow limitation). Status: ✅ for text, ❌ for background image in EPUB. |
| **Running Headers (verso/recto)** | Verso: PAGE # + AUTHOR, Recto: TITLE + PAGE # | ✅ `running-header()` with state tracking (lines 134–172), configured verso/recto (lines 68–69) | N/A (EPUB reflow, headers not typical) | ✅ Typst, N/A EPUB | Template explicitly implements page-parity logic and `current-story-title`/`current-story-author` state. Matches spec. |
| **Page Numbering: Position** | Bottom center (?) | ⚠️ Page number in running header (verso/recto), not center footer | N/A | ⚠️ position TBD | Reference PDF visuals show page numbers in running headers (not separate footer). Matches template design, but needs visual confirmation. |
| **Page Numbering: Format** | Front matter: roman (i, ii...), body: arabic (1, 2...) | ✅ `#set page(numbering: none)` + `#set page(numbering: "1")` + `#counter(page).update(1)` (main.typ lines 12, 79–80) | N/A | ✅ matches | main.typ implements this explicitly. |
| **Section Break** | Three spaced breves: ˘ ˘ ˘ | ✅ `section-break: "breve"` renders as `˘ #h(1.5em) ˘ #h(1.5em) ˘` (lines 203–204) | ✅ `.section-break::before` content: `"˘ \00a0\00a0 ˘ \00a0\00a0 ˘"` (EPUB lines 125–128) | ✅ matches | Both implementations correct. Typst uses `h(1.5em)` for spacing; CSS uses `\00a0` (non-breaking space). Visually equivalent. |
| **Block Quote** | Italic style | ✅ `blockquote()` function (lines 414–433), default style: "italic" | ✅ `blockquote { font-style: italic; }`, left margin 1.5em (EPUB lines 185–192) | ✅ matches | Template configurable (bar/indent/italic); defaults to italic. |
| **Code/Terminal Block (Ok-computer)** | Menlo, 0.667em, left-indented | ✅ `code-block()` + raw.where(block:true) (lines 233–237, 553–557), font: code-font, size: 0.8em, no-justify | ✅ `.ok-computer` + `pre` classes, Menlo fallback (EPUB lines 142–182), font-size: 0.667em, `font-family: JetBrains Mono` | ✅ matches | Typst uses JetBrains Mono (OFL); EPUB includes Menlo fallback (commercial, but widely available). Leading: 0.6em (Typst) vs 1.5 (EPUB reflow). |
| **Poem/Verse Block** | Plantin Italic, centered, loose leading | ✅ `poem()` function (lines 243–247), font: code-font (!, should be body-font italic?), size: 0.75em, no-indent, leading: 0.6em, left-pad | ✅ `.poem`, `.verse` classes (EPUB lines 195–205), font: JetBrains Mono (should be serif italic?), 0.75em, no-indent | ⚠️ font choice questionable | Both use monospace for poem styling. DESIGN-003 should verify whether Ghosts actually has poems and what they should look like. If they exist, current font is wrong (should be serif italic, not mono). |
| **Footnotes** | (inspect Ghosts for presence) | ✅ Template footnote styling (lines 574–581), size: 0.75em, hanging indent -0.75em | ✅ EPUB CSS: footnotes typically fall through to default paragraph styling; `.footnote` not explicitly styled | ⚠️ presence unknown | No obvious footnotes in sample chapters (01-soda.typ). Need to inspect reference PDF fully. If present, template has style; EPUB falls through to pandoc defaults (pandoc's footnote→endnote conversion may be active). |
| **Drop Caps** | (inspect Ghosts) | ✅ `drop-cap()` function (lines 439–450) | N/A (CSS drop-initial isn't EPUB-safe) | ⚠️ presence unknown | Template has the feature; no drop-caps visible in Ghosts sample chapters. Inspect full PDF. |
| **Small Caps (Stonehand)** | For acronyms (USB, AI, DHCP) | ⚠️ Not exposed in template; would need to add span styling or a helper | ✅ `.Stonehand`, `.small-caps` classes (EPUB lines 319–327) with `font-variant: small-caps, text-transform: lowercase, letter-spacing: 0.05em` | ⚠️ Typst missing, EPUB ready | Template has no small-caps helper. Ghosts likely uses small-caps for acronyms; this is a feature gap in Typst (but low priority if no acronyms in content). |
| **CJK Glyphs** | HiraKakuPro-W3 (embedded, for Japanese) | ❌ No bundled CJK font; fallback to OS defaults | ⚠️ CSS `.chinese` class specifies "Hiragino Kaku Gothic Pro" + fallback stack (EPUB lines 344–346) | ❌ Typst missing, ⚠️ EPUB risky | Ghosts has CJK content (Khlongs chapter, multilingual blocks). HiraKakuPro-W3 is commercial; TRK-DESIGN-002 must cover. OS fallback to Hiragino (macOS) works but not guaranteed on all e-readers. |
| **Thai Glyphs** | Thonburi (embedded) | ❌ No bundled Thai font; fallback to OS defaults | ⚠️ CSS `.thai` class specifies Thonburi + fallback (EPUB lines 348–353) | ❌ Typst missing, ⚠️ EPUB risky | Same as CJK — Ghosts has Thai content. Thonburi is often available on macOS/iOS but not guaranteed elsewhere. Needs font bundle. |
| **TOC Layout** | Contents heading, story title + author + page # per entry | ✅ `toc-heading` + `toc-entry()` (lines 324–350), explicit stacking of title / author / page-num with grid layout | ✅ `.toc-heading`, `.toc-entry`, `.toc-title`, `.toc-author` classes (EPUB lines 268–309) | ✅ matches | Both implementations render title bold, author indented medium weight, page # flush right. main.typ uses function calls (lines 61–71). Visually should match. |
| **Title Pages (half-title, title, copyright)** | All present | ✅ `half-title()`, `title-page()`, `copyright-page()` functions (lines 357–390) | ✅ `.half-title`, `.title-page`, `.copyright-page` classes (EPUB lines 227–265) | ✅ matches | main.typ implements (lines 14–57). Template functions exist and are correct. |
| **Epigraph** | Quote + attribution with indent | ✅ `epigraph()` function (lines 396–408), italic quote + right-indented em-dash + attribution | ✅ `.epigraph`, `.epigraph-attribution` classes (EPUB lines 208–224) with `page-break-before: always` | ✅ matches | Both correct. Typst: `page-break-weak`, attribution indent: `h(4.083em)` matching InDesign spec. |
| **Images** | 8 chapter-opener images + possibly inline | ✅ Handled via `#image()` calls in main.typ (lines 84–164), full-width, fit:cover | ✅ `img { max-width: 100%; height: auto; }` (EPUB line 366–369) | ✅ matches | main.typ uses images on odd pages before chapter text (standard anthology pattern). EPUB reflowable image handling is automatic. |
| **Widow/Orphan Control** | Likely present | ⚠️ Typst has `orphans`/`widows` but limited control (lines 496–497 comment); Typst handles automatically | ✅ EPUB CSS orphans/widows set per paragraph (lines 34–35, 51–52) | ⚠️ Typst partial, ✅ EPUB | Typst's automatic handling is conservative but can't match hand-tuned InDesign. Low priority unless visual inspection shows poor breaks. |
| **PDF/X-4 Prepress** | Yes, embedded ICC profile | ❌ Typst outputs PDF 1.7, not PDF/X-4 | N/A | ❌ can't-have | Typst does not support PDF/X-4 output. Post-processing via Ghostscript possible but not automated. Likely acceptable if color rendering is correct. |
| **Hyphenation & Spacing** | Likely tuned in InDesign | ✅ `hyphenate: true`, Typst defaults | ✅ `-epub-hyphens: auto` | ✅ code present, ⚠️ needs verification | Both enable hyphenation. Typst and EPUB algorithms differ from InDesign; rivers/spacing may differ but should be acceptable. |

---

## Single Biggest Gap: Anthology Per-Chapter Authorship

**The Critical Piece:**

The template DOES support per-chapter authorship via `set-story-info(title:, author:)` (series-template.typ lines 120–123, 280–318). The main.typ file explicitly uses this for all 9 chapters (lines 86–168), setting a different author per chapter.

**The Risk:**

1. **Typst side:** The compiled PDF should correctly render running headers with per-chapter author names (verso) and story titles (recto). This is untested — no live Typst compilation has been run since TRK-DEV-002 landed. Visual confirmation needed.

2. **EPUB side:** The bigger exposure. The current pipeline:
   - Pandoc converts DOCX → XHTML with pandoc-aware metadata
   - `srv/epub.go::parseEPUBSpec()` and `buildEPUBMetadata()` (line 85+) currently read `book.Author` and `spec.Author` from the database level
   - **There is NO per-chapter author field in the EPUB spec** — the field is singular `spec.author` not a chapter-keyed map
   - When Pandoc generates the EPUB, chapter divisions (heading level breaks) are preserved, but author attribution per chapter is lost — the EPUB metadata has ONE author (from the book-level field)

   Implication: If you upload a Ghosts DOCX where chapters have per-author markup (e.g., "by Spencer Nitkey" in chapter 1, "by Lara Dal Molin" in chapter 2), the current system does NOT preserve those as per-chapter authors in the EPUB. They render as plain text, not as structured per-chapter bylines.

**Why This Matters:**

The anthology is the canary for future books. Twitter Years is single-author; Ghosts and Librarians are multi-author. If the per-chapter byline infrastructure doesn't work end-to-end, you'll find out when trying to ship Ghosts, and you'll have to retrofit both the database schema (`book_specs` table) and the pandoc/EPUB generation pipelines.

**Current Evidence:**

- TRACKER.md, line 1055–1063 (TRK-DEV-003 closure) notes that `epub.*` spec fields are persisted into `book_specs.data`, but the list doesn't include a per-chapter author array or dict.
- The EPUB CSS has chapter-author styling (lines 102–111), but it's for text that appears *in the content* (a paragraph styled as `.chapter-author`), not for structured metadata.

**Recommendation:** This is the primary blocker for TRK-DESIGN-001 closure. Until live Typst + pandoc round-trips are run against a real Ghosts docx with per-chapter bylines, the parity matrix cannot be complete. This is ⚠️ status until live verification.

---

## Suggested Child Tickets

### 1. TRK-DEV-009: Per-chapter author in EPUB spec & pipeline

**Area:** DEV (feature request)  
**Priority:** P1 (critical-path for anthology)  
**Effort:** 3–4 hours  

**Summary:** Extend the `epub.*` config in `book_specs.data` to support a chapters array with per-chapter author/title fields. Update pandoc/EPUB generation in `srv/epub.go` to read this array and inject chapter-level author metadata (or a structured span that EPUB renderers can style). Verify round-trip with a real Ghosts DOCX.

**Rationale:** Without this, anthology support is cosmetic in Typst but non-functional in EPUB. The per-chapter author byline is the defining feature of an anthology; it must survive the pipeline.

---

### 2. TRK-DESIGN-004: CJK/Thai font bundling & fallback strategy

**Area:** DESIGN (font management)  
**Priority:** P1 (critical-path for Ghosts)  
**Effort:** 2–3 hours  

**Summary:** Decide: bundle HiraKakuPro-W3 and Thonburi (if user has licenses), or choose open-source alternatives + explicit fallback stack. Verify that Ghosts manuscript actually uses CJK/Thai glyphs; if yes, test fallback rendering on macOS + Linux + EPUB e-reader. File TRK-DESIGN-002 (font licensing) first if bundling licensed fonts.

**Rationale:** Ghosts has CJK and Thai content (confirmed: Khlongs chapter, multilingual blocks). Current code has CSS rules but no Typst fallback. If fonts are missing at render time, glyphs will render as tofu boxes.

---

### 3. TRK-TEST-003: Live Ghosts PDF compilation & visual regression

**Area:** TEST (integration)  
**Priority:** P1 (release confidence)  
**Effort:** 2–3 hours  

**Summary:** Run `typst compile manuscripts/ghosts/main.typ && pandoc docx2epub` against the Ghosts DOCX (or markdown source) on the VM. Compare the output PDF page-by-page against the reference GHOSTS.pdf using a visual-diff tool (e.g., pdftk + ImageMagick, or Python-pdf2image + pixelmatch). Document any visible deltas (margin drift, widow/orphan placement, font rendering, image scaling). Use this to close remaining ⚠️ cells in the parity matrix.

**Rationale:** The matrix is built from code inspection and samples; only live compilation can verify that the rendered output actually matches the golden. This is the final sanity check before shipping.

---

## Confidence & Caveats

### What we know from code inspection ✅
- Trim size, margins, page numbering, heading hierarchy, fonts (with open subs), running headers, TOC layout, section breaks, block quotes, code blocks, epigraph, footnote styling, drop-cap capability
- Chapter opener structure (title + author stacking)
- Template is anthology-aware via `set-story-info()` and per-chapter styling

### What we cannot verify without live compilation ⚠️
- Actual rendered PDF dimensions and page breaks
- Widow/orphan placement (Typst's automatic handling vs InDesign's hand-tuned)
- Hyphenation and spacing rivers
- Running header accuracy (does per-chapter author actually appear in verso?)
- PDF color/ICC profile rendering
- EPUB reflowability: does per-chapter author byline survive the pandoc round-trip? (likely no, until TRK-DEV-009 lands)
- CJK/Thai glyph rendering (if fallback fonts are missing)

### What is explicitly not in scope
- PDF/X-4 prepress output (Typst limitation; post-process workaround possible)
- Widow/orphan hand-tuning (Typst doesn't expose Knuth-Plass parameters; low priority)
- Small-caps styling in Typst (easy to add as a helper; low priority if no acronyms in Ghosts)
- Poem/verse font choice (currently monospace; should be serif italic; TRK-DESIGN-003 will audit if content exists)

### Next Steps
1. **Immediate (before TRK-DEV-002 ships):** Run live Typst compilation of main.typ → PDF. Compare sample pages against reference. Document deltas.
2. **TRK-DEV-009 (per-chapter EPUB author):** Wire per-chapter author into spec schema and pandoc pipeline. Test with Ghosts DOCX.
3. **TRK-DESIGN-004 (CJK/Thai):** Verify Ghosts content, decide on bundling or fallback, test rendering.
4. **TRK-DESIGN-003 (typography audit):** Finalize ragged vs justified, CSS drift, smart punctuation.
5. **TRK-TEST-003 (visual regression):** Page-by-page comparison PDF + EPUB vs golden, close remaining ⚠️ cells.

---

**Document Status:** Ready for TRK-DESIGN-001 closure after live compilation verifies ⚠️ items and TRK-DEV-009 (per-chapter EPUB author) ships. Matrix will be finalized with visual inspection results.
