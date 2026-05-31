# Session prompt — TRK-DESIGN-002 (print-only licensed font wiring)

> Use this as the kick-off prompt for a fresh Claude Code session.
> Concurrent-safe with TRK-DEV-012 Phase C (different code zones —
> this is `series-template.typ` + `.gitignore` + `srv/epub.go` guard;
> Phase C is upload pipeline + admin SPA).

Run `jpull` first. Then pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  ls /home/exedev/prodcal/typesetting/fonts/ | head -10 && \
  fc-list | grep -iE "plantin|proxima" | head -5'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `active`, OFL font dirs present (libertinus/, noto/, source-sans-3/, jetbrains-mono/), `licensed/` likely absent on first run, HTTP/2 200. `fc-list` shows nothing for Plantin/Proxima until populated.

## What you're building

Print-only licensed-font wiring. EPUB stays on OFL fonts; the licensed family files never enter the EPUB zip, never enter git, never leak server-side beyond `typst compile`.

User owns Plantin MT Pro + Proxima Nova under standard perpetual desktop licenses — same model as InDesign/Quark for decades. Print-PDF embedding (subset, rendered output) is covered. EPUB redistribution is not.

Full ticket in `docs/TRACKER.md` → TRK-DESIGN-002.

## Implementation order

### Step 1 — gitignore + placeholder README (~5 min)

```bash
echo "" >> .gitignore
echo "# Licensed fonts — print-only, never committed (TRK-DESIGN-002)" >> .gitignore
echo "typesetting/fonts/licensed/" >> .gitignore
```

Create `typesetting/fonts/licensed/README.md`:

```markdown
# Licensed fonts (print-only)

This directory is gitignored. Drop OTF/TTF files here under per-family
subdirectories (e.g. `plantin-mt-pro/`, `proxima-nova/`).

These fonts are licensed for desktop print use only. They MUST NOT be:
- committed to git (the repo is a distribution channel)
- embedded in EPUB output (zip distribution = redistribution)
- shared with anyone outside the license holder

The EPUB pipeline (`srv/epub.go`) has a runtime guard refusing any
font path containing `/licensed/`.

To populate: drop the files locally, then run
`scripts/sync-licensed-fonts.sh` to mirror to the VM.
```

### Step 2 — sync script (~10 min)

Create `scripts/sync-licensed-fonts.sh`:

```bash
#!/usr/bin/env bash
# Sync print-licensed fonts from Mac to VM. Never committed to git.
# TRK-DESIGN-002.
set -euo pipefail
SRC="${HOME}/jd-projects/jdbbs/typesetting/fonts/licensed/"
DST="exedev@jdbbs.exe.xyz:/home/exedev/prodcal/typesetting/fonts/licensed/"
if [ ! -d "$SRC" ] || [ -z "$(ls -A "$SRC" 2>/dev/null)" ]; then
  echo "No licensed fonts at $SRC — nothing to sync."
  exit 0
fi
echo "Syncing $SRC -> $DST"
rsync -avz --delete "$SRC" "$DST"
ssh exedev@jdbbs.exe.xyz "ls -la /home/exedev/prodcal/typesetting/fonts/licensed/"
```

`chmod +x scripts/sync-licensed-fonts.sh`.

### Step 3 — spec fields + admin UI (~30 min)

Add two new fields to the typesetting spec form in `srv/static/admin.html` under a new "Print fonts" subsection of the Typesetting card (NOT the EPUB card — these are print-only):

- `pdf.body_font` (text input, placeholder "Libertinus Serif")
- `pdf.heading_font` (text input, placeholder "Source Sans 3")

Helper text: "Family name as Typst sees it (use `fc-query` on the font file to confirm). Leave blank for OFL defaults. Licensed fonts only; never appears in EPUB output."

Wire through `tsPopulateForm` + `tsMarkDirty` patterns already used for other text fields. No backend changes needed — `book_specs.data` is JSON (schema-flexible).

### Step 4 — series-template.typ pickup (~20 min)

Read the new spec values via the existing pandoc-metadata route (the chapter mechanism added in TRK-DEV-012 Phase B used `--metadata-file` — extend that, or use the existing config.typ flow if cleaner).

In `series-template.typ`, find the existing body/heading font definitions and add conditional fallback:

```typst
#let body-font = if pdf-body-font != none and pdf-body-font != "" {
  (pdf-body-font, "Libertinus Serif", "Noto Serif CJK TC", "Noto Serif Thai")
} else {
  ("Libertinus Serif", "Noto Serif CJK TC", "Noto Serif Thai")
}
```

Same shape for heading font with Source Sans 3 fallback. Keep the existing CJK + Thai fallbacks in the list so multilingual content still renders.

### Step 5 — EPUB hard-separation guard (~15 min)

In `srv/epub.go::handleGenerateEPUB`, find where `--epub-embed-font` args are assembled (the Noto wiring from TRK-DEV-010). Add the assertion BEFORE appending each path:

```go
for _, p := range fontsToEmbed {
    if strings.Contains(p, "/licensed/") {
        return fmt.Errorf("refusing to embed licensed font in EPUB: %s", p)
    }
    if _, err := os.Stat(p); err == nil {
        args = append(args, "--epub-embed-font", p)
    }
}
```

Add a unit test in `srv/epub_packaging_test.go` (or sibling) that constructs a fontsToEmbed list containing a `/licensed/` path and asserts the function errors.

Also: confirm the EPUB CSS (`srv/epub.go::buildCSS`) does NOT reference Plantin or Proxima as fallbacks. Per [[jdbbs-epub-css-inline]] memory, the CSS is constructed in Go — check the literal strings, ensure only OFL families appear.

### Step 6 — Smoke test (no licensed fonts yet) (~5 min)

Before user drops the actual font files, verify fallback path works:

```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && \
  go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal'
# Compile Ghosts with empty pdf.body_font → should fall back to Libertinus
# Verify PDF still compiles, body font is Libertinus
```

### Step 7 — Smoke test (with licensed fonts, user-driven) (~10 min)

After user has dropped OTF/TTF into `typesetting/fonts/licensed/plantin-mt-pro/` and run `scripts/sync-licensed-fonts.sh`:

```bash
ssh exedev@jdbbs.exe.xyz 'fc-query /home/exedev/prodcal/typesetting/fonts/licensed/plantin-mt-pro/PlantinMT-Regular.otf | head -10'
# Confirms exact family name Typst will see
```

User sets `pdf.body_font = "<exact family name>"` in Ghosts spec, regenerates PDF, verifies with `pdffonts /tmp/ghosts.pdf | grep -i plantin` that Plantin is embedded (subset) in the PDF.

Then regenerates EPUB and verifies:
- `unzip -l /tmp/ghosts.epub | grep -i plantin` → zero matches
- `grep -i plantin /tmp/epub-x/EPUB/styles/stylesheet1.css` → zero matches

## Acceptance

- `.gitignore` excludes `typesetting/fonts/licensed/`.
- `typesetting/fonts/licensed/README.md` committed (explains the convention).
- `scripts/sync-licensed-fonts.sh` committed + executable.
- Admin SPA has "Print fonts" subsection with two inputs (body + heading).
- `series-template.typ` reads spec values, falls back cleanly to OFL when empty.
- `srv/epub.go` runtime guard rejects any `/licensed/` path before embedding.
- Unit test covers the guard.
- Fresh clone without `licensed/` populated → PDF compile still succeeds with Libertinus fallback.
- (User-side smoke) Plantin-embedded PDF verifies via `pdffonts`; EPUB shows zero Plantin presence.

## Wrap-up

1. Single commit (cohesive feature):
   ```
   TRK-DESIGN-002: print-only licensed font wiring (gitignored licensed/, spec fields, EPUB guard)
   ```
2. Push to main. Deploy on VM (Go binary rebuild required).
3. TRACKER: close TRK-DESIGN-002. Update Resume here.
4. User drops licensed font files locally; runs `scripts/sync-licensed-fonts.sh`; sets spec values; verifies print PDF.

## Non-goals

- **Don't touch EPUB CSS to reference licensed families.** EPUB stays OFL-only.
- **Don't commit any OTF/TTF licensed files.** Even accidentally — `git status` check before commit is essential.
- **Don't wire `Dockerfile`.** The VM build is from source; no container yet.
- **Don't auto-detect font family names.** User runs `fc-query` once and pastes into the spec field. Auto-detection adds complexity for a once-per-family operation.

## Pitfalls

- **Family name mismatch.** Typst looks up by the font's internal name table, which may differ from the filename. `fc-query` is the authoritative source. Typing "Plantin MT" instead of "Plantin MT Pro" silently falls back to the next family in the list.
- **Subsetting.** Typst subsets by default in PDF output. Don't disable; subset embedding is the licensed mode.
- **Cache invalidation.** If Typst caches fonts in a system location (`~/.cache/typst/`), a font swap may need a cache clear. Unlikely on the VM but possible.
- **Multilingual fallback chain.** Make sure CJK + Thai stay in the fallback list AFTER the licensed family. Plantin has no CJK glyphs; if a Ghosts chapter contains Japanese characters, Typst needs to fall through to Noto.
- **EPUB CSS leakage.** If a future ticket adds Plantin to the CSS as a "premium fallback for capable readers," the guard becomes incomplete. The `srv/epub.go::buildCSS` literal strings are the single source of truth — keep them OFL-only.
- **License paperwork.** Confirm before this lands that the user's Plantin/Proxima licenses are perpetual desktop (not subscription). If subscription, the license terminates with the subscription and archival PDFs become questionable.
