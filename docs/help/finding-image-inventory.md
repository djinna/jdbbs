---
title: "Images"
summary: "The images Inspect found in your file, and how large each will print."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 160
---

**Why it's flagged.** Inspect lists each inline image with the size it will print. The count also fills in your transmittal's checklist. Colour images stay colour in the EPUB and turn grey in print unless the transmittal says *Colour interior*.

**How serious.** Just noting.

**How to fix it.** In any editor:

1. Put each image in its own paragraph, inline rather than floating or wrapped.
2. Put the caption in the next paragraph.
3. Use a wide enough image; 1100 pixels wide is suggested for a full-width figure.

**When to leave it.** Leave it. This is a list, not a problem.

[All Inspect findings](inspect.md)
