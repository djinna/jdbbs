---
title: Send your manuscript
summary: Upload one .docx from your editor. Heading 1 for chapters, named styles or [[markers]] for the rest.
audience: authors
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - srv/static/factory.html
  - srv/static/factory.js
  - srv/books.go
  - typesetting/scripts/apply-style-markers.py
  - typesetting/scripts/generate-word-template.py
last_verified: efed34a
owner: jenna
order: 40
---

Step 2 on your Factory page. You send **one `.docx`**. Any editor that saves
one will do (LibreOffice, Google Docs, Pages and the rest). Uploading is free:
send a new version every time you revise.

## What the factory reads

The factory reads the **name** of each paragraph's style. It ignores fonts,
sizes and spacing, and sets everything in the book's typography. That means:

- **Chapter titles in Heading 1.** Sub-heads in Heading 2 and Heading 3. If
  your book has parts, Heading 1 is the part and Heading 2 the chapter (set
  this in the [transmittal](transmittal.md)).
- **Italic and bold** come through as you typed them. So do footnotes made
  with your editor's footnote command.
- **Special paragraphs** such as quotations, poetry, code, epigraphs and scene
  breaks need a style name or a marker (below).

## The thirteen factory styles

| Style | Use it for |
|---|---|
| Normal | Running text |
| First Paragraph | The paragraph after a heading or break (no indent) |
| Heading 1 | Chapter titles (part titles if the book has parts) |
| Heading 2 · Heading 3 | Sub-heads; they never start a new page |
| Block Quote | A quotation longer than a few words, set off from the text |
| Epigraph | A quotation at the head of a chapter or the book |
| Verse | Poetry or lyrics; one line per line (soft returns) |
| Code Block | Code, terminal output, anything whose spacing must stay exact |
| Section Break | A scene break; the mark comes from your transmittal |
| Copyright | Copyright and permissions lines |
| Signature | The name, title and place closing a foreword by another hand |
| Glossary Entry | One term and its definition per paragraph |

Your book can have more styles than these. Declare them in the
transmittal's **Custom styles**, each based on one of the thirteen.

## No styles? Use markers

Some editors (Google Docs, for one) can't hold custom paragraph styles.
Start the paragraph with a marker instead:

```
[[quote]] The rest of the paragraph…
[[verse]] · [[epigraph]] · [[code]] · [[break]] · [[signature]]
```

Close a run of several paragraphs with the matching end marker, e.g.
`[[/code block]]`. A custom style name works too: `[[style:Letter]]`. The
factory removes the markers and sets the real style. [Inspect](inspect.md)
lists every marker it found.

## Or start from the template

When your transmittal is final, the factory generates an **authoring
template** (`.docx`, or `.odt` for LibreOffice) with every style already set
up for your book. It's optional, and the markers do the same job.

## Upload

1. Title and author are filled in for you; check them.
2. Drop the file on the upload box, or choose it. **`.docx` only, up to
   50 MB.**
3. The newest upload becomes your **current manuscript**, the one that is
   inspected and built. Earlier uploads stay listed.

Then [Inspect](inspect.md) it.

**"Upload didn't work"** usually means the file isn't a real `.docx` (a
renamed `.doc`, `.pages` or `.odt`). Export it again from your editor as a
`.docx` and upload that.

## Cover (optional)

The EPUB gets your front cover if you upload one here: JPEG or PNG,
portrait, at least 1600 × 2400 px, up to 10 MB. It takes effect on your next
build. The print PDF is the interior only.
