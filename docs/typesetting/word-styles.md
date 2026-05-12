# Word Style Guide for Book Production

Use these Word styles for clean conversion to the series template.

## Required Styles

| Word Style | Maps To | Usage |
|------------|---------|-------|
| **Heading 1** | Chapter title | Start of each chapter/story |
| **Heading 2** | Section heading | Major sections within chapter |
| **Heading 3** | Subsection | Minor sections |
| **Normal** | Body paragraph | Regular prose (auto-indented) |
| **No Spacing** | First paragraph | Use after headings (no indent) |

## Optional Styles

| Word Style | Maps To | Usage |
|------------|---------|-------|
| **Quote** | Block quote | Epigraphs, quoted passages |
| **Intense Quote** | Epigraph | Attributed quotes |
| **Code** | Code block | Terminal output, code |
| **Subtitle** | Author name | Story author (after Heading 1) |

## Section Breaks

Insert a **horizontal line** (Insert → Horizontal Line) or type `---` on its own line.
This converts to the centered three-dot break: ˘ ˘ ˘

## Formatting

| Word Format | Result |
|-------------|--------|
| *Italic* | Emphasis |
| **Bold** | Strong |
| `Code` (font) | Inline code |

## Document Structure

```
[Title - Heading 1]
[Author - Subtitle style]

[First paragraph - No Spacing style]

[Body paragraphs - Normal style]
[Body paragraphs - Normal style]

---

[After break - No Spacing style]
[Continue - Normal style]

[New Chapter - Heading 1]
...
```

## Creating a Word Template

1. Open Word → New Blank Document
2. Modify these styles (right-click style → Modify):
   - **Heading 1**: Your chapter title font
   - **Normal**: Your body font, justified
   - **No Spacing**: Same as Normal but no first-line indent
3. Save As → Word Template (.dotx)
4. Use this template for all manuscripts

## Tips

- **Don't** use manual formatting (select text → make bold)
- **Do** use styles consistently
- **Don't** use tabs for indentation
- **Do** use Word's built-in styles or custom styles based on them
- **Don't** insert manual page breaks for chapters
- **Do** use Heading 1 which triggers automatic page breaks

## Troubleshooting

**Text not indenting correctly?**
- First paragraph after heading should use "No Spacing" style
- Subsequent paragraphs use "Normal" style

**Code blocks not formatting?**
- Apply "Code" character style, or
- Use a monospace font (Courier, Consolas)

**Section breaks not appearing?**
- Use horizontal line, not manual "* * *"
- Or use page break for hard chapter breaks
