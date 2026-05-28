# Session handoff — TRK-DESIGN-003 (smart punctuation + ragged-left EPUB body)

**Date:** 2026-05-28 (orchestrator logged; session ran across 2026-05-26 / 27 UTC).
**Status:** Done.
**Commits:** `b33b36c` (initial) + `8b1c6c4` (fixup).

## Shipped

Both parts verified end-to-end against book 9 (Ghosts test 002).

**Code (final state):**
- `srv/books.go:231,235` — `--from=docx+styles` + `-t typst+smart`
- `srv/epub.go:137,138` — `--from=docx+styles` + `--to=epub3+smart`
- `srv/epub.go::buildCSS` — unconditionally prepends `body, p { text-align: left; }` to override pandoc's default justified epub3 stylesheet.

**Part B decision:** picked option (a) — flat ragged-left default in CSS, no per-book toggle. Rationale: both in-flight titles want ragged per InDesign reference; reflowable EPUB readers handle justification poorly without good hyphenation. A per-book `epub.justify` toggle can be added later if a justified-default title arrives.

## Environmental traps hit (and resolved)

### 1. Wrong systemd unit serving requests

`srv.service` (legacy stub, no `WorkingDirectory` → cwd `/home/exedev`, empty `db.sqlite3`) and `prodcal.service` (canonical, cwd `/home/exedev/prodcal/`, 354 MB real DB) both exist on the VM and were both `enabled`. `sudo systemctl restart srv` after deploy kept the wrong unit on `:8000` — API returned `[]` from `/api/books`, `/api/books/9` returned 404, even though `sqlite3 /home/exedev/prodcal/db.sqlite3 "SELECT … FROM books"` showed 7 rows.

**Diagnosis:** inspect `/proc/$MainPID/cwd` and `lsof` on the listener — confirms which file the running binary actually has open.

**Immediate fix:** `sudo systemctl stop srv.service && sudo systemctl disable srv.service && sudo systemctl start prodcal.service`.

**Followup:** TRK-OPS-010 filed — remove `srv.service` from disk entirely.

### 2. `+smart` placement (reader vs writer)

Initial implementation put `+smart` on the docx **reader** (`--from=docx+styles+smart`); pandoc rejects with `exit 23: The extension smart is not supported for docx`. Smart is a **writer-side** extension only.

**Symptom was silent.** The EPUB pipeline returned a `{"status":"generating_epub"}` response, error-pathed in the background, and the next `/api/books/9/outputs` call returned the previous successful output unchanged. Glyph counts on that stale EPUB looked "correct" (552 `'`, 614 `"`, 616 `"`, 141 `—`) because the source DOCX already has curly quotes (Word auto-converts at input time).

**Caught by inspecting the prodcal journal:** `sudo journalctl -u prodcal -n 60`. The error line was unmissable once you looked.

**Fix:** move `+smart` to the writer flag: `-t typst+smart` (PDF path) and `--to=epub3+smart` (EPUB path).

**Lesson:** when verifying a code change end-to-end, confirm the output_id is new (not a stale earlier compile). The async generate-epub endpoint silently retains the previous successful output on background failure.

## Follow-ups filed / referenced

- **TRK-DEV-004 Phase C/D** — `+smart` on writer will bleed into any future preserved-block content (terminal transcripts, ASCII art). Neither Ghosts nor Twitter Years has any; file when a future manuscript adds it.
- **TRK-OPS-010** — remove `srv.service` unit file + stub DB (foot-gun for next restart). P2, ~20 min.
- **EPUB-side verse styling** — small follow-up to DESIGN-009 (pandoc class-markup work).

## Architectural note (worth remembering)

The EPUB stylesheet shipped in `EPUB/styles/stylesheet1.css` is **constructed inline as a Go string** by `srv/epub.go::buildCSS`. It is NOT the contents of `typesetting/templates/epub/epub-styles.css` (which exists but is not part of the runtime EPUB pipeline). Editing the template file does not affect production output. CSS changes for shipped EPUBs go in `buildCSS` in Go.

## Verification artifacts

- Fresh `book_outputs` row: id=29, 14,205,558 bytes, created 2026-05-27T01:54:32Z.
- Pandoc compiled clean — no `exit 23` errors in current journal tail; only the historical pre-fix error remains.
- `EPUB/styles/stylesheet1.css` line 2 = `body, p { text-align: left; }` (proves Part B reaches the shipped EPUB).
- Glyph counts unchanged vs prior outputs (source already curly) — Part A wiring is verified by error-free compile + correct extension placement per pandoc docs; a true visible delta requires a straight-quote manuscript.
