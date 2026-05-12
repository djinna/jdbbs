#import "../../templates/series-template.typ": *

#show: book.with(
  title: "Ghosts in Machines",
  subtitle: "A Protocolized Anthology",
)

// ============================================================================
// FRONT MATTER
// ============================================================================

#set page(numbering: none)

// Half-title
#v(1fr)
#align(center)[
  #set text(font: config.heading-font, size: 1.5em, weight: "bold")
  Ghosts in Machines
]
#v(2fr)
#pagebreak()

// Blank (verso)
#pagebreak()

// Title page
#v(1fr)
#align(center)[
  #set text(font: config.heading-font, weight: "bold")
  #text(size: 2em)[Ghosts in Machines]
  #v(0.5em)
  #text(size: 1.2em, weight: "regular", style: "italic")[A Protocolized Anthology]
  #v(2em)
  #text(size: 0.9em, weight: "regular")[protocolized.summerofprotocols.com]
]
#v(2fr)
#pagebreak()

// Copyright page
#v(1fr)
#set text(size: 0.75em)
#set par(first-line-indent: 0em, leading: 0.6em)

©️ 2025 Ethereum Foundation. All contributions are the property of their respective authors and used by Ethereum Foundation under license.

All contributions licensed under CC BY-NC 4.0. After 2028-12-13, all contributions will be licensed under CC BY 4.0.

Learn more at summerofprotocols.com/ccplus-license-2023

#v(1em)
Interior design/layout: Jenna Dixon \
Cover design: James Langdon

#v(1em)
Printed in Argentina | October 2025
#v(2fr)
#pagebreak()

// Table of Contents
#set text(size: config.base-size)
#toc-heading

#toc-entry("Khlongs, Subaks, Beaings: From Ancient Agriculture To Artificial Ghosts", "Sam Chua", "9")
#toc-entry("Soda Sweet as Blood", "Spencer Nitkey", "17")
#toc-entry("In Every Lifetime", "Lara Dal Molin", "31")
#toc-entry("In the Garden of Eden, Baby", "Sisyphus", "41")
#toc-entry("We Shape Our Tools and Then Our Tools Shape Us", "Tongzhou Yu", "67")
#toc-entry("The House That Paid Its Own Bills", "Elizabeth Maher", "77")
#toc-entry("Latency", "Rafael Fernández", "91")
#toc-entry("Genius in the Bottle", "Claire Pichelin", "101")
#toc-entry("Loyalty", "Zach Hyman", "119")

#pagebreak()

// ============================================================================
// BODY - Enable page numbering
// ============================================================================

#set page(numbering: "1")
#counter(page).update(1)

// Chapter 0: Introduction (Khlongs)
#pagebreak(to: "odd")
#image("ghosts_00_SBAcover.png", width: 100% + 1.3in, height: 100% + 1.5in, fit: "cover")
#pagebreak()
#no-header()
#set-story-info(title: "Khlongs, Subaks, Beaings", author: "Sam Chua")

#include "00-intro.typ"

// Chapter 1: Soda Sweet as Blood
#set-story-info(title: none, author: none)  // Clear for opener pages
#pagebreak(to: "odd")
#image("ghosts_01_SODA.jpg", width: 100%)
#pagebreak()
#no-header()
#set-story-info(title: "Soda Sweet as Blood", author: "Spencer Nitkey")

#include "01-soda.typ"

// Chapter 2: In Every Lifetime
#set-story-info(title: none, author: none)
#pagebreak(to: "odd")
#image("ghosts_02_EVERY_LIFETIME.jpg", width: 100%)
#pagebreak()
#no-header()
#set-story-info(title: "In Every Lifetime", author: "Lara Dal Molin")

#include "02-lifetime.typ"

// Chapter 3: In the Garden of Eden, Baby
#set-story-info(title: none, author: none)
#pagebreak(to: "odd")
#image("ghosts_03_GARDEN.jpg", width: 100%)
#pagebreak()
#no-header()
#set-story-info(title: "In the Garden of Eden, Baby", author: "Sisyphus")

#include "03-garden.typ"

// Chapter 4: We Shape Our Tools
#set-story-info(title: none, author: none)
#pagebreak(to: "odd")
#image("ghosts_04_WE_SHAPE.jpg", width: 100%)
#pagebreak()
#no-header()
#set-story-info(title: "We Shape Our Tools and Then Our Tools Shape Us", author: "Tongzhou Yu")

#include "04-tools.typ"

// Chapter 5: The House That Paid Its Own Bills
#set-story-info(title: none, author: none)
#pagebreak(to: "odd")
#image("ghosts_05_HOUSE.jpg", width: 100%)
#pagebreak()
#no-header()
#set-story-info(title: "The House That Paid Its Own Bills", author: "Elizabeth Maher")

#include "05-house.typ"

// Chapter 6: Latency
#set-story-info(title: none, author: none)
#pagebreak(to: "odd")
#image("ghosts_06_LATENCY.jpg", width: 100%)
#pagebreak()
#no-header()
#set-story-info(title: "Latency", author: "Rafael Fernández")

#include "06-latency.typ"

// Chapter 7: Genius in the Bottle
#set-story-info(title: none, author: none)
#pagebreak(to: "odd")
#image("ghosts_07_GENIUS.jpg", width: 100%)
#pagebreak()
#no-header()
#set-story-info(title: "Genius in the Bottle", author: "Claire Pichelin")

#include "07-genius.typ"

// Chapter 8: Loyalty
#set-story-info(title: none, author: none)
#pagebreak(to: "odd")
#image("ghostts_08_LOYALTY.jpg", width: 100%)
#pagebreak()
#no-header()
#set-story-info(title: "Loyalty", author: "Zach Hyman")

#include "08-loyalty.typ"
