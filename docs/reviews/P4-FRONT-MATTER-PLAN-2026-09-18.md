# P4 — front-matter ingestion: pipeline map + implementation plan (2026-09-18)

Decisions (taken with Jenna 2026-09-17) are in `docs/runs/RUN-2026-09-17-protocol-institute.md`
under **P4 → Decisions**. This note is the *how*, written after mapping the code and before
changing it, so the implementing session can start from here.

## What exists today

| Piece | Where | State |
|---|---|---|
| Print build | `srv/books.go` `runConversion` (~L370–580): pandoc `docx+styles` → `typst+smart` with `typesetting/filters/docx-to-typst-enhanced.lua` → header swap (`generatedHeader` → `#show: book.with(...)`) → `typst compile --root /` | Body only. `book()` in `series-template.typ` (L547) emits **no** half-title / title / © / TOC and never switches folio numbering. |
| Front-matter typst helpers | `series-template.typ` L448–500: `half-title`, `title-page`, `copyright-page`, `epigraph` | Defined, **never called** by the pipeline. Folios: `running-header()` / `drop-folio-footer()` with states `suppress-header-pages`, `drop-folio-pages`. |
| Lua filter | `docx-to-typst-enhanced.lua`: `Header(el)` only injects `#set-story-info()` per H1 when spec has anthology chapters; `Pandoc(doc)` prepends the placeholder header | No classification of headings. |
| Spec | `srv/bookspecs.go` L57 `front_matter{half_title,series_title,title_page,copyright_page,dedication,epigraph,toc,foreword,preface,acknowledgments,introduction}`, `back_matter{notes,appendix,glossary,bibliography,index}`; filled from the transmittal `checklist[]` (L276) | Booleans exist; nothing reads them at build time. `page_iv.copyright_year`, `cover.credit` captured in the transmittal (`transmittal.js` ~L840, ~L1036). |
| Inspect | `typesetting/scripts/detect-edge-cases.py` (`EdgeCaseDetector.detect_*`, 696 lines — never `cat`) | Has `detect_heading_lookalikes`, `detect_observed_styles`; no book map. |
| EPUB | `srv/epub.go` pandoc `--to=epub3`, `--toc` | pandoc generates nav; no landmarks for front/back matter. |
| Word template | `typesetting/scripts/generate-word-template.py` (`_tidy_styles_pane`) | Title as H1 + author as First Paragraph — the model P4 replaces. |

## Plan (in commit order; each step builds + tests green + smoke on a real DOCX)

1. **Book map (Go, pure).** `srv/bookmap.go`: read the DOCX paragraphs (style + text + page-break flags), apply the rule:
   - drop `Title`/`Subtitle`-styled paragraphs (report them);
   - blocks before the first H1, split on page breaks → untitled front matter, named in transmittal order (dedication, epigraph);
   - H1 is *front* if text matches the closed vocabulary (≤ 4 words, case-insensitive: Foreword, Preface, Prologue, Acknowledgments/Acknowledgements, Note on the Text, A Note on …, List of Figures/Tables/Illustrations, Contents, Dedication, Epigraph, Author's Note, Translator's Note) **and** precedes the first non-matching H1; *back* if it matches (Appendix, Notes, Endnotes, Bibliography, References, Works Cited, Glossary, Index, About the Author, Acknowledgments-after-body, Colophon, Afterword, Epilogue?) and follows the last body H1. **Introduction = body.**
   - output `BookMap{Dropped, UntitledFront[], Front[], Body[], Back[], Parts bool}` + human summary string ("Front matter: Foreword, Preface · Body starts at 'Chapter 1' · Back matter: Notes, About the Author").
   - table-driven test with a synthetic DOCX (helper `writeFontTestDOCX` in `epub_fonts_test.go` shows the minimal zip; extend to many paragraphs with `w:pStyle` + `w:br w:type="page"`).
2. **Inspect shows the map + cross-check** against `front_matter`/`back_matter` booleans: "listed but not found", "found but not listed". Surface in the preflight JSON (`manuscript_preflights`) and the factory page Inspect panel (`srv/static/factory.js` — never `cat`; `rg` for where findings render).
3. **Print build.** Pass the map to the lua filter via the existing `--metadata-file` (`writePandocMetadata`): list of H1 titles → `front|body|back`. Filter wraps each H1 in a typst call (`#front-matter-heading[...]` / `#back-matter-heading[...]`) or emits `#start-body()` before the first body H1. Template: `book()` gains `front-matter: (half-title:, title-page:, copyright:, dedication:, epigraph:)` from the spec (generated pages i–iv from the transmittal: title, subtitle, author, publisher, ISBN, © year, cover credit), roman folios until `#start-body()`, which does `pagebreak(to: "odd")` + `counter(page).update(1)` + `set page(numbering: "1")`. Parts opt-in: `--top-level-division=part` when `spec.parts == true`.
4. **EPUB.** Same map → pandoc: strip Title/Subtitle paragraphs, prepend a generated title page + © page XHTML (pandoc `--epub-title-page` is default; © via a small `--include-before-body` or a metadata `rights`), `epub:type` landmarks via a lua filter on the same classification.
5. **Word template + guidance.** Remove the Title-as-H1 model; template opens with a "Front matter" section: instruction text ("Don't type a title or copyright page; start with dedication or foreword; page break between untitled pieces; every section head is Heading 1"). Never embed licensed fonts.
6. **Transmittal**: nothing new to ask — the checklist already carries the booleans; add the "This book has parts" opt-in (`spec.parts`).

Touches the factory path (1, 3, 4). Freeze: Fri/Sat OK to land with tests + smoke; Sun–Tue hotfix only.

## Smoke fixtures on the VM (no credits)

- Admin builds are free on pass-less projects: **7** (Twitter Years, book 8), **14** (Zoothesia/Ghosts test 002, book 9). `POST localhost:8799/api/books/{8|9}/convert` (print) and `/generate-epub`.
- Project 22 (Obliquities, pass 7) has 2 credits — do not use.

## Progress (2026-09-18, same day)

- **Step 1 done** `srv/bookmap.go` (+ tests). Extra rules found on real files: first H1 that repeats the book
  title is dropped with its section (old template model, Twitter Years); a typed "Contents" H1 is dropped;
  a byline ("by Author" / "Author") before the first H1 is dropped; `Copyright`-styled paragraphs dropped;
  `Epigraph`/`Dedication`-styled paragraphs start their own piece without needing a page break.
  Warnings (medium) vs Notes (low) split.
- **Step 2 done** Inspect: factory panel "How the build reads your file", HTML report section `#book-map`,
  findings `book_map` (uncounted) / `book_map_warning` / `book_map_note`; transmittal cross-check.
- **Step 3 done** Print build. Template: `book(front-matter: (...))` generates half-title / blank / title /
  copyright from the spec; `#front-piece(kind:)`, `#contents-page()`, `#front-section()`, `#start-body()`,
  `#start-back()` emitted by the lua filter from `book_map` metadata; folios roman → arabic 1 on a recto;
  **running heads and drop folios now appear on every book** (before P4, non-anthology factory PDFs had none).
  Contents comes after dedication/epigraph (Chicago order). Blank versos are detected after the fact
  (`break-to-recto`) so they carry nothing. `epigraph()` no longer page-breaks (chapter epigraphs).
  Smoked on Twitter Years (628 pp) and Ghosts (anthology, per-story running heads intact).
- **Step 4 done** EPUB. New `typesetting/filters/docx-to-epub.lua` reads the same `book_map` metadata:
  drops the byline / typed copyright / typed Contents / repeated-title section; dedication and epigraph
  become their own sections (hidden `h1.fm-piece-head` for the nav entry + `epub:type`, centred block);
  every H1 gets an `epub:type` from its text (foreword, preface, chapter, appendix, endnotes, …), so
  pandoc writes `<body epub:type="frontmatter|bodymatter">` per file. Pandoc only infers backmatter for
  appendix/colophon/bibliography/index — endnotes/glossary/afterword bodies stay `bodymatter` (section
  types are right; no override exists). epubcheck clean on Ghosts, Twitter Years and the fixture.
- **Step 5 done** Word template guide: new "How the book is assembled" section (three zones; what is
  generated and must not be typed; dedication/epigraph before the first Heading 1; recognised front
  and back heading names; Introduction/Epilogue = body). Copyright style text now says it is dropped.
- **Step 6 done** Parts opt-in. Signal = transmittal "Parts" count ≥ 1 (`checklist_stats.parts`) or
  `structure.parts: true` (`specHasParts`). Print: `config.parts` → H1 part opener (own recto, title
  centred, no head/folio, blank verso), H2 chapter opener, H3/H4 take the H2/H3 look; Contents depth 2
  and its title is our own heading (outline's built-in title is a level-1 heading and became a part
  opener). Lua demotes front/back H1s to H2 so Foreword/Notes still read as chapters. EPUB:
  `--epub-chapter-level=2`, body H1 `epub:type="part"`. `TestPartsBook`. Not `--top-level-division`:
  pandoc's typst writer ignores it.
- **Found, not fixed**: module-level template helpers (`running-header`, `drop-folio-footer`,
  `contents-page`) read the *default* `config`, not the merged one passed to `book()`. So spec
  `running_heads.enabled: false` does not turn running heads off. Same pattern as the parts fix
  (`parts-state`): move the needed keys into a state set by `book()`. Post-workshop.
- **Not yet**:  (Word template +
  guide text), step 6 (Parts opt-in `spec.parts` → `--top-level-division=part`; not started), logo on the
  title page (`logo: none` always), `cover.credit` is placed on p. iv only if filled in.
