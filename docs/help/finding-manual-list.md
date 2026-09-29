---
title: "Lists typed by hand"
summary: "Bullets or numbers typed as characters (-, *, 1.) instead of a list style."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 120
---

**Why it's flagged.** Lines start with a typed dash, asterisk or number rather than a real list. The characters print as typed, and the lines don't get list spacing.

**How serious.** Just noting.

**How to fix it.** In any editor:

1. Select the lines and apply your editor's bulleted or numbered list.
2. Delete the typed dashes or numbers.
3. Upload again.

**When to leave it.** If you want the typed characters exactly as they are, for example a numbered list inside a quotation, leave them.

[All Inspect findings](inspect.md)
