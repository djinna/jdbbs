# Session prompt — TRK-DESIGN-004 (CJK/Thai font bundling)

> Use this as the kick-off prompt for a fresh Claude Code session.
> Concurrent-safe with TRK-DEV-009 (per-chapter EPUB author — touches
> `srv/epub.go`, `srv/bookspecs.go`, `srv/static/admin.html` EPUB section).
> Do NOT pick up TRK-DESIGN-003 in this session — they share
> `epub-styles.css` and will merge-conflict.

Run `jpull` first. Then standard pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  fc-list | grep -iE "(libertinus|source sans|jetbrains|noto|hiragino|thonburi)" | head -20'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `prodcal active`, font list (Libertinus + Source Sans + JetBrains Mono bundled in repo; Hiragino/Thonburi possibly present as OS fonts on macOS-derived images but NOT on Linux VM; Noto likely absent), `HTTP/2 200`.

## What you're building

The Ghosts parity audit (2026-05-26, `docs/GHOSTS_PARITY_2026-05-26.md`) flagged that Ghosts has CJK and Thai content (the Khlongs chapter and other multilingual passages). The reference InDesign PDF embeds HiraKakuPro-W3 (Japanese) and Thonburi (Thai). Current state:

- **Typst:** no CJK/Thai font bundled in `typesetting/fonts/` — would fall back to OS defaults. Works inconsistently across hosts; on the Linux VM, most CJK/Thai glyphs render as tofu.
- **EPUB CSS:** `epub-styles.css` lines ~344-353 declare `.chinese` / `.thai` classes referencing Hiragino/Thonburi by name with fallback stacks. Reader rendering depends on whether the e-reader has the fonts installed.

Goal: bundle OFL-licensed CJK + Thai fonts (Noto Serif CJK JP + Noto Sans Thai recommended), update both Typst and EPUB to use them as primary with the original commercial fonts as upstream fallbacks. Same pattern as the Libertinus bundle (`d451aa4`, TRK-MIG-007).

Full ticket: `docs/TRACKER.md` → `TRK-DESIGN-004`.

## Implementation steps

### 1. Audit Ghosts content (~10 min, read-only)

Confirm what's actually needed before downloading multi-megabyte fonts:

```bash
cd ~/jd-projects/jdbbs/manuscripts/ghosts
# Scan for CJK ranges
python3 -c "
import unicodedata, glob
for f in sorted(glob.glob('*.md')):
    with open(f) as fh:
        text = fh.read()
    cjk = sum(1 for ch in text if '一' <= ch <= '鿿')
    hiragana = sum(1 for ch in text if '぀' <= ch <= 'ゟ')
    katakana = sum(1 for ch in text if '゠' <= ch <= 'ヿ')
    thai = sum(1 for ch in text if '฀' <= ch <= '๿')
    if cjk or hiragana or katakana or thai:
        print(f'{f}: cjk={cjk} hira={hiragana} kata={katakana} thai={thai}')
"
```

Note what you find — drives whether you need CJK JP (Japanese: hiragana + katakana + some kanji), CJK SC (Simplified Chinese), CJK TC (Traditional Chinese), or all three. Thai is its own.

### 2. Pick + download fonts (~15 min)

**Recommended:** Noto Serif CJK + Noto Sans Thai (Google / Adobe, OFL licensed).

- Noto Serif CJK: https://github.com/notofonts/noto-cjk/releases — grab the "OTF/Japanese" zip (or whatever variant matches the content). Each variant is ~6-15MB.
- Noto Sans Thai (or Noto Serif Thai if available): https://github.com/notofonts/thai/releases.

Drop into `typesetting/fonts/noto/{Japanese,Thai}/` mirroring the existing `libertinus/OTF/` shape. Include the `OFL.txt` license file alongside each family.

If audit found ONLY Japanese, skip SC/TC. If audit found ONLY Thai, skip CJK entirely. Don't bundle fonts the manuscript doesn't need.

### 3. Wire into Typst (`typesetting/templates/series-template.typ`, ~30 min)

The current font-fallback in the template is short. Typst's `text()` `font:` argument accepts a list — multiple families tried in order. Update the body font setup to include the bundled fonts:

```typst
#set text(
  font: (
    "Libertinus Serif",
    "Noto Serif CJK JP",  // bundled CJK fallback
    "Noto Sans Thai",     // bundled Thai fallback
  ),
  size: ...,
)
```

Path resolution: `srv/books.go::runConversion` already invokes `typst compile --font-path typesetting/fonts` (recursive scan). New fonts in `typesetting/fonts/noto/` will be discovered automatically — no code change in Go.

Smoke this with a tiny .typ file containing CJK + Thai before touching the full Ghosts manuscript:

```bash
cat > /tmp/font-smoke.typ <<'EOF'
#set text(font: ("Libertinus Serif", "Noto Serif CJK JP", "Noto Sans Thai"))
English text.
日本語のテキスト.
ภาษาไทย.
EOF
typst compile --font-path ~/jd-projects/jdbbs/typesetting/fonts /tmp/font-smoke.typ /tmp/font-smoke.pdf
open /tmp/font-smoke.pdf  # macOS; or use a PDF viewer
```

All three lines should render correctly.

### 4. Wire into EPUB (`typesetting/templates/epub/epub-styles.css`, ~20 min)

Find the existing `.chinese` and `.thai` rules (around lines 344-353 per the parity doc). Two changes:

- Update the `font-family` to list the bundled Noto families first, then the original commercial fonts as fallbacks for systems that have them.
- Add `@font-face` declarations at the top of the file pointing at the bundled OTF files (relative path — they need to ship inside the EPUB).

The pandoc EPUB pipeline (`srv/epub.go`) already passes `--css=typesetting/templates/epub/epub-styles.css`. But for `@font-face` to work the font files need to be **packaged into the EPUB**. Check the existing pandoc invocation — if it uses `--epub-embed-font`, just add the new font file paths. If not, add them. Each `--epub-embed-font /path/to/file.otf` flag adds one font to the package.

After change, an EPUB will be ~5-15MB larger per CJK family, ~1-2MB larger for Thai. That's the cost of self-contained rendering.

### 5. Smoke test EPUB (~15 min)

Same temporary multi-lingual content as the Typst smoke, but go through the EPUB path. Compile, unzip the .epub, verify:

- Font files are present in the EPUB package
- CSS references the bundled families
- Opening in Calibre on a CLEAN device (or a device known not to have Hiragino) shows correct rendering

### 6. Acceptance checks

- Typst compile of a CJK + Thai sample produces visually correct glyphs (no tofu).
- EPUB compile of the same content embeds the bundled fonts.
- Opening the EPUB on a device without Hiragino/Thonburi (or in Calibre with bundled-font priority) renders correctly.
- License files (`OFL.txt`) shipped alongside each font family.
- `Dockerfile` (if it installs `fonts-libertinus` or similar) doesn't need to change — Typst font path is repo-relative.

## Deploy

```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && \
  go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal && \
  sleep 2 && systemctl is-active prodcal'
curl -sI https://jdbbs.exe.xyz | head -1

# Verify the bundled fonts arrived
ssh exedev@jdbbs.exe.xyz 'ls -la /home/exedev/prodcal/typesetting/fonts/noto/'
```

Don't touch the systemd unit (TRK-OPS-005). Push to `main` yourself.

## Wrap-up

1. `docs/TRACKER.md`: mark TRK-DESIGN-004 done with the font choice + file paths. Update Resume here.
2. If a TRACKER conflict on pull, it's almost certainly the parallel DEV-009 session — accept both sides (different ticket sections).
3. Note in TRACKER whether the audit (step 1) showed Japanese-only / Thai-only / multiple CJK variants needed — informs future titles.
4. Commit, push, deploy.

## Non-goals

- **Don't change Plantin/Proxima** (TRK-DESIGN-002 covers commercial-font bundling; user has licenses but distribution rights need separate verification).
- **Don't start TRK-TEST-002 (live Ghosts regression)** — that's the natural next step but a separate ticket.
- **Don't touch `srv/*.go` or admin.html** — DEV-009 (parallel session) is in those zones.
- **Don't bundle Hiragino or Thonburi directly** — they're Apple system fonts; distribution outside macOS is typically not permitted. OFL Noto fonts are the right call.

## Concurrent-work awareness

At session-start time, **TRK-DEV-009 (per-chapter EPUB author)** is likely running in another fresh Claude Code session on this Mac. They're touching:
- `srv/epub.go` (post-pandoc XHTML byline injection)
- `srv/bookspecs.go` (epubSpec struct + parsing)
- `srv/static/admin.html` (EPUB section: Chapters editor)

Zero overlap with your zone (`typesetting/fonts/`, `typesetting/templates/`). The only shared file is `docs/TRACKER.md` for the close commits — if your `git pull` hits a conflict there, rebase and accept both sides (different ticket sections).
