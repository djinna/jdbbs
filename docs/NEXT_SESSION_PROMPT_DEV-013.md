# Session prompt — TRK-DEV-013 (epubcheck PKG-005 packaging fix)

> Use this as the kick-off prompt for a fresh Claude Code session.
> Standalone — touches `srv/epub.go` only.
>
> **Concurrency:** safe to run in parallel with ANY Typst-side session
> (DESIGN-005+007+008, DESIGN-006+009). Zero file overlap.

Run `jpull` first. Then pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  git -C /home/exedev/prodcal log --oneline -1 && \
  which epubcheck || echo "epubcheck not on PATH"'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `active`, recent HEAD, `epubcheck` available (install via `brew install epubcheck` on Mac if you'll validate locally; VM may not have it).

## What you're doing

Fix a packaging error reported by `epubcheck` on every EPUB the prod pipeline generates:

```
ERROR(PKG-005): The mimetype file has an extra field of length 9.
The use of the extra field feature of the ZIP format is not permitted
for the mimetype file.
```

**EPUB spec requirement:** the `mimetype` entry must be:
- the **first** file in the EPUB zip
- stored uncompressed (STORE method, not DEFLATE)
- have **no extra field** in its local file header

Pandoc's zip writer apparently adds a 9-byte extra field — likely a Unix-extra-field with mtime/atime metadata. Strict validators (epubcheck, some Kindle paths, some commercial e-reader QA pipelines) reject; permissive readers (Apple Books, ADE, Calibre) accept.

Reference `reference/GHOSTS.epub` passes epubcheck clean.

## Implementation

### Step 1 — Reproduce locally

```bash
# Generate a fresh test EPUB via the running prod (or compile any existing
# project's EPUB output through the SPA, then download it to scratch/)
# Confirm the error reproduces:
epubcheck scratch/ghosts-app.epub 2>&1 | grep PKG-005
```

If you don't have `epubcheck`, install: `brew install epubcheck` (Mac) or `sudo apt install epubcheck` (Linux).

### Step 2 — Locate the mimetype rewrite point

The pattern lives in `srv/epub.go`. The existing `injectChapterAuthors` function already does post-pandoc EPUB rewriting — open the zip, modify entries, write it back out preserving order. The mimetype-extra-field fix is the same shape, slotted into the same pipeline.

**Implementation choice:** wrap a new helper `stripMimetypeExtraField(epubBytes []byte) ([]byte, error)` that:

1. Reads the zip from bytes (or from the on-disk path, depending on how `injectChapterAuthors` is structured today).
2. Locates the `mimetype` entry (must be the first file, name === `"mimetype"`).
3. Rewrites it: same name, same `application/epub+zip` body, **uncompressed** (Method=0), **no extra field** in the local file header.
4. Preserves all other entries byte-for-byte (don't rewrite them — only mimetype).
5. Returns the corrected bytes.

Either:
- Use `archive/zip` to read entries, then construct a new zip with `zip.Writer` writing mimetype first via `CreateHeader` with `Method: 0` and `Extra: nil`, then copy remaining entries with `CopyFrom`-style passthrough.
- OR splice the 30-byte mimetype local-file-header at the start of the byte stream, recompute the central directory entry to match (no extra field), recompute central-directory offsets if header size changed. Surgical but error-prone.

Go with the `archive/zip` approach unless you find it doesn't preserve other entries byte-exactly (some implementations recompress on copy — verify with a diff against pre-rewrite for non-mimetype entries).

### Step 3 — Wire into the EPUB pipeline

Call `stripMimetypeExtraField` from `runEPUBGeneration` (or wherever the final EPUB byte buffer is produced) **after** `injectChapterAuthors` has done its work, before the bytes go to the response writer / output storage.

Order matters: `injectChapterAuthors` rewrites chapter XHTML; `stripMimetypeExtraField` rewrites the mimetype entry. They don't conflict but the latter should run last so it sees the final byte stream.

### Step 4 — Test

Add to `srv/epub_chapter_test.go` (or a new `srv/epub_packaging_test.go`):

```go
func TestStripMimetypeExtraField(t *testing.T) {
    // Build a zip with a mimetype entry carrying an extra field, plus a
    // couple other entries.
    // Call stripMimetypeExtraField.
    // Verify:
    //  - mimetype is still first
    //  - mimetype's local file header has zero-length extra field
    //  - mimetype is stored uncompressed
    //  - other entries are byte-identical to pre-strip
    //  - the resulting bytes are a valid zip
}
```

If you have `epubcheck` available in the test environment (probably not — it's Java), skip an automated epubcheck assertion. Manual verification via `epubcheck scratch/ghosts-app-fixed.epub` is sufficient.

### Step 5 — Deploy + verify

```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && \
  go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal && \
  sleep 2 && systemctl is-active prodcal'
curl -sI https://jdbbs.exe.xyz | head -1

# Regenerate Ghosts EPUB through the SPA (or trigger via curl with
# X-ExeDev-UserID header). Download, run epubcheck:
epubcheck scratch/ghosts-app-v2.epub
```

Expect: zero PKG-005 errors. Other epubcheck output (info-level RSC-004 about commercial-font encryption) is unrelated and acceptable.

## Acceptance

- `epubcheck` on a freshly-generated Ghosts EPUB returns zero PKG-005 errors.
- `srv/epub_chapter_test.go` (or sibling) has a unit test asserting the mimetype entry has no extra field after strip.
- Existing tests still pass (the strip should be invisible to readers that already accepted the original).
- DEV-009's `injectChapterAuthors` still works — strip runs after it, doesn't undo the chapter-XHTML rewrites.
- Reader-side smoke: open the fixed EPUB in Calibre or Apple Books, confirm chapter bylines still present, fonts still embedded.

## Wrap-up

1. Commit:
   ```
   TRK-DEV-013: strip mimetype extra field for epubcheck PKG-005 compliance
   ```
2. Push to main.
3. Build + redeploy on VM (this is Go code; binary rebuild required).
4. Update `docs/TRACKER.md`: close DEV-013, update Resume here.
5. **If DESIGN-006+009 has also landed by now:** close TRK-DESIGN-001 (the parent parity audit) — zero ❌ remaining in the matrix.

## Non-goals

- **Don't refactor pandoc's invocation** — the bug is in pandoc's zip writer; we work around it post-hoc.
- **Don't try to upstream a fix to pandoc** — separate, slow path.
- **Don't touch `injectChapterAuthors`** — it's working as intended; just chain after it.
- **Don't strip extra fields from other entries** — only mimetype is spec-restricted. Other entries can carry extra fields freely.

## Pitfalls

- **Go's `archive/zip` may not preserve byte-identical local-file-header order or extra fields on other entries.** If a test detects drift on non-mimetype entries, switch to manual byte-level splice (parse the 30-byte LFH for mimetype, rewrite, leave the rest of the byte stream untouched, fix up central-directory offsets if the LFH size changed).
- **The 9-byte extra field's content** — if it's a Unix mtime/atime field (`UT` signature), pandoc adds it via its zip writer's default options. There's no flag to disable on the pandoc side that we've found.
- **EPUB readers that DO require mimetype-first.** Some older Kindle versions strict-check this; the current behavior may already break for those users. Fix is forward-compatible only — no regressions.
- **CentralDirectory offsets.** If the LFH for mimetype shrinks (extra field gone → smaller header), all subsequent entries' offsets in the central directory must decrement by the same amount. `archive/zip` handles this automatically if you rebuild the zip; manual splice does not.
- **mimetype body itself** — must be exactly `application/epub+zip` (20 bytes, no trailing newline). Verify the bytes match exactly; some pandoc versions include trailing whitespace.
