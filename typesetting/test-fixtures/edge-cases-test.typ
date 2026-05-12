#import "/templates/series-template.typ": *

#show: book.with(
  title: "TITLE",
  author: "AUTHOR",
)

Testing Word → EPUB conversion edge cases

= 1. Manual Formatting
<manual-formatting>
This paragraph contains #strong[manually bolded text] and #emph[manually
italicized text] and #emph[#strong[both bold and italic]].

This has red text and highlighted yellow

= 2. Font Chaos
<font-chaos>
Comic Sans text mixed with Times New Roman and tiny text and HUGE text

= 3. Spacing Issues
<spacing-issues>
Multiple spaces between words

Tab at start of paragraph

Line with \
manual break \
in middle

After multiple empty paragraphs

= 4. Section Break Variations
<section-break-variations>
End of section one

#section-break
Start of section two

#section-break
Another section

#section-break
= 5. List Formatting
<list-formatting>
Manual numbered list:

// Editorial review: manual list item detected; keep/strip/convert decision required
1. First item

// Editorial review: manual list item detected; keep/strip/convert decision required
2. Second item

// Editorial review: manual list item detected; keep/strip/convert decision required
3. Third item

Manual bullets:

// Editorial review: manual list item detected; keep/strip/convert decision required
\* Bullet one

// Editorial review: manual list item detected; keep/strip/convert decision required
- Bullet two

// Editorial review: manual list item detected; keep/strip/convert decision required
• Bullet three

= 6. Special Characters
<special-characters>
\"Curly quotes\" vs \"straight quotes\"

Em—dash, en–dash, and hyphen-ated

Ellipses... vs ellipsis…

Café, niño, über, Москва

Math: 2×2\=4, ±10%, π≈3.14, ∑

Emoji: 😊 📚 ✨

= 7. Tables
<tables>
#align(center)[#table(
  columns: 3,
  align: (col, row) => (auto,auto,auto,).at(col),
  inset: 6pt,
  [#strong[Header 1]], [#strong[Header 2]], [#strong[Header 3]],
  [Data with \*emphasis\*],
  [Normal data],
  [More data],
  [],
  [],
  [],
)
]

= 8. Custom Styles
<custom-styles>
Book titles: #emph[The Great Gatsby] and #smallcaps[NINETEEN
EIGHTY-FOUR]

= 9. Kitchen Sink Paragraph
<kitchen-sink-paragraph>
This paragraph has #emph[#strong[everything]]: multiple spaces,
different font sizes, and ends with manual line break. \
Plus another line.

= 10. Nested Manual Lists
<nested-manual-lists>
// Editorial review: manual list item detected; keep/strip/convert decision required
1. Parent one

// Editorial review: manual list item detected; keep/strip/convert decision required
a) Child alpha

// Editorial review: manual list item detected; keep/strip/convert decision required
b) Child beta

// Editorial review: manual list item detected; keep/strip/convert decision required
2. Parent two

// Editorial review: manual list item detected; keep/strip/convert decision required
\* Child bullet one

// Editorial review: manual list item detected; keep/strip/convert decision required
- Child bullet two
