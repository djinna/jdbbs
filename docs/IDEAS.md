# Ideas / to-do log

Not a roadmap — a place to park ideas so they survive sessions. Dated on
entry; move to a spec or TRACKER when picked up.

| date | idea | context |
|---|---|---|
| 2026-09-11 | **Per-client interactive stylesheet instances.** The `/stylesheet/` tool becomes a template: each client/project gets its own accept/reject instance seeded from the universal (fiction or nonfiction) stylesheet. | Came up while deciding to split `/stylesheet` into the internal tool (`/stylesheet-pi/`) and a public anonymized reference. |
| 2026-09-11 | **Sitemap-review tool → reusable "review table" pattern.** `~/sitemap-review` (port 8102) is a JSON table + per-row reply box. Generalise for future review rounds (e.g. stylesheet curation, page QA sign-off). | Built to collect page-normalization decisions; an example, neither binding nor exhaustive. |
| 2026-09-11 | **Client-visible on-disk documents.** `pi-client/{slug}/` root + `serveClientDoc` with `checkClientAuth`. Recipe is in the standard §7; build when the first real client doc arrives. | Deferred deliberately — no queued need. |
| 2026-09-11 | **Magic-link client login** (email → one-time link → cookie) to replace emailed passwords. | From SESSION-HANDOFF-2026-09-03; post-workshop. |
| 2026-09-18 | **Print cover / spine calc + "how to get it printed".** Cover is EPUB-only today; the factory has no printing step. Separate product question. Printers to point people to when this is picked up: **Bookmobile**, **Accutrack** (Jenna's research, 2026-09-18). | From the 2026-09-18 punch list §6.5; keep on the list, don't build into the factory before the workshop. |
| 2026-09-18 | **LibreOffice preview of the uploaded DOCX** inside the factory. Fidelity note: it would show what *Word* shows (what Inspect reads), not what the typeset PDF will look like — so it answers "did my styles land" not "what will my book look like". | §6.4 decided 2026-09-18: **no in-app preview** (fidelity is to Word, not to the book — more trouble than it is worth). LibreOffice (no-GUI, 231 MB) stays installed as a studio tool for eyeballing the Word template (`scripts/docx-preview.sh`); nothing in the app depends on it. |
| 2026-09-18 | **Index as a factory add-on.** LLM writes a conceptual index (headings, subentries, see/see-also) from the manuscript; locators resolved by Typst at compile time via `#index[]` markers so rebuilds stay true. Jev is not the tool (classifier, not generator) — optional confidence gate only. Feasibility: `docs/reviews/INDEX-ADDON-FEASIBILITY-2026-09-18.md`. | §5.13; post-workshop, ~5 days. |
| 2026-09-18 | **Jev (typesafe.ai) pilot: heading front/body/back classification in the book map.** Behind a flag, vocab-list heuristic kept as fallback, answers cached per manuscript hash, model pinned (`jev-1.13.0`, not `jev-latest`). Everything else in the review (lookalikes, transmittal pre-fill, finding triage) parked. Review: `docs/reviews/TYPESAFE-REVIEW-2026-09-18.md`. | Jenna agreed the recommendation 2026-09-18 (punch list 5.6 / 0.1). **After the workshop.** |
| 2026-09-18 | **API testing workshop, post-symposium.** Look through the Protocolize workshop and SIGP4B session participants for anyone with enough production going on (a series, a press, a recurring publication) to be worth inviting to a small "your factory calls my factory" session that exercises the machine-callable endpoint (5.8) against real work. Source lists: `registrations` table / Admin → Cohorts, plus who asks about the API in the talk Q&A. | Jenna, punch list 5.8 note, 2026-09-18: "lgtm" on claiming the API in the talk; remind her to scout participants for an API testing workshop after Sep 23. |


## Done (struck 2026-09-18)

| date | idea | landed |
|---|---|---|
| 2026-09-14 | **Admin doc editor `/admin/docs/`** — pick a `jdbbs-public` HTML file, edit, save → publishes instantly, auto-commits. | Exists as Docs in the admin nav. |
