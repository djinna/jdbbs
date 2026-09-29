---
title: "Bracketed style name"
summary: "A custom style declared in the transmittal with [[ ]] in its name."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 190
---

**Why it's flagged.** A custom style in your transmittal has square brackets in its name, for example "[[Letter]]". The factory already reads it as "Letter", so nothing breaks.

**How serious.** Just noting.

**How to fix it.** In any editor:

1. In the transmittal's Custom styles, remove the brackets from the name.
2. Keep the brackets in your manuscript, where they mark the paragraph: `[[Letter]]`.

**When to leave it.** It's tidy-up only. You can leave it.

[All Inspect findings](inspect.md)
