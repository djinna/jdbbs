---
title: "Mixed fonts or sizes in one paragraph"
summary: "A paragraph with more than one font or size in it."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 60
---

**Why it's flagged.** One paragraph mixes several fonts or sizes, which usually means pasted text from different sources. The factory sets the whole book in its own fonts, so those differences are dropped and won't show.

**How serious.** Worth a look.

**How to fix it.** In any editor:

1. Select the paragraph and clear its formatting, or paste it as plain text.
2. Then reapply bold or italic where you want it.
3. Upload again.

**When to leave it.** If the mix is only pasted-in leftovers, you don't need to do anything. The book comes out in one face either way.

[All Inspect findings](inspect.md)
