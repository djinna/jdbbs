# Session prompt — 2026-06-02: bring the Ghosts print production flow online

You're continuing the **Ghosts print-PDF golden-parity** work. The goal **today** is to take
the proven-on-one-chapter pipeline and finish it end-to-end: every chapter + front matter at
golden parity, compiled on the VM with the licensed fonts, producing a **shippable print PDF**.
That readies the *fiction production flow* for the next title (Zoothesia, same layout).

Target to match: `reference/GHOSTS.pdf` (InDesign golden, 136pp).
The single source of design truth is `typesetting/templates/series-template.typ`; the book is
assembled in `manuscripts/ghosts/main.typ`.

---

## Orientation (read first)

- **Local compile loop (~0.4s):**
  `typst compile --root . --font-path typesetting/fonts manuscripts/ghosts/main.typ scratch/x.pdf`
- **Throwaway files → `scratch/`** (gitignored). Last night's closeups/review PDFs are in
  `scratch/opener/`. Never `/tmp`, never `tmp/` inside the source tree.
- **VM ssh is USER-RUN, not assistant-run.** Hand the user one concatenated
  `ssh exedev@jdbbs.exe.xyz '…'` command and work from what they paste back. VM repo path is
  still legacy `/home/exedev/prodcal/`. The doc pipeline (typst/pandoc/python-docx) lives on the VM.
- **Commit only when asked.** Trailer: `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`.
- Use the gstack `/browse` skill for any web; never `mcp__claude-in-chrome__*`.

---

## What landed last night (commit `a99ea70`, pushed to main)

Built the reusable **chapter-opener machinery** in `series-template.typ` and matched it to the
golden "Garden of Eden" chapter (the standard all chapters now adopt):

- **`chapter-opener(title:, author:, art:)`** — opener recto: 130pt square hero (upper-left at
  inside margin), ragged-left title (Proxima bold 20pt in a 78pt column), author below. No
  running head, no folio. Measured to golden p41 (trim coords).
- **`chapter-body-start(title:, author:)`** — blank verso, then body sinks 96pt, centered drop
  folio at the foot, no running head. New `drop-folio-pages` footer state mirrors
  `suppress-header-pages`. Measured to golden p43.
- **Running-head sides fixed template-wide** — now keyed off `calc.even(here().page())`
  (physical parity), not folio parity. Our body folio resets to 1, so its parity drifts from the
  physical page and was flipping every verso/recto head.
- **Running-head typography matched to the golden's RENDERED output** (see gotcha below).

Garden is wired in `main.typ`; title/author `#text` lines were stripped from `03-garden.typ`.
**The other 8 chapters still use the OLD full-page-image opener** — that's job #1 today.

---

## Today's work, in order

### 1. Roll the golden opener out to the other 8 chapters  *(~45 min)*
For each chapter below, in `main.typ` replace the old block
(`#set-story-info(none,none)` / `#pagebreak(to:"odd")` / `#image("…", width:100%)` /
`#pagebreak()` / `#no-header()` / `#set-story-info(…)`) with:

```typst
#set-story-info(title: none, author: none)
#chapter-opener(title: "<Title>", author: "<Author>", art: "/manuscripts/ghosts/<file>")
#chapter-body-start(title: "<Title>", author: "<Author>")
#include "<NN-name>.typ"
```

Then strip that chapter file's own title/author `#text` lines at the top (as done in
`03-garden.typ`). **Use root-relative art paths** (`/manuscripts/…`) so `image()` resolves
regardless of the calling file.

| Ch | file | title | author | art |
|----|------|-------|--------|-----|
| 1 | `01-soda.typ` | Soda Sweet as Blood | Spencer Nitkey | `ghosts_01_SODA.jpg` |
| 2 | `02-lifetime.typ` | In Every Lifetime | Lara Dal Molin | `ghosts_02_EVERY_LIFETIME.jpg` |
| 4 | `04-tools.typ` | We Shape Our Tools and Then Our Tools Shape Us | Tongzhou Yu | `ghosts_04_WE_SHAPE.jpg` |
| 5 | `05-house.typ` | The House That Paid Its Own Bills | Elizabeth Maher | `ghosts_05_HOUSE.jpg` |
| 6 | `06-latency.typ` | Latency | Rafael Fernández | `ghosts_06_LATENCY.jpg` |
| 7 | `07-genius.typ` | Genius in the Bottle | Claire Pichelin | `ghosts_07_GENIUS.jpg` |
| 8 | `08-loyalty.typ` | Loyalty | Zach Hyman | `ghostts_08_LOYALTY.jpg` ⚠️ note the `ghostts` typo |

**Chapter 0 / Khlongs (`00-intro.typ`) is the exception** — its art `ghosts_00_SBAcover.png` is a
**493×203 banner, not a square hero**. Check golden **p9** for the intro opener layout before
forcing it into the square `chapter-opener()`; it likely needs a banner variant (or a separate
opener function). Don't blindly square-crop the banner.

Verify each: titles wrap ragged-left like Garden, openers land recto, blank verso, body drops,
drop folio centered, running heads correct sides. Spot-check 2–3 openers against the golden.

### 2. Front matter (i–vi)  *(~60–90 min)*
Currently `main.typ` has rough half-title / title / copyright / contents. Bring them to golden
parity (pp. i–vi). The **copyright page needs the LC CIP block** generated from metadata.
Design thread + CIP rules (em-dash delimiter caution, space-padded separators, three-fields →
id.loc.gov lookup) are in **`docs/notes/2026-06-01-front-matter-cip-worknotes.md`** — read it first.

### 3. Per-chapter cleanups  *(~30 min)*
- **TRK-DESIGN-009** — poem/verse blocks render as plain body; wrap them with `poem()` in the
  affected chapter files (body-italic, not mono).
- **Khlongs multilingual** — CJK/Thai passage (`林家商店47號…ร้านยามาลิน สาขา ๔๗`) falls back; golden used
  Hira/Thonburi. Fallback chain is in `series-template.typ` `book()` set-text; verify glyphs render.

### 4. Prove the pipeline on the VM → shippable PDF  *(~30 min, user runs ssh)*
Hand the user a single concatenated command to: `git pull --ff-only`, `typst compile` with
`--font-path typesetting/fonts`, then `pdffonts` (expect subsetted **Plantin MT Pro** + **Proxima
Nova** faces, **no substitutions**) and `pdfinfo` (page count → goal ~136, page size 311.81×504.57
trim). `scp` the PDF back to `scratch/` and spot-check key pages vs the golden.

### 5. (Stretch) ready the flow for the next title
Zoothesia is the next fiction title and **reuses the Ghosts layout** — once Ghosts is shippable,
a drop-in test proves the production flow. Tweet Book is a *separate, partial* layout (different
org) — out of scope today.

---

## Gotchas that will bite you (hard-won last night)

- **"10pt ≠ 10pt" (Adobe vs our Proxima cut).** The InDesign panel says running head = Proxima
  Nova Semibold 10pt caps / 9pt folio, but our **bought Proxima OTF renders wider + heavier** than
  Adobe's at the same nominal pt. Running heads are therefore matched to the golden's **rendered**
  output, not the nominal: `running-heads-size: 0.82em` (caps cap-height **5.52pt**),
  `running-heads-folio-size: 0.89em` (folio figure-height **6.0pt** — the folio is *larger* than the
  caps in the render), `weight: 500` (Medium), **no tracking**. Don't "correct" these back to 10/9pt.
- **Measure at 300dpi, not 150.** At 150dpi a ~10pt cap is ~14px tall and threshold-ink
  measurement under-reads by ~1pt — that sent an earlier pass the wrong way. Render comparison
  crops at `-r 300`.
- **Font sizes don't nest.** `text(size: 0.89em)` inside a block that already set `0.82em` =
  0.89×8.2pt, not 8.9pt. Size caps and folio **each relative to the base**, not compounded.
- **Image paths** in the template must be **project-root-relative** (`/manuscripts/…`); `image()`
  resolves relative to the calling file otherwise.
- DESIGN-005 (import path), 007 (ALL CAPS heads), 008 (first-line indent) are **already done** —
  don't re-open them.

## Definition of done (today)
`manuscripts/ghosts/main.typ` compiles clean to ~136pp; all 9 chapters use the golden opener;
front matter i–vi matches golden; VM compile shows licensed fonts with no substitutions; a
side-by-side of the opener, a body page, and the copyright/CIP page reads as a match to
`reference/GHOSTS.pdf`. Then it's a shippable print PDF and the fiction production flow is proven.
