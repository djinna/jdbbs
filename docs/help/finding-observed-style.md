---
title: "Styles seen in the file"
summary: "Custom styles Inspect noticed in your file."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 170
---

**Why it's flagged.** Inspect lists custom paragraph and character styles it saw in your file. Where you've declared a style in the transmittal, the finding says it's kept.

**How serious.** Just noting.

**How to fix it.** In any editor:

1. If you meant to use the style, declare it in the transmittal's **Custom styles**.
2. If not, apply a factory style instead.

**When to leave it.** Leave it if the style is declared or you don't mind how it looks.

[All Inspect findings](inspect.md)
