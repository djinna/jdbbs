# Session prompt — Typst quick-win triple (TRK-DESIGN-005 + 008 + 007)

> Use this as the kick-off prompt for a fresh Claude Code session.
> Three small Typst fixes that together close the biggest perceived-quality
> gap in Ghosts vs reference. Roughly one session.

Run `jpull` first. Then pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  git -C /home/exedev/prodcal log --oneline -1 && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT id, name FROM projects" && \
  typst --version'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `active`, HEAD `e5b5f09` or later, two projects (7=Twitter Years, 14=Ghosts TEST-002), `typst <version>`, HTTP/2 200.

## What you're doing

Three Typst-side fixes filed during TRK-TEST-002 (2026-05-26). Each is small in isolation; together they close the most visible Ghosts parity gaps. Recompile after each, diff against `scratch/ghosts-typst.pdf` (last session's output) and `scratch/ghosts-reference.pdf` (the InDesign golden).

**Tickets covered (full bodies in `docs/TRACKER.md`):**

1. **TRK-DESIGN-005 (P1, 5 min)** — `manuscripts/ghosts/main.typ:1` import path is stale; fix to `../../typesetting/templates/series-template.typ`. Lets `typst compile` work without the `/tmp/ghosts-build/` staging-tree workaround.

2. **TRK-DESIGN-008 (P1, ~20 min)** — Body paragraph first-line indent not rendering despite `#set par(first-line-indent: 0.75em)` in every chapter `.typ`. Likely needs the new struct form `first-line-indent: (amount: 0.75em, all: true)`. The biggest visible regression vs reference.

3. **TRK-DESIGN-007 (P3, 5 min)** — Running header title is Title Case; reference is ALL CAPS. Wrap title in `upper()` inside `running-header()` in `series-template.typ`. Author byline already correct.

## Implementation order

### Step 1 — TRK-DESIGN-005 (unblocks the rest)

Edit `manuscripts/ghosts/main.typ`:

```diff
-#import "../../templates/series-template.typ": *
+#import "../../typesetting/templates/series-template.typ": *
```

Smoke compile (no staging-tree hack needed now):

```bash
ssh exedev@jdbbs.exe.xyz '\
  cd /home/exedev/prodcal && \
  typst compile --root . --font-path typesetting/fonts \
    manuscripts/ghosts/main.typ /tmp/ghosts-typst.pdf && \
  pdfinfo /tmp/ghosts-typst.pdf | grep -E "Pages|Page size"'
```

Expect: 101±5 pages (same ballpark as last session), 353.811 × 546.567 pt trim. Push + redeploy not required — this only affects manuscript-direct compiles, not the SPA pipeline.

### Step 2 — TRK-DESIGN-008 (the indent fix)

First, check Typst version on VM (`typst --version`). The fix depends on which `first-line-indent` syntax the VM's Typst expects.

Test the fix in **one** chapter file first to verify the syntax works (cheap iteration):

```diff
# manuscripts/ghosts/01-soda.typ
-#set par(justify: true, leading: 0.6em, first-line-indent: 0.75em)
+#set par(justify: true, leading: 0.6em, first-line-indent: (amount: 0.75em, all: true))
```

(If that form errors, fall back to keeping `first-line-indent: 0.75em` and adding a separate `#set par(first-line-indent-all: true)` — or check Typst docs for the version's actual shape.)

Compile, render page covering ch1 body to a PNG, eyeball: do paragraphs now have a 0.75em first-line indent on all but the first paragraph after each heading? If yes, roll the fix across all 9 chapter files (00-intro through 08-loyalty). Better: lift the `#set par(...)` into `series-template.typ::book()` so chapter files don't need to repeat it.

### Step 3 — TRK-DESIGN-007 (running header case)

In `typesetting/templates/series-template.typ`, find `running-header()` (around lines 134-172 per the parity doc). Wrap the title render in `upper()`:

```diff
-#current-story-title
+#upper(current-story-title)
```

(Match the actual code shape — `current-story-title` may be in a `text(...)[...]` block; just wrap whatever renders the title in `upper()`.)

Verify: author byline (verso) was already ALL CAPS, so don't double-upper it.

### Step 4 — Recompile + re-diff vs reference

```bash
ssh exedev@jdbbs.exe.xyz '\
  cd /home/exedev/prodcal && git pull --ff-only && \
  typst compile --root . --font-path typesetting/fonts \
    manuscripts/ghosts/main.typ /tmp/ghosts-typst-v2.pdf'

scp exedev@jdbbs.exe.xyz:/tmp/ghosts-typst-v2.pdf scratch/
mkdir -p scratch/diff-typst-v2
pdftoppm -r 100 -png scratch/ghosts-typst-v2.pdf scratch/diff-typst-v2/page
```

Spot-check pages 8, 9, 10, 15 (chapter opener + body) vs the same pages in `scratch/diff-ref/` and `scratch/diff-typst/`. Want to see:

- Indents now visible on body paragraphs (was the biggest miss).
- Running header title in ALL CAPS.
- Import path no longer requires staging-tree.

## Acceptance

- `typst compile --root <repo-root> manuscripts/ghosts/main.typ` succeeds without the `/tmp/ghosts-build/` workaround.
- Body paragraphs in `scratch/diff-typst-v2/page-009.png` (and similar) show a 0.75em first-line indent on continuing paragraphs.
- Running header verso/recto on a body page shows ALL CAPS title.
- `docs/GHOSTS_PARITY_2026-05-26.md` and `docs/TRACKER.md` updated: DESIGN-005/007/008 closed.
- Page-count delta narrows (was 101 vs 136; expect 105-115 after these fixes, depending on whether indent fix shifts breaks).

## Then

After this triple lands, the remaining DESIGN-001 child tickets are:

- TRK-DESIGN-006 (P2, ~30-60 min) — chapter opener images not rendering. Bigger investigation; may need `main.typ` refactor.
- TRK-DESIGN-009 (P2, ~45 min) — poem/verse blocks rendering as plain body. Two-part fix (chapter `.typ` files + `poem()` styling).
- TRK-DEV-013 (P3, ~30-60 min) — epubcheck PKG-005 packaging error from pandoc's zip writer.

Once those land, DESIGN-001 closes with zero ❌ in the parity matrix.

Parallel queues that haven't moved:
- TRK-DESIGN-003 — body alignment ragged-vs-justified + smart punctuation audit.
- TRK-DEV-012 Phase C — chapter auto-detection on upload (still deferred).

## Wrap-up

1. Commit all three fixes as one commit (they're cohesive) with message: `TRK-DESIGN-005+007+008: typst quick wins (import path, header case, paragraph indent)`.
2. Push.
3. Update `docs/TRACKER.md`: close DESIGN-005/007/008; update Resume here; note remaining DESIGN-001 children.
4. Write `docs/NEXT_SESSION_PROMPT_<date>.md` for the next priority (probably DESIGN-006 + 009 combined, since they both touch chapter `.typ` files).

## Non-goals

- **Don't fix DESIGN-006 (images) here** — it's a deeper investigation.
- **Don't touch the EPUB pipeline** — these are all manuscript+template Typst-side fixes.
- **Don't deploy** — these don't affect the running prodcal binary at all. The SPA pipeline compiles via the Lua filter + `series-template.typ`; the running-header change would propagate, but you're not changing the prod binary, just template files. If you want belt-and-braces, `ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only'` after pushing — no rebuild/restart needed.

## Pitfalls

- **Typst version drift.** If `first-line-indent: (amount: 0.75em, all: true)` errors, the VM Typst may be older than expected. Check `typst --version`, consult docs for that version.
- **Roll-out blast radius.** Lifting `#set par(...)` into the template affects every series-template user (right now only Ghosts and Twitter Years). Twitter Years is single-author justified — verify it still renders cleanly after the change.
- **`upper()` on already-uppercase author.** Don't apply it to `current-story-author` — it's already correctly uppercased in the existing code.
