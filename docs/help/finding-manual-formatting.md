---
title: "Bold/italic applied by hand"
summary: "Bold, italic or underline set directly on the text rather than by a style."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 110
---

**Why it's flagged.** Bold, italic or underline was applied straight to the words instead of through a character style. This is how most people format, so it's very common.

**How serious.** Usually Carried through automatically. The build keeps your bold and italic as you meant them.

**How to fix it.** In any editor:

1. Nothing needed for ordinary bold and italic.
2. Underline is worth a second look: books rarely underline. If it stands for italic, change it to italic.

**When to leave it.** Leave it. It's the normal way to emphasise words.

[All Inspect findings](inspect.md)
