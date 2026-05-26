# Next session — 2026-05-30

> Use this as the kick-off prompt for a fresh Claude Code session.

Run `jpull` first. Then standard pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  cat ~/backups/.HEALTH-OK && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT migration_number FROM migrations ORDER BY migration_number DESC LIMIT 3"'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `prodcal active`, `OK`, `16/15/14`, `HTTP/2 200`.

## What's done as of 2026-05-29

- **TRK-DEV-009 (per-chapter EPUB author bylines)** — shipped. `srv/epub.go` post-processes the pandoc EPUB zip to inject `<p class="chapter-author">{author}</p>` after each chapter's `<h1>`, matched in spine order against `book_specs.data.epub.chapters[]`. Empty/missing chapters → byte-identical output. Admin SPA has a "Chapters (anthology bylines)" repeating-row editor in the EPUB card.
- **TRK-DESIGN-004 (CJK/Thai font bundling)** — if the parallel session updated its TRACKER block, also shipped. Verify via `git log --oneline -10`.

## Top candidate for this session: TRK-TEST-002 (live Ghosts visual regression)

Now unblocked because both DEV-009 (EPUB per-chapter authorship) and DESIGN-004 (CJK/Thai fonts) are in. Without those, the EPUB diff would have been noise and the PDF would have been missing glyphs.

Plan:
1. Compile `manuscripts/ghosts/main.typ` to PDF via Typst (local or on VM).
2. Pandoc-compile the same source path to EPUB through the admin SPA (upload Ghosts DOCX as a new project — `id != 7` to avoid touching Twitter Years).
3. Configure the new project's `spec.epub.chapters` with the 9 authors from `manuscripts/ghosts/main.typ` (the `set-story-info(...)` calls).
4. Diff page-by-page against `reference/GHOSTS.pdf` and `reference/GHOSTS.epub`.
5. Close the 10 ⚠️ cells in `docs/GHOSTS_PARITY_2026-05-26.md`. File any visual regressions as fresh TRK-DESIGN-* tickets.

## Alternative: TRK-DESIGN-003 (smart punctuation / body alignment)

Same `srv/epub.go` pandoc invocation site as DEV-009 — was deliberately deferred to avoid conflict. Now safe to pick up.

## Smoke test for DEV-009 (if you want to verify before TEST-002)

The DEV-009 unit tests cover the rewriter in isolation but nothing exercises the full pandoc → inject → calibre path. Quick check:

1. Pick any existing Twitter Years compile (or upload any multi-`<h1>` DOCX as a throwaway project).
2. In the admin SPA, EPUB tab, add 2-3 chapters with distinct authors.
3. Compile EPUB. Download the artifact.
4. `unzip -p artifact.epub 'EPUB/*.xhtml' | grep -c chapter-author` → should equal the configured chapter count.

## Non-goals

- Don't auto-detect chapters from DOCX heading runs — explicitly deferred.
- Don't extend transmittal → spec.epub.chapters mapping yet — wait until transmittals carry structured chapter data.
- Don't touch the prodcal systemd unit (TRK-OPS-005).
