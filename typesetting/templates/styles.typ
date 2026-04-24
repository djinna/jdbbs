// =============================================================================
// CHARACTER-LEVEL STYLES
// =============================================================================
// Extensible registry for inline text styles.
// Add new styles here as needed for future books.

// -----------------------------------------------------------------------------
// SMALL CAPS (c2sc + smcp for proper capital-to-smallcap conversion)
// Usage: #sc[AI], #sc[USB], #sc[DHCP]
// -----------------------------------------------------------------------------
#let sc(content) = {
  text(features: ("c2sc", "smcp"))[#content]
}

// -----------------------------------------------------------------------------
// ITALIC (explicit, for when #emph doesn't suffice)
// Usage: #ital[emphasized text]
// -----------------------------------------------------------------------------
#let ital(content) = {
  text(style: "italic")[#content]
}

// -----------------------------------------------------------------------------
// BOLD (explicit)
// Usage: #bold[strong text]
// -----------------------------------------------------------------------------
#let bold(content) = {
  text(weight: "bold")[#content]
}

// -----------------------------------------------------------------------------
// BOLD ITALIC
// Usage: #bold-ital[very emphasized]
// -----------------------------------------------------------------------------
#let bold-ital(content) = {
  text(weight: "bold", style: "italic")[#content]
}

// -----------------------------------------------------------------------------
// SPACED ELLIPSIS (. . .)
// Usage: text #ellipsis more text
// InDesign class: Ellipsis-60
// -----------------------------------------------------------------------------
#let ellipsis = [ . . . ]

// Also as function for custom spacing
#let spaced-ellipsis(spacing: 0.3em) = {
  h(spacing)
  [.]
  h(spacing)
  [.]
  h(spacing)
  [.]
  h(spacing)
}

// -----------------------------------------------------------------------------
// SUPERSCRIPT / SUBSCRIPT
// Usage: x#super[2], H#sub[2]O
// -----------------------------------------------------------------------------
#let super(content) = {
  text(baseline: -0.4em, size: 0.7em)[#content]
}

#let sub(content) = {
  text(baseline: 0.2em, size: 0.7em)[#content]
}

// -----------------------------------------------------------------------------
// UNDERLINE (use sparingly in book typography)
// Usage: #uline[underlined text]
// -----------------------------------------------------------------------------
#let uline(content) = {
  underline[#content]
}

// -----------------------------------------------------------------------------
// STRIKETHROUGH
// Usage: #strike[deleted text]
// -----------------------------------------------------------------------------
#let strike(content) = {
  text(stroke: 0.5pt)[#content]
}

// -----------------------------------------------------------------------------
// LETTERSPACED / TRACKED TEXT
// Usage: #tracked[S P A C E D]
// Useful for titles, headers
// -----------------------------------------------------------------------------
#let tracked(amount: 0.1em, content) = {
  text(tracking: amount)[#content]
}

// -----------------------------------------------------------------------------
// ALL CAPS (with optional tracking)
// Usage: #allcaps[chapter one]
// -----------------------------------------------------------------------------
#let allcaps(tracking: 0.05em, content) = {
  text(tracking: tracking)[#upper(content)]
}

// -----------------------------------------------------------------------------
// HIGHLIGHT / MARK
// Usage: #highlight[important passage]
// -----------------------------------------------------------------------------
#let highlight(color: yellow.lighten(60%), content) = {
  box(fill: color, inset: (x: 0.1em), content)
}

// -----------------------------------------------------------------------------
// SANS SERIF INLINE
// Usage: #sans[UI element]
// For technical terms, UI references, etc.
// -----------------------------------------------------------------------------
#let sans(font: "Source Sans 3", content) = {
  text(font: font)[#content]
}

// -----------------------------------------------------------------------------
// MONO / CODE INLINE
// Usage: #mono[function_name]
// For inline code, filenames, commands
// -----------------------------------------------------------------------------
#let mono(font: "JetBrains Mono", content) = {
  text(font: font, size: 0.9em)[#content]
}

// -----------------------------------------------------------------------------
// FOREIGN / NON-ENGLISH TEXT
// Usage: #foreign(lang: "fr")[c'est la vie]
// Typically italicized
// -----------------------------------------------------------------------------
#let foreign(lang: "und", content) = {
  text(lang: lang, style: "italic")[#content]
}

// -----------------------------------------------------------------------------
// ACRONYM (with optional expansion)
// Usage: #acronym[NASA] or #acronym(expand: "National Aeronautics...")[NASA]
// Uses small caps by default
// -----------------------------------------------------------------------------
#let acronym(expand: none, content) = {
  let styled = text(features: ("c2sc", "smcp"))[#content]
  if expand != none {
    // Could add tooltip or footnote in future
    styled
  } else {
    styled
  }
}

// -----------------------------------------------------------------------------
// PROPER NOUN / NAME STYLING
// Usage: #name[Claude Shannon]
// Some books use small caps for names
// -----------------------------------------------------------------------------
#let name(use-smallcaps: false, content) = {
  if use-smallcaps {
    text(features: ("smcp",))[#content]
  } else {
    content
  }
}

// -----------------------------------------------------------------------------
// BOOK/WORK TITLE (italic by convention)
// Usage: #booktitle[War and Peace]
// -----------------------------------------------------------------------------
#let booktitle(content) = {
  text(style: "italic")[#content]
}

// -----------------------------------------------------------------------------
// ARTICLE/CHAPTER TITLE (quotes by convention)
// Usage: #articletitle[The Structure of Scientific Revolutions]
// -----------------------------------------------------------------------------
#let articletitle(content) = {
  ["#content"]
}

// -----------------------------------------------------------------------------
// PULL QUOTE / CALLOUT TEXT
// Usage: #pullquote[Key insight here]
// Larger, emphasized for visual hierarchy
// -----------------------------------------------------------------------------
#let pullquote(content) = {
  text(size: 1.1em, style: "italic")[#content]
}

// -----------------------------------------------------------------------------
// OLDSTYLE FIGURES (if font supports)
// Usage: #oldstyle[1234567890]
// -----------------------------------------------------------------------------
#let oldstyle(content) = {
  text(features: ("onum",))[#content]
}

// -----------------------------------------------------------------------------
// LINING FIGURES (force tabular numbers)
// Usage: #lining[1234567890]
// -----------------------------------------------------------------------------
#let lining(content) = {
  text(features: ("lnum", "tnum"))[#content]
}
