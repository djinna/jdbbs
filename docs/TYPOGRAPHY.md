# Typography & Book Design Reference

A reference for book production decisions, grounded in classical principles.

**Philosophy**: Default to proven, classical book design. Break rules only intentionally, for specific project reasons documented in that project's brief.

---

## The Foundation: Why Rules Exist

Typographic conventions aren't arbitrary. They evolved over 500+ years of printing because they *work* — they reduce fatigue, increase comprehension, and respect the reader's time. As Bringhurst writes:

> "Typography exists to honor content."

When we follow classical principles, we're not being conservative for its own sake. We're using solutions that have been tested by millions of readers across centuries.

---

## Page Geometry

### Classical Proportions

The page itself has ideal proportions, developed long before printing:

| Ratio | Proportion | Character |
|-------|------------|----------|
| 1:1.414 | √2 (ISO/A4) | Modern, rational |
| 2:3 | 1:1.5 | Renaissance ideal |
| 1:1.618 | Golden section | Classical harmony |
| 3:4 | 1:1.333 | Sturdy, compact |
| 4:5 | 1:1.25 | Wide, modern |

The **2:3 proportion** is the workhorse of book design. Our template uses approximately this (4.91" × 7.59" ≈ 1:1.55).

### The Text Block

The text block (the rectangle containing your type) should relate harmoniously to the page. Classical approaches:

**Van de Graaf canon**: The text block is the same proportion as the page, positioned so that:
- Inner margin = 1 unit
- Top margin = 1.5 units  
- Outer margin = 2 units
- Bottom margin = 3 units

This creates a text block that sits high and toward the spine, with generous bottom margin — where the reader's thumbs go.

**Bringhurst's guidance**: 
- Text block height ≈ page width (creates a square relationship)
- Back margin (gutter) ≈ inner margin × 1.5 to account for binding
- The two-page spread is the fundamental unit, not the single page

### Margins Are Not Empty

Margins do active work:
- **Inner/gutter**: Keeps text from disappearing into binding
- **Outer**: Provides thumb room; space for marginal notes
- **Top**: Houses running headers; provides visual "air"
- **Bottom**: Largest margin; anchors the text block; page numbers live here

> "White space is to be regarded as an active element, not a passive background." — Jan Tschichold

---

## Horizontal Motion: The Line

### Line Length (Measure)

**The rule**: 45–75 characters per line, with 66 as the ideal.

This isn't aesthetic preference — it's cognitive science. The eye struggles to track back to the next line if the measure is too long. Too short, and reading becomes choppy, with excessive hyphenation.

```
Too short:          Ideal:                      Too long:
This is very        This line length allows     This line goes on and on and the
hard to read        the eye to track back       reader's eye struggles to find
because the         comfortably to the next     the beginning of the next line
lines break         line without losing its     which creates fatigue and reduces
constantly.         place in the text.          comprehension significantly.
```

**Measure and leading are linked**: Longer lines need more leading (line spacing) to help the eye track back. Shorter lines can be set tighter.

### Alignment

**Justified text** (flush left and right):
- Traditional for books
- Requires good hyphenation and spacing algorithms
- Creates calm, even texture
- Risk: "rivers" of white space if poorly set

**Ragged right** (flush left only):
- More relaxed, contemporary feel
- Even word spacing throughout
- Better for narrow columns
- No rivers

**Bringhurst's view**: Both are valid. Justified is traditional for body text in books. But "ragged right is more honest" — it doesn't stretch or compress to fill space.

**Our default**: Justified, with careful attention to hyphenation.

### Hyphenation

Rules for good hyphenation:
- Minimum 2 characters before hyphen, 3 after
- No more than 2–3 consecutive hyphenated lines
- Never hyphenate proper nouns if avoidable
- Never hyphenate the last word of a paragraph
- Never hyphenate across a page break

---

## Vertical Motion: The Page

### Leading (Line Spacing)

**The rule**: Leading should be 120–145% of type size.

- 10pt type → 12–14.5pt leading
- Expressed as "10/12" (ten on twelve) or line-height: 1.2

**Adjust for**:
- Longer lines → more leading
- Darker/heavier type → more leading
- Sans serif → often needs more leading
- x-height (larger x-height → more leading)

**Bringhurst**: "Continuous text usually needs positive leading... text for reading in bursts can often be set solid [no extra leading]." Headings can be set tighter than body.

### Vertical Rhythm

Ideal: Everything on the page aligns to a baseline grid.

- Body text sits on consistent baselines
- Headings occupy whole multiples of the body leading
- Block quotes, lists indent but maintain the rhythm
- Figures and captions fit the grid or span exact multiples

This creates a harmonious vertical rhythm when you hold the page to light — all text aligns through the sheet.

**Practical compromise**: Strict baseline grids are hard in complex books. At minimum, maintain consistent leading within text types.

### Paragraph Spacing

**The rule**: Use indentation OR vertical space, never both.

**Indentation** (traditional for books):
- First paragraph after heading: no indent
- Subsequent paragraphs: 1–1.5em indent
- No extra space between paragraphs
- Maintains vertical rhythm

**Block paragraphs** (common in business/web):
- No indent
- Half-line to full-line space between paragraphs
- Breaks vertical rhythm
- Acceptable for manuals, not ideal for continuous prose

**Our default**: Indentation, 0.75em, no space between paragraphs.

---

## Choosing & Using Type

### Selecting a Text Face

A good text face for books:

1. **Designed for extended reading** — not a display face
2. **Has true italics** — not slanted (oblique) romans
3. **Complete character set** — ligatures, small caps, multiple figure styles
4. **Even color** — no letters that jump out as too dark or light
5. **Appropriate x-height** — not too large (looks clunky) or small (hard to read)
6. **Restrained personality** — shouldn't call attention to itself

> "A good font is like a good waiter: efficient, unobtrusive, and there when you need it."

**Our default**: Libertinus Serif (open-source Palatino-influenced design)

### Combining Fonts

**The rule**: Use one serif and one sans-serif, maximum. Contrast, don't conflict.

Good combinations have:
- Similar x-heights
- Compatible proportions and weight
- Different enough to create clear hierarchy

**Classic pairings**:
- Serif body + Sans headings (our approach)
- One family with enough weights (less common in books)

**Avoid**:
- Two serifs that are similar but not identical
- Fonts that fight for attention
- More than 2–3 fonts total

### Type Size

| Element | Size | Notes |
|---------|------|-------|
| Body | 10–12pt | Smaller trim = smaller type |
| Block quote | Body size | Indent signals difference |
| Footnote | 8–9pt | ~80% of body |
| Caption | 8–9pt | May be sans serif |
| Running head | 8–9pt | Often small caps |
| Chapter title | 18–24pt | 1.5–2× body |
| Subhead | 12–14pt | 1.2× body |

---

## Character-Level Typography

### Small Capitals

**Use for**:
- Acronyms: nasa not NASA (full caps are too loud)
- A.D., B.C., A.M., P.M.
- Some proper nouns (house style)
- Running headers (sometimes)

**Rules**:
- Use real small caps (OpenType), never scaled-down capitals
- Add slight tracking: +3–5%
- Small caps are usually slightly taller than x-height

**Never**:
- Fake small caps by reducing font size
- Use small caps for emphasis (use italic)
- Mix small caps and full caps in same word

### Letterspacing (Tracking)

> **"Never letterspace lowercase."** — Frederic Goudy (and Bringhurst)

Letterspacing lowercase destroys the word shapes that enable fast reading.

**Do letterspace**:
- All capitals: +5–10% (they're designed to begin words, not be adjacent)
- Small caps: +3–5%
- Very large display type: may need reduction
- Very small text: may need increase for legibility

### Figures (Numerals)

| Type | Use | Example |
|------|-----|------|
| **Oldstyle (text)** | Body text | 1234567890 with ascenders/descenders |
| **Lining (titling)** | Tables, caps context | 1234567890 all same height |
| **Tabular** | Columns of numbers | Fixed-width for alignment |
| **Proportional** | Running text | Variable-width |

**Our default**: Oldstyle proportional for body text.

### Punctuation & Dashes

**Quotation marks**: Always curly, never straight.
- "Double quotes" for speech, titles
- 'Single quotes' for quotes within quotes, British style
- Apostrophe: it's not it's

**Dashes**:
- Hyphen (-): compound words, line breaks
- En dash (–): ranges (1990–1999), scores (3–2), compound adjectives when one part is multiple words ("pre–World War II")
- Em dash (—): interruption, parenthetical—like this—or attribution

**Bringhurst prefers spaced en dashes** ( – ) over tight em dashes for parentheticals. We follow the em dash convention for this series but it's a valid stylistic choice.

**Ellipsis**:
- Three spaced periods ( . . . ) — Bringhurst's preference
- Or the ellipsis character (…) with space before and after
- Not three unspaced periods (...)

### Ligatures

Use standard ligatures: fi, fl, ff, ffi, ffl

These prevent collisions between f and following letters. Most OpenType fonts have them; enable by default.

**Discretionary ligatures** (ct, st, etc.): Usually too archaic for contemporary books. Use only if deliberately evoking historical style.

---

## Book Structure

### Front Matter

Traditional order:

1. **Half-title** (recto) — title only, no subtitle or author
2. **Blank** (verso) — or frontispiece, card page, series list
3. **Title page** (recto) — full title, author, publisher
4. **Copyright** (verso) — legal info, ISBN, printer
5. **Dedication** (recto, optional)
6. **Blank** (verso, if dedication present)
7. **Contents** (recto)
8. **Acknowledgments** (recto, or back matter)
9. **Introduction/Preface** (recto)

Front matter traditionally uses lowercase roman numerals (i, ii, iii...). Main text restarts at 1.

### Chapter Openings

**Traditional conventions**:
- Start on recto (right-hand page)
- "Sink" or "drop": text begins 1/3 to 1/2 down the page
- Chapter number above title, or integrated
- No running header on opening page
- Page number at foot, centered

**First paragraph**:
- No indent
- May use drop cap (2–3 lines tall)
- May use small caps for first few words (lead-in)

### Chapter Title Length Guidelines

**The challenge**: Chapter/story titles vary widely in length. A design that works for "Loyalty" may look sparse for "We Shape Our Tools and Then Our Tools Shape Us".

**Strategies for accommodating variety**:

1. **Stacked/broken lines** (our approach for Protocolized series):
   - Break titles into 2-3 words per line
   - Creates consistent visual weight regardless of title length
   - Short titles: 1-2 lines; Long titles: 4-5 lines
   - Author name follows same pattern
   - Works well with large type sizes

2. **Fixed measure with size adjustment**:
   - Set a maximum line width (e.g., 60% of text block)
   - Reduce point size for longer titles to fit
   - Risk: Very long titles become too small

3. **Centered with natural breaks**:
   - Let titles wrap naturally at a set width
   - Insert manual breaks at logical phrase boundaries
   - Requires editorial attention per title

**Editorial guidance for authors**:

| Length | Words | Examples | Notes |
|--------|-------|----------|-------|
| Ideal | 2-5 | "Soda Sweet as Blood", "In Every Lifetime" | Works with any design approach |
| Acceptable | 6-8 | "The House That Paid Its Own Bills" | May need line breaks or size reduction |
| Long | 9+ | "We Shape Our Tools and Then Our Tools Shape Us" | Requires stacked design; consider subtitle |
| Very short | 1 | "Latency", "Loyalty" | May look sparse; can work with stacked author |

**Butterick's guidance**: "Suppress hyphenation in headings, and use the keep lines together and keep with next paragraph options to prevent headings from breaking awkwardly."

**Bringhurst on display type**: Large type (chapter titles) can be set tighter than body text - reduce leading to 100-110% for stacked titles. The letters are large enough that extra leading isn't needed for legibility.

**Our standard for Protocolized Anthology**:
- Max ~8 words strongly preferred
- Titles over 10 words: work with author to shorten or create subtitle
- All titles set in stacked format with consistent line width (~12 characters)

### Running Headers

**Verso (left page)**: Book title, part title, or author
**Recto (right page)**: Chapter title

**Style**:
- Small caps or caps-and-small-caps
- Smaller than body (8–9pt)
- Separated from text by space and/or rule

**Suppress on**:
- Chapter openings
- Blank pages
- Full-page illustrations
- Front matter (some styles)

### Page Numbers (Folios)

**Position options**:
- **Top, outside corner**: Most common for running-head designs
- **Bottom, centered**: Traditional, used on chapter openers
- **Bottom, outside corner**: Alternative

**Suppress on**:
- Blank pages
- Full-page images
- Part-title pages (sometimes)

### Section Breaks

Within chapters, scenes or sections divide with:

- **White space**: One blank line. Simple but can be missed at page break.
- **Ornament**: *** or # or ˘ ˘ ˘ or fleuron. Visible at page boundaries.
- **Number**: 2. or II. For numbered sections.

**Our default**: Three breves with space: ˘ ˘ ˘

---

## Widows, Orphans, and Breaks

### Definitions

- **Widow**: Last line of paragraph alone at top of page/column. *Always fix.*
- **Orphan**: First line of paragraph alone at bottom of page. *Fix if possible.*
- **Runt**: Very short last line of paragraph (one word or less). *Fix if easy.*

### Solutions

1. **Rewrite**: Add or remove a few words to reflow
2. **Tracking adjustment**: Tiny letterspacing change (±1%) to pull/push line
3. **H&J changes**: Adjust hyphenation or justification for paragraph
4. **Vertical adjustment**: Add/remove line from facing page (keep spread balanced)

**Never**:
- Leave widows (they're always wrong)
- Create rivers or bad spacing to fix orphans
- Adjust leading to fix breaks (destroys vertical rhythm)

### Keep Together

Certain elements should never separate:

- Heading and at least 2 lines of following paragraph
- Subhead and first line of text
- List item number and content
- Figure and caption
- Table head and first row

---

## When to Break the Rules

Rules exist for reasons. Break them when you have *better* reasons.

### Valid Reasons to Deviate

| Default | Override When |
|---------|---------------|
| Serif body text | Technical/code-heavy content; modernist design brief |
| Justified alignment | Narrow columns; heavily hyphenated language |
| Recto chapter starts | Saving pages for cost; continuous narrative feel |
| Traditional margins | Design intent emphasizing marginalia; art book |
| Oldstyle figures | All-numbers context (tables, forms) |
| Em dashes | House style specifies spaced en dashes |
| Indented paragraphs | Block paragraph design language (rare for books) |

### Document Deviations

When breaking rules for a project, document why:

```markdown
## Project: [Title]

### Design Deviations from Standard

1. **Ragged right alignment**
   - Reason: Narrow trim size (4" wide); justified causes rivers
   
2. **Sans-serif body text**
   - Reason: Client brand guidelines require [Font Name]
   - Mitigation: Increased leading to 1.4; generous margins

3. **Chapters don't start recto**
   - Reason: Cost constraint; saves ~20 pages
```

---

## Quick Reference Checklist

### Before Typesetting
- [ ] Page proportions appropriate for content?
- [ ] Margins: gutter accounts for binding? Thumb room?
- [ ] Font: designed for text? Complete character set?
- [ ] Line length: 45–75 characters?

### During Typesetting  
- [ ] Leading: 120–145% of type size?
- [ ] Paragraphs: indent XOR space, not both?
- [ ] First paragraphs: no indent after heads?
- [ ] Small caps: real (not faux)? Tracked?
- [ ] Acronyms: consistent small caps?
- [ ] Quotes: curly, correct direction?
- [ ] Dashes: hyphen/en/em used correctly?
- [ ] Figures: appropriate style for context?

### Before Output
- [ ] Widows eliminated?
- [ ] Orphans addressed?
- [ ] Running headers correct and consistent?
- [ ] Page numbers present, positioned correctly?
- [ ] Chapters start correctly (recto if required)?
- [ ] Blank pages truly blank (no headers/folios)?
- [ ] Front matter ordered correctly?
- [ ] Final read-through for rivers, spacing issues?

---

## Sources

- **Robert Bringhurst**, *The Elements of Typographic Style* (4th ed.) — The canonical reference
- **Matthew Butterick**, *Practical Typography* (practicaltypography.com) — Accessible, actionable
- **Jan Tschichold**, *The Form of the Book* — Modernist principles from a master
- **Ellen Lupton**, *Thinking with Type* — Visual, contemporary introduction
- **Jost Hochuli**, *Detail in Typography* — Deep dive on micro-typography
- **Fiona Raven & Glenna Collett**, *Book Design Made Simple* — Practical self-publishing guide
