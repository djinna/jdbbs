# Ideas / to-do log

Not a roadmap — a place to park ideas so they survive sessions. Dated on
entry; move to a spec or TRACKER when picked up.

| date | idea | context |
|---|---|---|
| 2026-09-11 | **Per-client interactive stylesheet instances.** The `/stylesheet/` tool becomes a template: each client/project gets its own accept/reject instance seeded from the universal (fiction or nonfiction) stylesheet. | Came up while deciding to split `/stylesheet` into the internal tool (`/stylesheet-pi/`) and a public anonymized reference. |
| 2026-09-11 | **Sitemap-review tool → reusable "review table" pattern.** `~/sitemap-review` (port 8102) is a JSON table + per-row reply box. Generalise for future review rounds (e.g. stylesheet curation, page QA sign-off). | Built to collect page-normalization decisions; an example, neither binding nor exhaustive. |
| 2026-09-11 | **Client-visible on-disk documents.** `pi-client/{slug}/` root + `serveClientDoc` with `checkClientAuth`. Recipe is in the standard §7; build when the first real client doc arrives. | Deferred deliberately — no queued need. |
| 2026-09-11 | **Magic-link client login** (email → one-time link → cookie) to replace emailed passwords. | From SESSION-HANDOFF-2026-09-03; post-workshop. |
| 2026-09-14 | ~~**Admin doc editor `/admin/docs/`.**~~ **Done** — exists as `/admin/docs/` (Docs in the admin nav). Pick a `jdbbs-public` HTML file, edit in a textarea, save → publishes instantly (`servePublicDoc` reads disk), auto-commits to jdbbs-public. Raw HTML, so for sentence fixes, not restructuring. Lets the user tweak handout/why-book/deck copy during the Sep 19–22 freeze without a session. ~30–40 min. | Asked for while reviewing the session guide; **do after the Book 2 smoke**. |
| 2026-09-18 | **Print cover / spine calc + "how to get it printed".** Cover is EPUB-only today; the factory has no printing step. Separate product question. Printers to point people to when this is picked up: **Bookmobile**, **Accutrack** (Jenna's research, 2026-09-18). | From the 2026-09-18 punch list §6.5; keep on the list, don't build into the factory before the workshop. |
| 2026-09-18 | **LibreOffice preview of the uploaded DOCX** inside the factory. Fidelity note: it would show what *Word* shows (what Inspect reads), not what the typeset PDF will look like — so it answers "did my styles land" not "what will my book look like". | From §6.4; decide with that framing. |
| 2026-09-18 | **Jev (typesafe.ai) pilot: heading front/body/back classification in the book map.** Behind a flag, vocab-list heuristic kept as fallback, answers cached per manuscript hash, model pinned (`jev-1.13.0`, not `jev-latest`). Everything else in the review (lookalikes, transmittal pre-fill, finding triage) parked. Review: `docs/reviews/TYPESAFE-REVIEW-2026-09-18.md`. | Jenna agreed the recommendation 2026-09-18 (punch list 5.6 / 0.1). **After the workshop.** |
