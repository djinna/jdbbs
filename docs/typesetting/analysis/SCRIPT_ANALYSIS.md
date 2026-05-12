# Scripts Directory Analysis

Generated analysis of all scripts in `/home/exedev/book-production/scripts/`.

---

## 1. `build.sh` (5,955 bytes, executable)

**Purpose:** Central build orchestrator / CLI dispatcher. Provides subcommands for the full book production pipeline: `convert`, `review`, `compile`, `epub`, `preview`, `full`, `help`.

**Inputs:**
- Word `.docx` manuscripts
- Optional edge-case decision JSON files (from the review step)
- Typst `.typ` source files (for compile/preview)

**Outputs:**
- Typst `.typ` files (from convert)
- PDF files (from compile)
- HTML + JSON edge-case review reports (from review)

**Dependencies:**
- `pandoc` (with Typst writer)
- `typst` compiler
- `python3` (inline Python script for TSV generation + calls to `detect-edge-cases.py`)
- `scripts/docx-to-typst-enhanced.lua` (Pandoc Lua filter)
- `scripts/detect-edge-cases.py`

**Error Handling: GOOD**
- Uses `set -e` for fail-fast
- Validates input file existence before operations
- Validates decisions file existence
- Color-coded log/warn/error functions; `error()` exits with code 1
- Cleans up temp files (`rm -f` on the TSV map file)

**Notable Issues:**
1. The `convert()` function embeds an inline Python script (heredoc `PY`) to transform JSON decisions into a TSV map file. This is clever but fragile — hard to test/debug independently.
2. `compile_epub()` is a stub — just prints a suggestion to use pandoc directly. The `warn "EPUB export is experimental"` is honest but the function doesn't actually produce output.
3. The `full` pipeline only does docx→typ→pdf. No EPUB in the full pipeline.
4. The inline Python uses `sys.argv[1], sys.argv[2]` positional args passed through `python3 -` which is correct but unusual.
5. `EDGE_DECISIONS_FILE` env var support is a nice touch for CI/scripting.

**Code Quality: HIGH** — Well-structured dispatcher pattern, good help text, clean argument handling.

---

## 2. `build-ghosts.sh` (4,645 bytes, executable)

**Purpose:** Project-specific build script for "Ghosts in Machines" anthology. Converts Word docs to markdown, copies images from a reference EPUB extraction, and generates a hardcoded `main.typ` Typst file with full front matter, TOC, and chapter includes.

**Inputs:**
- Word docs from `manuscripts/ghosts/8000 MS/*.docx`
- Images from `reference/ghosts_epub/OEBPS/image/`

**Outputs:**
- Markdown files in `src/ghosts/*.md`
- `src/ghosts/main.typ` (the master Typst document)

**Dependencies:**
- `pandoc` (docx→markdown)
- `typst` (mentioned in final echo but not invoked)

**Error Handling: FAIR**
- Uses `set -e` for fail-fast
- No validation of input directory/file existence
- No check that `reference/ghosts_epub/` exists before `cp`

**Notable Issues:**
1. **Typo in source filename preserved intentionally:** `ghostts_08_LOYALTY.jpg` (double 't') — this matches the actual filename in the reference EPUB, so it's correct but confusing.
2. The script generates `main.typ` but does NOT create the individual chapter `.typ` files (e.g., `00-intro.typ`, `01-soda.typ`). The final echo says "Next: Create individual chapter .typ files from the markdown" — so this is an incomplete pipeline.
3. **Hardcoded page numbers** in the TOC entries (e.g., `"9"`, `"17"`, `"31"`). These will be wrong if content changes.
4. The generated Typst file uses a relative import `../../templates/series-template.typ` which is fragile.
5. Image paths in the `.typ` reference `ghosts_00_SBAcover.png` without the `src/ghosts/` prefix — this works because it's an `#include` from within `src/ghosts/`.
6. Missing Chapter 0 intro image copy — copies numbered images but the intro image `ghosts_00_SBAcover.png` must already exist.
7. The `for doc in manuscripts/ghosts/8000\ MS/*.docx` glob uses escaped space — correct but brittle.

**Code Quality: MEDIUM** — Functional for its specific purpose but not generalizable. Essentially a one-shot project script.

---

## 3. `detect-edge-cases.py` (44,620 bytes)

**Purpose:** Comprehensive Word document analyzer that detects manual/local formatting that might need human review before automated conversion. Generates a polished HTML report and optional JSON output.

**Inputs:**
- Word `.docx` file (positional arg)
- Optional `--declared-styles` JSON file listing known/expected custom styles
- Optional `-o` output path, `--json` flag

**Outputs:**
- HTML report (styled with IBM Plex fonts, CSS grid lines, severity badges)
- JSON file (when `--json` flag used, same path with `.json` extension)

**Dependencies:**
- `python-docx` (`from docx import Document`)
- Standard library: `argparse`, `json`, `re`, `unicodedata`, `html`, `datetime`, `pathlib`

**Detection Categories (10 detectors):**
1. `detect_manual_formatting()` — bold/italic/underline applied directly (not via style)
2. `detect_unusual_fonts()` — non-standard fonts with emoji aggregation
3. `detect_colored_text()` — non-black text color, highlight colors
4. `detect_manual_lists()` — bullet/numbered/lettered lists via text patterns
5. `detect_manual_breaks()` — section breaks via `* * *`, `---`, etc.
6. `detect_mixed_styles()` — paragraphs with multiple fonts/sizes
7. `detect_direct_formatting()` — spacing, indentation, alignment overrides
8. `detect_image_inventory()` — inline images with dimensions
9. `detect_observed_styles()` — Word styles not in the declared set
10. `detect_special_typography()` — ASCII art / preformatted blocks
11. `detect_language_scripts()` — Thai, CJK, non-Latin text

**Error Handling: FAIR**
- No try/except around Document loading (will crash with traceback on bad file)
- No validation that input file exists before opening
- The `_is_from_style()` method is a stub that always returns `False` — this means ALL bold/italic is flagged as "manual", producing many false positives

**Notable Issues:**
1. **`_is_from_style()` always returns False** — This is the biggest issue. Every bold/italic run is reported as "manual formatting" even when it comes from a character style. The comment says "simplified check" but it's actually non-functional.
2. The HTML report is ~300 lines of inline CSS with a sophisticated design (noise texture, grid lines, IBM Plex fonts via Google Fonts). This is impressive but the entire HTML is built via string concatenation — no template engine. Makes maintenance harder.
3. `_best_declared_style_match()` only checks for styles containing 'ascii' — very crude heuristic.
4. The `_should_skip_paragraph()` method hardcodes specific example text strings (like "Shall I compare thee to a summer's day?") to skip template guide content. This is brittle.
5. `detect_manual_lists()` and `detect_manual_breaks()` import `re` inside the method body despite it being imported at module level.
6. Type hint `List[Dict] | None` uses Python 3.10+ union syntax — won't work on Python 3.9.
7. No `if __name__ == '__main__'` guard around the class definitions, but the main() call is guarded.

**Code Quality: HIGH** — Very thorough detection logic, well-organized class structure, good HTML output. The `_is_from_style()` stub is the main functional gap.

---

## 4. `docx-to-typst-enhanced.lua` (16,760 bytes)

**Purpose:** Pandoc Lua filter that maps Word paragraph/character styles to Typst function calls, with edge-case decision support (keep/strip/convert) for manual lists, colored text, and highlighted text.

**Inputs:**
- Pandoc AST (from Word docx)
- Optional metadata: `edge-decisions-map-file` (TSV) or `edge-decisions-file` (JSON)

**Outputs:**
- Modified Pandoc AST with Typst raw blocks/inlines

**Dependencies:**
- Pandoc (Lua filter API)
- Typst writer (`-t typst`)

**Style Mappings:**
- Character: smallcaps→`#sc`, booktitle→`#booktitle`, foreign→`#foreign`, tracked→`#tracked`, allcaps→`#allcaps`, metadata-c→`#metadata-c`
- Paragraph: sectionbreak→`#section-break`, firstparagraph→`#first-para`, epigraph→`#epigraph`, poem→`#poem`, tweet-p→`#tweet-p`, metadata-p→`#metadata-p`, tweet-p-ascii→`#tweet-p-ascii`

**Edge Case Decision System:**
- Loads decisions from TSV map file (generated by build.sh) or JSON file
- Supports three actions per finding: `keep` (preserve as-is), `strip` (remove), `convert` (transform to proper Typst list)
- `apply_strip_substrings()` — removes strip-decided colored text inline
- `normalize_converted_list_blocks()` — coalesces consecutive converted list items into a single Typst list block

**Error Handling: FAIR**
- Writes warnings to stderr for missing decision files
- Uses `pcall` for JSON parsing (safe)
- No error on malformed TSV lines (silently skips)

**Notable Issues:**
1. The `Span` function for the enhanced version preserves original inline nodes (good), but the original `docx-to-typst.lua` stringifies content (loses formatting). This is a deliberate improvement.
2. `BlockQuote` handler stringifies content via `pandoc.utils.stringify()` — this loses all inline formatting (bold, italic, links) inside blockquotes. Should use the approach from `Span` (preserving AST nodes).
3. The `Para` function's section break detection is thorough (handles `* * *`, `***`, `---`, `———`, ornament chars `❦❧⁂§`) but the regex `^[*%-–—]+%s*[*%-–—]+%s*[*%-–—]+$` is overly permissive — would match `*-*` or `—*—`.
4. The filter adds a `#import "/templates/series-template.typ": *` header with absolute-ish path using `/templates/` (Typst root-relative). This requires `--root .` when compiling.
5. Manual list detection in `Para` uses heuristic patterns — could false-positive on paragraphs starting with "A. Lincoln said..."
6. The `normalize_converted_list_blocks` post-processor is sophisticated but only handles `+` and `-` markers — no nested list support.

**Code Quality: HIGH** — Well-structured with clear separation of concerns. The decision system is well-designed.

---

## 5. `docx-to-typst.lua` (4,827 bytes)

**Purpose:** Original/simpler version of the Pandoc Lua filter. Maps Word styles to Typst functions without edge-case decision support.

**Inputs/Outputs:** Same as enhanced version but simpler.

**Dependencies:** Pandoc Lua filter API.

**Notable Issues:**
1. **Content stringification bug:** `Span` handler does `pandoc.utils.stringify(el.content)` which strips all nested formatting. If a span contains bold text inside a custom style, the bold is lost. The enhanced version fixes this.
2. **Same BlockQuote bug** as enhanced version — stringifies content.
3. Uses relative import path `"templates/series-template.typ"` (no leading `/`) vs. enhanced version's `"/templates/series-template.typ"`. Inconsistency.
4. No edge-case handling, no manual list detection, no section break pattern matching in Para.
5. This appears to be the v1 that was superseded by `docx-to-typst-enhanced.lua`.

**Error Handling: POOR** — No error handling at all.

**Code Quality: MEDIUM** — Clean but limited. Should probably be removed or marked as deprecated since the enhanced version exists.

---

## 6. `docx2epub.sh` (2,121 bytes, executable)

**Purpose:** Converts Word `.docx` to EPUB3 using Pandoc directly (no Typst intermediate).

**Inputs:**
- Word `.docx` file (required)
- Optional output path
- `EPUB_TITLE` env var (or auto-extracted from docx)
- `EPUB_EMBED_FONTS` env var (optional, default 0)

**Outputs:** EPUB3 file.

**Dependencies:**
- `pandoc` (docx→epub3)
- `python3` + `python-docx` (for title extraction)
- CSS: `templates/epub/epub-styles.css`
- Optional: woff2 fonts in `fonts/sourcesans/WOFF2/` and `fonts/jetbrainsmono/fonts/webfonts/`

**Error Handling: GOOD**
- `set -e` for fail-fast
- Validates argument count
- Title extraction has try/except fallback to basename
- Font embedding checks directory existence before globbing

**Notable Issues:**
1. Title extraction uses an inline Python heredoc — uses `raise SystemExit(0)` instead of `sys.exit(0)` which is unusual but functionally equivalent.
2. Does NOT use the Lua filter — goes straight docx→epub3 via Pandoc. This means Word style mappings are not applied.
3. No `--lua-filter` option means custom styles (tweet-p, metadata-c, etc.) won't be handled.
4. The `ls -la "$OUTPUT"` at the end is a nice touch for verification.

**Code Quality: GOOD** — Clean, well-documented font policy.

---

## 7. `docx2pdf.sh` (1,545 bytes, executable)

**Purpose:** Converts Word `.docx` to PDF via Pandoc→Typst→PDF pipeline (without the Lua filter).

**Inputs:** Word `.docx` file, optional output path.

**Outputs:** PDF file (also leaves intermediate `.typ` file in output/).

**Dependencies:**
- `pandoc` (docx→typst)
- `typst` compiler
- Fonts: `fonts/sourcesans/OTF`, `fonts/jetbrainsmono/fonts/ttf`

**Error Handling: FAIR**
- `set -e`
- Validates argument count
- No validation that input file exists

**Notable Issues:**
1. **Does NOT use the Lua filter** — runs plain `pandoc -f docx -t typst` without `--lua-filter`. This means all custom Word styles are ignored.
2. **Template import path mismatch:** Uses `#import "../templates/series-template.typ"` (relative) but the enhanced Lua filter uses `#import "/templates/series-template.typ"` (root-relative). These will break depending on where the .typ file lives.
3. The intermediate `.typ.body` file is created then concatenated and deleted — good cleanup.
4. Hardcodes title as `"Untitled"` — no extraction from docx metadata.
5. The comment `#let horizontalrule = section-break` is syntactically a Typst `let` binding, but Pandoc's Typst writer emits `#line()` for `<hr>`, not `#horizontalrule` — so this binding has no effect.
6. Leaves the intermediate `.typ` file behind (unlike `.typ.body` which is cleaned up).

**Code Quality: MEDIUM** — Functional but the non-functional `horizontalrule` binding and missing Lua filter reduce its usefulness. Likely superseded by `build.sh convert`.

---

## 8. `md2pdf.sh` (1,228 bytes, executable)

**Purpose:** Converts Markdown to PDF via Pandoc→Typst→PDF pipeline.

**Inputs:** Markdown `.md` file, optional output path.

**Outputs:** PDF file.

**Dependencies:** `pandoc`, `typst`, fonts.

**Error Handling: FAIR** — Same as docx2pdf.sh.

**Notable Issues:**
1. Same `#let horizontalrule = section-break` non-functional binding.
2. Same relative import path issue.
3. Hardcodes title as `"Story Collection"`.
4. Leaves intermediate `.typ` file.
5. No Lua filter (less relevant for markdown input, but still means no custom style handling).

**Code Quality: MEDIUM** — Nearly identical to `docx2pdf.sh`. Could be refactored to share code.

---

## 9. `md2epub.sh` (2,037 bytes, executable)

**Purpose:** Converts Markdown to EPUB3 using Pandoc.

**Inputs:**
- Markdown `.md` file (required)
- Optional: output path, `--title`, `--author` flags
- `EPUB_EMBED_FONTS` env var

**Outputs:** EPUB3 file.

**Dependencies:** `pandoc`, CSS template, optional woff2 fonts.

**Error Handling: FAIR**
- `set -e`
- Validates argument count
- Silent skip of unknown args (`*) shift ;;`) — could hide mistakes

**Notable Issues:**
1. The arg parser uses a while loop with `shift` — positional output arg detection (`*.epub`) is done by glob pattern which could match unintended args.
2. `TITLE` defaults to `$BASENAME` even if not provided — no way to produce an EPUB without a title.
3. Nearly identical font embedding logic to `docx2epub.sh` — code duplication.

**Code Quality: GOOD** — Clean arg parsing, well-documented font policy.

---

## 10. `generate-word-template.py` (19,411 bytes, executable)

**Purpose:** Generates a styled Word `.docx` template from a book specification JSON. The template serves as a starting point for copyeditors, with all paragraph/character styles pre-configured to match the Typst production template.

**Inputs:**
- Book spec JSON (from stdin or `--spec-file`)
- JSON schema: `{ metadata: {title, author, publisher}, typography: {body_font, heading_font, code_font, base_size_pt, leading_pt, paragraph_indent_em, justify}, headings: {h1_size_em, h1_weight, ...}, elements: {blockquote_style, code_block_size_em, section_break, poem_size_em}, custom_styles: [{word_style, name, type, description}] }`

**Outputs:**
- Word `.docx` file (to `--output` path or stdout)
- Contains: Normal, First Paragraph, Heading 1-3, Block Quote, Code Block, Section Break, Verse, Copyright, Epigraph styles + any custom styles from spec
- Includes a "Template Guide" section with sample content demonstrating each style

**Dependencies:**
- `python-docx` (`from docx import Document`)
- Standard library: `argparse`, `io`, `json`, `sys`

**Error Handling: FAIR**
- No validation of spec JSON structure (missing keys will raise KeyError)
- No try/except around file I/O
- `_weight_to_bold()` has numeric fallback with try/except — good
- Duplicate style names are checked (`if cs_name in [s.name for s in doc.styles]`)

**Notable Issues:**
1. **The style name check iterates all styles into a list on every custom style** — O(n²) but irrelevant at typical scale.
2. `bool | None` type hint on `_weight_to_bold` return type — Python 3.10+ only.
3. The `_ensure_font_exists()` sets `eastAsia` font via raw XML manipulation (`qn("w:eastAsia")`) — correct but fragile if python-docx internals change.
4. Section break character lookup (`SECTION_BREAK_CHARS`) is a nice design — supports breve, asterism, dinkus, fleuron, rule.
5. The sample content in the template guide includes specific example text that `detect-edge-cases.py` skips via `_should_skip_paragraph()` — these are coordinated.
6. Output to stdout uses `sys.stdout.buffer.write()` — correct for binary .docx data.
7. Custom character styles are created but no font/formatting is applied to them — they're bare stubs.

**Code Quality: HIGH** — Clean, well-organized helper functions, good separation of style creation from sample content.

---

## 11. `md-to-chapter.py` (9,169 bytes, executable)

**Purpose:** Converts a Pandoc-generated markdown file (from docx with `+styles +fenced_divs`) into a Typst chapter file. Handles custom paragraph styles, footnotes, inline formatting, and section breaks.

**Inputs:**
- Markdown file (from `pandoc --from=docx+styles --to=markdown+fenced_divs`)
- Title string (argv[2])
- Author string (argv[3])

**Outputs:** Typst `.typ` file (to stdout).

**Dependencies:** Python 3 standard library only (no external packages).

**Supported Styles:**
- First Paragraph, Verse, Code Block/Code-Block, Block Quote/Block-Quote, Epigraph, Section Break/Section-Break

**Error Handling: FAIR**
- Validates argc (prints usage to stderr, returns 1)
- No try/except around file I/O
- Unknown styles fall through to plain paragraph rendering (resilient)

**Notable Issues:**
1. **Hardcoded absolute import path:** `#import "/home/exedev/book-production/templates/series-template.typ": *` — This won't work on any other machine or deployment. Should use a relative or root-relative path.
2. **Hardcoded font settings:** `#set text(font: "Libertinus Serif", size: 10pt)` — duplicates what the template already sets. If typography changes in the spec, this file must be manually updated.
3. `escape_typst_text()` only escapes `#` and `$` — misses other Typst special chars like `@`, `<`, `>`, backticks in certain contexts.
4. Bold/italic conversion: `**text**` → `*text*` (Typst bold), `*text*` → `_text_` (Typst italic). The italic regex uses negative lookbehind `(?<![\\*])` which is correct but the bold must be processed first to avoid conflicts — and it is.
5. Footnote extraction parses continuation lines (indented with 4 spaces or tab) — good.
6. `render_segment()` for 'code-block' wraps in both `#code-block[` and triple backticks — this assumes the Typst template defines a `code-block` function that expects raw content.
7. The YAML frontmatter stripping is fragile — looks for `---` at position 0, then finds the next `---` and strips everything before it.
8. Strips leading headings (H1/H2) from content — assumes title/author come from argv, not from the markdown.

**Code Quality: MEDIUM-HIGH** — Good regex-based parsing, clean segment rendering, but the hardcoded paths/fonts are problematic.

---

## 12. `series-template-manager.py` (15,737 bytes)

**Purpose:** Manages series-level template versioning and book registration. Provides CRUD operations for series, template version history, book-to-series association, and template comparison.

**Inputs:**
- CLI subcommands: `list`, `create`, `info`, `get-template`
- Book spec JSON files
- Optional Word template `.docx` files

**Outputs:**
- `series_registry.json` (persistent registry)
- Versioned template spec JSON files
- Copied template .docx files

**Dependencies:** Python 3 standard library only (`json`, `shutil`, `pathlib`, `datetime`, `hashlib`).

**Error Handling: FAIR**
- Raises `ValueError` for missing series/versions
- No try/except around file I/O in `_load_registry()`, `_save_registry()`
- No CLI error handling (argparse handles basic validation)

**Notable Issues:**
1. **Default base_path is wrong:** `"./book-production/series-templates"` — this is relative to CWD and assumes you're running from the parent of `book-production/`. Should probably be relative to the script's location.
2. **No file locking** — concurrent access to `series_registry.json` could corrupt it.
3. **Template version format** uses `datetime.now().strftime("%Y-%m-%d %H%M")` with a space — filenames become `template spec 2025-04-14 1530.json`. Spaces in filenames can cause issues.
4. `_calculate_checksum()` uses SHA-256 truncated to 16 hex chars — adequate for integrity checking but not collision-resistant at that length.
5. `_compare_dicts()` doesn't handle list comparisons — if a spec has a list value, it only reports "changed" without showing which list elements differ.
6. The `SeriesTemplateAPI` class wraps `SeriesTemplateManager` but `apply_series_template_to_book()` does a shallow `.copy()` on the spec dict — nested dicts (metadata, typography) are still shared references, so mutations would affect the original.
7. The CLI only exposes 4 of the many available operations (no `add-book`, `save-template`, `compare` subcommands).
8. No `--base-path` CLI option to override the default path.

**Code Quality: MEDIUM-HIGH** — Well-structured class hierarchy with registry pattern, but the shallow copy bug and hardcoded path are real issues.

---

## Cross-Cutting Observations

### Redundancy / Overlap
| Script | Superseded by | Notes |
|--------|--------------|-------|
| `docx-to-typst.lua` | `docx-to-typst-enhanced.lua` | Original has content stringification bug |
| `docx2pdf.sh` | `build.sh full` | Doesn't use Lua filter |
| `md2pdf.sh` | `build.sh compile` | Doesn't use Lua filter |
| `docx2epub.sh` | Could be integrated into `build.sh epub` | Currently independent |
| `md2epub.sh` | Could be integrated into `build.sh epub` | Currently independent |

### Path Inconsistencies
- `docx-to-typst.lua`: `#import "templates/series-template.typ"` (relative)
- `docx-to-typst-enhanced.lua`: `#import "/templates/series-template.typ"` (root-relative)
- `md-to-chapter.py`: `#import "/home/exedev/book-production/templates/..."` (absolute)
- `docx2pdf.sh` / `md2pdf.sh`: `#import "../templates/series-template.typ"` (relative up)

These are **four different import paths** for the same template. Only the enhanced Lua filter's root-relative path is portable (requires `--root .`).

### Python Version Compatibility
- `detect-edge-cases.py` and `generate-word-template.py` use `X | None` union syntax (Python 3.10+)
- `md-to-chapter.py` uses `from __future__ import annotations` (works on 3.7+)
- `series-template-manager.py` uses `Optional[str]` from typing (works on 3.7+)

### Shared Dependencies
- **python-docx**: Used by `detect-edge-cases.py`, `generate-word-template.py`, `docx2epub.sh` (inline)
- **pandoc**: Used by all shell scripts and both Lua filters
- **typst**: Used by `build.sh`, `docx2pdf.sh`, `md2pdf.sh`

### Missing Infrastructure
1. No `requirements.txt` or `pyproject.toml` for Python dependencies
2. No test suite for any script
3. No CI/CD configuration
4. No version pinning for pandoc/typst
