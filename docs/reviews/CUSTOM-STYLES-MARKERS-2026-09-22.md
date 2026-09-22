# Custom styles brought in by `[[markers]]` — design decision (2026-09-22)

Decided in chat with Jenna on workshop day 2, using Toby Shorin's *Devotion*
(project 29, poetry collection, Pages export, every paragraph `Body A`) as the
worked example. **Design only during the freeze; build Thursday 24 Sep.**

## What already exists (built for vgr's first book, spring 2026)

- Transmittal → **Custom Styles**: `name · type (paragraph|character) · purpose`.
  Stored in the book spec (`custom_styles[]`), added to the generated Word
  template, and to the names the `[[marker]]` pre-pass resolves
  (`apply-style-markers.py --declared-styles`).
- Filter emits `#ident[…]`; `bookspecs.go` generates `#let ident(content) = { content }`
  (a plain body paragraph) unless the spec JSON carries a hand-written `preset`
  / `typst` snippet (vgr's `tweet`, `metadata-p`, `metadata-c`, `ascii`). **No UI
  shows the preset or snippet** — that is the part that "got lost".
- Inspect: `style_marker` rows (resolved → low, unresolved → medium, "declare it
  on the transmittal"); `undeclared_custom_style` (high) for Word styles not
  declared.

So the **channel** (declare on the transmittal, mark in the text) was decided
on 18 Sep and works. What was never designed is **what a declared style looks
like**. Toby's 28 `[[verse2]]`/`[[verse3]]`/`[[break2]]` printed literally because
he never declared them; had he declared them they would have silently become
body prose — worse.

## Decision: option C — "based on" a factory style + structural deltas

Three directions were weighed:

| | Verdict |
|---|---|
| A. Typography controls on the declaration (face, size, indent…) | No. Too much UI, and author-chosen sizes/faces never sit right against a set body. |
| B. Honour the paragraph's direct formatting from the manuscript | No. Toby's 508 paragraphs carry the *identical* 18 pt hanging indent — Pages/Docs exports don't carry intent, and in Word it rewards sloppy files. |
| **C. Parent + one or two structural deltas** | **Yes.** BookFactory owns typography (face, size, leading, alignment); the author owns structure (which element, nesting level, break weight). |

Rule of thumb for Inspect wording and for us: *the factory fixes anything about
how a thing looks; the author fixes anything about what a thing is.* A stanza
gap that vanished was ours; an undeclared marker is theirs.

### The declaration row (transmittal, Custom Styles)

`Name · Based on · [Indent] · [Space before] · Purpose`

- **Based on** — required; one of the 13 factory paragraph styles (Normal, First
  Paragraph, Heading 1–3, Block Quote, Epigraph, Verse, Code Block, Section
  Break, Copyright, Signature, Glossary Entry). Default Normal, shown as such.
- **Indent** — offered only for Verse, Block Quote, Code Block:
  `none · one level (1.5 em) · two levels (3 em)`. Values are printed in the UI
  (no explanation of em). Relative to the parent's own indent.
- **Space before** — offered only for Section Break:
  `as usual · 3× the stanza gap · 6× the stanza gap`. Whatever the normal
  space after a stanza is, these are multiples of it.
- No other deltas. Anything C cannot express (a boxed sidebar, a tweet that
  looks like nothing else) stays an admin-designed snippet — Jenna's design
  time, priced like an add-on; until she writes it the style renders as its
  parent, and the transmittal *says so*.

Toby's book under C: `verse2` = based on Verse, one level; `verse3` = two
levels; `break2` = based on Section Break, 3×. Zero new vocabulary.

### Generated Typst

`#let verse2(c) = poem(indent: 1, c)` — the factory functions gain the
parameters (`poem(indent:)`, `blockquote(indent:)`, `code-block(indent:)`,
`section-break(gap:)`). Rule of precedence: **a hand-written `typst` snippet
wins; else parent + deltas; else parent alone.** EPUB: `class="verse verse-indent-1"`
etc. in `docx-to-epub.lua` + `epub-styles.css`. Word template: declared styles
get `w:basedOn` the parent and the indent, so Word/Pages authors applying the
style directly see roughly the right thing while drafting.

### Character styles — inline markers (new, decided)

Jenna's lead client's books need character styles that are not optional
(`[[code]]` for inline computer text, small caps, book titles…). The marker
grammar grows an **inline form**: an opener that is *not* at the start of a
paragraph, or a closer that is *not* at the end, marks a run:

```
Type [[code]]ls -la[[/code]] to list the directory.
The [[sc]]NATO[[/sc]] alphabet …
```

- Paragraph form (opener at start, no closer or closer at a paragraph end) is
  unchanged. `[[code]]` at paragraph start still means Code Block.
- Resolves against the factory character styles the filter already maps
  (`char_style_map` in `docx-to-typst-enhanced.lua`: Small Caps/`sc`, Book Title,
  Article Title, Foreign, Tracked, Bold-Ital, Code) plus declared `character`
  custom styles. A declared character style has **Based on** ∈ {italic, small
  caps, code, none} — no other deltas.
- Pre-pass: split the run, apply the character style, strip the markers;
  unresolved stay literal (visible in the proof) and Inspect lists them.

### Inspect wording (tone: helpful, not scolding)

Unresolved marker, today: "Not a factory style and not declared on the
transmittal — will print as written. Use one of: …". Proposed:

> You've used `[[verse2]]` in 21 paragraphs — we don't know that one yet. Add it
> under **Custom styles** on your transmittal (name it `verse2`, based on Verse,
> indent one level) and it will carry straight through. Until then it prints as
> written so you can see where it is.

A general tone pass over every Inspect / build-error string is a separate
§5 row.

### Surfacing what vgr's book built (transmittal + admin)

- Transmittal: each custom style shows its **resolved form** — "verse2 → Verse,
  indent one level"; for snippet styles "tweet → designed by the studio
  (preset: tweet-block)". Read-only for the customer.
- Admin book-spec view: same line plus the `preset` select and `typst` snippet
  textarea (today JSON-only). How much of this the customer may touch is a
  **separate future pass**.

### Verse itself — fixed today (1d08665)

Centred italic 0.75 em was wrong for any poem. `#poem` is now roman,
left-aligned, 1.5 em left pad, 1.5 em hanging indent for run-overs; EPUB
`.verse` drops the monospace face. Size (`poem_size_pt`, default 7.5 on 10)
untouched — revisit when a collection is actually in flow. 5.23 (collection
preset: poem-per-page, lighter title opener) stays parked; poem titles are the
author's Heading 1 (Toby's first attempt predates his understanding of style
sheets — disregard).

## Build list (Thu 24 Sep, after the freeze)

1. Spec + transmittal: `based_on`, `indent` (0/1/2), `space_before` (1/3/6) on
   `custom_styles[]`; row UI with parent-aware deltas and printed values;
   resolved-form line. Migration-free (JSON), but `typochoices`-style
   normalisation in `bookspecs.go`.
2. `bookspecs.go`: generate `#let ident(c) = parent(indent: n, c)` /
   `section-break(gap: n)`; snippet precedence; tests.
3. `series-template.typ`: `poem(indent:)`, `blockquote(indent:)`,
   `code-block(indent:)`, `section-break(gap:)`; sampler page.
4. EPUB filter + CSS classes; Word template `basedOn` + indent.
5. Marker pre-pass: inline character form; `detect-edge-cases.py` in step;
   friendlier unresolved wording; `TestApplyStyleMarkersScript` cases.
6. Admin book-spec: preset select + snippet textarea; transmittal resolved line.
7. Docs: `/factory#bring`, `/workshop#bring`, WORD-FREE-AUTHORING note update.

Estimate: one day. Verify on mcheck with `scratch/verse/in.docx` after
declaring `verse2`, `verse3`, `break2` on project 17's transmittal.

## Open

- Does the Word template also need the inline character styles as real Word
  character styles (yes, cheap) — and should Inspect suggest `[[code]]…[[/code]]`
  when it sees a monospace run (it already suggests the paragraph form)?
- `break2` at 6× — confirm on a real page that 6 stanza gaps does not read as a
  page break.
