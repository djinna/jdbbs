---
title: "Non-Latin script"
summary: "Text in a non-Latin script, such as Greek, Cyrillic or Arabic."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 100
---

**Why it's flagged.** A paragraph contains writing in a non-Latin script. Such text may need its own font or language setting, so the studio wants you to confirm how it should be handled.

**How serious.** Worth a look.

**How to fix it.** In any editor:

1. Check the marked passage in a proof to see it printed.
2. Mention the language in the transmittal's Typography notes so the studio knows it's intended.

**When to leave it.** If the script is deliberate and shows correctly in a proof, leave it.

[All Inspect findings](inspect.md)
