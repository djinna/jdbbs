# Session prompt — TRK-TEST-002 (Ghosts live visual regression) + TRK-DEV-010

> Use this as the kick-off prompt for a fresh Claude Code session.
> No concurrent sessions assumed — this one's broad enough to touch
> srv/epub.go (DEV-010 fold-in), manuscripts/ghosts/ (compile), and
> docs/ (matrix closure). Run it alone.

Run `jpull` first. Then standard pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  git -C /home/exedev/prodcal log --oneline -1 && \
  ls typesetting/fonts/noto/{CJK-TC,Thai}/*.{otf,ttf} 2>/dev/null | wc -l && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT migration_number FROM migrations ORDER BY migration_number DESC LIMIT 3" && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 "SELECT id, name FROM projects"'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `active`, HEAD at `a02c1f4` or later, 4 Noto font files, migrations `16/15/14`, one project (`7|The Twitter Years: 2007–22`), HTTP/2 200.

## What you're doing

Two tickets folded into one session because they share the same code zone.

**Part A — TRK-DEV-010 (~15 min):** add `--epub-embed-font` flags to the pandoc invocation in `srv/epub.go::handleGenerateEPUB` so EPUBs ship the bundled Noto fonts inside the package. Without this, EPUBs reference Noto via CSS but rely on the reader having Noto installed. Self-contained EPUBs render correctly on any device.

**Part B — TRK-TEST-002 (~2-3 hours):** the live visual regression that closes TRK-DESIGN-001. Compile Ghosts end-to-end (Typst PDF + pandoc EPUB), diff page-by-page against `reference/GHOSTS.{pdf,epub}`, fill in the 10 ⚠️ cells of the parity matrix (`docs/GHOSTS_PARITY_2026-05-26.md`) with ✅ / ❌ / acceptable-deviation. File child tickets for any ❌; do NOT try to fix issues in this session.

Both ticket bodies in `docs/TRACKER.md`.

## Implementation steps

### Part A — TRK-DEV-010 (do this first; small, unblocks Part B)

In `srv/epub.go::handleGenerateEPUB`, find the `exec.Command("pandoc", args...)` invocation (~line 157). Add `--epub-embed-font` args for the bundled Noto fonts before the pandoc call:

```go
// Embed bundled multilingual fonts so EPUBs render correctly on any device
fontPaths := []string{
    filepath.Join(typesetDir, "fonts/noto/CJK-TC/NotoSerifTC-Regular.otf"),
    filepath.Join(typesetDir, "fonts/noto/CJK-TC/NotoSerifTC-Bold.otf"),
    filepath.Join(typesetDir, "fonts/noto/Thai/NotoSerifThai-Regular.ttf"),
    filepath.Join(typesetDir, "fonts/noto/Thai/NotoSerifThai-Bold.ttf"),
}
for _, fp := range fontPaths {
    if _, err := os.Stat(fp); err == nil {
        args = append(args, "--epub-embed-font", fp)
    }
}
```

Where `typesetDir` is whatever resolver the file already uses (search for `typesettingRoot()` or similar — the Typst path already uses it). Skip if a font file is missing (defensive against partial deploys).

Smoke locally: not really possible without Go + a manuscript. Verify via the EPUB compile in Part B.

### Part B — TRK-TEST-002 setup: create a Ghosts project

The Ghosts manuscript lives in the repo (`manuscripts/ghosts/`) with `.md` and `.typ` sources but NOT as a DB project. To exercise the EPUB pipeline end-to-end (admin SPA → spec → compile, the DEV-009 path), create one:

1. Open `https://jdbbs.exe.xyz/admin/` → Projects tab → New Project. Name it something like "Ghosts (TEST-002)". Project will get a new id (likely 14+, since 1-13 were used and 1-6+8-13 were deleted leaving only 7).
2. Move to the new project's Books tab. You need a single DOCX containing all 9 chapters with chapter-style headings. Two options:
   - **(Recommended)** Build a minimal DOCX from the `manuscripts/ghosts/*.md` files. Pandoc can do this: `pandoc *.md -o /tmp/ghosts.docx`. Upload that DOCX.
   - **(Alternative)** Skip the DOCX path; just compile from .md directly via `typesetting/scripts/md2epub.sh` for the EPUB-side test, and `typst compile` for the PDF-side test. Bypasses the SPA but exercises the same Go EPUB-generation code.
3. Go to the Typesetting tab, select the new project. Configure:
   - Trim: Protocolized (4.91 × 7.59 in)
   - Body font: Libertinus Serif (or Plantin if licensed and bundled — likely Libertinus for now)
   - Heading font: Source Sans 3
   - EPUB chapters: populate the `Chapters (anthology bylines)` table with the 9 chapter title + author pairs from `manuscripts/ghosts/main.typ` (look for `set-story-info(title:, author:)` calls). Save.

### Part B — PDF compile + diff

Run the Typst PDF compile against the in-repo Ghosts source directly (faster than going through the SPA, and `main.typ` already orchestrates the 9 chapters with `set-story-info`):

```bash
ssh exedev@jdbbs.exe.xyz '\
  cd /home/exedev/prodcal/manuscripts/ghosts && \
  typst compile --font-path ../../typesetting/fonts main.typ /tmp/ghosts-typst.pdf && \
  ls -la /tmp/ghosts-typst.pdf'
```

If `typst` errors, that's the first finding — file as TRK-DESIGN-NNN child and continue.

Pull both PDFs locally for the diff:

```bash
scp exedev@jdbbs.exe.xyz:/tmp/ghosts-typst.pdf /tmp/
# Reference is in repo
cp ~/jd-projects/jdbbs/reference/GHOSTS.pdf /tmp/ghosts-reference.pdf
```

Page-by-page diff via ImageMagick + pdftoppm:

```bash
mkdir -p /tmp/diff-{typst,ref,delta}
pdftoppm -r 150 /tmp/ghosts-typst.pdf /tmp/diff-typst/page
pdftoppm -r 150 /tmp/ghosts-reference.pdf /tmp/diff-ref/page

# Per-page diff
for i in /tmp/diff-typst/page-*.ppm; do
  base=$(basename "$i" .ppm)
  if [[ -f "/tmp/diff-ref/$base.ppm" ]]; then
    ae=$(compare -metric AE -fuzz 5% "$i" "/tmp/diff-ref/$base.ppm" "/tmp/diff-delta/$base.png" 2>&1)
    echo "$base: AE=$ae"
  fi
done
```

`-fuzz 5%` swallows JPEG-like noise; `AE` is absolute pixel difference. Threshold: a page with `AE < 50000` (out of ~7.5M pixels at 150dpi, ~0.7%) is likely visually equivalent; `AE > 500000` is a real difference worth investigating.

If page counts differ (Typst output != 136 pages), the diff loop will skip orphan pages — that's a finding in itself. Note page-count mismatch first.

### Part B — EPUB compile + diff

Use the admin SPA "Compile EPUB" button on the Ghosts project (this exercises DEV-009's per-chapter injection + DEV-010's font-embedding). Download via the compile-history panel.

```bash
# After download, place at /tmp/ghosts-app.epub
# Reference:
cp ~/jd-projects/jdbbs/reference/GHOSTS.epub /tmp/ghosts-reference.epub
```

EPUB structural validation:

```bash
# Install epubcheck if needed: brew install epubcheck
epubcheck /tmp/ghosts-app.epub 2>&1 | head -20
epubcheck /tmp/ghosts-reference.epub 2>&1 | head -20  # baseline
```

EPUB content diff — focus on per-chapter author preservation (the DEV-009 acceptance):

```bash
mkdir -p /tmp/ghosts-app-unzip /tmp/ghosts-ref-unzip
unzip -o /tmp/ghosts-app.epub -d /tmp/ghosts-app-unzip
unzip -o /tmp/ghosts-reference.epub -d /tmp/ghosts-ref-unzip

# Verify the embedded fonts arrived
find /tmp/ghosts-app-unzip -name '*.otf' -o -name '*.ttf'

# Verify per-chapter author bylines
grep -l 'class="chapter-author"' /tmp/ghosts-app-unzip/**/*.xhtml | wc -l
# Expect: 9 chapter files with author byline

# Compare per-chapter content structure
for f in /tmp/ghosts-app-unzip/EPUB/ch*.xhtml; do
  echo "=== $(basename $f) ==="
  grep -E '(class="chapter-author"|<h1)' "$f" | head -2
done
```

Visual EPUB diff via Calibre `ebook-convert`:

```bash
ebook-convert /tmp/ghosts-app.epub /tmp/ghosts-app-rendered.pdf
ebook-convert /tmp/ghosts-reference.epub /tmp/ghosts-reference-rendered.pdf
# Then same pdftoppm + compare loop as above
```

### Part B — close the parity matrix

For each ⚠️ cell in `docs/GHOSTS_PARITY_2026-05-26.md`, assign ✅ / ❌ / acceptable-deviation based on the diffs. The 10 ⚠️ cells from the audit:

1. Actual rendered PDF dimensions and page breaks
2. Widow/orphan placement (Typst auto vs InDesign hand-tuned)
3. Hyphenation and spacing rivers
4. Running header accuracy (per-chapter author in verso?)
5. PDF color/ICC profile rendering
6. EPUB reflowability: per-chapter author byline survives pandoc?
7. CJK/Thai glyph rendering
8. Body alignment (justified vs ragged-right per TRK-DESIGN-003)
9. Poem/verse block font (currently mono; should be serif italic?)
10. Footnote presence + styling

For each ❌, file a child ticket (TRK-DESIGN-005 onward) with: matrix-cell name, observed-vs-expected, severity, link to a diff image. Do **NOT** try to fix issues in this session — the goal is observation + classification, not implementation. Save the fixes for future tickets.

Drop the populated matrix into `docs/GHOSTS_PARITY_2026-05-26.md` (update the existing rows; add a "verified 2026-05-MM" line at top).

## Acceptance

- TRK-DEV-010: VM `srv/epub.go` includes the embed-font invocations; new EPUB downloads contain the Noto OTF/TTF files inside the package.
- TRK-TEST-002: all 10 ⚠️ cells in the parity doc are ✅ / ❌ / acceptable-deviation; any ❌ has a child ticket filed; v1 release-confidence for Ghosts-like titles is documented (either "ship-ready" or "ship-blocked on N child tickets").
- TRK-DESIGN-001 close: this session's findings let DESIGN-001 flip to `done` if no ❌ remain, or stay `in-progress` with a clear punch list.

## Deploy

```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && \
  go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal && \
  sleep 2 && systemctl is-active prodcal'
curl -sI https://jdbbs.exe.xyz | head -1
```

Don't touch the systemd unit (TRK-OPS-005). Push to `main` yourself.

## Wrap-up

1. Update `docs/GHOSTS_PARITY_2026-05-26.md` with the ⚠️ → ✅/❌ resolutions, evidence (diff metric numbers, child-ticket links).
2. `docs/TRACKER.md`:
   - Close TRK-DEV-010 with the file changes.
   - Close TRK-TEST-002 with the matrix-close summary.
   - If no ❌ → close TRK-DESIGN-001 too. Else update its status with the open child tickets.
   - Update Resume here.
3. Write `docs/NEXT_SESSION_PROMPT_2026-MM-DD.md` with the next priority (depends entirely on the diff findings — could be a child ticket, could be "ship Twitter Years now," could be TRK-DEV-004 special-typography).
4. Commit + push + deploy. Test artifacts (`/tmp/diff-*`) don't need to ship.

## Non-goals

- **Don't fix anything found** — document, classify, file child tickets. Resist the urge to chase a single diff.
- **Don't redesign the parity matrix format** — extend the existing rows; don't restructure.
- **Don't run Calibre/epubcheck installs on the VM** — do that locally on your Mac.
- **Don't commit `/tmp/` artifacts** — they're test outputs.
- **Don't enable PDF/X-4 conversion** — that's a deferred can't-have per the parity doc.

## Pitfalls to expect

- **Page-count mismatch.** Typst's auto widow/orphan + leading rounding will produce a slightly different page count than InDesign's hand-tuned 136. ±3 pages is acceptable; ±10 is a finding.
- **Font metric drift.** Libertinus Serif ≠ Plantin MT Pro exactly. Body text will have similar but not identical width per line, which cascades to all page breaks. This will inflate every per-page AE; tune your threshold for it.
- **Embedded JPEG noise** in the reference PDF (InDesign exports JPEGs at quality settings; Typst may not). The `-fuzz 5%` should swallow this; if you see AE inflated only on image-heavy pages, raise to 10%.
- **EPUB reader differences.** If you render via Calibre vs Apple Books vs Kindle, font handling differs even for "embedded" fonts. Calibre + Adobe Digital Editions are the most-deterministic combo.
- **Per-chapter heading detection.** DEV-009's injection finds the first `</h1>` per chapter file. If pandoc emits a different heading level for chapter titles in this manuscript, the injection may miss. Verify by grepping for `chapter-author` in the unzipped XHTML.
