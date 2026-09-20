# Word-free authoring for the book factory — decision note

2026-09-19. Punch-list item 0.25. Question from Jenna: does the factory still
need Word at all? Friday's testers had no Word; Pages renamed styles on export,
Word for the web felt janky. This note grounds the answer in the pipeline code
and in experiments run today on the VM (`scratch/wordfree/`, nothing built,
no credits spent). Item 0.6 listed workarounds; this is the product question.

## 1. What the pipeline actually needs from the input

Short version: **the factory does not need Word. It needs a `.docx` whose
paragraphs carry the 13 template style *names*.** Everything else is optional.

The contract, stage by stage:

- **Upload** (`srv/books.go` `handleUploadBook`; `srv/static/factory.js:1034`)
  accepts one file; the browser side refuses anything not named `*.docx`. The
  server does not sniff the file type — a bad file fails later in pandoc.
- **Inspect** (`srv/preflight.go` → `typesetting/scripts/detect-edge-cases.py`)
  opens the file with python-docx and reads `paragraph.style.name`. Names in
  `BUILTIN_PARAGRAPH_STYLES` (Word defaults + the 13 factory styles) are
  quiet; any other name becomes an `observed_style` and then an
  `undeclared_custom_style` (**high**) in `appendUndeclaredStyleWarnings`
  unless the transmittal declared it. This is exactly what "broke Inspect" for
  the Pages export: a renamed style is, to Inspect, an undeclared custom one.
  Inspect also reports `[[style]]` markers as `style_marker` (see below).
- **Book map** (`srv/bookmap.go` `readDOCXParagraphs`) parses
  `word/document.xml` + `word/styles.xml` directly and keys on the normalised
  `w:name` (`heading1`, `title`, …). It decides front/body/back matter from
  Heading 1 text and finds dedication/epigraph pages from hard page breaks —
  two signals pandoc discards, which is why it reads the XML itself.
- **Style markers pre-pass** (`srv/stylemarkers.go` →
  `typesetting/scripts/apply-style-markers.py`): before pandoc, `[[name]]` at
  the start of a paragraph is resolved (case/space-insensitive) against the 13
  factory names, an alias table (`quote`→Block Quote, `code`→Code Block,
  `poem`→Verse, `h1`…) and the transmittal's declared styles; the paragraph's
  style is set and the marker removed. Unresolved markers stay in the proof.
  python-docx again, so `.docx` only. Runs in the build, **not** in Inspect.
- **Corrections** (`srv/corrections_apply.go` → `apply-corrections-docx.py`)
  and **chapter detection** (`srv/chapter_detect.go`, `pandoc --from=docx+styles
  -t json`) are both `.docx`-specific.
- **Print PDF** (`srv/books.go:676`): `pandoc --from=docx+styles … --lua-filter
  docx-to-typst-enhanced.lua -t typst`. `+styles` wraps every paragraph in a
  Div with `custom-style="<Word style name>"`; the Lua filter
  (`typesetting/filters/docx-to-typst-enhanced.lua` `para_style_map`) maps
  normalised names to Typst calls (`firstparagraph`→`#first-para`,
  `epigraph`, `verse`→`#poem`, `codeblock`, `sectionbreak`, `copyright`,
  `signature`, `glossaryentry`, plus spec-declared styles). Heading 1–3 and
  Block Quote are not in the map: pandoc's docx reader turns them into native
  headers/blockquotes *by style name*. Unmapped names (Normal, Body Text) are
  unwrapped to plain paragraphs. So a style **name** is the whole signal;
  fonts, sizes and direct formatting in the file are ignored.
- **EPUB** (`srv/epub.go`): pandoc reads the same `.docx`, plus
  `epubScanDOCXText` for font-subsetting — `.docx`-specific too.

The template itself (`typesetting/scripts/generate-word-template.py`
`FACTORY_STYLES`) ships exactly: Normal, First Paragraph, Heading 1/2/3,
Block Quote, Epigraph, Verse, Code Block, Section Break, Copyright, Signature,
Glossary Entry. Anything that emits a `.docx` using those names is a
first-class input. Anything that emits a different container (Markdown, ODT,
HTML) hits five `.docx`-only stages, so the cheap way in is always "convert to
a template-conformant `.docx` first", never "teach each stage a new format".

## 2. Candidate paths

Effort is ours (S = an afternoon, M = a few days, L = weeks). "Pipeline
change" means code under `srv/`, `scripts/`, `typesetting/`.

**(a) Google Docs + `[[style]]` markers — already works end to end.**
Author writes in Docs with Heading 1/2/3, types `[[quote]]`, `[[code]]`,
`[[verse]]`, `[[first paragraph]]`, `[[signature]]…[[/signature]]` etc.,
File → Download → .docx, uploads. Verified today: a Docs-shaped file (Word
built-ins only, 12 markers) produced **byte-identical Typst** to the Word
baseline after the pre-pass, and Inspect via the live API on the zoo project
(book 38) came back 0 high / 12 `style_marker` / correct book map (Foreword
front, chapter body, Glossary back). Fidelity risk: low for paragraph styles;
no inline (character) styles — italics/bold come through natively, small caps
or `booktitle` spans do not. Support burden: low, but authors never *see* the
styles; the proof is the only preview. Docs still has no native custom
paragraph styles (community threads still asking in Jan 2026; add-ons write
direct formatting, not named styles), so markers remain the right mechanism.
Effort: **S** (documentation only: a one-page "Docs path" with the marker
list; the factory page already has the short version). No pipeline change.
One nit for later: Inspect runs on the un-marked file, so the book map in
Inspect can misfile a `[[epigraph]]` page as "dedication" while the build
(markers first, then map) gets it right.

**(b) Apple Pages export + a normaliser.** I could not run Pages here, and
the tester's exported file is not in the DB (Friday's uploads were Word for
the web and Word), so the exact renaming is unverified — Pages' own defaults
are Title/Heading/Heading 2/Heading 3/Body/Caption, and "Heading" for
"Heading 1" is the likely culprit. Simulated it: renaming Heading 1→Heading
and Normal→Body gives **zero chapters** in the Typst and two high Inspect
findings. A 30-line python-docx normaliser that maps foreign names back
(`scratch/wordfree/normalize_styles.py`) restored byte-identical Typst.
Fidelity risk: medium — we'd be guessing Pages' mapping until we see one real
file, and Pages may also drop custom styles it doesn't recognise. Support:
medium (Mac only; every Pages release can move). Effort: **S** for the
normaliser once we have one Pages file to key it on; it would sit next to
`apply-style-markers.py` as a second pre-pass (pipeline change, small).
Honest read: worth doing only if Pages users actually show up; ask the Friday
tester for the `.docx` first.

**(c) LibreOffice Writer with our template.** Verified: opening the
template-based manuscript in LibreOffice 24.2 (headless, on the VM), saving
as ODT, then back to `.docx`, preserved all 13 style names; pandoc output was
**byte-identical**; Inspect found one low `direct_spacing`. `soffice
--convert-to ott template.docx` also works, so we could serve an `.ott` — but
it is unnecessary: Writer opens the `.docx` template with styles intact
(Styles sidebar, F11). Fidelity: high. Support: low-medium (free, offline,
all platforms; UI is unfamiliar to Mac users). Effort: **S**, documentation
only. No pipeline change.

**(d) Markdown with a small convention.** pandoc's Markdown reader turns
`::: {custom-style="Epigraph"}` fenced divs into the *same* AST the docx
reader produces, so the Lua filter already understands it. Two ways in:
(d1) Markdown straight to Typst — works, but bypasses Inspect, book map,
markers, corrections, chapter detection and the EPUB font scan (all
`.docx`-only); **M–L** to make those format-agnostic, and not worth it.
(d2) Markdown → `pandoc -t docx --reference-doc=<project template>` → the
normal pipeline. Verified today: byte-identical Typst to the Word baseline,
Inspect clean, every downstream stage untouched. Effort: **S–M** — a
`scripts/md-to-template-docx.sh` we run for people is S; accepting `.md` at
upload and converting server-side is M (upload validation, a tiny converter,
the conventions page). Fidelity: high for prose; images and footnotes need
testing. Support: low for technical authors, hopeless for the rest. This is
the right path for the API/CLI audience (`docs/API-CLI-RECIPE-2026-09-19.md`),
not for workshop authors.

**(e) A browser editor of our own.** **L**, and I recommend against it now.
We would be rebuilding paragraph-style editing, autosave, images, footnotes,
revision handling and offline behaviour to reach parity with tools authors
already have — and its only job would be to emit the same 13-style `.docx`
(or fenced-div Markdown) that (a), (c) and (d) already produce. A one-person
studio would carry the support forever. Revisit only if a year of factory use
shows authors want to *finish* books inside the studio rather than write them.

**(f) "Paste your text, we style it" heuristics.** The pieces exist
(`detect_heading_lookalikes`, `chapter_detect.go`), and the Friday experience
showed "hand the template and your text to Claude" works too. Fidelity: low
as an automatic path — a wrong chapter break or a missed epigraph is exactly
the damage the template exists to prevent, and the author has no way to see
it before the proof. **M** to do honestly (a review step where the author
confirms the guesses). Better framed as a service Jenna offers ("send me the
text, I'll put it in the template") than as a product feature.

**(g) Keep Word as the recommended path, with documented tiers.** Word
remains the tool where the template is visible, styles are one click and
nothing is renamed. Friday's Word for the web upload (book 21) kept every
template style name — the "janky" was the editor, not the file (it did add
direct spacing to 727/771 paragraphs, which the build ignores). Effort: **S**
(a "No Word? Three ways in" box). No pipeline change.

## 3. Recommendation

**(i) Workshop Mon/Tue 21–22 Sep (freeze on).** Say it plainly: *you do not
need Word; you need a .docx that uses the template's style names.* Offer, in
this order: Word or Word for the web if you have it; LibreOffice Writer (free,
opens the template as-is); Google Docs with `[[style]]` markers (free, already
live, Inspect shows them under "Marked styles"). Say Pages is the one tool to
avoid for now. None of this touches pipeline code — the public `/workshop`
and `/factory` pages live in `jdbbs-public/` and publish on save, and the
factory page already carries the Docs paragraph. If Jenna says yes, the
"three ways in" box from 0.6 is twenty minutes and is copy, not code. Bring
the marker list on one slide; that is the whole Docs path.

**(ii) Next month.** In order of value per effort:
1. Make Docs + markers a documented first-class path (S): a page with the
   full alias table, an example, and the caveat about inline styles; add a
   Docs example to the API recipe. Fix the Inspect nit (markers before book
   map in preflight — one call, mirrors the build).
2. Get one real Pages export; if the mapping is as expected, add the style
   normaliser as a second pre-pass (S) and downgrade Pages from "avoid" to
   "works".
3. Markdown → template `.docx` as a script for the CLI audience (S); decide
   later whether upload should accept `.md` (M).
4. Do **not** start a browser editor.

## 4. Open decisions for Jenna

1. Monday framing: ".docx with our style names" instead of "Word" — yes?
   And the "No Word? Three ways in" copy on `/factory` + `/workshop` (no
   code, publishes on save) — go?
2. Docs + markers as an advertised path (with the inline-style caveat), or
   keep it as a documented workaround?
3. Pages: ask Friday's tester for the exported `.docx` so we can build the
   normaliser on a real file — or drop Pages support outright?
4. Markdown: CLI/API audience only, or should the factory page accept `.md`?
5. Confirm: no browser editor this year.

---
Evidence: `scratch/wordfree/` — `ms-word.docx` (13-style baseline),
`ms-pages-sim.docx`/`ms-pages-fixed.docx`, `lo/rt/ms-lo.docx`,
`ms.md`/`ms-from-md.docx`, `ms-docs-markers.docx`/`markers.json`,
`api-inspect-docs.json` (zoo book 38), `styles_of.py`, `normalize_styles.py`.

---

## 5. Outcome (2026-09-20) — what shipped, what is left

Jenna's answers (punch list 0.25 thread, 19 Sep): the fragile step is moving
*style definitions* between documents, not the editor. So the advertised
contract shrank to what the pipeline actually needs and what every editor
has: **Heading 1** for chapter titles, **Heading 2** for sub-heads, a plain-text
`[[quote]]` / `[[verse]]` / `[[epigraph]]` / `[[code]]` / `[[break]]` /
`[[signature]]` marker at the start of a special paragraph (`[[/quote]]` to
close a run), then export `.docx` and run Inspect (free). A built-ins-only
file with markers builds byte-identical Typst to the templated file.

Shipped (all live, in `jdbbs-public` + server):
- `/factory#bring` — "Bring your manuscript — from any editor": the four
  steps, the marker list (only names the pre-pass resolves), export
  instructions per editor, small type for template `.docx`/`.odt`, the
  book-a-week case (house style guide + API intake).
- `/workshop#bring` — "What to bring — and what you don't need".
- Transmittal handoff card leads with "Already have a draft? Heading 1 /
  Heading 2 / [[quote]]… No template needed."; template offered second.
- `GET /api/projects/{id}/word-template?format=odt` (LibreOffice headless on
  the VM, `docxToODT` in `srv/bookspecs.go`).

Decisions from §4: (1) yes — ".docx with our style names", not "Word";
(2) Docs + markers is the advertised path; (4) Markdown stays CLI/API-only;
(5) no browser editor. Still open: (3) Pages — waiting on a real export from a
tester before writing the style normaliser. Next-month items in §3(ii) stand.
