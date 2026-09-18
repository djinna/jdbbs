# Session handoff — Fri 18 Sep 2026 (punch-list block, pre-workshop)

Workshop Mon/Tue 21–22 Sep; talk Wed 23 Sep 17:30 UTC ("Have your factory
call my factory", deck `~/jdbbs-public/2026-pi-symposium/factory-talk.html`;
never edit `talk.html` there). Store flips to live Stripe Wed 23 00:00 HKT by timer.

## How we work
Shared punch list https://jdbbs.exe.xyz:8766/ (source `scratch/run/CHECKLIST.md`,
notes `scratch/run/notes.json`, relay tmux `runpage`, restart with
`RUNPAGE_CHAT_CONV=<conversation id>`). §0 = Jenna's inbox; answer ON the item
(`POST localhost:8766/note {"id","text","who":"shelley"}`); tick with sed;
`python3 scripts/punchlist-export.py 2026-09-18` → commit
`docs/runs/PUNCHLIST-2026-09-18.md` + `docs/runs/README.md`. **Jenna's rule
(17:45 UTC): work through the whole list — pick a smart order, do not triage
it down; watch context and compact at good stopping points.**

## State at handoff
- prodcal `main` = `7371666` + punch-list export commit, pushed to GitHub; tree clean.
- jdbbs-public clean at `26e1c77`.
- Service healthy; post-deploy free Ghosts build (book 9) OK: 61 pp, PDF + EPUB.
- Cache buster in `srv/static/transmittal.html` is now **20260918f**.

## Done today (all pushed)
5.4 C13 transmittal · 0.5 Format · 0.6 What-you-get · 5.8 machine factory
(docs/API.md, `srv/machine_factory_test.go`) · 6.4 · 0.7 Signature style ·
0.10 Glossary Entry (template = 13 styles) · "breaks belong to what follows"
rule · 0.8 drafts (`docs/brand/ABOUT-THE-BOOK-FACTORY-2026-09-18.md`, NOT
ticked) · 0.9 → 5.13 parked · **5.14 (new, this block):** checklist trap —
blank/"Coming later" rows for half-title/title/copyright/Contents switched
those generated pages off in the build; now default on, own group in the UI,
Included/Leave out only, CIP row hidden, test
`TestPullTransmittalGeneratedPagesDefaultOn`.
§2 pre-check as pinstitute (real one-time login link inserted then deleted;
`ip='smoke'`): portal, perception/transmittal, book-001/factory, free Inspect
all fine; template has 13 qFormat styles, Georgia/Arial/Courier New,
protocolized trim 4.91 × 7.59 in. Notes posted on 2.6, 2.7, 0.8, 5.14.

## Open catch
Plantin MT Pro **Bold** missing on VM (Regular + Italic only) — bold body
silently sets regular in Plantin print builds. In docs/IDEAS.md; Jenna to supply OTF.

## Waiting on Jenna
0.8 pick A/A′/B → then wire as transmittal checkbox (off by default; colophon
on © page and/or back-matter page) · 6.1–6.3, 6.5 keep/kill · 3.2 cuts ·
3.4 run-through · 2.5–2.7 eyeball · wording pass on Format / What you get /
Manuscript Checklist group labels / `/factory/terms` · one InDesign-set page for 5.5.

## Next for the agent (order)
1. Check §0 inbox first — new items take priority.
2. **5.5** on Typst 0.12: hyphenation/justification/keep-with-next levers,
   measured with `typesetting/scripts/compscore.py` on Ghosts (book 9, free)
   and Obliquities PDF (do NOT build project 22 — use its existing output).
   `costs` needs 0.13 → Wed 24+ (see `docs/reviews/P3-COMPOSITION-2026-09-18.md`).
3. **1.5 re-tag**: checkpoint tag per CHECKPOINTS.md + `go test ./srv/` +
   one pinstitute smoke (there have been ~10 deploys since `0cd67fd`).
4. 6.6 IDEAS tidy once 6.x decided; 0.8 wiring once picked; 3.5 after 3.4.
5. 5.9 only if queueing shows Mon/Tue.

## Rules carried
Never `cat` >350-line files; explicit `git add -- paths`; `make build` before
`sudo systemctl restart prodcal`; free test builds only on book 9 via
`localhost:8799/api/books/9/convert`; never burn project 22 credits; no
licensed-font embedding; no GitHub PAT; Typst 0.13 deferred to Sep 24+.

Metrics: ~45 % context at handoff; files read in full (guard bypasses): 0.
