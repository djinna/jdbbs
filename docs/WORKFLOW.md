# Book Production Workflow

## Overview

```
Word (.docx)  →  Pandoc  →  Typst (.typ)  →  PDF / EPUB
     ↓                           ↑
  Styles Map              Style Functions
```

You edit in Word using consistent paragraph/character styles. Pandoc converts to Typst, mapping Word styles to Typst functions via a custom filter.

## 1. Word Style Conventions

Use these style names in Word (case-insensitive, Pandoc normalizes them):

### Paragraph Styles
| Word Style | Purpose | Typst Output |
|------------|---------|--------------|
| `Normal` or `Body Text` | Regular paragraphs | Default paragraph |
| `First Paragraph` | After headings/breaks | `#first-para[...]` |
| `Heading 1` | Chapter title | `= Title` |
| `Heading 2` | Section heading | `== Section` |
| `Block Quote` | Quoted passages | `#blockquote[...]` |
| `Code` or `Ok-computer` | Terminal/code | ``` code block ``` |
| `Poem` or `Verse` | Poetry | `#poem[...]` |
| `Epigraph` | Opening quotes | `#epigraph[...]` |
| `Section Break` | Scene dividers | `#section-break` |

### Character Styles
| Word Style | Purpose | Typst Output |
|------------|---------|--------------|
| `Emphasis` or `Italic` | Italics | `_text_` or `#ital[...]` |
| `Strong` or `Bold` | Bold | `*text*` or `#bold[...]` |
| `Small Caps` or `Acronym` | Acronyms (AI, USB) | `#sc[...]` |
| `Code Char` or `Monospace` | Inline code | `` `code` `` |
| `Book Title` | Book/film titles | `#booktitle[...]` |
| `Foreign` | Non-English phrases | `#foreign[...]` |

## 2. Conversion Pipeline

### Basic conversion (no style mapping)
```bash
pandoc manuscript.docx -o manuscript.typ
```

### With style-aware filter
```bash
pandoc manuscript.docx \
  --lua-filter=scripts/docx-to-typst.lua \
  -o manuscript.typ
```

### Full pipeline to PDF
```bash
# Convert and compile
pandoc manuscript.docx --lua-filter=scripts/docx-to-typst.lua -o src/book.typ
typst compile --font-path fonts/ src/book.typ output/book.pdf
```

## 3. Project Structure

```
book-production/
├── src/
│   └── book.typ          # Generated from Word, may hand-edit
├── manuscripts/
│   └── book.docx         # Your Word source
├── templates/
│   ├── series-template.typ
│   ├── styles.typ
│   └── images.typ
├── fonts/
├── images/               # Book images (covers, figures)
├── output/
│   ├── book.pdf
│   └── book.epub
└── scripts/
    ├── docx-to-typst.lua # Pandoc filter
    └── build.sh          # Build script
```

## 4. Iterative Workflow

1. **Edit in Word** - Apply styles consistently as you copyedit
2. **Convert** - Run pandoc to regenerate Typst
3. **Preview** - Compile PDF, check typography
4. **Adjust** - Either fix in Word and reconvert, or hand-edit Typst for final tweaks
5. **Repeat** until satisfied

### Tips
- Word is the "source of truth" during copyediting
- Once copyediting is done, Typst becomes the source for final layout tweaks
- Use Word's Style Inspector to verify consistent style application
- Keep a style guide document alongside your manuscript

## 5. Style Verification in Word

Before converting, verify style consistency:

1. **View → Navigation Pane** - Check heading structure
2. **Home → Styles → Style Inspector** - Verify paragraph/character styles
3. **Find & Replace → Format → Style** - Find unstyled or wrong-styled text
4. **Review → Compare** - Compare against previous version

### Common issues to check:
- Direct formatting instead of styles (italics applied manually vs. Emphasis style)
- Inconsistent acronym styling (some AI in small caps, some not)
- Missing first-paragraph style after headings
- Section breaks as empty paragraphs vs. proper style

## 6. Word Template Files

```
templates/word/
├── author-template.docx      # Clean starting template for authors
├── protocolized-style-guide.docx  # Detailed examples of each style
└── default-reference.docx   # Pandoc's reference doc (for debugging)
```

### For Authors

Send authors `author-template.docx` as a starting point. It contains:
- Pre-defined paragraph styles (First Paragraph, Body Text, Verse, etc.)
- Brief instructions in the document itself
- Correct structure for chapter openings

### Adding New Styles

If an author needs a style not in our standard set:

1. **Define in Word**: Create a new paragraph or character style with a clear name
2. **Document it**: Note what it's for and how it should look
3. **Update converter**: Add handling in `scripts/md-to-chapter.py`
4. **Update template**: Add the Typst function to `templates/series-template.typ`

Pandoc preserves custom styles automatically (with `+styles` extension), so the pipeline handles new styles gracefully—they just need corresponding Typst output.

## 7. Customizing for a New Book

1. Copy/adapt the template for book-specific needs
2. Add any new character styles to `styles.typ`
3. Update the conversion script to recognize new Word styles
4. Test with a sample chapter before full conversion
