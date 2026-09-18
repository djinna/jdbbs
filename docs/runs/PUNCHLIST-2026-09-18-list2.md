<!-- exported 2026-09-18 18:08 UTC by scripts/punchlist-export.py; source scratch/run/CHECKLIST.md + notes.json -->

# Punch list 2 · Fri 18 Sep → workshop (Mon 21 / Tue 22) → talk (Wed 23)

Legend: **YOU** = Jenna's turn · **ME** = Shelley's turn · **BOTH** = look together.
Ticked = done · ◐ = in progress. Click a box to cycle ☐ → ☑ → ◐. **note** opens a reply box under any item (⌘↵ saves) — I read those back. Bottom bar adds a new item to the Inbox (now at the top). Refreshes every 15 s (pauses while you type).

Previous list (35 of 55 ticked, incl. 5.14 front matter, 5.15 images, 5.16 `@`) archived at `docs/runs/PUNCHLIST-2026-09-18.md` → readable at `/admin/runs/`. Item numbers carry over so the note threads stay attached. Checkpoint: `checkpoint-2026-09-18-pre-workshop-v2`.

**Freeze:** Sat 19 – Sun 20 fix / test / clean allowed (build → `go test ./srv/` → restart → one factory smoke as `pinstitute`). Mon 21 – Wed 23 hotfix-only. Store flips to live Stripe by timer **Wed 23 Sep 00:00 HKT**.

## 0 · Inbox — new items, untriaged (Shelley moves them into a section)

## 2 · Your side — YOU

- [ ] 2.6 Unlock the `pinstitute` portal and eyeball the new cards (Transmittal · Factory → · Calendar, `[FACTORY PASS]`)

  > **shelley** · 2026-09-18 17:24 UTC  
  > Pre-checked as pinstitute just now (signed in with a real one-time link, minus the email): portal shows both projects with the Transmittal · Factory → · Calendar cards and [FACTORY PASS] on Obliquities; perception/transmittal loads with the new Format + What-you-get sections; book-001/factory shows 2 of 3 builds left, and Inspect ran free in 1.5 s (1156 findings on the test rollup, book map printed, no build spent). Nothing broke. Your turn is only the eyeball.

- [ ] 2.7 Re-download the Word template on a final transmittal, confirm in Word — Styles pane shows the **13** factory styles, fonts Georgia/Arial/Courier New, page = trim (server side pre-checked; the Word look is yours)

  > **shelley** · 2026-09-18 17:24 UTC  
  > Pre-checked the file itself: the template downloaded from book-001/factory has 13 qFormat styles (Normal, Heading 1–3, First Paragraph, Block Quote, Code Block, Section Break, Verse, Copyright, Epigraph, Signature, Glossary Entry), fonts Georgia / Arial / Courier New, page = the protocolized trim (4.91 × 7.59 in). Item text says 11 — read it as 13. What only Word can confirm is the Styles pane.

  > **shelley** · 2026-09-18 18:07 UTC  
  > Pre-checked the plumbing: the template download correctly refuses until the transmittal is marked final (409, ‘fill in the transmittal and mark it final first’). The Word-side look (Styles pane, fonts, page size) is yours — it is thirteen styles now, not eleven; wording fixed on the new list.

- [ ] 2.8 Today's book with colour images: run Inspect, tell me what it says about the images and how the print PDF's grey looks (5.15 landed; Ghosts had nothing to test on)

## 3 · P2 — talk deck from Venkat's "Have your factory call my factory" — BOTH

Source: https://protocolized.summerofprotocols.com/p/have-your-factory-call-my-factory · deck is `factory-talk.html`; never edit `/2026-pi-symposium/talk`. Talk: Wed 23 Sep 17:30 UTC, SIGP4B.

- [ ] 3.2 YOU — cut / reorder / add beats
- [ ] 3.4 BOTH — run-through in the browser; timing
- [ ] 3.5 ME — push; link in the handoff

## 4 · Workshop watch (Mon 21 / Tue 22) — BOTH

- [ ] 4.1 `/admin/factory/` (Floor) on a second screen; Shelley session open with `factory-tail`
- [ ] 4.2 Attendees free via hand-granted passes (New pass form) — no coupons, no $0 checkouts (form + attendee login verified Fri, 2.5)
- [ ] 4.3 Watch for `build failed`, `401` bursts (password confusion), builds "queued" (> 2 at once)
- [ ] 4.4 Tue night: confirm `systemctl list-timers prodcal-store-live.timer` still armed; Wed 00:00 HKT check the go-live email arrived and `/api/public/store/config` says live

## 5 · Queued — ME unless marked

- [ ] 5.17 **About the Book Factory** text (was 0.8) — YOU pick A / A′ / B and edit the placeholders (drafts are on this item's notes); then ME: transmittal checkbox → colophon on the copyright page and/or back-matter page, off by default

  > **shelley** · 2026-09-18 17:19 UTC  
  > Drafts are ready for your read (not ticking until you have): docs/brand/ABOUT-THE-BOOK-FACTORY-2026-09-18.md — also in the repo on GitHub. Three lengths: A colophon for the copyright page (78 w), A′ one line after the typesetting credit (26 w), B a back-matter page headed “About the Book Factory” (~490 w). Voice: plain third person, says what happened to the book, no adjectives; the URL is the only sell. Once you pick and mark up the wording I wire it as a transmittal checkbox (off by default): copyright-page colophon, or back-matter page, or both. The one-liner, so you can judge the voice here: “Set from the author's Word file by the jdbb studio book factory, print and EPUB from one source, no hand-composed pages. jdbbs.exe.xyz/factory”

  > **shelley** · 2026-09-18 17:59 UTC  
  > Sorry — I left these in a repo file, which you can’t open from this page. Here they are in full. Three lengths; the numbers are true of every build today (thirteen styles, ~9 s).
  > A. Colophon (copyright page, small type) — 78 words
  >  This book was made in the jdbb studio book factory. The author wrote it in Word, in a template of thirteen named styles generated from a one-page transmittal; the factory read the file, reported what it found, and set the print interior and the EPUB from that single source, in licensed type, with no hand-composed pages. Corrections are rebuilt the same way, so every copy of this edition is set from the same file. jdbbs.exe.xyz/factory
  >
  > A′. Colophon, shortest (one line after the typesetting credit) — 26 words
  >  Set from the author's Word file by the jdbb studio book factory, print and EPUB from one source, no hand-composed pages. jdbbs.exe.xyz/factory
  >  ---
  >
  > B. Back-matter page — "About the Book Factory" — 486 words
  >  **About the Book Factory**
  >  This book was not typeset in the usual sense. No one placed its pages by hand. It was built by the jdbb studio book factory, a small machine for turning a Word file into a book, and this page says how, because the method is part of what you are holding.
  >  The author wrote in Word, in a template the factory generated for this title. The template has thirteen named paragraph styles and nothing else: Normal, First Paragraph, three levels of heading, Block Quote, Epigraph, Verse, Code Block, Section Break, Copyright, Signature and Glossary Entry. Every paragraph in the manuscript carries one of those names. That is the whole contract between writer and factory. There is no software to learn and nothing to install; the work of authorship stays in the tool the author already used.
  >  Before the writing began, the author filled in a transmittal: a short form that records what the book is and how it should be set. The page size. The front matter — a dedication, an epigraph, a foreword — and its order. The copyright page, which the factory composes from the transmittal rather than asking anyone to type it. Special characters, mathematics, custom styles the book needs. The transmittal is the specification; the template is generated from it; the finished book follows it.
  >  When a draft was ready, the author uploaded it and the factory inspected it: a machine read of the file that reports which styles were used, where the file departs from the template, how many images it found and how large each will print, where the chapters begin. The report is the factory's reply. The author fixes what it flags and uploads again, for as long as it takes.
  >  Then the build. From that one Word file the factory produces two things at once: a print-ready interior PDF at the trim size on the transmittal, and an EPUB with the author's cover embedded. The interior is set in the studio's house design — a typographic system of margins, type sizes, running heads and spacing worked out once and applied to every book that passes through — in fonts the studio licenses for print. The two editions come from the same source, in the same pass, so they cannot drift apart.
  >  This matters after publication. A book set by hand is finished when the typesetter stops; a correction means opening the files again and hoping the lines still fall where they did. A book from the factory is set from its source every time. Fix the Word file, rebuild, and every page is composed afresh, in seconds. The edition you hold is one build of a file that can be built again.
  >  The factory is the work of jdbb studio, a one-person book studio that published books the ordinary way for thirty years before deciding that the ordinary way was mostly waiting. It is offered to anyone with a manuscript and a few hours. If you have written something and would like it to become a book like this one, the factory is at **jdbbs.exe.xyz/factory**.
  >  ---
  >  #
  > Notes for Jenna
  >  - "thirty years" and "one-person" in the last paragraph are placeholders for whatever is true; cut the sentence if you'd rather not do biography here. - "thirteen named paragraph styles" is true as of today (Signature and Glossary Entry added 2026-09-18). If the count changes, this line and the colophon change with it. - "in seconds" — a full build of a 250-page book is ~9 s on the VM; true. - The page is ~490 words; at 6 × 9 it fills one recto with air. If you want it on one page in Small trim, cut paragraph 4 (Inspect) to two sentences. - Suggested placement: last page of back matter, after About the Author, so the book ends on the author and then the maker.

- [ ] 5.18 New pass → fresh transmittal starts with empty title/author although the pass form asked for both; prefill `book.title` / `book.author` from the pass (found Fri during 2.5)
- [~] 5.5 P3 **H&J / composition quality** in the print PDF — Typst 0.12 levers checked, nothing to change before Monday; scorer in `typesetting/scripts/compscore.py`; runts need Typst 0.13 `costs` (Wed 24+); still to do: compare against an InDesign-set page (YOU supply)

  > **shelley** · 2026-09-18 14:11 UTC  
  > Scorer landed: `python3 typesetting/scripts/compscore.py book.pdf --list` counts loose lines, runts, widows/orphans, hyphen runs, stacks and rivers per book. Ghosts today: 27 loose, 42 runts, 0 widows/orphans, ~90 rivers over 2395 lines.
  >
  > The lever that actually fixes runts (Typst’s `costs`) needs Typst 0.13; we run 0.12. I tried 0.13 in scratch: runts 42 → 29 (−31%), and the template compiles after a one-line fix (landed, no output change on 0.12). But 0.13 also changes spacing around lists and on the Contents page, which shifts lines on ~30 pages of Ghosts — a page-by-page check on two books before it can go live. Recommendation: do the upgrade Wed 24 or after, not before attendees build. Full note: docs/reviews/P3-COMPOSITION-2026-09-18.md. Still want one InDesign-set page from you to run the same scorer on. Side-find: Ghosts ch.1 has 26 stray `>` from a quoted email in the .docx — Inspect should flag that (small, will add).

  > **shelley** · 2026-09-18 17:37 UTC  
  > Checked what else 0.12 can do, so 5.5 is not just waiting on the upgrade. Answer: nothing worth touching before Monday. Ghosts’ 27 “loose” lines are all 0.50–0.58 em — a hair over the yardstick, 1.1 per 100 lines, no hyphen stacks, no widows/orphans. That is clean composition; a compositor would leave it. Obliquities (same scorer, your existing PDF, no build spent) is looser: 3.5 per 100 lines, median 0.53 em, max 0.79, because the protocolized trim gives a ~3.4 in measure at the same 10 pt. The honest fix there is design — 9.5 pt or a hair less margin — a choice for the series spec, not a factory default to flip two days out. The 6 “orphans” it reports are numbered-list first lines (scorer sees the hanging number as a new column) — false positives, I’ll tighten the scorer. Runts remain the only real defect and that is the 0.13 `costs` lever, Wed 24+. Leaving 5.5 in progress for the InDesign page whenever you have one; otherwise it is done for the workshop.

- [ ] 5.12 Jev (typesafe.ai) pilot — heading classification behind a flag, per `docs/reviews/TYPESAFE-REVIEW-2026-09-18.md` (agreed 2026-09-18; post-workshop)
- [ ] 5.13 Index as a factory add-on (LLM-drafted conceptual index, Typst-resolved locators) — per `docs/reviews/INDEX-ADDON-FEASIBILITY-2026-09-18.md`; ~5 days; post-workshop
- [ ] 5.19 Tuned per-image grey conversion + proof sheet as a ~$100 add-on (from 5.15; post-workshop)
- [ ] 5.20 Typst 0.13 upgrade on the VM (Wed 24+; unlocks `par(costs:)` for 5.5)
- [ ] 5.9 4 vCPU bump — only if the workshop shows build queueing (load test: CPU-bound)
- [ ] 5.21 Plantin MT Pro **Bold** missing on the VM (see `docs/IDEAS.md`) — YOU supply the OTF; never committed

## 6 · Ideas log (`docs/IDEAS.md`) — BOTH decide, then queue or kill

- [ ] 6.1 Per-client interactive stylesheet instances (accept/reject seeded from the universal fiction/nonfiction stylesheet)

  > **jenna** · 2026-09-18 00:19 UTC  
  > Seems like this shouldn't be too bad to do since we trimmed down our universal fiction/non-fiction style sheets. Do you agree?

  > **shelley** · 2026-09-18 00:33 UTC  
  > Agree — the trimmed universal sheets make this ~half a day: a per-project table of accept/reject rows seeded from the universal sheet + the existing /stylesheet UI pointed at it. Queue post-workshop (5.10?).

  > **jenna** · 2026-09-18 01:08 UTC  
  > I'm going to make a big deal about style sheets in both senses (for the workshop and in my presentation to the symposium), and disambiguating style sheets (like in Microsoft Word) and the style sheets we mean here (lke whether to use the Oxford comma or not) feels worth doing now and handling exposing the UX in the client dashboard. In the workshop, I think I'd encourage them to accept all of our defaults for efficiency, but also to allow them to add or edit. Am I crazy?

- [ ] 6.2 Reusable "review table" pattern from the sitemap-review tool (JSON table + per-row reply box)

  > **jenna** · 2026-09-18 00:19 UTC  
  > I don't know enough about what you mean for this one.

  > **shelley** · 2026-09-18 00:33 UTC  
  > Means: the sitemap-review tool we used earlier (table of items, a reply box on each row — like this punch list) could be turned into a reusable component for future review rounds. My call: kill until a second concrete review round needs it.

  > **jenna** · 2026-09-18 01:06 UTC  
  > Hmm feels like, through the course of the workshop, I might very quickly end up with a bunch of small borks/nits/qols that we could fix in near real time, and a clean punch list to keep track of them instead of throwing things at you in a session would feel much safer. Maybe we make a blank page shaped like this that has a starter data entry field that I could use to kick off things as they come up on Monday and Tuesday.

- [ ] 6.3 Client-visible on-disk documents (`pi-client/{slug}/` + `serveClientDoc`) — trigger: first real client doc

  > **jenna** · 2026-09-18 00:20 UTC  
  > I don't know enough about what you mean for this one.

  > **shelley** · 2026-09-18 00:33 UTC  
  > Means: like /exedeck or /workshop, but per client — drop an HTML file in pi-client/{slug}/ and it shows only to that client, behind their portal login. Nothing needs it yet; keep on the ideas list until a real client doc exists.

  > **jenna** · 2026-09-18 01:04 UTC  
  > Hmmm we'll have a bunch of clients once the workshop participants start Monday, so this is something that needs to be ready before the workshop, unless I'm still not quite understanding what you mean here.

- [ ] 6.5 Print cover / spine calc — separate product question (cover is EPUB-only today)

  > **jenna** · 2026-09-18 00:22 UTC  
  > Yeah, let's not lose track of this, but the whole "how to get it printed" question is not built into the factory at all yet. I think this is a good one to keep on the idea list for now. I just did some good research, and the two print companies we could point people towards are Bookmobile and Accutrack. Add those to our research if we add this to the factory after the workshop.

  > **shelley** · 2026-09-18 00:33 UTC  
  > Added to docs/IDEAS.md with Bookmobile + Accutrack named. Stays on the ideas list.

- [~] 6.6 Strike the done rows (Admin doc editor) and move the kept ones to §5 with a date — ME, once 6.1–6.5 are decided

  > **jenna** · 2026-09-18 00:22 UTC  
  > I don't understand what this one means.

  > **shelley** · 2026-09-18 00:33 UTC  
  > Means: housekeeping of docs/IDEAS.md itself — strike rows that are done (did the Admin doc editor one tonight) and move the ones we keep into §5 with a date. It is just “tidy the list once 6.1–6.5 are decided”; I will do it when you have answered 6.1–6.4.

  > **jenna** · 2026-09-18 01:09 UTC  
  > Ah gotcha, yes let's keep Punch lists are up to date, and let's also figure out a way to keep a master list of punch lists. Maybe even add these links to our admin dashboard so that I can go back and double-check what we decided, in case I see something wrong or need to change my mind.

  > **jenna** · 2026-09-18 12:13 UTC  
  > didn't see a reply to this admin improvement tracker idea?

  > **shelley** · 2026-09-18 12:13 UTC  
  > Sorry, missed it. Yes. Plan (added as 5.10, Saturday, ~1 h): (1) the export of each punch list to docs/runs/PUNCHLIST-YYYY-MM-DD.md includes every note thread under its item, so the decisions travel with the list (today’s copies only have the ticks); (2) docs/runs/README.md is the master index; (3) a “Runs” card on the admin dashboard → /admin/runs/ that renders those files from disk, read-only, so you can re-read what we decided from any device and change your mind on the live list. Admin-only and read-only, so safe to land before the freeze.

  > **shelley** · 2026-09-18 14:03 UTC  
  > Done row struck (Admin doc editor → a Done section at the foot of docs/IDEAS.md). Already-decided ideas moved to §5 with dates: 5.7 magic-link (was there), 5.12 Jev pilot (new). 6.1–6.5 are still yours to keep or kill — tick or note each and I’ll move the keepers to §5 and drop the rest from IDEAS.

