---
title: "Scene breaks done by hand"
summary: "A scene break made with typed characters or a row of empty paragraphs."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 140
---

**Why it's flagged.** Inspect saw either typed break characters or a run of empty paragraphs standing in for a scene break. Empty paragraphs and typed marks aren't a real break, so the gap may come out wrong or vanish.

**How serious.** Just noting.

**How to fix it.** In any editor:

1. Replace the typed marks or empty lines with one paragraph in the **Section Break** style.
2. No styles in your editor? Type `[[break]]` on its own line.
3. The mark itself (white space, breve, ornament) comes from your transmittal's Typography choice.

**When to leave it.** If the run of empty paragraphs is something else, such as space in a poem, tell the studio.

[All Inspect findings](inspect.md)
