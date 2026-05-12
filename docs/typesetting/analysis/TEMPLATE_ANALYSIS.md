# Template & Source File Analysis

Comprehensive technical analysis of all template files, source files, and documentation.

---

## 1. TEMPLATE FILES

### 1.1 `templates/styles.typ` — Character-Level Style Registry

**Purpose:** Extensible registry of inline text styling functions. Acts as a shared utility library imported by the main template.

**Functions defined (20 total):**

| Function | Signature | Purpose |
|----------|-----------|--------|
| `sc(content)` | OpenType `c2sc` + `smcp` | Small caps for acronyms |
| `ital(content)` | `style: "italic"` | Explicit italic |
| `bold(content)` | `weight: "bold"` | Explicit bold |
| `bold-ital(content)` | bold + italic | Combined |
| `ellipsis` | constant | Spaced ellipsis `. . .` |
| `spaced-ellipsis(spacing)` | default 0.3em | Parameterized ellipsis |
| `super(content)` | baseline -0.4em, 0.7em size | Superscript |
| `sub(content)` | baseline 0.2em, 0.7em size | Subscript |
| `uline(content)` | wraps `underline` | Underline |
| `strike(content)` | `stroke: 0.5pt` | Strikethrough |
| `tracked(amount, content)` | default 0.1em | Letterspaced text |
| `allcaps(tracking, content)` | default 0.05em | Uppercase + tracking |
| `highlight(color, content)` | default yellow 60% | Background highlight |
| `sans(font, content)` | default "Source Sans 3" | Sans-serif inline |
| `mono(font, content)` | default "JetBrains Mono", 0.9em | Monospace inline |
| `foreign(lang, content)` | italic + lang tag | Non-English text |
| `acronym(expand, content)` | small caps | Acronym styling |
| `name(use-smallcaps, content)` | optional small caps | Proper nouns |
| `booktitle(content)` | italic | Book/work titles |
| `articletitle(content)` | quotes | Article titles |
| `pullquote(content)` | 1.1em italic | Pull quote text |
| `oldstyle(content)` | `onum` feature | Old-style figures |
| `lining(content)` | `lnum` + `tnum` | Lining tabular figures |

**Issues identified:**
- `strike()` uses `stroke: 0.5pt` on text which draws a border around each glyph, NOT a strikethrough line. Should use `strike()` built-in or a line overlay.
- `super()` shadows Typst's built-in `super` function — may cause conflicts.
- `sub()` shadows Typst's built-in `sub` function — may cause conflicts.
- `articletitle()` outputs `["#content"]` which renders the literal string `#content` rather than the actual content. The quotes/hash syntax is wrong.
- `acronym()` has an unused `expand` parameter path (comment says "Could add tooltip or footnote in future").
- Hardcoded fonts: `sans()` defaults to "Source Sans 3", `mono()` defaults to "JetBrains Mono" — these should ideally pull from config.

---

### 1.2 `templates/series-template.typ` — Main Book Template

**Purpose:** The primary template for the "Protocolized Anthology" series. Provides page setup, running headers, chapter openers, front matter, and the main `book()` show rule.

**Imports:** `styles.typ: *`, `images.typ: *`

**Configuration system:**
- `default-config` dictionary with 30 keys covering page, margins, fonts, sizes, headings, running heads, and element styles
- `merge-config(overrides)` function for partial override
- However, `config` is bound to `default-config` at module level and the `book()` function uses the module-level `config` — meaning `merge-config()` returns a new dict but doesn't update the live config. **Callers cannot actually override config through the documented API.** They would need to reassign `config` after import.

**Page dimensions (default):**
- Width: 353.811pt (4.91"), Height: 546.567pt (7.59") — custom size, ~1:1.55 ratio
- Margins: top 0.75in, bottom 0.75in, inside 0.7in, outside 0.6in
- No `page-paper` named size used by default

**Typography:**
- Body: Libertinus Serif, 10pt
- Headings: Source Sans 3
- Code: JetBrains Mono
- Leading: 2pt (for 10/12 = 1.2 ratio)
- Paragraph indent: 0.75em
- Justified, hyphenated, English

**Running headers:**
- State-based system using `state()` for story title, author, and suppression list
- Verso (even): page number left, AUTHOR (uppercased) right
- Recto (odd): title left, page number right
- Font: Source Sans 3, 0.75em, medium weight
- `no-header()` marks pages for suppression (chapter openers)
- `set-story-info()` updates running header content
- Bug: `suppress-header-pages` uses `.final()` which reads the final value across the entire document — meaning if ANY page is suppressed, the list builds up. This works for suppression but is not the most efficient approach.

**Components provided:**

| Component | Type | Description |
|-----------|------|------------|
| `section-break` | block | Configurable: breve/asterism/dinkus/blank/fleuron |
| `section-break-stars` | block | Legacy asterisk variant |
| `code-block(content)` | block | JetBrains Mono, 0.8em, no justify |
| `poem(content)` | block | JetBrains Mono, 0.75em (NOTE: uses code font for poetry) |
| `chapter(title, author, ...)` | block | Chapter opener with optional stacking |
| `stacked-title(title, ...)` | block | Multi-line title using " / " split |
| `stacked-author(author, ...)` | block | Multi-line author using " / " split |
| `toc-heading` | block | "Contents" heading |
| `toc-entry(title, author, page)` | block | Manual TOC entry |
| `half-title(title)` | page | Half-title page |
| `title-page(title, subtitle)` | page | Full title page |
| `copyright-page(content)` | page | Copyright (0.667em) |
| `epigraph(quote, attribution)` | block | Italic quote + attribution |
| `blockquote(content)` | block | Configurable: italic/bar/indent |
| `drop-cap(letter, body)` | block | 3em drop capital |
| `first-para(content)` | inline | No-indent paragraph |
| `body-para(content)` | inline | Indented paragraph |
| `book(title, subtitle, author, ...)` | show rule | Main document template |

**`book()` show rule configures:**
- Page dimensions and margins from config
- Running header in page header (no footer/page numbers in footer)
- Base text settings (font, size, lang, hyphenation)
- Paragraph settings (justify, leading, indent)
- Heading styles (H1 forces pagebreak, H2/H3 add vertical space)
- Raw code block styling
- Emphasis/strong styling
- `line` → section-break mapping (horizontal rules)
- Footnote styling

**Issues identified:**
- `poem()` uses `config.code-font` (JetBrains Mono) for poetry — unusual choice. Docs don't mention this. Poetry traditionally uses the body or italic font.
- `book-page` is defined but never used anywhere — dead code.
- Footnote show rule references `super` which was shadowed by `styles.typ` import — potential conflict.
- `heading-align` config is a string ("left"/"center"/"right") compared with `==` — works but is fragile. No validation.
- `toc-entry()` takes page number as a string, not linked — manual TOC, not auto-generated.
- The `chapter()` function's `background-image` parameter is accepted but never used.
- Page numbers only appear in running headers — no standalone folio for chapter openers where headers are suppressed.

---

### 1.3 `templates/images.typ` — Image Placement Utilities

**Purpose:** Comprehensive library of image placement functions for various book layouts.

**Configuration:** `image-config` dict with `bleed` (0.125in), `gutter` (0.25in), `caption-size` (0.833em), `caption-style` ("italic").

**Functions (13 total):**

| Function | Purpose | Notes |
|----------|---------|-------|
| `full-bleed-image(path, ...)` | Extends to trim edges | Hardcoded margin offsets (-0.7in, -0.75in) |
| `full-page-image(path, ...)` | Within margins | Optional caption |
| `frontispiece(path, ...)` | Verso facing title | Forces even page |
| `chapter-opener-image(path, ...)` | Full-page opener | Forces odd page, hardcoded offsets |
| `figure-inline(path, ...)` | Within text flow | Uses `figure()` element |
| `figure-side(path, ...)` | Margin float | Simplified (no true float in Typst) |
| `image-grid(paths, ...)` | Grid layout | Configurable columns |
| `portrait(path, ...)` | Height-constrained | Centered |
| `ornament(path, ...)` | Decorative element | Section break replacement |
| `icon(path, ...)` | Inline icon | Baseline-adjusted |
| `page-background(path, ...)` | Watermark | Center placement |
| `bordered-image(path, ...)` | With border | Configurable stroke |
| `wrap-image(path, side, ...)` | Text wrap | Grid-based approximation |

**Issues identified:**
- `full-bleed-image()` and `chapter-opener-image()` have hardcoded margin values (-0.7in, -0.75in, 1.3in, 1.5in) that only work with the default config margins. Should compute from `config.margin-*`.
- `page-background()` doesn't actually apply opacity — the parameter is accepted but unused.
- `figure-side()` is a rough approximation — Typst limitation acknowledged in comments.
- `wrap-image()` is grid-based, not true text wrap — also a Typst limitation.
- None of these functions are actually used by `main.typ` (the ghosts book uses raw `image()` calls directly).

---

### 1.4 `templates/pandoc-typst.typ` — Pandoc Template

**Purpose:** Bridge template for Pandoc's `--template` flag when converting from Markdown to Typst.

**Structure:** Imports `series-template.typ`, applies `book.with()` using Pandoc template variables (`$title$`, `$subtitle$`, `$author$`). Optionally generates TOC using `#outline()`. Sets up roman numerals for front matter, arabic for body.

**Issues:**
- Very minimal — doesn't handle most of the rich components from the series template (chapters, epigraphs, running headers).
- The TOC uses Typst's built-in `#outline()` rather than the manual `toc-entry()` system.
- No handling of chapter-level metadata (story authors, images).

---

### 1.5 `templates/manuscript-transmittal-form.md` — Intake Form

**Purpose:** Comprehensive intake form for new book projects. Markdown-formatted, designed to be filled in by production staff.

**Covers:** Project metadata, ISBNs, series info, trim sizes, complexity level, 40+ style checkboxes organized by category (body text, headings, special content, character styles, front/back matter), custom style slots (5), typography preferences (with Libertinus Serif default), image requirements, table requirements, editorial notes.

**Trim size options:** 5.5"×8.5", 6"×9", 5"×8", Custom

**Issue:** The default trim in the template (4.91"×7.59") doesn't match any of the transmittal's checkbox options. The form should include the series' custom size or make it clearer.

---

### 1.6 `templates/epub/epub-styles.css` — EPUB Stylesheet

**Purpose:** Full CSS stylesheet for EPUB3 output, mirroring the print design.

**Coverage (~350 lines):**
- CSS reset
- Base typography: Libertinus Serif, 1em, line-height 1.2, justified, hyphenated
- Paragraphs: 0.833em, 0.75em indent, orphans/widows control
- First-paragraph selectors (6 different context selectors + 2 class selectors)
- Headings: Source Sans 3, proper hierarchy (1.667em, 1.333em, 1em)
- Chapter title/author classes matching InDesign names (`Chap-title-epub`, `Chap-author-epub`)
- Section breaks with CSS `::before` pseudo-element generating breve characters
- HR override with `::after` pseudo-element
- Code/terminal: JetBrains Mono, 0.667em, pre-wrap
- Block quotes: italic, no indent
- Poetry: JetBrains Mono, 0.75em
- Epigraph/attribution with InDesign class names
- Front matter (half-title, title page, copyright)
- TOC with heading/title/author classes
- Character styles: italic, small caps (with `font-variant` + `text-transform: lowercase`), ellipsis spacing, bold
- Multilingual support: Chinese (Hiragino/PingFang), Thai (Thonburi/Leelawadee)
- Images: max-width 100%, figure/figcaption
- Lists, links, page breaks
- Running headers (for PDF-like readers)

**Consistency with print template:**
- Font sizes match: 1.667em (H1), 1.333em (H2), 0.833em (body p), 0.667em (code/copyright)
- Font families match: Libertinus Serif, Source Sans 3, JetBrains Mono
- Section break character matches (˘ ˘ ˘)
- Paragraph indent matches (0.75em)

**Issues:**
- Body `font-size: 1em` but `p` has `font-size: 0.833em` — paragraphs are 83.3% of body size, which seems intentional (matching InDesign's relative sizing) but unusual for EPUB where 1em is already reader-controlled.
- No `@font-face` declarations — consistent with the "no embedded fonts" policy stated in clarifications.
- Small caps use `font-variant: small-caps` + `text-transform: lowercase` which is a workaround — true OpenType small caps (`font-feature-settings: "smcp"`) would be more precise but requires embedded fonts.
- Fallback font stacks are sensible.

---

### 1.7 `templates/word/` — Word Templates (Binary)

**Files:** `author-template.docx`, `default-reference.docx`, `protocolized-style-guide.docx`

Binary .docx files — cannot be read as text. Documented in `docs/word-styles.md`.

---

## 2. SOURCE FILES

### 2.1 `src/ghosts/main.typ` — Ghosts in Machines (Main Book)

**Purpose:** The primary real-world book being typeset with this system. A 9-story anthology.

**Structure:**
- Imports `series-template.typ: *`
- Applies `book.with(title, subtitle)` — no author (anthology)
- Front matter: hand-coded half-title, blank verso, title page, copyright, TOC
- 9 chapters, each with: `set-story-info()`, image page, `no-header()`, `#include` of chapter file
- Page numbering: none for front matter, arabic starting at 1 for body

**Pattern per chapter:**
```
#set-story-info(title: none, author: none)  // Clear for opener
#pagebreak(to: "odd")
#image("filename.jpg", width: 100%)
#pagebreak()
#no-header()
#set-story-info(title: "...", author: "...")
#include "NN-name.typ"
```

**Issues:**
- Front matter is entirely hand-coded rather than using the template's `half-title()`, `title-page()`, `copyright-page()` functions. The template functions exist but are unused.
- Images are placed with raw `image()` calls, not using `images.typ` functions like `chapter-opener-image()`.
- Chapter opener images use `width: 100%` without negative margins — they won't bleed. The `chapter-opener-image()` function in images.typ handles bleed but isn't used.
- Typo in Chapter 8 filename: `"ghostts_08_LOYALTY.jpg"` (double t).
- The TOC page numbers are hardcoded strings ("9", "17", etc.) — will go stale if content changes.
- The copyright page resets text size/leading manually instead of using `copyright-page()` template function.

### 2.2 `src/ghosts/00-intro.typ` through `08-loyalty.typ` — Chapter Files

**Common pattern across ALL chapter files:**
```typst
#set text(font: "Libertinus Serif", size: 10pt)
#set par(justify: true, leading: 0.6em, first-line-indent: 0.75em)

#text(font: "Source Sans 3", size: 1.667em, weight: "bold")[Title]
#text(font: "Source Sans 3", size: 1.333em, weight: "medium")[Author]

#v(2em)
#set par(first-line-indent: 0em)
```

**Critical issue:** Every chapter file re-declares base typography settings that the `book()` show rule already establishes. This means:
1. Fonts/sizes are hardcoded in each chapter rather than flowing from config
2. The `leading: 0.6em` in chapters differs from `config.leading: 2pt` in the template — 0.6em at 10pt = 6pt, vs config's 2pt. **The chapters have 3× the leading of the config.** The 0.6em value is likely the correct one (standard 10/16 or similar), meaning the config's `leading: 2pt` is probably wrong or misunderstood (Typst `leading` = inter-line gap, not total line-height).
3. Chapter titles/authors are manual `#text()` calls rather than using `chapter()` or heading syntax
4. `#set par(first-line-indent: 0em)` is set for the first paragraph but never reset back to 0.75em — subsequent paragraphs inherit 0em unless Typst's set rules scope correctly

**Content notes:**
- Chapters are complete short stories (literary fiction, sci-fi themes around AI/ghosts/protocols)
- Heavy use of em dashes, italic emphasis, some escaped special characters
- Chapter 03 (Garden of Eden) is the longest at ~1000+ lines
- Chapter 04 (Tools) includes Chinese text rendered inline
- Chapter 06 (Latency) uses decorative section dividers with em-dashes and asterisks
- Lists in chapters use `#set par(first-line-indent: 0em)` before each item — very manual

### 2.3 `src/sample-book.typ` — Demo/Test Book

**Purpose:** Demonstrates template usage with sample content.

**Uses template features:** `half-title`, `title-page`, `copyright-page`, `toc-heading`, `toc-entry`, `first-para`, `section-break`, `poem`, `epigraph`, `sc`. Uses standard Typst headings (`=`) for chapters.

**Issue:** Uses roman numeral page numbering for front matter (`"i"`) which the ghosts book doesn't do.

### 2.4 `src/edge-cases-test.typ` — Edge Case Test Document

**Purpose:** Tests conversion of problematic Word formatting: manual bold/italic, font chaos, spacing issues, section breaks, lists, special characters, tables, nested lists.

**Notable:** Contains editorial review comments (`// Editorial review: manual list item detected`) showing the edge-case detection system output. Uses `#smallcaps` (Typst built-in) rather than `#sc` from styles.typ.

### 2.5 `src/The-Twitter-Years-200722-template.typ` — Large Real Book

**Purpose:** A 2,879-line converted manuscript (Venkatesh Rao's Twitter compilation). Appears to be Pandoc-converted from Word.

**Structure:** Uses `book.with(title: "TITLE", author: "AUTHOR")` — placeholder metadata not filled in. Contains preface, style demonstrations, and extensive tweet/thread content. Demonstrates real-world conversion output including blockquotes, code blocks, and section breaks.

---

## 3. DOCUMENTATION FILES

### 3.1 `docs/TYPOGRAPHY.md` — Master Typography Reference

**Quality:** Excellent, comprehensive (500+ lines). Covers page geometry, line length, alignment, hyphenation, leading, vertical rhythm, paragraph spacing, type selection, small caps, letterspacing, figures, punctuation, ligatures, book structure, widows/orphans, and when to break rules.

**Accuracy vs implementation:**
- ✅ States 2:3 proportion, template is ~1:1.55 — close
- ✅ States 45-75 char line length as ideal
- ✅ States 120-145% leading — config's 10pt/2pt = 10/12 = 120% ✅ but chapters use 10pt/0.6em ≈ 10/16 = 160% ❌ (exceeds guidance)
- ✅ States indentation default 0.75em — matches
- ✅ States justified default — matches
- ✅ Three breves as default section break — matches
- ✅ Running header conventions (verso author, recto title) — matches
- ⚠️ States page numbers should suppress on blank pages, chapter openers — partially implemented (headers suppressed, but no separate folio system)
- ⚠️ States "oldstyle proportional for body text" default — `oldstyle()` function exists but isn't applied globally
- ⚠️ States Bringhurst prefers spaced periods for ellipsis — the `ellipsis` constant provides this, but it's not automatically applied

### 3.2 `docs/TYPOGRAPHY_PAIRINGS.md` — Font Pairing Guide

**Accuracy vs implementation:**
- Lists 8 themed pairings. Default (#8 "Modern Neutral") lists Libertinus Serif + **Libertinus Sans** for headings.
- ⚠️ **Inconsistency:** The actual template uses **Source Sans 3** for headings, not Libertinus Sans. The EPUB CSS also uses Source Sans 3. The pairing guide's default doesn't match the implementation.
- The guide's Pairing #3 (Scholarly: Crimson Pro + Source Sans 3) and #4 (Business: Source Serif 4 + Source Sans 3) are closer to what's actually implemented.

### 3.3 `docs/WORKFLOW.md` — Pipeline Documentation

**Accuracy:**
- ✅ Correctly describes Word → Pandoc → Typst → PDF flow
- ✅ Documents style mapping tables
- ✅ References `scripts/build.sh`, `scripts/docx-to-typst-enhanced.lua`, `scripts/detect-edge-cases.py`
- ⚠️ References `scripts/md-to-chapter.py` in "Adding New Styles" section — this file doesn't appear in the src tree (may exist in scripts/)
- ⚠️ States Word "No Spacing" style maps to first paragraph — but `docs/word-styles.md` says the same. Neither mentions `#first-para[]` mapping explicitly.

### 3.4 `docs/IMPLEMENTATION_SUMMARY.md` — Status Tracker

**Accuracy:**
- Documents 4 completed components: transmittal form, edge case detection, typography pairings, series template manager
- References `scripts/series-template-manager.py` and `scripts/detect-edge-cases.py`
- Lists correct key decisions (no embedded EPUB fonts, flag-don't-strip edge cases, themed typography choices)
- ⚠️ Lists "Ghosts in the Machine" as baseline book — actual title is "Ghosts in Machines"

### 3.5 `docs/TRANSMITTAL_TO_TEMPLATE_FLOW.md` — Data Flow Documentation

**Accuracy:**
- Documents the full pipeline from transmittal → spec JSON → Word template → dual output
- References a Go backend (`bookspecs.go`, API endpoints) that's part of a separate `prodcal` system
- Well-structured with mermaid diagrams
- Status: mostly "To Be Tested" — integration isn't complete

### 3.6 `docs/CLARIFICATION_QUESTIONS.md` — Blank Template
Empty response fields — this is the template version.

### 3.7 `docs/CLARIFICATION_ANSWERS.md` — Filled Responses
Key decisions captured:
- Style naming: 2-3 letter lowercase project codes, hyphens, max ~16 chars
- EPUB: No custom fonts, small files, but rich image support
- Typography: Inspired by Vellum app's streamlined approach
- Versioning: ISO date-time stamps in filenames
- Edge cases: Flag everything for editorial review
- Accessibility: Full EPUB accessibility metadata required

### 3.8 `docs/typst-paper-sizes.md` — Reference Table
107 named Typst paper sizes with dimensions. Correctly notes the series uses a custom size not matching any named paper. Excellent reference document.

### 3.9 `docs/word-styles.md` — Word Style Guide
Simple mapping table for authors. Accurate and consistent with WORKFLOW.md.

---

## 4. KEY INCONSISTENCIES & ISSUES SUMMARY

### Critical
1. **Leading mismatch:** `config.leading: 2pt` vs chapter files' `leading: 0.6em` (6pt at 10pt). The chapters override the template's value, suggesting 2pt is wrong.
2. **Config system doesn't work:** `merge-config()` returns a new dict but the module-level `config` used by all functions remains `default-config`. Callers can't actually override settings.
3. **Chapter files bypass the template:** All 9 ghosts chapters re-declare fonts, sizes, and leading locally rather than inheriting from `book()`.
4. **Typography pairing default mismatch:** Docs say Libertinus Sans for headings; code uses Source Sans 3.

### Moderate
5. **`styles.typ` bugs:** `strike()` doesn't strikethrough, `articletitle()` is broken, `super`/`sub` shadow builtins.
6. **`images.typ` unused:** None of the image utilities are used by the actual book.
7. **Hardcoded bleed margins** in `images.typ` only work with default page config.
8. **Front matter functions unused:** `half-title()`, `title-page()`, `copyright-page()` exist in template but ghosts/main.typ hand-codes everything.
9. **No page numbers on chapter openers** — headers are suppressed but no alternate folio placement.
10. **Manual TOC** with hardcoded page numbers that will drift.

### Minor
11. **Typo in image filename:** `ghostts_08_LOYALTY.jpg`
12. **Poem styling uses monospace font** — unusual choice.
13. **`chapter()` function has unused `background-image` param.**
14. **Twitter Years template** has placeholder title/author.
15. **`book-page` variable** defined but never used.
