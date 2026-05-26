# Session prompt — TRK-DEV-012 Phase A + Phase B

> Use this as the kick-off prompt for a fresh Claude Code session.
> No concurrent sessions — this one's broad (touches both Go pipelines,
> Lua filter, admin SPA).

Run `jpull` first. Then standard pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  git -C /home/exedev/prodcal log --oneline -1'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `active`, HEAD at `0705871` or later, HTTP/2 200.

## Context

TRK-DEV-009 (per-chapter EPUB authorship) shipped with two architectural shortcuts identified during TEST-002 setup:

1. **Schema placement:** chapters live at `spec.epub.chapters[]` in the EPUB card of the admin SPA. They should live at `spec.chapters[]` (top-level) because anthology authorship is a property of the book, not of the EPUB format.

2. **PDF pipeline gap:** the EPUB injection works (post-pandoc XHTML rewrite), but the Typst PDF pipeline reads NO per-chapter info from the spec. `typesetting/filters/docx-to-typst-enhanced.lua` emits no `set-story-info(title:, author:)` per chapter, so SPA-driven Typst output has one book-level author throughout. Running headers and per-chapter bylines on PDF are stuck for any anthology compiled via the SPA. The hardcoded `manuscripts/ghosts/main.typ` has the right shape (9 `set-story-info()` calls) but only works for direct `typst compile` invocation, not the SPA path.

Full ticket: `docs/TRACKER.md` → TRK-DEV-012. Two phases this session:

- **Phase A (~1h):** schema relocation + admin SPA UI move.
- **Phase B (~1-2h):** Lua filter emits `set-story-info()` per chapter from spec data.

Phase C (auto-detection from manuscript on upload) is filed but deferred — pick later when manual-entry friction warrants.

## Phase A — schema relocation + admin SPA UI move

### A1. Spec shape

Move chapters from `spec.epub.chapters[]` to `spec.chapters[]`. Each chapter entry stays the same shape: `{ title, author, file }` where `file` is optional.

### A2. Backend changes

**`srv/epub.go`:** in `parseEPUBSpec` (or wherever `epubSpec.Chapters` is populated — around line 286), read from `data.chapters` first; fall back to `data.epub.chapters` if present and `data.chapters` is empty (back-compat for any saved specs from the brief DEV-009 window). Log the fallback path so we can see if any real specs need migration.

**`srv/bookspecs.go::specToTypstConfig`:** no change needed yet; chapters don't flow through `config.typ` — they'll flow through pandoc metadata (Phase B). But verify `specToTypstConfig` doesn't try to write chapters into the Typst config dict; if it does (it shouldn't, this is new ground), drop that.

**`srv/books.go::buildTypstConfig`:** likely no change. Confirm.

### A3. Admin SPA UI

`srv/static/admin.html`:

- Move the existing "Chapters (anthology bylines)" subsection (around line 594) **out** of the EPUB card and **into a new top-level "Anthology" card** between Elements card and EPUB card.
- Rename the subsection from "Chapters (anthology bylines)" to just "Chapters" since it's no longer EPUB-specific. Update the helper-text copy: "For multi-author anthologies. Each row injects a per-chapter byline into both PDF and EPUB outputs."
- The `tsRenderEpubChapters` JS function probably reads `tsSpec.epub.chapters` — rename to `tsRenderChapters` and read from `tsSpec.chapters`. Add a back-compat one-time migration: on `tsPopulateForm`, if `tsSpec.epub?.chapters?.length > 0 && !tsSpec.chapters` then `tsSpec.chapters = tsSpec.epub.chapters; delete tsSpec.epub.chapters` and trigger save. Logs to console.
- All other refs to `epub-chapters` IDs / element selectors: rename to `chapters` / `anthology-chapters` to match the new card.

### A4. Verify

- Save a project's spec with chapters in the new location. Reload — chapters persist.
- A pre-existing spec with `spec.epub.chapters[]` (test by manually editing the DB or letting one ride from DEV-009-era saves) loads correctly and on first re-save migrates to `spec.chapters[]`.
- Twitter Years (no chapters) renders the empty-state copy.

## Phase B — Lua filter emits set-story-info per chapter

### B1. Decide chapter-data delivery mechanism

Pandoc Lua filters don't have direct access to external data without a route in. Three options, ranked:

1. **(Recommended) Pass via `--metadata-file`:** in `srv/books.go::runConversion`, after looking up the spec, write `{ "chapters": [...] }` to a JSON file in the working dir, pass `--metadata-file=chapters.json` to pandoc. The Lua filter's `Meta()` function reads `doc.meta.chapters` from the parsed metadata.
2. **Pass via `--metadata=chapters=<json>`:** simpler but the metadata value has to be a string pandoc can interpret; nesting structured data through the `=value=` form is finicky for arrays.
3. **Generate a small inline Lua file** containing the chapter data as a Lua table, pass `--lua-filter=chapters.lua` ahead of the main filter. Concrete but ugly: temp file per compile.

Go with option 1 unless something concrete pushes you elsewhere.

### B2. Go side wiring

In `srv/books.go::runConversion`, after `buildTypstConfig` returns the spec (or after a separate lookup if chapters are read separately):

```go
// Build chapters metadata file for the Lua filter, if anthology
if len(spec.Chapters) > 0 {
    chaptersJSON, _ := json.Marshal(map[string]any{"chapters": spec.Chapters})
    chaptersFile := filepath.Join(tmpDir, "chapters.json")
    if err := os.WriteFile(chaptersFile, chaptersJSON, 0644); err == nil {
        pandocCmd.Args = append(pandocCmd.Args, "--metadata-file", chaptersFile)
    }
}
```

(Adapt to actual struct shape — `spec.Chapters` here is sketch; you may need to fetch it from the loaded spec object separately.)

### B3. Lua filter side

`typesetting/filters/docx-to-typst-enhanced.lua`:

- In `Meta(meta)`: read `meta.chapters` into a Lua table for easy lookup. Track index via closure or module-level var.
- Add `Header(el)` (or extend the existing one if it exists) — when `el.level == 1`, emit a `RawBlock("typst", "#set-story-info(title: \"<title>\", author: \"<author>\")\n")` immediately before the heading, indexed by source-order count of h1s seen so far.

   Pseudo:
   ```lua
   local h1_count = 0
   local chapters_meta = nil

   function Meta(meta)
     if meta.chapters then
       chapters_meta = meta.chapters
     end
   end

   function Header(el)
     if el.level == 1 and chapters_meta then
       h1_count = h1_count + 1
       local c = chapters_meta[h1_count]
       if c then
         local title = pandoc.utils.stringify(c.title or "")
         local author = pandoc.utils.stringify(c.author or "")
         -- Escape Typst special chars in title/author (mostly " and \)
         title = title:gsub('\\', '\\\\'):gsub('"', '\\"')
         author = author:gsub('\\', '\\\\'):gsub('"', '\\"')
         local raw = string.format('#set-story-info(title: "%s", author: "%s")\n\n', title, author)
         return { pandoc.RawBlock("typst", raw), el }
       end
     end
     return nil
   end
   ```

   Verify the actual filter's API matches — it might use a different visitor pattern. Adapt.

### B4. Verify

- Recompile Ghosts via the SPA (or via `typesetting/scripts/build.sh` with a project arg) after entering 9 chapter title/author pairs. Output PDF should have correct running headers per chapter (verso=author, recto=title) — same as `manuscripts/ghosts/main.typ`'s direct output.
- Compile a single-chapter book (Twitter Years) — should be unchanged (no `set-story-info()` injected when chapters is empty).
- Confirm the generated `main.typ` (in the temp dir during compile, or via debug logging) has `#set-story-info()` calls at chapter boundaries.

## Acceptance

- `spec.chapters[]` is the canonical location; backend reads from there with fallback to `spec.epub.chapters[]` for back-compat.
- Admin SPA: Chapters editor is in a top-level "Anthology" card, not inside EPUB.
- One-time migration of any pre-existing `spec.epub.chapters[]` to `spec.chapters[]` on save.
- Compiled Ghosts PDF (via SPA) has per-chapter `set-story-info()` driving running headers — verso shows chapter author, recto shows chapter title, paged per `manuscripts/ghosts/main.typ`'s reference behavior.
- Twitter Years (no chapters) is byte-identical to today's PDF output.
- EPUB pipeline continues working — DEV-009's injection still functions, just reading from the new schema location.

## Deploy

```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && \
  go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal && \
  sleep 2 && systemctl is-active prodcal'
curl -sI https://jdbbs.exe.xyz | head -1
```

Don't touch the systemd unit (TRK-OPS-005). Push to `main` yourself.

## Wrap-up

1. Update `docs/TRACKER.md`:
   - Mark TRK-DEV-012 Phase A + Phase B done; status `in-progress` with Phase C still open.
   - Update Resume here block.
2. If you also have time to do TRK-DEV-011 finding #1 (move Chapters out of EPUB card) — note that DEV-012 Phase A already does this. Cross-reference and close that finding.
3. Write `docs/NEXT_SESSION_PROMPT_TEST-002.md` follow-up if it makes sense — the existing TEST-002 prompt at `docs/NEXT_SESSION_PROMPT_TEST-002.md` doesn't need to be rewritten, but a one-line note that DEV-012 A+B landed and the chapter manual-entry step now goes in the Anthology card (not EPUB) might help.
4. Commit + push + deploy.

## Non-goals

- **Don't do Phase C (auto-detection).** Filed; later session.
- **Don't restructure the rest of the spec.** Resist the urge to clean up adjacent areas; surgical change.
- **Don't introduce a new migration.** `book_specs.data` is JSON; schema-flexible. One-time client-side migration on form-load handles back-compat.
- **Don't touch DEV-009's `injectChapterAuthors` logic in `srv/epub.go`.** Only change where chapters are read from in spec parsing.
- **Don't run TEST-002 in this session.** Separate ticket.

## Pitfalls to expect

- **JSON encoding of authors with quotes/apostrophes/accents** (e.g. "Rafael Fernández", "Sam Chua"). The metadata file is JSON → Lua's metadata loader handles UTF-8 fine, but the Typst output emission must escape `"` and `\` properly (see escaping in the pseudo-code above). Test with the Fernández name as a smoke check.
- **Heading level confusion.** The chapter-prep we did for Ghosts uses `# Heading 1` per chapter. Confirm `el.level == 1` is the right filter trigger by examining the pandoc AST for a sample DOCX. If pandoc shifts heading levels for some reason (e.g. a leading title or front-matter), the indexing may not align with the spec's chapter array — verify before declaring victory.
- **h1 occurrences inside chapter bodies.** A chapter might contain its own h1 for some structural reason (probably not, but check). If so, the source-order indexing breaks. Mitigation: only inject for the FIRST h1 within each "chapter section" — requires tracking chapter boundaries. Skip until proven necessary.
- **The Lua filter's `Header` vs `Heading` visitor name** — different pandoc versions use different names. Check what the existing filter uses (`grep -n "^function" typesetting/filters/docx-to-typst-enhanced.lua`).
