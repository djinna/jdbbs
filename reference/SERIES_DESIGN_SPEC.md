# Protocolized Anthology Series - Design Specification

## Series Overview
Three fiction anthologies sharing consistent design:
1. **Ghosts in Machines** (136 pages)
2. **Terminological Twists** (96 pages)
3. **The Librarians** (80 pages)

## Page Dimensions
- **Trim size**: 353.811 × 546.567 pts (4.91" × 7.59")
- **Aspect ratio**: ~1:1.54 (close to trade paperback / A5)
- **Output**: PDF/X-4 (print-ready)

## Typography

### Fonts
| Role | Font | Weights |
|------|------|---------|
| Body | Plantin MT Pro | Regular, Italic |
| Headings | Proxima Nova | Bold, Semibold, Medium, Regular |
| Code/Terminal | Menlo | Regular |
| CJK (GHOSTS only) | Hiragino Kaku Gothic Pro | W3 |
| Thai (GHOSTS only) | Thonburi | Regular |

### Body Text
- Font: Plantin MT Pro Regular
- Size: ~10pt (0.833em relative)
- Line height: 1.2
- Alignment: Justified with hyphens
- First paragraph: no indent
- Subsequent paragraphs: 0.75em indent

### Chapter Titles
- Font: Proxima Nova Bold
- Size: 1.667em (~20pt)
- Case: Title case
- Position: Upper left of page

### Author Names (chapter openers)
- Font: Proxima Nova Medium
- Size: 1.333em (~16pt)

### Running Headers
- Font: Proxima Nova Medium
- Size: 0.75em (small caps style)
- Verso (even): PAGE NUMBER + AUTHOR NAME
- Recto (odd): STORY TITLE + PAGE NUMBER

### TOC
- "Contents" heading: Proxima Nova Semibold, 1.333em
- Story titles: Proxima Nova Bold, 0.833em
- Author names: Proxima Nova Medium, 0.833em, indented

### Page Numbers
- Font: Proxima Nova
- Position: Bottom center
- Roman numerals for front matter (i, ii, iii, vii)
- Arabic for body (1, 2, 3...)

## Special Elements

### Terminal/Code Blocks (class: Ok-computer)
- Font: Menlo Regular
- Size: 0.667em
- Line height: 1.5 (or 1.188 for multi-line)
- Left margin: 0.75em indent
- Hyphenation: disabled

### Section Breaks
- Three spaced dots, centered: ˘   ˘   ˘
- Vertical space: 0.5em above and below

### Multilingual Text (GHOSTS)
- Centered block
- Stacked: English / Chinese / Thai
- Each in appropriate font

### Poems/Verse
- Font: Plantin MT Pro Italic
- Extra top margin: 1em
- No indent

### Chapter Openers
- Full-bleed background image (faded/low opacity)
- Title and author overlaid in upper portion
- No page number on opener pages
- Force page break before

## Front Matter Structure
1. Half-title (book title only)
2. Frontispiece (image grid)
3. Title page (full title + "A Protocolized Anthology")
4. Copyright page
5. Contents (TOC)
6. [blank or epigraph]

## Margins (approximate from visual inspection)
- Top: ~0.75"
- Bottom: ~0.75" 
- Inside: ~0.7"
- Outside: ~0.6"
- Text block: ~3.6" × 5.8"
