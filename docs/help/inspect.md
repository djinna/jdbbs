---
title: Inspect
summary: What the factory found in your file, how serious each finding is, and what to do next.
audience: authors
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - srv/preflight.go
  - srv/static/factory.js
  - typesetting/scripts/detect-edge-cases.py
  - srv/bookmap_inspect.go
last_verified: efed34a
owner: jenna
order: 50
---

Step 3 on your Factory page. **Inspect** reads your current manuscript and
tells you what won't come through the way you meant: stray styles, headings
the factory can't see, odd spacing, images. It's **free and unlimited** and
never uses a build. A script checks your file against a list of rules. No AI
reads your manuscript.

Choose **Inspect manuscript**, then **Inspect again** after each revision.
**Open the full report →** shows every finding in detail.

## How serious is it?

| Label | Meaning |
|---|---|
| **Worth fixing** | Likely to come out wrong in the book. Fix it in your editor and upload again. |
| **Worth a look** | May be fine; check it. |
| **Just noting** | For your information. |
| **Carried through automatically** | The build keeps it as you meant it (your italic and bold, for example). Nothing to do. |

**Nothing Inspect finds blocks a build.** You can always build a free proof
and look.

## The book map

Inspect also shows **how the build reads your file**: what it takes as front
matter, body and back matter, from your Heading 1 titles and where they
fall. Your own title page and contents are dropped, because the factory
generates them from the transmittal. If a chapter lands in the wrong place,
check its heading.

## Common findings

- **Headings without a Heading style.** A line that looks like a chapter
  title (bold, larger, "Chapter 3") but is styled as body text, so the
  factory can't find it. Give it Heading 1.
- **Custom styles not in your transmittal.** Your file uses a style the
  factory doesn't know. Declare it in the transmittal's Custom styles,
  change it to a factory style, or ask the studio.
- **Marked styles.** The `[[markers]]` it found and what each became.
- **Manual lists, manual breaks, direct spacing.** Typed bullets, hand-made
  scene breaks, spacing set by hand. Usually harmless; styles are cleaner.
- **Coloured or highlighted text, unusual fonts.** Fonts are ignored (the
  book has its own). Colour and highlighting may not survive.
- **Images.** How many, and how large each will print. Colour images stay
  colour in the EPUB and turn grey in print unless the transmittal says
  *Colour interior*.

One page per finding, with how to fix each in any editor, is coming.

## The counts

Inspect also counts chapters, words and images. They fill in the
transmittal's checklist.

If Inspect **couldn't read the file**, export it again from your editor as a
`.docx` and upload it. Nothing is charged.
