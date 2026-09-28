# Snowmoon (Vitalik Buterin) — can the factory ingest it? Assessment, 2026-09-28 11:49 (iteration 1 — superseded by `-1230`)

Source: https://vitalik.eth.limo/snowmoon/ (index = front matter; 32 chapter
pages under `html/chapter-N.html`). Nothing was ingested; the 33 HTML files are
parked in `scratch/snowmoon/` (gitignored) for the analysis below.

## Verdict in three lines

1. **The prose maps cleanly** onto the 13 factory styles + option-C custom
   styles (`docs/reviews/CUSTOM-STYLES-MARKERS-2026-09-22.md`). HTML is just
   another pandoc reader, so the route is the one already decided for Markdown:
   *convert to a template-conformant `.docx`, then the normal pipeline* — not
   "teach five `.docx`-only stages a new format" (`WORD-FREE-AUTHORING-2026-09-19.md` §1).
2. **It is the best test case we have seen for the style-absorption idea**: the
   page's CSS is literally a stylesheet with eight classes, and the dialogue is
   coloured *per fictional character* — 22 hues, 2,541 spans — which is exactly
   the "semantics disguised as presentation" that option C is designed to keep
   (as a named character style) while dropping the look.
3. **Two things do not fit any factory style and dominate the book**: 287
   `device-view` panels (in-world screens: tables, sliders, buttons, emoji,
   56 SVG game boards) and 8 `dz-card` conlang placards in a calligraphic font.
   Under option C these are "admin-designed snippets" — but at ~9 per chapter
   they are the novel's central device, not an edge case. That is the real
   design job, and it is Jenna's, not the pipeline's.

Also: it is not a short novella. **~102,800 words** across 32 chapters
(1,500–4,400 each; 4,572 of those words are inside device panels). Roughly
330–360 pp at trade-paperback density before the panels are sized.

## What the source is made of

One inline `<style>` per page (core + a "veridia plugin" block), no external
CSS, no images (`<img>` = 0), no `<h2>`, chapter titles are just "Chapter N".

| Source construct | Count | Factory mapping | Fit |
|---|---|---|---|
| Index page: title, chapter list, License (GPL v3 + a paragraph of legal theory), AI-usage declaration, 爱-usage declaration | 1 page | Front matter. Title/author → spec (author is implicit from the domain, not on the page). Chapter list → generated Contents. The three declarations → Copyright page (`Copyright` style) or a `[[preface]]`-tier front piece; the licence text is unusual (GPL, not CC) and should be carried verbatim. No dedication, epigraph, subtitle, ISBN, © year. | Good |
| `<h1>Chapter N</h1>` | 32 | Heading 1 | Exact |
| `div.dateline.chapter-open` — *Place, Country · 3724 Snowmoon 3* under each H1 | 32 | Custom paragraph style **Dateline**, based on Heading 3 (or Epigraph). Structure only; the factory sets it. | Good — option C as designed |
| `div.dateline.scene-break` — a date (sometimes place) mid-chapter | 12 | Custom style **Scene Dateline**, based on Section Break (carries text — Section Break today is an ornament; needs the base to accept content, or base on Heading 3 with space-before) | Fair — small template gap |
| `<hr>` (plain) | 73 | Section Break | Exact |
| `<p>` prose | ~3,000 | Normal / First Paragraph (first-after-heading/break rule already in the filter) | Exact |
| `<em>` / `<b>` | 125 / 13 | native italic / bold | Exact |
| `<blockquote>` — song verses in caps, two lines | 30 | Verse (not Block Quote — they are lyrics). Note the source is malformed: first line is bare text, second is a `<p>`; pandoc copes. | Good |
| `<ul>` | 11 | pandoc lists → Typst lists; factory has no list style today | Fair |
| `<sup>` (some nested `<sup><sup>`) | 10 | native superscript; nested ones are inside device panels | Fine |
| **`<span style="color:oklch(0.7 0.15 H)">…</span>` around dialogue** | **2,541 spans, 22 hues** | One **character** custom style per hue (`speaker-h93`, `speaker-h347`, …; the author can rename to Zei/Fin/Bai). Based on `none` → prints as plain roman; EPUB keeps a class so colour can survive on screen. This is exactly option C's rule — *the author owns what a thing is, the factory owns how it looks* — but 22 declaration rows by hand is too many: an HTML importer must **derive the declarations from the stylesheet** and offer them for naming. | **Good — and the case that justifies the importer** |
| `div.device-view` (`wide-`/`narrow-`) — dark monospace panels holding `<table>` (69, all inside panels), `<input type=range>` (3), `<button>` (26), `<center>` (64), 200%-size emoji rows, and 56 inline `<svg>` boards | **287** | No factory style. Option C: admin-designed snippet — a boxed monospace Typst block (`#device-panel[...]`, wide/narrow). Tables inside are fine; sliders/buttons need a static print convention (`[ Select ]`, a glyph slider); SVGs go through pandoc as images → Typst `image()` (rasterise or keep SVG; Typst renders SVG). One blinking `<b style="animation: blinker…">` must be flattened. | **Weak — the design job** |
| `div.dz-card.dz-mono` — Dzegoban text-as-object, TeX Gyre Chorus (free, GUST licence), per-word colour chips and `white-space:nowrap` groupings | 8 | Custom style based on Code Block or Verse + a designed snippet for the calligraphic face; add TeX Gyre Chorus to `typesetting/fonts`. Colour chips → drop in print, keep in EPUB. | Fair |
| `nav.chapter-nav`, dark-mode toggle, `wc-exclude` | chrome | strip | n/a |

## How this maps to the style-absorption plan

What we have today (all verified in the 19 Sep note): pandoc's Markdown path
(d2) produces byte-identical Typst via `pandoc -t docx --reference-doc=<project
template>`, and pandoc's docx writer honours `custom-style="Name"` on Divs and
Spans. pandoc's **HTML reader** yields Divs with *classes* and Spans with
*style attributes*, not `custom-style` — so the missing piece is one small
Lua filter (or python pre-pass) between them:

```
HTML pages ─► pandoc html reader ─► [absorb-stylesheet.lua] ─► pandoc docx writer
                                          │  class → custom-style             (reference-doc = project template,
                                          │  inline colour → character style   declared styles present)
                                          │  chrome stripped
                                          ▼
                              proposed transmittal Custom Styles rows  ─►  normal pipeline
                              (name · based on · purpose · count · sample)  (Inspect, book map, markers,
                                                                              print, EPUB — untouched)
```

The "absorb" step is the same for any source with a stylesheet (HTML/CSS,
InDesign-exported HTML, Scrivener, Substack/Ghost exports): **read the class and
inline-style inventory, propose one declaration row per distinct class/colour
with counts and a sample, let the author pick the parent and name it.** The
transmittal already has the declaration UI; Inspect already has the
`observed_style`/`undeclared_custom_style` vocabulary. Snowmoon exercises every
cell of that grid:

- paragraph class → paragraph custom style (datelines);
- inline colour → character custom style with base `none` (speakers) — proves
  the `none` base earns its place;
- container class that no parent expresses → "renders as its parent until the
  studio designs it, and the transmittal says so" (device panels, dz-cards).

What it does *not* map to: honouring the CSS typography (option A/B were
rejected on 22 Sep and this file confirms why — the web face is system-UI on a
violet gradient, the panel colours are screen-only, `oklch()` is meaningless
on a one-colour interior).

## Effort, if we did it

| Piece | Size | Notes |
|---|---|---|
| Fetch + stitch 32 chapters, strip chrome, build front matter from index | S | Shell/python; front matter authored by hand from the three declarations |
| `absorb-stylesheet` pre-pass: class/colour inventory → proposed declarations + `custom-style` rewrite → `.docx` via reference-doc | S–M | The reusable part; generalises to any HTML source. Keep it a script first (`scripts/html-to-template-docx.py`), decide later whether upload accepts `.html`/`.zip` (same open question as `.md`, decision 4 of 20 Sep: CLI/API only) |
| Section Break that carries text (scene datelines) | S | Template + lua + epub css |
| Character style base `none` end-to-end (Typst identity fn, EPUB class) | S | Already decided, check it is wired |
| **Device-panel design**: boxed mono block, wide/narrow, table styling, slider/button glyph conventions, SVG sizing, 287 instances | **M, design-led** | Jenna's call. Without it the book "works" but 4,500 words of screens read as code blocks. Also a mono face question — Courier New in the source; factory mono face + a display face for Dzegoban (TeX Gyre Chorus, free) |
| Licence page for GPL v3 text | S | Copyright style; verbatim |

Total without the panel design: about a day, all of it reusable. With it: a
studio design pass first.

## Recommendation

- Treat Snowmoon as the **fixture** for the stylesheet-absorption importer, not
  as a book to rush through: it has the richest class/colour inventory we've
  seen, is GPL-licensed (we may transform it, and must publish the pipeline we
  use — which we do anyway), and is long enough to exercise the book map,
  running heads and Contents at scale.
- Build the absorb step as a script on the Markdown-(d2) pattern; do not add an
  HTML upload path yet.
- Put "device panel" and "placard" on the designed-presets list next to
  `tweet-block` / `ascii-block` — same family (text-as-object) and the same
  pricing logic (studio design time).
- Open questions for Jenna: (1) is a per-speaker character style worth keeping
  at all in print (roman vs. the web's colour) or should the importer collapse
  the 22 hues to `none` and only the EPUB keep them; (2) device panels — design
  once for this book, or as a factory preset; (3) do we want the GPL condition
  in our own catalogue copy.
