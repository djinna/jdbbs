// Protocolized Anthology Series Template
// Replicates design from GHOSTS, TT, and LIBRARIANS books
// Uses open-source font alternatives

// Import extensible style and image modules
#import "styles.typ": *
#import "images.typ": *

// =============================================================================
// CONFIGURATION (defaults — override via merge-config or generated config.typ)
// =============================================================================

#let default-config = (
  // Page: either a Typst built-in paper name (e.g. "us-digest") OR explicit
  // width/height. If page-paper is non-none, it takes precedence and the
  // width/height fields are ignored.
  //
  // Defaults are the Protocolized Anthology trim (Ghosts, TT, Librarians):
  // 353.811 × 546.567 pt = 124.8 × 192.8 mm (= 4.914 × 7.591 in). No Typst
  // built-in matches this trim, so page-paper stays `none` by default.
  page-paper: none,
  page-width: 353.811pt,
  page-height: 546.567pt,
  
  // Margins (from spec)
  margin-top: 0.75in,
  margin-bottom: 0.75in,
  margin-inside: 0.7in,
  margin-outside: 0.6in,
  
  // Font families (open-source alternatives)
  // Original: Plantin MT Pro → Libertinus Serif
  // Original: Proxima Nova → Source Sans 3  
  // Original: Menlo → JetBrains Mono
  body-font: "Libertinus Serif",
  heading-font: "Source Sans 3",
  code-font: "JetBrains Mono",
  
  // Base font size (0.833em relative to 12pt = ~10pt)
  base-size: 10pt,
  
  // Paragraph: leading is inter-line gap (for 10/12, leading = 2pt)
  leading: 2pt,
  paragraph-indent: 0.75em,

  // Headings (em relative to base-size)
  h1-size: 1.667em,
  h1-weight: "bold",
  h2-size: 1.333em,
  h2-weight: 600,
  h3-size: 1em,
  h3-weight: "medium",

  // Running heads
  running-heads-enabled: true,
  running-heads-size: 0.75em,
  running-heads-verso: "author",
  running-heads-recto: "title",

  // Elements
  section-break: "breve",
  blockquote-style: "italic",
  poem-size: 0.75em,
  code-block-size: 0.8em,
  footnote-size: 0.75em,
)

// Merge overrides into defaults. Caller can pass a partial dict.
#let merge-config(overrides) = {
  let result = default-config
  for (key, val) in overrides {
    result.insert(key, val)
  }
  result
}

// Active config — set to defaults, can be overridden by callers
// who import this template and call merge-config().
#let config = default-config

// =============================================================================
// PAGE SETUP
// =============================================================================

// Page dimensions. We use explicit width/height so a Typst `set page(...)`
// stays a single set-rule and its scope isn't split across an if/else.
// Callers (e.g. the Go server's specToTypstConfig) resolve Typst named paper
// sizes like "us-digest" into width/height via the trim registry before
// emitting config, so width/height here is always authoritative.
#let book-page = page.with(
  width: config.page-width,
  height: config.page-height,
  margin: (
    top: config.margin-top,
    bottom: config.margin-bottom,
    inside: config.margin-inside,
    outside: config.margin-outside,
  ),
)

// =============================================================================
// RUNNING HEADERS (STATE-BASED)
// =============================================================================

// State to track current story info for running headers
#let current-story-title = state("story-title", none)
#let current-story-author = state("story-author", none)

// State to track pages that should have no header (keyed by physical page number)
// We store a set of page numbers to suppress headers on
#let suppress-header-pages = state("suppress-pages", ())

// Call this at start of each chapter to set header info
#let set-story-info(title: none, author: none) = {
  current-story-title.update(title)
  current-story-author.update(author)
}

// Mark current page as header-suppressed (call in content flow)
#let no-header() = context {
  let current = here().page()
  suppress-header-pages.update(pages => {
    if current not in pages { pages + (current,) } else { pages }
  })
}

// Running header renderer - called from page header
#let running-header() = context {
  // If running heads are disabled, return immediately
  if not config.running-heads-enabled { return }

  let current-page = here().page()
  let suppress-list = suppress-header-pages.final()
  
  // Check if this page should have header suppressed
  if current-page in suppress-list { return }
  
  let title = current-story-title.get()
  let author = current-story-author.get()
  
  // No header if no story info set (front matter)
  if title == none { return }
  
  let page-num = counter(page).get().first()
  let is-even = calc.rem(page-num, 2) == 0
  
  set text(font: config.heading-font, size: config.running-heads-size, weight: "medium")
  
  if is-even {
    // Verso (left/even): page number left, AUTHOR right
    grid(
      columns: (auto, 1fr),
      align: (left, right),
      [#page-num],
      upper(author),
    )
  } else {
    // Recto (right/odd): TITLE left, page number right
    grid(
      columns: (1fr, auto),
      align: (left, right),
      title,
      [#page-num],
    )
  }
}

// =============================================================================
// PARAGRAPH STYLES
// =============================================================================

// First paragraph (no indent) - use after headings, breaks, etc.
#let first-para(content) = {
  set par(first-line-indent: 0em)
  content
}

// Normal paragraph (0.75em indent)
#let body-para(content) = {
  set par(first-line-indent: 0.75em)
  content
}

// =============================================================================
// SECTION BREAK
// =============================================================================

// Section break renderer — style determined by config.section-break
// Supported styles: "breve", "asterism", "dinkus", "blank", "fleuron"
#let section-break = {
  if config.section-break == "blank" {
    v(1em)
  } else {
    v(0.5em)
    align(center)[
      #set text(size: 0.833em)
      #if config.section-break == "breve" [
        ˘ #h(1.5em) ˘ #h(1.5em) ˘
      ] else if config.section-break == "asterism" [
        ⁂
      ] else if config.section-break == "dinkus" [
        \* #h(1.5em) \* #h(1.5em) \*
      ] else if config.section-break == "fleuron" [
        ❧
      ] else [
        ˘ #h(1.5em) ˘ #h(1.5em) ˘
      ]
    ]
    v(0.5em)
  }
}

// Alternative with asterisks (kept for backward compatibility)
#let section-break-stars = {
  v(0.5em)
  align(center)[
    \* #h(1.5em) \* #h(1.5em) \*
  ]
  v(0.5em)
}

// =============================================================================
// CODE BLOCKS ("Ok-computer" style)
// =============================================================================

// Terminal/code output - slightly smaller font, generous line-height
#let code-block(content) = {
  set text(font: config.code-font, size: config.code-block-size)
  set par(leading: 0.6em, justify: false, first-line-indent: 0em)
  pad(left: 0.75em, top: 0.333em, bottom: 0.333em, content)
}

// =============================================================================
// POEM/VERSE BLOCK
// =============================================================================

#let poem(content) = {
  set text(font: config.code-font, size: config.poem-size)
  set par(first-line-indent: 0em, leading: 0.6em, justify: false)
  pad(left: 0.75em, top: 0.5em, bottom: 0.5em, content)
}

// =============================================================================
// CHAPTER OPENER
// =============================================================================

// Stacked title display - breaks title into multiple lines
// like the reference design (2-3 words per line)
// Use " / " in the title string to indicate manual line breaks
#let stacked-title(title, size: config.h1-size, weight: config.h1-weight) = {
  set text(font: config.heading-font, size: size, weight: weight)
  set par(leading: 0.3em, first-line-indent: 0em, justify: false)
  
  // Split on " / " for manual breaks, otherwise display as-is
  let parts = title.split(" / ")
  for (i, part) in parts.enumerate() {
    part
    if i < parts.len() - 1 { linebreak() }
  }
}

// Stacked author display
#let stacked-author(author, size: config.h2-size, weight: config.h2-weight) = {
  set text(font: config.heading-font, size: size, weight: weight)
  set par(leading: 0.3em, first-line-indent: 0em, justify: false)
  
  let parts = author.split(" / ")
  for (i, part) in parts.enumerate() {
    part
    if i < parts.len() - 1 { linebreak() }
  }
}

#let chapter(
  title: none,
  author: none,
  background-image: none,
  stacked: false,  // Use stacked multi-line display
  body,
) = {
  // Force page break to odd page, suppress header on opener
  pagebreak(weak: true, to: "odd")
  
  // Chapter title
  if title != none {
    if stacked {
      stacked-title(title)
    } else {
      set text(font: config.heading-font, size: config.h1-size, weight: config.h1-weight)
      set par(leading: 0.4em, first-line-indent: 0em)
      title
    }
  }
  
  // Author name
  if author != none {
    v(0.25em)
    if stacked {
      stacked-author(author)
    } else {
      set text(font: config.heading-font, size: config.h2-size, weight: config.h2-weight)
      set par(first-line-indent: 0em)
      author
    }
  }
  
  v(2em)
  
  // Chapter body - first para has no indent
  set par(first-line-indent: 0em)
  body
}

// =============================================================================
// TABLE OF CONTENTS STYLES
// =============================================================================

#let toc-heading = {
  set text(font: config.heading-font, size: 1.333em, weight: 600)
  set par(first-line-indent: 0em)
  [Contents]
  v(3em)
}

#let toc-entry(title, author, page-num) = {
  // Story title - bold
  {
    set text(font: config.heading-font, size: 0.833em, weight: "bold")
    title
  }
  linebreak()
  // Author (indented 0.75em under title)
  {
    set text(font: config.heading-font, size: 0.833em, weight: "medium")
    h(0.75em)
    author
  }
  h(1fr)
  {
    set text(font: config.heading-font, size: 0.833em)
    page-num
  }
  v(0.5em)
}

// =============================================================================
// FRONT MATTER PAGES
// =============================================================================

// Half-title page
#let half-title(title) = {
  pagebreak(weak: true)
  v(1fr)
  align(center)[
    #set text(font: config.heading-font, size: 1.5em, weight: "bold")
    #title
  ]
  v(2fr)
}

// Title page
#let title-page(title, subtitle: none) = {
  pagebreak(weak: true)
  v(1fr)
  align(center)[
    #set text(font: config.heading-font, size: 1.917em, weight: "bold")
    #title
    
    #if subtitle != none {
      v(0.5em)
      set text(size: 1em, weight: "bold")
      upper(subtitle)
    }
  ]
  v(2fr)
}

// Copyright page
#let copyright-page(content) = {
  pagebreak(weak: true)
  set text(size: 0.667em)
  set par(leading: 0.6em, first-line-indent: 0em)
  content
}

// =============================================================================
// EPIGRAPH
// =============================================================================

#let epigraph(quote, attribution: none) = {
  pagebreak(weak: true)
  set par(first-line-indent: 0em)
  set text(style: "italic")
  quote
  if attribution != none {
    linebreak()
    h(4.083em)  // matches InDesign Epi-sig indent
    set text(style: "normal")
    [— #attribution]
  }
  v(1.333em)
}

// =============================================================================
// BLOCK QUOTE
// =============================================================================

#let blockquote(content) = {
  set par(first-line-indent: 0em)
  if config.blockquote-style == "bar" {
    block(
      inset: (left: 1em, top: 0.5em, bottom: 0.5em, right: 0em),
      stroke: (left: 1.5pt + luma(120)),
      content,
    )
  } else if config.blockquote-style == "indent" {
    pad(left: 1.5em, right: 1.5em, top: 0.5em, bottom: 0.5em)[
      #content
    ]
  } else {
    // Default: "italic"
    pad(left: 1.5em, right: 1.5em, top: 0.5em, bottom: 0.5em)[
      #set text(style: "italic")
      #content
    ]
  }
}

// =============================================================================
// DROP CAP
// =============================================================================

#let drop-cap(letter, body) = {
  let cap = {
    set text(size: 3em, weight: "regular")
    box(baseline: 0.5em, letter)
  }
  grid(
    columns: (auto, 1fr),
    gutter: 0.2em,
    cap,
    body,
  )
}

// =============================================================================
// MAIN DOCUMENT TEMPLATE
// =============================================================================

#let book(
  title: "Untitled",
  subtitle: none,
  author: none,
  font-path: none,
  body,
) = {
  // Document metadata
  set document(title: title, author: if author != none { (author,) } else { () })
  
  // Page setup with running headers (no footer)
  set page(
    width: config.page-width,
    height: config.page-height,
    margin: (
      top: config.margin-top,
      bottom: config.margin-bottom,
      inside: config.margin-inside,
      outside: config.margin-outside,
    ),
    header: running-header(),
  )
  
  // Base typography - JUSTIFIED text (matches original)
  set text(
    font: config.body-font,
    size: config.base-size,
    lang: "en",
    hyphenate: true,
  )
  
  // Paragraph settings - justified with 0.75em indent
  set par(
    justify: true,
    leading: config.leading,
    first-line-indent: config.paragraph-indent,
  )
  
  // Orphan/widow control
  // Note: Typst handles this automatically, but we set conservative values
  
  // Heading styles
  show heading.where(level: 1): it => {
    pagebreak(weak: true)
    set text(font: config.heading-font, size: config.h1-size, weight: config.h1-weight)
    set par(first-line-indent: 0em, leading: 0.4em)
    it.body
    v(0.5em)
  }
  
  show heading.where(level: 2): it => {
    v(1em)
    set text(font: config.heading-font, size: config.h2-size, weight: config.h2-weight)
    set par(first-line-indent: 0em)
    it.body
    v(0.25em)
  }
  
  show heading.where(level: 3): it => {
    v(0.75em)
    set text(font: config.heading-font, size: config.h3-size, weight: config.h3-weight)
    set par(first-line-indent: 0em)
    it.body
    v(0.25em)
  }
  
  // Raw/code blocks - "Ok-computer" style
  show raw.where(block: true): it => {
    set text(font: config.code-font, size: config.code-block-size)
    set par(leading: 0.6em, justify: false, first-line-indent: 0em)
    pad(left: 0.75em, top: 0.333em, bottom: 0.333em, it)
  }
  
  show raw.where(block: false): it => {
    set text(font: config.code-font, size: 0.85em)
    it
  }
  
  // Emphasis (italic)
  show emph: set text(style: "italic")
  
  // Strong (bold)
  show strong: set text(weight: "bold")
  
  // Horizontal rule → section break
  show line: section-break
  
  // Footnote styling
  show footnote.entry: it => {
    set text(size: config.footnote-size)
    set par(leading: 0.5em, first-line-indent: -0.75em)
    pad(left: 0.75em, it)
  }
  
  body
}
