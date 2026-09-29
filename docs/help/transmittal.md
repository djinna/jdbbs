---
title: The transmittal
summary: The one-page form about your book. It sets the title pages, trim size, typography and copyright page.
audience: authors
status: draft
routes: ["/{client}/{project}/factory/", "/{client}/{project}/transmittal/"]
covers:
  - srv/static/transmittal.js
last_verified: efed34a
owner: jenna
order: 30
---

The transmittal is step 1 of your Factory page. In publishing, a transmittal
is the sheet that travels with a manuscript into production and says what the
book is. Here it also drives the build: the factory generates your half-title,
title page, copyright page and contents from it.

It **saves as you type**. The header shows how much is filled in and whether
it is **Draft** or **Final**. **History** keeps earlier versions (at most one
every five minutes).

## What it asks

- **Book information:** author, title, subtitle, series, publisher, ISBNs,
  and an optional target date.
- **Manuscript checklist:** which parts of the book are in the manuscript,
  coming later, or not in it. Leave the title page, copyright page and
  contents included unless your book shouldn't have one. **Parts:** if your
  book has parts, Heading 1 is a part and Heading 2 a chapter. The chapter,
  word and image counts fill in when you [Inspect](inspect.md).
- **Illustrations:** one image per paragraph, placed inline (not floating),
  caption in the next paragraph. At least 1100 px wide for a full-width
  figure; PNG for line art and screenshots, JPEG for photos. The EPUB keeps
  colour. The print interior is black and white unless you tick *Colour
  interior*.
- **Rights:** a checkbox confirming the text is yours or you have permission.
  The factory typesets what you send; clearing rights is up to you.
- **Copyright page:** year, rights holder, licence (all rights reserved or a
  Creative Commons licence), credits. You see a live preview.
- **Format:** trim size. **6 × 9 in** is the standard trade paperback; choose
  it if you're unsure. Your printer supplies the spine and cover template.
- **Typography:** typeface, text size, section-break mark, indented or block
  paragraphs. The first answer in each row is the studio default, and the
  sampler PDF shows them all.
- **Custom styles:** paragraph or character styles your book needs beyond the
  factory's thirteen. Each one is based on a factory style. See
  [Send your manuscript](send-your-manuscript.md).
- **Cover:** notes only. The EPUB cover is uploaded in step 2. The print
  cover goes to your printer.

## Mark Final

When the form is right, choose **Mark Final**. The factory generates your
**authoring template** (`.docx`, or `.odt` for LibreOffice) and emails you
the link. You don't have to use the template. You can switch back to
**Draft** any time.
