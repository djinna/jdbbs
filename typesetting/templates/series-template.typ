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
  // Page: named paper or explicit dimensions
  // If page-paper is set (e.g. "us-trade"), width/height are ignored.
  page-paper: none,
  page-width: 311.81pt,   // 110mm — the golden's TRIM (TrimBox). Was 353.811pt,
  page-height: 504.57pt,  // 178mm — which was the MediaBox (trim + 21pt bleed/crop).
                          // 110x178mm = UK A-format paperback.

  // Margins — measured from the golden's TRIM, not its media box. reference/GHOSTS.pdf
  // is a print PDF: MediaBox 353.811x546.567, BleedBox inset 13pt, TrimBox inset 21pt.
  // Text-block position (verso L=57 / recto L=60 etc., in media coords) minus the 21pt
  // trim offset gives these. Measure stays 237pt (311.81 - 38.8 - 35.8).
  margin-top: 39.8pt,
  margin-bottom: 57.2pt,
  margin-inside: 38.8pt,
  margin-outside: 35.8pt,

  // Font families — licensed print fonts (TRK-DESIGN-002) in typesetting/fonts/licensed/.
  // The open-source faces (Libertinus Serif / Source Sans 3) remain in the text
  // fallback chain below so a fresh checkout without the licensed files degrades
  // gracefully instead of substituting silently.
  // Original (print): Plantin MT Pro (body), Proxima Nova (display), Menlo → JetBrains Mono.
  body-font: "Plantin MT Pro",
  heading-font: "Proxima Nova",
  tweet-font: "Source Sans 3",
  code-font: "JetBrains Mono",
  
  // Base font size (0.833em relative to 12pt = ~10pt)
  base-size: 10pt,
  
  // Paragraph: leading is the gap between line bounding boxes. Tuned so the
  // baseline-to-baseline matches the golden's measured 12pt (classic 10/12).
  leading: 4.8pt,
  paragraph-spacing: 4.8pt,  // == leading: continuous text, no extra paragraph gap
  paragraph-indent: 0.75em,

  // Text flow
  justify: true,
  hyphenate: true,

  // Headings (em relative to base-size)
  h1-size: 1.667em,
  h1-weight: "bold",
  h2-size: 1.333em,
  h2-weight: 600,
  h3-size: 1em,
  h3-weight: "medium",
  heading-align: "left",
  // Parts (P4 step 6): when true, Heading 1 is a part opener (own recto,
  // blank verso) and Heading 2 is the chapter opener; Heading 3 takes the
  // H2 look. Front/back-matter H1s are demoted by the pipeline so they still
  // read as chapters.
  parts: false,
  part-size: 2em,

  // Spacing before/after headings
  h1-above: 0em,      // H1 gets pagebreak, so above-space is rarely needed
  h1-below: 0.5em,
  h2-above: 1em,
  h2-below: 0.25em,
  h3-above: 0.75em,
  h3-below: 0.25em,

  // Running heads — matched to the golden PDF's RENDERED output (reference/GHOSTS.pdf),
  // not the .indd nominal. The .indd panel reads Proxima Nova Semibold 10pt / 9pt folio,
  // but the golden PDF renders smaller + lighter (our bought Proxima cut is wider/heavier
  // than the Adobe cut at the same pt — "10pt ≠ 10pt"). Scaled by the measured cap-height
  // ratio (golden 5.52 / ours 6.72 ≈ 0.82) and dropped to Medium to match the strokes.
  running-heads-enabled: true,
  running-heads-size: 0.82em,        // caps ≈ 8.2pt (matches golden cap-height 5.52pt)
  running-heads-folio-size: 0.89em,  // folio figure-height 6.0pt — in the golden render the
                                     // page number is slightly LARGER than the caps, not smaller
  running-heads-weight: 500,         // Medium (golden strokes are lighter than Semibold)
  running-heads-verso: "author",     // no tracking — our narrower cut stays at natural spacing
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

// State to track pages that should show a centered DROP FOLIO in the footer.
// Chapter-body-start pages (golden p43) carry a centered folio at the foot and no
// running head — the inverse of normal body pages (running head draws the folio,
// no footer).
#let drop-folio-pages = state("drop-folio-pages", ())

// Front matter / folio machinery (P4, docs/reviews/P4-FRONT-MATTER-PLAN-2026-09-18.md).
// Physical page numbers are the key everywhere; the displayed folio is the page
// counter (reset to 1 at body start) rendered in folio-style.
#let blank-pages = state("blank-pages", ())          // pages that must stay empty (no folio, no head)
#let generated-end-page = state("generated-end-page", 0) // last generated page (i–iv); folios start after it
#let body-start-page = state("body-start-page", none) // physical page where arabic 1 lands
#let break-from = state("break-from", 0)
#let parts-state = state("parts", false)             // book() copies config.parts here for module-level helpers

#let mark-page(st) = context {
  let p = here().page()
  st.update(s => if p in s { s } else { s + (p,) })
}

// Page break to a recto, recording any verso left blank on the way so the
// header/footer leave it empty. Robust to being a no-op (already on an empty
// recto): the blank range is computed after the break, not guessed before.
#let break-to-recto() = {
  context { break-from.update(here().page()) }
  pagebreak(weak: true, to: "odd")
  context {
    let from = break-from.get()
    let to = here().page()
    if to - from >= 2 {
      blank-pages.update(s => s + range(from + 1, to))
    }
  }
}

// An intentionally blank page (p. ii when there is no frontispiece).
#let blank-page() = {
  pagebreak(weak: true)
  mark-page(blank-pages)
  v(0pt)  // makes the page count as non-empty so the strong break below is honoured
  pagebreak()
}

// Folio rendering: roman before body-start-page, arabic from it.
#let folio-text(physical, n) = {
  let bs = body-start-page.final()
  if bs == none or physical < bs { numbering("i", n) } else { numbering("1", n) }
}
#let page-carries-folio(p) = {
  p > generated-end-page.final() and p not in blank-pages.final()
}
#let page-in-front-matter(p) = {
  let bs = body-start-page.final()
  bs == none or p < bs
}

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
  // Blank pages and front matter carry no running head (front matter gets a
  // centred folio from the footer instead).
  if current-page in blank-pages.final() { return }
  if page-in-front-matter(current-page) { return }
  
  let title = current-story-title.get()
  let author = current-story-author.get()
  
  // No header if no story info set
  if title == none { return }
  
  let page-num = folio-text(current-page, counter(page).get().first())
  // Side is driven by PHYSICAL page parity (recto/verso), NOT folio parity. Our folio
  // (reset to 1 at body start) can differ in parity from the physical page, so keying
  // off the folio flips every running head onto the wrong side. The DISPLAYED number
  // stays the folio; only the layout side follows the physical page.
  let is-even = calc.even(current-page)

  // Caps at Semibold 10pt; folio rendered one step down at 9pt (same family/weight).
  // Size caps and folio INDEPENDENTLY, each relative to the inherited base size (10pt),
  // so 0.82em = 8.2pt caps and 0.89em = 8.9pt folio. Don't set size on the outer text, or
  // the folio's em compounds against the caps size and comes out too small.
  set text(font: config.heading-font, weight: config.running-heads-weight)
  let folio = text(size: config.running-heads-folio-size)[#page-num]
  let caps(s) = text(size: config.running-heads-size)[#upper(s)]

  if is-even {
    // Verso (even): folio + AUTHOR grouped at the outside (left).
    [#folio #h(1.35em) #caps(author)]
  } else {
    // Recto (odd): TITLE + folio grouped at the outside (right).
    align(right)[#caps(title) #h(1.35em) #folio]
  }
}

// Footer renderer — centered DROP FOLIO, drawn only on chapter-body-start pages.
// Matches golden p43: bold Proxima folio, centered, in the bottom margin. Same font
// size as the running-head folio (calibrated: identical bbox height in the golden).
#let drop-folio-footer() = context {
  if not config.running-heads-enabled { return }
  let current-page = here().page()
  if not page-carries-folio(current-page) { return }
  // Front matter: every page carries a centred roman folio. Body: only the
  // pages marked as drop-folio (chapter openers).
  if not page-in-front-matter(current-page) and current-page not in drop-folio-pages.final() { return }
  // Page number: 9pt Semibold, same as the running-head folio.
  set text(font: config.heading-font, size: config.running-heads-folio-size,
           weight: config.running-heads-weight)
  align(center)[#folio-text(current-page, counter(page).get().first())]
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
  set text(font: config.body-font, size: config.poem-size, style: "italic")
  set par(first-line-indent: 0em, leading: 0.8em, justify: false)
  align(center, pad(top: 0.5em, bottom: 0.5em, content))
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

// -----------------------------------------------------------------------------
// CHAPTER OPENER PAGE (golden GHOSTS layout — "Garden of Eden" standard)
// -----------------------------------------------------------------------------
// One recto page: a ~130pt square hero image upper-left, the title set ragged-left
// in a narrow column to its right, the author below it. Otherwise blank — no running
// head, no folio. All offsets are MEASURED from golden p41 in TRIM coordinates and
// expressed as dx/dy from the text-area origin (margin-inside, margin-top).
//
//   square: 130x130pt at trim (59.6, 122.5)         -> dx 20.8 / dy 82.7
//   title : Proxima bold ~20pt, col x=198.75, ~78pt wide, cap-top y=158.4, pitch 21pt
//   author: Proxima medium ~15pt, cap-top y=266.7
//
// `art` is a PROJECT-ROOT-RELATIVE path (e.g. "/manuscripts/ghosts/x.jpg") so the
// built-in image() resolves it regardless of which file invokes this template.
#let chapter-opener(title: none, author: none, art: none) = {
  pagebreak(weak: true, to: "odd")  // openers are recto

  if art != none {
    place(top + left, dx: 20.8pt, dy: 82.7pt,
      image(art, width: 130pt, height: 130pt, fit: "cover"))
  }

  if title != none {
    // top-edge: cap-height makes the placed box-top coincide with the cap-top, so
    // dy maps directly to the measured cap-top; leading sets the 21pt baseline pitch.
    place(top + left, dx: 160pt, dy: 121.1pt,
      box(width: 78pt)[
        #set text(font: config.heading-font, size: 2.0em, weight: "bold",
                  top-edge: "cap-height", bottom-edge: "baseline")
        #set par(leading: 7.7pt, first-line-indent: 0em, justify: false)
        #title
      ])
  }

  if author != none {
    place(top + left, dx: 160pt, dy: 228.5pt,
      box(width: 78pt)[
        #set text(font: config.heading-font, size: 1.6em, weight: "medium",
                  top-edge: "cap-height", bottom-edge: "baseline")
        #set par(first-line-indent: 0em, justify: false)
        #author
      ])
  }
}

// First body page of a chapter (golden p43): a blank verso precedes it, the body
// drops ~96pt from the top margin, and a centered drop folio sits at the foot with
// NO running head. Call between the opener and the chapter body include. Pass the
// story title/author here (set AFTER the page break so the blank verso stays blank).
#let chapter-body-start(title: none, author: none) = {
  pagebreak(weak: true, to: "odd")  // blank verso (even) + land body on recto
  set-story-info(title: title, author: author)
  context {
    let p = here().page()
    suppress-header-pages.update(s => if p in s { s } else { s + (p,) })
    drop-folio-pages.update(s => if p in s { s } else { s + (p,) })
  }
  v(96pt)  // body sink: first line lands at trim ~135.7 (top margin 39.8 + 95.9)
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

// -----------------------------------------------------------------------------
// GENERATED FRONT MATTER (P4) — pages i–iv come from the transmittal, not the
// manuscript. `fm` is a dict built by the Go pipeline (srv/books.go):
//   half-title, title-page, copyright-page: bool  (toc is emitted by the
//   pipeline via contents-page() after the untitled pieces)
//   subtitle, publisher, isbn-paper, isbn-epub, copyright-year,
//   copyright-holder, credit-lines, cover-credit: str or none
//   logo: project-root-relative image path or none
// -----------------------------------------------------------------------------

#let fm-get(fm, k, default: none) = {
  let v = fm.at(k, default: default)
  if v == "" { default } else { v }
}

// Title page: title, subtitle, author, publisher (+ logo) — p. iii.
#let title-page-full(title, subtitle: none, author: none, publisher: none, logo: none) = {
  pagebreak(weak: true)
  set par(first-line-indent: 0em, justify: false)
  v(1fr)
  align(center)[
    #set text(font: config.heading-font, size: 1.917em, weight: "bold")
    #title
    #if subtitle != none {
      v(0.6em)
      set text(size: 1em, weight: "medium")
      subtitle
    }
    #if author != none {
      v(2.5em)
      set text(size: 1.1em, weight: "medium")
      author
    }
  ]
  v(2fr)
  align(center)[
    #if logo != none { image(logo, height: 2.4em); v(0.5em) }
    #if publisher != none {
      set text(font: config.heading-font, size: 0.9em, weight: "medium")
      upper(publisher)
    }
  ]
  v(0.1fr)
}

// Copyright page — p. iv, set small at the foot of the page.
#let copyright-page-generated(fm, title, author) = {
  pagebreak(weak: true)
  set text(size: 0.75em)
  set par(leading: 0.55em, first-line-indent: 0em, justify: false, spacing: 0.9em)
  v(1fr)
  let holder = fm-get(fm, "copyright-holder", default: author)
  let year = fm-get(fm, "copyright-year")
  let cline = if year != none and holder != none [Copyright © #year #holder. All rights reserved.]
    else if holder != none [Copyright © #holder. All rights reserved.]
    else [All rights reserved.]
  [#strong(title)]
  parbreak()
  cline
  let publisher = fm-get(fm, "publisher")
  if publisher != none { parbreak(); [Published by #publisher.] }
  let credits = fm-get(fm, "credit-lines")
  if credits != none { parbreak(); credits }
  let cover = fm-get(fm, "cover-credit")
  if cover != none { parbreak(); cover }
  let isbn-p = fm-get(fm, "isbn-paper")
  let isbn-e = fm-get(fm, "isbn-epub")
  if isbn-p != none or isbn-e != none {
    parbreak()
    if isbn-p != none [ISBN #isbn-p (paperback)]
    if isbn-p != none and isbn-e != none { linebreak() }
    if isbn-e != none [ISBN #isbn-e (ebook)]
  }
}

// Contents — recto, after any dedication / epigraph (Chicago order). Emitted
// by the pipeline (lua filter) so it lands after the untitled pieces. Entry
// folios are rendered roman/arabic by position, like the pages themselves.
#let contents-page() = {
  break-to-recto()
  show outline.entry: it => {
    set text(font: config.heading-font, size: 0.9em)
    set par(first-line-indent: 0em, justify: false)
    let loc = it.element.location()
    block(above: 0.9em, below: 0em, context {
      let folio = folio-text(loc.page(), counter(page).at(loc).first())
      link(loc, it.body + h(1fr) + folio)
    })
  }
  // Own title rather than outline's: the built-in title is a level-1 heading
  // and would become a part opener in a parts book. Level 2 there = chapter look.
  context {
    let parts = parts-state.get()
    heading(level: if parts { 2 } else { 1 }, outlined: false, numbering: none)[Contents]
    outline(title: none, depth: if parts { 2 } else { 1 }, indent: 0em)
  }
}

// Untitled front-matter piece (dedication, epigraph) — each on a fresh recto.
#let front-piece(kind: "dedication", body) = {
  break-to-recto()
  set par(first-line-indent: 0em, justify: false)
  if kind == "dedication" {
    v(1fr)
    align(center, body)
    v(2fr)
  } else if kind == "epigraph" {
    v(0.3fr)
    pad(left: 12%, right: 4%, { set text(style: "italic"); body })
    v(1fr)
  } else {
    v(1fr)
    body
    v(2fr)
  }
}

// Titled front-matter section (Foreword, Preface, …): starts recto. The H1
// that follows keeps the ordinary heading style; folios stay roman.
#let front-section() = break-to-recto()

// Book info for the default running heads.
#let book-info = state("book-info", (title: none, author: none))

// Body start: recto, folio 1, arabic from here, running heads on.
#let start-body() = {
  break-to-recto()
  context { body-start-page.update(here().page()) }
  counter(page).update(1)
  // Default running heads = book title / author, unless an anthology has
  // already named the first story.
  context {
    if current-story-title.get() == none {
      let bi = book-info.get()
      set-story-info(title: bi.title, author: bi.author)
    }
  }
}

// Back matter start: hook for later (different running heads, etc.).
#let start-back() = break-to-recto()

// Everything before the manuscript body: generated pages, then the point from
// which folios are carried. Called by book() when front-matter is a dict.
#let generated-front-matter(fm, title, subtitle, author) = {
  let on(k) = fm.at(k, default: false) == true
  if on("half-title") {
    half-title(title)
    blank-page()
  }
  if on("title-page") {
    title-page-full(title, subtitle: subtitle, author: author,
      publisher: fm-get(fm, "publisher"), logo: fm-get(fm, "logo"))
  }
  if on("copyright-page") {
    copyright-page-generated(fm, title, author)
  }
  // Folios begin on the first page after the generated ones. No page break
  // here: the first piece of content (contents page, dedication, section,
  // body) brings its own break-to-recto.
  context { generated-end-page.update(here().page()) }
}

// =============================================================================
// EPIGRAPH
// =============================================================================

// No page break of its own: a chapter-opening epigraph sits under the heading,
// and a book epigraph is placed on its page by front-piece (P4).
#let epigraph(quote, attribution: none) = {
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
  config: config,  // accepts caller's merged config, defaults to module-level config
  front-matter: none,  // dict (see generated-front-matter) → P4 front matter + folios;
                       // none → legacy: body starts on page 1, arabic throughout
  body,
) = {
  // Document metadata
  set document(title: title, author: if author != none { (author,) } else { () })
  book-info.update((title: title, author: author))
  parts-state.update(config.parts)
  
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
    footer: drop-folio-footer(),
  )
  
  // Base typography - JUSTIFIED text (matches original)
  // Fallback chain: licensed/selected body-font first, then OFL Libertinus so
  // a print-only licensed family (TRK-DESIGN-002) that is missing on disk
  // (fresh checkout, family-name typo) degrades to a real serif instead of a
  // CJK face. CJK (Traditional Chinese) + Thai stay last so multilingual
  // passages (e.g. Ghosts ch. 8 "林家商店47號---ร้านยามาลิน สาขา ๔๗") render
  // correctly on hosts without OS-installed CJK/Thai fonts.
  set text(
    font: (config.body-font, "Libertinus Serif", "Noto Serif TC", "Noto Serif Thai"),
    size: config.base-size,
    lang: "en",
    hyphenate: config.hyphenate,
  )
  
  // Paragraph settings
  set par(
    justify: config.justify,
    leading: config.leading,
    spacing: config.paragraph-spacing,
    first-line-indent: config.paragraph-indent,
  )
  
  // Orphan/widow control
  // Note: Typst handles this automatically, but we set conservative values
  
  // Heading styles. Three looks — chapter opener, sub-head, sub-sub-head —
  // mapped onto levels 1/2/3, or 2/3/4 when the book has parts (level 1 is
  // then the part opener).
  let aligned(body) = {
    if config.heading-align == "center" {
      align(center, body)
    } else if config.heading-align == "right" {
      align(right, body)
    } else {
      align(left, body)
    }
  }
  let chapter-opener(it) = {
    pagebreak(weak: true)
    // Opener page: no running head, centred drop folio.
    mark-page(suppress-header-pages)
    mark-page(drop-folio-pages)
    set text(font: config.heading-font, size: config.h1-size, weight: config.h1-weight)
    set par(leading: 0.4em, first-line-indent: 0em, justify: false)
    aligned(it.body)
    v(config.h1-below)
    par(first-line-indent: 0em)[]
  }
  let sub-head(it) = {
    v(config.h2-above)
    block(breakable: false, below: 0em)[
      #set text(font: config.heading-font, size: config.h2-size, weight: config.h2-weight)
      #set par(first-line-indent: 0em, justify: false)
      #aligned(it.body)
      #v(config.h2-below)
      // Invisible anchor keeps heading with following content
      #box(height: 1em)
    ]
    par(first-line-indent: 0em)[]
  }
  let sub-sub-head(it) = {
    v(config.h3-above)
    block(breakable: false, below: 0em)[
      #set text(font: config.heading-font, size: config.h3-size, weight: config.h3-weight)
      #set par(first-line-indent: 0em, justify: false)
      #aligned(it.body)
      #v(config.h3-below)
      #box(height: 1em)
    ]
    par(first-line-indent: 0em)[]
  }
  // Part opener: own recto, title centred on the page, no head or folio,
  // verso left blank so the first chapter starts recto.
  let part-opener(it) = {
    break-to-recto()
    mark-page(suppress-header-pages)
    mark-page(blank-pages)
    set text(font: config.heading-font, size: config.part-size, weight: config.h1-weight)
    set par(leading: 0.4em, first-line-indent: 0em, justify: false)
    v(1fr)
    align(center, it.body)
    v(2fr)
    pagebreak(weak: true, to: "odd")
    context {
      // The verso we just skipped stays empty.
      let p = here().page()
      blank-pages.update(s => if (p - 1) in s { s } else { s + (p - 1,) })
    }
  }

  show heading.where(level: 1): it => if config.parts { part-opener(it) } else { chapter-opener(it) }
  show heading.where(level: 2): it => if config.parts { chapter-opener(it) } else { sub-head(it) }
  show heading.where(level: 3): it => if config.parts { sub-head(it) } else { sub-sub-head(it) }
  show heading.where(level: 4): it => if config.parts { sub-sub-head(it) } else { it }
  
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
  show footnote: it => {
    text(font: config.heading-font, size: 0.72em)[#super[#it.numbering]]
  }
  show footnote.entry: it => {
    set text(font: config.body-font, size: config.footnote-size)
    set par(leading: 0.5em, first-line-indent: -0.75em)
    pad(left: 0.75em, it)
  }
  
  if front-matter == none {
    // Legacy: no generated pages; folios from page 1, arabic.
    context { body-start-page.update(here().page()) }
  } else {
    generated-front-matter(front-matter, title, subtitle, author)
  }
  body
}
