# Session prompt — TRK-DESIGN-006 + TRK-DESIGN-009 (chapter images + poem styling)

> Use this as the kick-off prompt for a fresh Claude Code session.
> Combined because both touch `manuscripts/ghosts/*.typ` chapter files and
> `typesetting/templates/series-template.typ` — they would merge-conflict
> if split.
>
> **Sequencing:** must run AFTER `NEXT_SESSION_PROMPT_2026-05-30.md`
> (DESIGN-005 + 007 + 008) lands. DESIGN-005 fixes the import path so
> direct `typst compile` works without a staging-tree workaround — this
> session relies on that.
>
> **Concurrency:** safe to run in parallel with TRK-DEV-013 (that touches
> `srv/epub.go`, zero overlap with this session's zone).

Run `jpull` first. Then pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  git -C /home/exedev/prodcal log --oneline -1 && \
  ls -la /home/exedev/prodcal/manuscripts/ghosts/*.jpg | head -10 && \
  typst --version'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `active`, HEAD includes DESIGN-005/007/008 commit (post-2026-05-30 session), 8 chapter JPEGs in `manuscripts/ghosts/`, typst version, HTTP/2 200.

## What you're doing

Two Typst-side fixes from the TEST-002 parity matrix (both filed 2026-05-26). Tackling them together because both edit chapter `.typ` files and `series-template.typ`.

**Tickets (full bodies in `docs/TRACKER.md`):**

1. **TRK-DESIGN-006 (P2, ~30-60 min)** — Chapter opener images not rendering. 8 JPEGs (`ghosts_01_SODA.jpg` ... `ghosts_08_LOYALTY.jpg`) sit next to chapter `.typ` files but the compiled PDF shows chapter openers with only title + byline, no image. Likely fixed by DESIGN-005's import-path correction; verify and shim if not.

2. **TRK-DESIGN-009 (P2, ~45 min)** — Poem/verse blocks render as plain justified body. Two-part fix:
   - **Styling:** `poem()` in `series-template.typ` currently uses code-font (JetBrains Mono). Should be body-font italic, smaller, centered, looser leading.
   - **Wrapping:** chapter `.typ` files don't actually wrap verse content in `#poem[...]`. Need to refactor `01-soda.typ` (Spencer Nitkey, 4-line verse) and any other chapters with embedded verse.

## Implementation order

### Step 1 — TRK-DESIGN-006 (chapter images)

Start by recompiling Ghosts with the post-DESIGN-005 import path to confirm whether images now render — the parity-doc ticket flags this as the first hypothesis.

```bash
ssh exedev@jdbbs.exe.xyz '\
  cd /home/exedev/prodcal && \
  typst compile --root . --font-path typesetting/fonts \
    manuscripts/ghosts/main.typ /tmp/ghosts-img-check.pdf'
scp exedev@jdbbs.exe.xyz:/tmp/ghosts-img-check.pdf scratch/
mkdir -p scratch/diff-img-check
pdftoppm -r 100 -png scratch/ghosts-img-check.pdf scratch/diff-img-check/page
open scratch/diff-img-check/page-008.png  # ch1 opener
```

**If images now render** → close DESIGN-006 as fixed-by-side-effect of DESIGN-005, move on to step 2.

**If still no images** → walk the hypotheses:

1. **Path resolution.** Inspect `main.typ` lines ~84-164 (per parity doc — chapter opener block). Look at the `#image()` calls:
   ```typst
   #image("ghosts_01_SODA.jpg")     // resolves relative to main.typ's dir
   #image("./ghosts_01_SODA.jpg")   // same
   #image("/manuscripts/ghosts/ghosts_01_SODA.jpg")  // absolute via --root
   ```
   The `--root .` flag (set in step 1) means `/foo/bar` paths are rooted at the repo root. If `main.typ` uses absolute-style paths, `--root` must match. If it uses relative paths, they resolve from `main.typ`'s directory.

2. **Background placement covered by text.** Search `main.typ` for `set page(background: ...)` — if the image is set as page background and the text frame is opaque or fills the page, the image is hidden. Switch to `place(top + left)` or `image()` as a foreground block.

3. **File-name case mismatch.** Linux is case-sensitive; `main.typ` might reference `ghosts_01_soda.jpg` but the file is `ghosts_01_SODA.jpg`. Check `ls manuscripts/ghosts/*.jpg` against the strings in `main.typ`.

Once you identify the bite, fix and recompile. Verify all 9 chapter openers visually (pages 8, ~20, ~32, etc. — chapter-opener page-numbers shift as upstream fixes change page-counts).

**Acceptance for DESIGN-006:**
- All 9 chapter-opener pages in `scratch/ghosts-img-check.pdf` show their JPEG.
- Image sizing/placement matches reference (full-bleed or top-half, depending on the design — eyeball against `reference/GHOSTS.pdf`).

### Step 2 — TRK-DESIGN-009 (poem styling + wrapping)

**Part A: fix `poem()` in `series-template.typ`** (~10 min)

Find `poem()` around lines 243-247 (per parity doc). Likely current shape:

```typst
#let poem(body) = {
  set text(font: "JetBrains Mono", size: 0.9em)  // wrong: code-font
  body
}
```

Replace with:

```typst
#let poem(body) = {
  set text(
    font: body-font,            // resolves to Libertinus Serif (or licensed body font)
    size: 0.875em,
    style: "italic",
  )
  set par(
    leading: 0.8em,             // looser than body's 0.6em
    first-line-indent: 0em,     // verse should not indent
    justify: false,
  )
  align(center, body)
}
```

(Adapt to actual `body-font` variable name in the template — could be `body-font`, `bodyfont`, `$body-font`, or just literal `"Libertinus Serif"` depending on how it's parameterized.)

**Part B: wrap verse in chapter `.typ` files** (~20-30 min)

Find verse content in `manuscripts/ghosts/01-soda.typ`. The 4-line poem looks like (per parity doc):

```
Where have I gone? / High waves sighing, /
erasing, placing life on... who? /
It is Thursday and I miss you. /
Without true light, hues grow so dull.
```

Wrap it as:

```typst
#poem[
  Where have I gone? \
  High waves sighing, \
  erasing, placing life on... who? \
  It is Thursday and I miss you. \
  Without true light, hues grow so dull.
]
```

Use Typst's `\` for line breaks inside `poem[...]`. Don't use `/` — that was the manuscript's transcription convention, not Typst syntax.

Then `grep -i "verse\|poem\|stanza" manuscripts/ghosts/*.typ` to catch any other verse blocks across the 9 chapters. If found, wrap each.

**Smoke compile:**

```bash
ssh exedev@jdbbs.exe.xyz '\
  cd /home/exedev/prodcal && git pull --ff-only && \
  typst compile --root . --font-path typesetting/fonts \
    manuscripts/ghosts/main.typ /tmp/ghosts-poem.pdf'
scp exedev@jdbbs.exe.xyz:/tmp/ghosts-poem.pdf scratch/
pdftoppm -r 100 -png scratch/ghosts-poem.pdf scratch/diff-poem/page
```

Verify on the page containing chapter 1's verse (was page ~15 pre-fixes; could be a different page now). Want:
- Verse centered (not justified left)
- Italic body-font (not mono)
- Slightly smaller than surrounding body
- Looser leading between verse lines than between prose lines

### Step 3 — Re-diff vs reference

```bash
mkdir -p scratch/diff-typst-v3
pdftoppm -r 100 -png scratch/ghosts-poem.pdf scratch/diff-typst-v3/page
# Spot-check pages: every chapter opener, the verse page, a typical body page
```

Compare with `scratch/diff-ref/` to confirm chapter openers and verse now match.

## Acceptance

- All 9 chapter openers display their image.
- The chapter-1 verse renders as centered italic body-font, visibly distinct from surrounding prose.
- `poem()` styling change does not break Twitter Years (no verse, no chapter images — should be byte-identical or page-count-stable).
- `docs/GHOSTS_PARITY_2026-05-26.md` updated: DESIGN-006 + 009 closed.
- `docs/TRACKER.md` updated: close DESIGN-006 + 009; update Resume here; note remaining DESIGN-001 children (only DEV-013 if that hasn't landed yet).
- Page-count delta narrows further (was 101 vs 136 pre-005/007/008; chapter images should add ~9 pages or shift breaks significantly).

## Then

After this session lands, the only remaining ❌ from TEST-002 is:

- **TRK-DEV-013 (P3, ~30-60 min)** — epubcheck PKG-005 (mimetype extra-field). Different code zone (`srv/epub.go`); see separate prompt at `docs/NEXT_SESSION_PROMPT_DEV-013.md`.

Once DEV-013 lands, **TRK-DESIGN-001 closes** with zero ❌ in the parity matrix — Ghosts-like titles are release-confident on visual quality.

Parallel queues that haven't moved:
- TRK-DESIGN-003 — smart-punctuation conversion + ragged-vs-justified audit.
- TRK-DEV-012 Phase C — chapter auto-detection on upload (deferred).

## Wrap-up

1. Single combined commit (both tickets touch related files; cohesive):
   ```
   TRK-DESIGN-006 + 009: chapter images + poem styling
   ```
2. Push to main.
3. VM pull (no rebuild needed — both changes are in templates/manuscripts, not Go code):
   ```bash
   ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only'
   ```
4. Update `docs/TRACKER.md`: close DESIGN-006 + 009, update Resume here.
5. Update `docs/GHOSTS_PARITY_2026-05-26.md`: flip these cells to ✅.

## Non-goals

- **Don't touch the EPUB pipeline.** EPUB-side has the same poem gap but that needs pandoc class-markup work — separate ticket.
- **Don't auto-detect verse from markdown.** The parity doc flags this as a follow-up (Lua filter work) — out of scope here.
- **Don't restructure `main.typ`.** If DESIGN-006 needs a substantial refactor (not just a path or placement fix), file a follow-up rather than yak-shaving.
- **Don't deploy a new binary.** `series-template.typ` is read at compile-time from disk; VM `git pull` is sufficient.

## Pitfalls

- **`poem()` may not exist in `series-template.typ` yet** — if the parity-doc line-number reference is stale, grep first: `grep -n "^#let poem\|^let poem" typesetting/templates/series-template.typ`. If absent, add it.
- **`body-font` variable name unknown** — inspect the template's existing `#set text()` calls to find the variable name in use. If hard-coded as `"Libertinus Serif"`, match that.
- **Twitter Years regression** — `series-template.typ` changes affect both books. After the `poem()` fix, recompile Twitter Years and confirm output is unchanged (or only changed in ways `poem()` modifications would explain — i.e., nothing, since Twitter Years has no verse).
- **Chapter-opener page numbers shift** — once images render, each chapter opener may push subsequent content one page. Don't hardcode page numbers; visually identify openers by content.
- **Image aspect-ratio mismatch** — if the JPEGs are landscape and the page is portrait, Typst will scale. Match reference's placement (full-bleed vs framed) by inspecting `main.typ`'s `#image()` `width:`/`height:` args.
