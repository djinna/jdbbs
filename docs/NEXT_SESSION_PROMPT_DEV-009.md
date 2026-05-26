# Session prompt — TRK-DEV-009 (per-chapter author in EPUB)

> Use this as the kick-off prompt for a fresh Claude Code session.
> Concurrent-safe with TRK-DESIGN-004 (font bundling — touches
> `typesetting/fonts/`, `series-template.typ`, `epub-styles.css`).
> Do NOT pick up TRK-DESIGN-003 or TRK-DEV-004 in this session —
> DESIGN-003 touches the same pandoc invocation site as you (`srv/epub.go`);
> DEV-004 will eventually touch admin.html EPUB section like you.

Run `jpull` first. Then standard pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT migration_number FROM migrations ORDER BY migration_number DESC LIMIT 3" && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT id, name FROM projects"'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `prodcal active`, `16/15/14`, one project (`7|The Twitter Years: 2007–22`), `HTTP/2 200`.

## What you're building

The Ghosts parity audit (2026-05-26, `docs/GHOSTS_PARITY_2026-05-26.md`) found one critical blocker for shipping any Ghosts-like anthology: **per-chapter author bylines work on the Typst side but break in the EPUB pipeline.** The Typst template uses `set-story-info(title:, author:)` per chapter and main.typ configures 9 different authors across 9 chapters. The EPUB pipeline has no per-chapter author field — `book_specs.data.epub` has only a singular `author`, so pandoc emits one book-level author for the whole EPUB.

Twitter Years (single-author) ships fine without this. Ghosts and every future series title needs it.

Full ticket: `docs/TRACKER.md` → `TRK-DEV-009`.

## Implementation steps

### 1. Schema (~15 min)

No migration needed — `book_specs.data` is JSON, schema-flexible. Add a `chapters` array to `data.epub`:

```jsonc
"epub": {
  "...existing fields...": "...",
  "chapters": [
    { "title": "Soda Sweet as Blood", "author": "Spencer Nitkey", "file": "01-soda.md" },
    { "title": "In Every Lifetime",   "author": "Lara Dal Molin", "file": "02-lifetime.md" }
  ]
}
```

`file` is optional (used later for source-file-matching); empty array preserves current behavior (book-level `author` for the whole EPUB).

### 2. Backend wiring (`srv/bookspecs.go` + `srv/epub.go`, ~1 hour)

Read the existing `parseEPUBSpec` (in `srv/bookspecs.go`) and `handleGenerateEPUB` (in `srv/epub.go`, line ~85+) — they're the path that DEV-003 closed.

- Extend `epubSpec` (Go struct in `srv/epub.go`) with a `Chapters []epubChapter` field; `epubChapter` struct = `{ Title, Author, File string }`.
- Extend `parseEPUBSpec` to read `data.epub.chapters` (empty/missing → empty slice → existing behavior).
- In `handleGenerateEPUB`, after `parseEPUBSpec` and after the pandoc DOCX→XHTML pass: if `spec.Chapters` is non-empty, post-process the generated XHTML to inject a `<p class="chapter-author">{{author}}</p>` block under each chapter heading.

Implementation note for the post-process: pandoc with `--split-level=1` already emits separate XHTML files per chapter. Either:
- (Simpler) After pandoc but before epub-zip, walk the chapter files, match each to a `spec.Chapters[i]` by file order, regex-insert the byline after the first `<h1>` tag.
- (Cleaner) Use pandoc's Lua filter system — extend an existing filter or add a small new one — to inject the byline at AST level. Use this if the order-match feels brittle.

Default to the simpler approach unless ordering edge cases bite you.

### 3. UI in admin SPA (`srv/static/admin.html`, ~1 hour)

In the Typesetting tab, under the EPUB section (search for `ts-epub-` ID prefixes), add a "Chapters" subsection. Reuse the corrections-table pattern (search `tsCorrectionsTable` and adjacent helpers for the shape). Each row: title input + author input + optional file dropdown (populated from project files, or just a text input for now). Add-row / remove-row buttons. Auto-save like other spec fields (the existing debounced save in `tsSaveSpec` should pick up `data.epub.chapters` automatically if the form-to-JSON serialization is set up the standard way — verify).

### 4. Smoke test (~30 min)

Twitter Years (project 7) is single-author so it's not a good smoke target for this. Two options:

- **(Recommended)** Create a temporary multi-author project for testing. Upload any DOCX with multiple chapter-style headings, configure 2-3 chapters with different authors, compile EPUB, verify each chapter's XHTML has the right byline.
- **(More involved)** Upload the Ghosts manuscript as a real project. `manuscripts/ghosts/` has both .md and .typ sources but no consolidated DOCX. Skip this unless you want to assemble one from the .md files — that's effectively starting TRK-TEST-002 which is a separate ticket.

Either way: open the produced EPUB in Calibre (or unzip and inspect XHTML directly with `unzip -p artifact.epub ch1.xhtml | grep chapter-author`).

### 5. Acceptance checks

- A book with empty/missing `data.epub.chapters` produces an EPUB identical to today's output (no regression for Twitter Years).
- A book with populated chapters produces an EPUB where each chapter shows its own `<p class="chapter-author">` paragraph styled per the existing CSS (epub-styles.css lines 102-111).
- The `.chapter-author` rendering is visually equivalent to what the InDesign reference EPUB shows for Ghosts (visual confirmation against `reference/GHOSTS.epub` if you have a moment).
- No new API endpoint needed — extending the existing `book_specs` PUT + EPUB generate suffices.

## Deploy

```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && \
  go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal && \
  sleep 2 && systemctl is-active prodcal'
curl -sI https://jdbbs.exe.xyz | head -1
```

Don't touch the systemd unit (TRK-OPS-005). Push to `main` yourself (auto-mode classifier blocks direct push from inside Claude Code sessions), then ssh deploy.

## Wrap-up

1. `docs/TRACKER.md`: mark TRK-DEV-009 done with call trace + commit refs. Update Resume here block.
2. If a TRACKER conflict on pull, it's almost certainly the parallel DESIGN-004 session also marking itself done — accept both sides (different ticket sections).
3. Write `docs/NEXT_SESSION_PROMPT_2026-05-27.md` or the equivalent date. Next priority likely **TRK-TEST-002** (live visual regression — now meaningful because both EPUB and font issues are addressed).
4. Commit, push.

## Non-goals

- **Don't bundle fonts** — that's TRK-DESIGN-004 (concurrent session).
- **Don't touch TRK-DESIGN-003 (smart punctuation / body alignment)** — same `srv/epub.go` pandoc invocation, will conflict. Separate session.
- **Don't start TRK-TEST-002 (live regression)** — that's the natural next step but a separate ticket. Stop after smoke confirms per-chapter bylines render.
- **Don't auto-detect chapters from DOCX** — that was noted as a follow-up in the ticket; keep it manual for v1.
- **Don't migrate the schema** — `book_specs.data` is JSON, schema-flexible. Just add the field and use it.

## Concurrent-work awareness

At session-start time, **TRK-DESIGN-004 (CJK/Thai font bundling)** is likely running in another fresh Claude Code session on this Mac. They're touching:
- `typesetting/fonts/noto/` (new directory)
- `typesetting/templates/series-template.typ` (font fallback chain)
- `typesetting/templates/epub/epub-styles.css` (`.chinese` / `.thai` classes near lines 344-353)

Zero overlap with your zone (`srv/*.go`, `srv/static/admin.html` EPUB section). The only shared file is `docs/TRACKER.md` for the close commits — if your `git pull` hits a conflict, rebase and accept both ticket sections.
