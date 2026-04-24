// =============================================================================
// IMAGE PLACEMENT UTILITIES
// =============================================================================
// Flexible image handling for various book layouts.
// Covers frontispiece, chapter openers, inline figures, full-bleed, etc.

// -----------------------------------------------------------------------------
// CONFIGURATION (can be overridden per-book)
// -----------------------------------------------------------------------------
#let image-config = (
  // Default margins/gutters
  bleed: 0.125in,        // Standard print bleed
  gutter: 0.25in,        // Space around inline images
  
  // Caption styling
  caption-size: 0.833em,
  caption-style: "italic",
)

// -----------------------------------------------------------------------------
// FULL-BLEED IMAGE (extends to trim edges)
// Usage: #full-bleed-image("path/to/image.jpg")
// For chapter openers, section dividers, dramatic moments
// -----------------------------------------------------------------------------
#let full-bleed-image(
  path,
  alt: none,
  fit: "cover",
) = {
  // Negative margins to extend to bleed
  place(
    top + left,
    dx: -0.7in,   // margin-inside
    dy: -0.75in,  // margin-top
    image(path, 
      width: 100% + 1.3in,  // inside + outside margins
      height: 100% + 1.5in, // top + bottom margins
      fit: fit,
      alt: alt,
    )
  )
}

// -----------------------------------------------------------------------------
// FULL-PAGE IMAGE (within margins)
// Usage: #full-page-image("path/to/image.jpg", caption: "Description")
// -----------------------------------------------------------------------------
#let full-page-image(
  path,
  caption: none,
  alt: none,
  fit: "contain",
) = {
  pagebreak(weak: true)
  
  align(center)[
    #image(path, 
      width: 100%, 
      height: if caption != none { 90% } else { 100% },
      fit: fit,
      alt: alt,
    )
  ]
  
  if caption != none {
    v(0.5em)
    align(center)[
      #set text(size: image-config.caption-size, style: image-config.caption-style)
      #caption
    ]
  }
}

// -----------------------------------------------------------------------------
// FRONTISPIECE (facing title page, traditional placement)
// Usage: #frontispiece("art.jpg", caption: "Illustration by...")
// -----------------------------------------------------------------------------
#let frontispiece(
  path,
  caption: none,
  alt: none,
) = {
  pagebreak(weak: true, to: "even")  // Verso page
  
  v(1fr)
  align(center)[
    #image(path, width: 80%, fit: "contain", alt: alt)
  ]
  
  if caption != none {
    v(1em)
    align(center)[
      #set text(size: 0.75em, style: "italic")
      #caption
    ]
  }
  v(1fr)
}

// -----------------------------------------------------------------------------
// CHAPTER OPENER IMAGE (collage/art above chapter start)
// Usage: #chapter-opener-image("collage.jpg")
// As seen in GHOSTS book - full width, takes whole page
// -----------------------------------------------------------------------------
#let chapter-opener-image(
  path,
  alt: none,
) = {
  pagebreak(weak: true, to: "odd")  // Recto page for chapter start
  
  // Full page, no margins
  place(
    top + left,
    dx: -0.7in,
    dy: -0.75in,
    image(path,
      width: 100% + 1.3in,
      height: 100% + 1.5in,
      fit: "cover",
      alt: alt,
    )
  )
  
  // Force page break after image
  pagebreak()
}

// -----------------------------------------------------------------------------
// INLINE FIGURE (within text flow)
// Usage: #figure-inline("diagram.png", caption: "Figure 1: Process flow")
// -----------------------------------------------------------------------------
#let figure-inline(
  path,
  caption: none,
  alt: none,
  width: 100%,
  placement: center,
) = {
  v(1em)
  
  align(placement)[
    #figure(
      image(path, width: width, fit: "contain", alt: alt),
      caption: if caption != none { 
        text(size: image-config.caption-size, style: "normal")[#caption]
      },
    )
  ]
  
  v(1em)
}

// -----------------------------------------------------------------------------
// SIDE FIGURE (floated to margin - note: limited Typst support)
// Usage: #figure-side("small-image.png", caption: "Detail")
// -----------------------------------------------------------------------------
#let figure-side(
  path,
  caption: none,
  alt: none,
  side: right,
  width: 40%,
) = {
  // Typst doesn't have true floats yet, so this is a simplified version
  let img-block = box(width: width)[
    #image(path, width: 100%, fit: "contain", alt: alt)
    #if caption != none {
      set text(size: 0.75em, style: "italic")
      caption
    }
  ]
  
  if side == right {
    place(right, dx: 0.5em, img-block)
  } else {
    place(left, dx: -0.5em, img-block)
  }
}

// -----------------------------------------------------------------------------
// IMAGE GRID (multiple images in grid layout)
// Usage: #image-grid(("a.jpg", "b.jpg", "c.jpg"), columns: 3)
// -----------------------------------------------------------------------------
#let image-grid(
  paths,
  columns: 2,
  gutter: 0.5em,
  caption: none,
) = {
  v(1em)
  
  let cells = paths.map(p => image(p, width: 100%, fit: "cover"))
  
  grid(
    columns: (1fr,) * columns,
    gutter: gutter,
    ..cells
  )
  
  if caption != none {
    v(0.5em)
    align(center)[
      #set text(size: image-config.caption-size, style: image-config.caption-style)
      #caption
    ]
  }
  
  v(1em)
}

// -----------------------------------------------------------------------------
// PORTRAIT IMAGE (constrained height, centered)
// Usage: #portrait("headshot.jpg", caption: "The Author")
// -----------------------------------------------------------------------------
#let portrait(
  path,
  caption: none,
  alt: none,
  max-height: 60%,
) = {
  align(center)[
    #image(path, height: max-height, fit: "contain", alt: alt)
  ]
  
  if caption != none {
    v(0.5em)
    align(center)[
      #set text(size: image-config.caption-size, style: image-config.caption-style)
      #caption
    ]
  }
}

// -----------------------------------------------------------------------------
// DECORATIVE ELEMENT (small ornament, flourish, dinkus)
// Usage: #ornament("flourish.svg")
// For section breaks, chapter endings
// -----------------------------------------------------------------------------
#let ornament(
  path,
  width: 3em,
) = {
  v(0.5em)
  align(center)[
    #image(path, width: width)
  ]
  v(0.5em)
}

// -----------------------------------------------------------------------------
// ICON INLINE (small image within text)
// Usage: Check #icon("checkmark.svg") when done
// -----------------------------------------------------------------------------
#let icon(
  path,
  size: 1em,
  baseline: 0.1em,
) = {
  box(baseline: baseline)[
    #image(path, height: size)
  ]
}

// -----------------------------------------------------------------------------
// BACKGROUND IMAGE (behind text)
// Usage: #page-background("watermark.png", opacity: 20%)
// -----------------------------------------------------------------------------
#let page-background(
  path,
  opacity: 30%,
) = {
  place(
    center + horizon,
    image(path, width: 80%, fit: "contain"),
  )
}

// -----------------------------------------------------------------------------
// IMAGE WITH BORDER
// Usage: #bordered-image("photo.jpg", stroke: 1pt)
// -----------------------------------------------------------------------------
#let bordered-image(
  path,
  caption: none,
  alt: none,
  width: 100%,
  stroke: 0.5pt + black,
  inset: 0.5em,
) = {
  v(1em)
  align(center)[
    #box(stroke: stroke, inset: inset)[
      #image(path, width: width, fit: "contain", alt: alt)
    ]
  ]
  
  if caption != none {
    v(0.5em)
    align(center)[
      #set text(size: image-config.caption-size, style: image-config.caption-style)
      #caption
    ]
  }
  v(1em)
}

// -----------------------------------------------------------------------------
// WRAPAROUND TEXT (image with text flowing around - limited support)
// Usage: #wrap-image("small.jpg", side: left)[Paragraph text here...]
// Note: Typst doesn't fully support text wrap yet
// -----------------------------------------------------------------------------
#let wrap-image(
  path,
  side: left,
  width: 35%,
  gutter: 0.75em,
  content,
) = {
  // Simplified: image and text in columns
  let img = image(path, width: 100%, fit: "contain")
  
  if side == left {
    grid(
      columns: (width, 1fr),
      gutter: gutter,
      img,
      content,
    )
  } else {
    grid(
      columns: (1fr, width),
      gutter: gutter,
      content,
      img,
    )
  }
}
