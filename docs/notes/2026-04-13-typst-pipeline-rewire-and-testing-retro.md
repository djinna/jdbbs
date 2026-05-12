# 2026-04-13 Typst pipeline rewire and testing retro

## What happened

We resumed manual QA asking whether the Typst workflow was ready for another test. A local smoke build in `book-production` succeeded via the newer direct DOCX -> Typst path, which should have been treated as the primary signal for where to continue.

Manual QA then reported:
- Inspect Manuscript report visible, but lacking an obvious timestamp / recency marker
- EPUB pull passed
- PDF pull failed in Prodcal with a Python syntax error in `scripts/md-to-chapter.py`

Investigation showed Prodcal's PDF button was still wired to the older path:

`DOCX -> pandoc markdown+fenced_divs -> md-to-chapter.py -> main.typ -> typst compile`

That path had two problems:
1. Immediate breakage: `md-to-chapter.py` had become syntactically invalid
2. Deeper design debt: even after repairing syntax, the markdown-bridge path remained fragile on real manuscript content (headings, footnotes, literal hashtags, escaping, include/import scope, etc.)

Meanwhile the newer path had already succeeded locally:

`DOCX -> pandoc --lua-filter=docx-to-typst-enhanced.lua -> full Typst document -> typst compile`

## Efficiency lesson / what we should have done sooner

We should have pivoted to the direct Typst pipeline immediately after the smoke test passed on the real manuscript.

Once there were two paths and only one was healthy, continuing to debug the legacy path during manual QA created churn and extended the testing cycle.

Better decision rule for future sessions:
- If the newer direct pipeline succeeds on the target manuscript and the integrated app path fails in an older bridge layer, stop debugging the older bridge unless there is an explicit reason to preserve it.
- Rewire the app to the healthier path first, then resume QA.

## Concrete inefficiencies from this session

1. We validated that direct Typst worked, but did not immediately switch Prodcal to use it.
2. We spent additional cycles repairing `md-to-chapter.py`, only to uncover more brittle edge cases in the legacy route.
3. Manual QA continued against a known stale architecture path instead of the path that had already passed a smoke build.

## Recommendations for future QA loops

1. Prefer architectural pivot over iterative patching when:
   - there are duplicate paths
   - one path already passes the real manuscript
   - the failing path is older / more adapter-heavy
2. During phase-gated QA, treat a proven local path as a checkpoint that can justify immediate rewire work before more user-side retesting.
3. Add visible timestamps / run recency indicators to the preflight report page so testers can verify they are looking at the latest inspection result.
4. Keep explicit notes in repo docs whenever a QA cycle spends time on the wrong layer, so we can review the pattern later.

## Immediate product follow-ups

- Rewire Prodcal PDF conversion to the direct DOCX -> Typst pipeline
- Add a visible timestamp / run metadata to the preflight report page
- Continue QA from the Typesetting tab after the PDF path is rewired, instead of reopening older bridge issues
