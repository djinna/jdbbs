# Front matter + copyright/CIP worknotes — 2026-06-01

Design thread for the print-side **front matter** (half-title → contents) and a
reusable, metadata-driven **copyright / Publisher's-CIP** page. Captured here
because it currently lives only in a chat session. Parallel to — not part of —
the Ghosts print-parity throughline (see `docs/notes` parity work + the
`ghosts-print-parity-status` memory).

## Targets / first instances
- **First fiction title: "Zoothesia"** — manuscript arriving ~2026-06-02. First real CIP test case (fiction is the easy case for classification + subjects).
- **First nonfiction: the "Tweet Book"** — to be worked after Ghosts print-parity is done.
- Normal front-matter set going forward: **i–vi** (half-title, blank, title, copyright, contents, [blank/start]) then **Arabic 1** for the body.

## The page-split decision (resolves a contradiction in the original plan)
The original plan said both "swap in finished InDesign pages for i–iv" *and*
"generate the copyright page (iv) from book info." Those conflict on iv — a flat
InDesign PDF can't be generated from metadata. Resolution by page nature:

| Page | Nature | Approach |
|---|---|---|
| Frontispiece / image collage | Art | Raster: high-res PNG via `#image()` full-bleed, or part of a swap PDF |
| i–iii (half-title, blank, title) | Type-only | Typst-native is likely ~95% once licensed fonts land; build the swap *mechanism* but may not need it here |
| **iv (copyright/CIP)** | **Data** | **Must be Typst-native + templated** — only way "generate from book info" works |
| v–vi (contents) | Data | Typst-native — Ghosts already has a working TOC (`toc-entry(...)`); make it data-driven |

## Swap mechanism (for the InDesign-authored pages we *do* keep external)
- **Merge, don't embed.** Post-compile PDF merge (`pdfunite` / `qpdf`, poppler-adjacent — already on the VM and the MBP16). Typst compiles the body; InDesign exports front matter as a **vector PDF**; a build step concatenates. Preserves vector type + CMYK + InDesign's embedded fonts. Slots into `build-ghosts.sh`.
- **Avoid `muchpdf`** (Typst PDF-embed package): it rasterizes via pdfium → worse for type-heavy pages.
- **TIFF is unusable** in Typst (`#image()` takes PNG/JPEG/SVG/GIF/WebP only). So "full-page TIFFs" → convert, or better keep vector via merge.
- **Recto/verso parity gotcha (the classic split-assembly bug):** if front matter is merged externally, Typst doesn't know those pages exist. `main.typ` starts body numbering at 1 and uses `#pagebreak(to: "odd")` per opener — prepend N front-matter pages and every should-be-recto opener can flip to verso. Reserve the front-matter page count in the body, or reconcile parity at merge time. Surface before the first merged proof, not after.

## Typst title-page affordances (more than the current `main.typ` uses)
`tracking` (letterspacing), small caps, `line`/rules + ornaments, precise
`place`/`grid` positioning. The real gap on the title pages is font + small
optical nudges, most of which closes once Plantin/Proxima land.

## Copyright / Publisher's-CIP (P-CIP) page — the meaty one
Reusable, metadata-driven Typst template. Format is **ISBD** in a labeled
display; the punctuation/spacing is load-bearing (see encoding rules below).

**Hard rule:** the block header must read **"Publisher's Cataloging-in-Publication data"** — never "Library of Congress." LC neither creates nor vets these; claiming otherwise misrepresents the record. Anyone may produce a P-CIP block.

**Auto-fillable from our book metadata** (mechanical ISBD fields):
- Title proper ` : ` subtitle ` / ` statement-of-responsibility (transcribed from the *title page*, not the application)
- Description: edition | imprint/series note | `place ; place : publisher, year` | content notes (`Includes index` etc.)
- Identifiers: `ISBN:` (list multiple if present)
- Titles are **sentence case** (capitalize first word + proper nouns only)

**NOT auto-fillable — needs live cataloging judgment, cannot be a generated guess** (look up at **id.loc.gov**):
1. **Authorized name access points** (NAF) — disambiguate same-name authors (middle initial / fuller name / birth year) per RDA.
2. **LCSH subject headings** — authorized headings + subdivisions; fiction takes the `--Fiction` form subdivision. BISAC is a *separate, commercial* taxonomy (BISG list), not LCSH.
3. **Classification** — LCC + DDC. The genuine bottleneck for nonfiction. **Fiction is near-trivial** (`DDC 813.6` American fiction / `LCC PS` + author cutter), and libraries override to `FIC` regardless.

**LCCN / PCN eligibility — verify before encoding any "omit LCCN" logic:**
- The **CIP Program** (official block) is **ineligible** for self-/small publishers (needs ≥3 prior titles by 3 authors widely acquired by US libraries).
- The **PCN / Preassigned Control Number** program gives just the LCCN, free, via PrePub Book Link / Author portal — for US-**published**, not-yet-published **print** books. Ebook-only / already-published → not accepted; **omit the LCCN entirely** (an ISBN-only P-CIP block is fully valid).
- **CAUTION (category error to avoid):** eligibility keys on **place of *publication* (publisher's location), NOT place of *printing* (country of manufacture)**. "Printed in Argentina" does **not** by itself disqualify a US-published book. (Moot for Ghosts — it uses a CC-license block, not CIP — but matters for Zoothesia / the Tweet Book.)

### Encoding rules — addendum from the Claude session (2026-06-01)
These will bite at implementation time; pin them in the template spec:

- **The `--` subdivision delimiter is NOT a literal double-hyphen.** In real
  LCSH/CIP practice the subdivision delimiter is **space-em-dash-space**
  (` — `, U+2014), sometimes plain-ASCII `--` as a stand-in. The Rare Bird
  examples render it as what looks like two en-dashes glued together — a
  font/copy-paste artifact. **Pick one canonical internal representation**
  (store ` -- ` ASCII, render to ` — ` U+2014 on output) rather than matching
  the visual exactly, or you'll get Unicode inconsistency across blocks.
- **Spaces in the separators are significant.** The ` | ` field separator and
  the ISBD ` : ` / ` / ` / ` ; ` are all space-padded; the padding is part of
  the grammar. Easy to lose to a `trim()` or a naive `join`.
- **Three fields can't be solved in code.** Names / LCSH / Classification need
  live lookups against **id.loc.gov**, so however the harness is structured,
  those three should be a **human-or-model judgment step with the authority
  file open**, not a generated guess. The template fills itself; those three
  don't.
- Offer stands: formalize the ISBD punctuation grammar + field-ordering rules
  further if implementation needs it.

### Reusable template skeleton
```
Publisher's Cataloging-in-Publication data

Names: [Surname, Forename], [author|editor|illustrator…] | [Surname, Forename], [role].
Title: [title proper] : [subtitle] / [statement of responsibility exactly as on the title page].
Description: [edition statement]. | [Imprint/series note]. | [Place, ST] : [Publisher], [year]. | [Includes index. / Includes bibliographical references and index.]
Identifiers: LCCN: [only if a PCN was obtained] | ISBN: [paperback] | ISBN: [hardcover] | ISBN: [ebook]
Subjects: LCSH: [Heading -- Subdivision -- Subdivision]. | [Heading -- Subdivision]. | BISAC: [CATEGORY / Subcategory].
Classification: LCC [class number] | DDC [number]-dc23
```
Drop any field that doesn't apply (no index → no index note; no series → no series line; no PCN → no LCCN).
