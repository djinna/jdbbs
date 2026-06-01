# Manuscript pipeline — local vs VM

How a manuscript becomes an **EPUB + print PDF**, where each step runs, and what it
takes to do it locally vs on the VM (`jdbbs.exe.xyz`).

> **This is a different system from [`DEV-SETUP.md`](DEV-SETUP.md).** That doc sets up the
> **Go server** (`cmd/`, `srv/`) — the web app at `jdbbs.exe.xyz` that *manages* book
> projects. **This** doc is the **typesetting pipeline** that actually *produces* the
> books. Setting up the Go toolchain does **not** set up this pipeline.

## At a glance

Inputs live in `manuscripts/<book>/` (Word `.docx`, Markdown `.md`, or hand-authored
Typst `.typ`). Two output targets, two toolchains:

- **Print PDF** — `pandoc` (DOCX/MD → Typst) → `typst compile` (Typst → PDF), using the
  templates in `typesetting/templates/` and fonts in `typesetting/fonts/`. Real books
  (e.g. `ghosts`) are authored directly in Typst — `manuscripts/ghosts/main.typ`
  `#include`s per-chapter `.typ` files — and compiled with `typst` directly.
- **EPUB** — `pandoc` (DOCX/MD → EPUB3) with `typesetting/templates/epub/epub-styles.css`.
  No Typst. Fonts not embedded by default (set `EPUB_EMBED_FONTS=1` to embed).

### What each step needs

| Step | Tool |
|------|------|
| DOCX/MD → Typst, DOCX/MD → EPUB | `pandoc` |
| Typst → PDF | `typst` + the fonts in `typesetting/fonts/` |
| DOCX title extraction, edge-case detection, Word-template generation | `python3` + `python-docx` |

## The scripts (`typesetting/scripts/`)

| Command | Does |
|---------|------|
| `./build.sh full <in.docx> [decisions.json]` | DOCX → Typst → PDF (edge-case aware) |
| `./docx2pdf.sh <in.docx> [out.pdf]` | DOCX → PDF (simple: pandoc → typst) |
| `./md2pdf.sh <in.md> [out.pdf]` | Markdown → PDF |
| `./docx2epub.sh <in.docx> [out.epub]` | DOCX → EPUB3 |
| `./md2epub.sh <in.md> [out.epub]` | Markdown → EPUB3 |
| `./build-ghosts.sh` | Build the *Ghosts* anthology (real-book example) |
| _direct_ | `typst compile --root . --font-path typesetting/fonts manuscripts/<book>/main.typ output/<book>.pdf` |

Output lands in `output/`.

## Local vs VM — where it runs today

The **VM (`jdbbs.exe.xyz`)** has the whole pipeline assembled (pandoc, typst,
python-docx, all fonts) and is the **canonical / parity / deploy** environment. Real
books are built there. Per [`../CLAUDE.md`](../CLAUDE.md) you run ssh yourself:

```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && typst compile --root . --font-path typesetting/fonts manuscripts/ghosts/main.typ /tmp/ghosts.pdf' && scp exedev@jdbbs.exe.xyz:/tmp/ghosts.pdf scratch/
```

### What works locally right now (MBP16)

Present: `pandoc`, `python3`, all fonts (including the licensed `plantinMTpro` /
`proximanova`, which sit locally as gitignored files). Missing: `typst`, `python-docx`.

| Path | Local now? | Gap |
|------|:----------:|-----|
| Markdown → EPUB | ✅ | — |
| DOCX → EPUB | ⚠️ | `python-docx` (title extraction) |
| Markdown → print PDF | ⚠️ | `typst` |
| DOCX → print PDF (simple) | ⚠️ | `typst` |
| DOCX → print PDF (`build.sh full`) | ⚠️ | `typst` + `python-docx` |
| Authored-Typst book (e.g. ghosts) → PDF | ⚠️ | `typst` (fonts + source already present) |

### Going fully local — two installs

```bash
brew install typst                       # the print-PDF typesetter
python3 -m venv ~/.venvs/jdbbs && source ~/.venvs/jdbbs/bin/activate && pip install python-docx
```

Then the scripts above run locally. Caveats:

- **Fonts are licensed and gitignored.** `plantinMTpro` / `proximanova` exist on the
  MBP16 and the VM but do **not** travel via `git`. A fresh machine needs them copied in
  manually, or it builds print PDFs on the VM only.
- **Parity.** The VM is the canonical build environment; local `typst`/`pandoc` versions
  can differ. For final, print-ready output, build (or re-verify) on the VM.
- **The venv must be active.** The Python scripts call a bare `python3`, so
  `source ~/.venvs/jdbbs/bin/activate` in the shell where you run them.

## Recommendation

- **EPUB** (especially from Markdown) is cheap and safe to do locally — go for it now.
- **Print PDF** locally is great for **drafts and proofing** once `typst` is installed,
  but treat the **VM as the source of truth** for final, print-ready PDFs (font parity +
  the canonical build).
