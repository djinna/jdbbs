---
title: "Colored text"
summary: "Text set in a colour other than black."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 70
---

**Why it's flagged.** Some text has a colour. Colour may not survive into the book. Print is black and white inside unless your transmittal says *Colour interior*.

**How serious.** Worth fixing for clearly coloured text; Just noting for near-black.

**How to fix it.** In any editor:

1. Select the text and set the colour to Automatic or black.
2. If the colour marks something meaningful, like a link or an instruction to yourself, remove it or describe it in the transmittal's notes.

**When to leave it.** Near-black text is harmless. If the colour is deliberate, tell the studio.

[All Inspect findings](inspect.md)
