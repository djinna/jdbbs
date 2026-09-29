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

- **[Headings without a Heading style](finding-heading-lookalike.md).** A
  line that looks like a chapter title but is styled as body text, so the
  factory can't find it. Give it Heading 1.
- **[Custom styles not in your transmittal](finding-undeclared-custom-style.md).**
  Your file uses a style the factory doesn't know. Declare it, change it, or
  ask the studio.
- **[Marked styles](finding-style-marker.md).** The `[[markers]]` it found
  and what each became.
- **[Lists](finding-manual-list.md), [breaks](finding-manual-break.md) and
  [spacing](finding-direct-spacing.md) made by hand.** Usually harmless;
  styles are cleaner.
- **[Coloured](finding-colored-text.md) or
  [highlighted](finding-highlighted-text.md) text, [unusual
  fonts](finding-unusual-font.md).** Fonts are ignored (the book has its
  own). Colour and highlighting may not survive.
- **[Images](finding-image-inventory.md).** How many, and how large each
  will print.

Every finding has its own page, with why it's flagged and how to fix it in
any editor: see **Inspect findings** in [all help](/help/).

## The counts

Inspect also counts chapters, words and images. They fill in the
transmittal's checklist.

If Inspect **couldn't read the file**, export it again from your editor as a
`.docx` and upload it. Nothing is charged.
