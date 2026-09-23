<!-- exported 2026-09-23 13:42 UTC by scripts/punchlist-export.py; source scratch/run/CHECKLIST.md + notes.json -->

# Punchlist 5 · Tue 22 Sep work block · talk Wed 23

Legend: **YOU** = Jenna · **ME** = Shelley · **BOTH** = together. Click a box to cycle ☐ → ☑ → ◐; **note** opens a reply box (⌘↵ saves; paste screenshots in). Bottom bar adds an item to the Inbox. Short list on purpose — everything carried from list 4 lives in the archive: [list 4](/admin/runs/PUNCHLIST-2026-09-22-list4) · [all runs](/admin/runs/). Item numbers continue (next: 0.49 / 5.28) so threads stay attached.

**Freeze:** hotfix-only through Wed 23. Hotfix path: fix → `go test ./srv/` → 0 builds converting → `make build && sudo systemctl restart prodcal` → smoke on `mcheck/book-001`. Store flips to live Stripe **Wed 23 Sep 00:00 HKT**.

## 0 · Inbox — new items, untriaged
- [~] 0.36 **Seapunk before/after review** (project 32, books 46–55) — compare source .docx vs proof; suggestions for the factory + guidance Jenna can pass to the Seapunk team. Subagent, **next conversation** (Jenna).

  > **jenna** · 2026-09-22 14:24 UTC  
  > . after we start a new conversation --

  > **shelley** · 2026-09-22 14:25 UTC  
  > Understood — parked for the next conversation. Brief when it runs: project 32 (seapunkstudios), books 46–55 (source .docx in the DB blobs, proofs in book_outputs), compare before/after like the Fotis review, output: factory suggestions + a note you can forward to the Seapunk team.

  > **jenna** · 2026-09-22 14:28 UTC  
  > Pick this up now.

  > **shelley** · 2026-09-22 14:31 UTC  
  > Started. Scope: six uploads (46,47,48,49,50,55 — not 46–55; 51–54 are mcheck/Toby). Sources + proofs extracted to scratch/seapunk/books/. Early smell from sizes: 47+48 proofs (and the 47 FINAL) are template-sized (~30 KB) from 1.9 MB sources; 49 onward is the full book. Subagent cOI7OI4 digging in; report → docs/reviews/SEAPUNK-REVIEW-2026-09-22.md.

  > **shelley** · 2026-09-22 14:43 UTC  
  > Report: docs/reviews/SEAPUNK-REVIEW-2026-09-22.md (§7 = forwardable note). Gist: the team did everything right; bookmap.go title-as-H1 rule dropped the entire body (one H1 = title, five H2s) → 6-page proofs for 46/47/48 and TWO finals of an empty book (16:46 template test, 16:50). Inspect saw it ("dropped with 61 paragraphs") but filed it as low/ready. 49 fixed it by accident (first H2 → H1); 50/55 were busywork on lows. 55 is a real 26-page book with 3 issues: footnote marks all ¹ (series-template.typ:1249, one-liner), captions adrift from full-page images, one stray empty H2. At 19:08 they downloaded the EMPTY 47 final — they may think that is the deliverable. Hotfix candidates: (1) title rule never drops content, (2) refuse final when body words = 0 + show pages on confirm, (3) Inspect: empty body = high, (4) footnote mark. (9) reinstate 2 finals on pass 17 — your call.

- [ ] 0.41 Think about adding an affordance on the house style sheet for a book for adding new rules that they want to track.  ·  _added Tue 15:28 UTC_

  > **shelley** · 2026-09-22 15:29 UTC  
  > Checking what exists: the book sheet (/{client}/{project}/stylesheet/) already has “+ Add a rule to <Section>” at the foot of each section. Is the ask (a) make that more visible / put one Add at the top, (b) a free-form “decisions we are tracking” list (word list: e-mail not email; Seapunk capitalised) not tied to a house section, or (c) an Add on the public house sheet /stylesheet/ when reached from a book? Parked for Thu unless you say hotfix.

- [ ] 0.42 Give me a sentence or two I can use to explain to the workshop how footnotes are handled in the Book Factory.  ·  _added Tue 15:44 UTC_

  > **shelley** · 2026-09-22 15:44 UTC  
  > Use Word’s own footnotes (References → Insert Footnote) — nothing else to do. In the print PDF they are set as true footnotes at the foot of the page they belong to, numbered through the book, in the book’s body face. In the EPUB the same notes become tap-to-read pop-up notes (EPUB 3 footnote semantics, with a Notes list at the end for readers that don’t pop up). Endnotes typed by hand as a “Notes” section at the back are also kept, as a back-matter section. (Known issue we are fixing today: in the print PDF every in-text mark currently prints as ¹ instead of ¹ ² ³; the notes themselves are numbered correctly.)

- [ ] 0.43 Can we add an affordance for native LaTeX for complicated formulas, mathematical formulas, etc.?  ·  _added Tue 15:52 UTC_

  > **shelley** · 2026-09-22 15:53 UTC  
  > Good news: it already works, nothing to build. Word’s own equation editor (Insert → Equation; also accepts LaTeX typed into the equation box) comes through pandoc as real math → Typst sets it natively in the PDF (inline and display), and the EPUB carries MathML. Verified with a live mcheck proof just now (book 61: E=mc², a display integral — both render). Caveats: MathML support in reading apps varies (Apple Books fine, Kindle patchy); Inspect says nothing about equations yet. What we do NOT have: LaTeX typed as plain text ($…$ in a Normal paragraph) — that would be an inline marker [[tex]]…[[/tex]] in the 5.26 slot (Thu, ~2 h). Workshop line: “Use Word’s equation editor — you can type LaTeX into it — and the factory sets the maths natively in print and as MathML in the EPUB.”

- [ ] 0.44 Can we add an affordance for side notes, and if so, how?  ·  _added Tue 15:52 UTC_

  > **shelley** · 2026-09-22 15:54 UTC  
  > Not today; yes, and here is how. Author side: a factory paragraph style Side Note (marker [[sidenote]], or a declared custom style based on it) on the paragraph that follows the sentence it belongs to — same “styles are names” contract as everything else. Factory side, two renderings: (1) PRINT with a spec switch “margin notes”: widens the outer margin (template already has inside/outside margins as config, series-template.typ:153) and Typst places the note in the outer margin beside its anchor, small size, in the body face — a Tufte-style page; the switch changes the whole book’s text block, so it is a design decision per title, not per note. (2) Without the switch, and always in the EPUB (reflow has no margin): a set-off aside — indented, smaller, thin rule — as <aside epub:type=sidebar>. Cost: style in the 3 lists + Lua + one Typst function + EPUB css + template sample ≈ half a day, Thu/Fri after 5.25. Question for you: do you want the margin-notes page design at all, or is the set-off aside enough for the workshop books?

- [ ] 0.45 Let's look at whether the template we produce for people addresses character-level style stuff, formatting for bold and italic, to make clear that they don't need to create markers for those.  ·  _added Tue 16:18 UTC_

  > **shelley** · 2026-09-22 16:18 UTC  
  > Checked: it does not. The template guide (generate-word-template.py ~L584–750) covers every paragraph style and the three zones, and never says the words italic or bold; neither does /factory #bring nor the welcome/template emails. Fotis is the proof — she declared BOLD and ITAL as custom styles and typed [[ITAL]] mid-sentence 18 times. Proposal (copy only, freeze-safe, ~30 min): one short section “Inside a paragraph” in the guide, one line in #bring step 3, one line in the template email: “Italic, bold, small caps, footnotes and equations are not styles and need no markers — use Word’s own buttons (⌘I, ⌘B, References → Insert Footnote, Insert → Equation). Markers and styles are for whole paragraphs only.” Also: the sample body paragraph in the template should actually contain an italic word and a footnote so people see it. Ship with today’s hotfix build? (the guide text is in Python read at request time; #bring is live on save; only the email needs the build).

- [ ] 0.46 Let's look at Mike Casey's custom style for a sans serif version of Normal  ·  _added Tue 16:24 UTC_

  > **shelley** · 2026-09-22 16:25 UTC  
  > Looked (project 18 mcasey, book 62 built 15:58 today). Two findings. (1) HIS MARKERS DID NOT APPLY: he named the style on the transmittal literally “[[commentary]]” — brackets included — so the 4 [[commentary]] paragraphs stayed unresolved and the proof prints the literal “[[commentary]]” four times (log: resolved=0 unresolved=4). Tiny factory fix: strip [[ ]] (and a leading “style:”) when normalising declared names, both in apply-style-markers.py normalize() L79 and customstyles.go; Inspect should say “declared as ‘[[commentary]]’ — the brackets belong in the text, not the name”. Adding to today’s hotfix as (e). (2) EVEN RESOLVED IT WOULD RENDER AS PLAIN NORMAL: his description asks for a sans face to mark commentary vs original text; preset generic-paragraph = body, and the 5.25 design (based on + indent + space before) has no way to say “sans”. Proposal for Thu 5.25 slice: add a Face option per custom style — body / heading / code — mapped to the spec’s three fonts (heading is the sans in most of our specs), Word template gets the matching font, EPUB class gets font-family. That gives “sans Normal” with one dropdown and no free-form fonts. For the workshop today: I can set his typst snippet by hand in admin (text(font: config.heading-font)) once (e) lands, so his next proof shows the commentary in sans — say yes and I will.


- [ ] 0.47 **Store: PROTOCOL50 never created in Stripe.** Go-live bootstrap at 16:00 UTC stopped on WORKSHOP49 (its expiry, Tue 23:59 HKT, was already in the past → Stripe 400) before reaching PROTOCOL50. Prices are fine; checkout works; the alumni code does not exist yet. Hotfix: skip expired promos in `ensurePromo` (store.go:224). Also the go-live note to Jenna failed (`source` does not export AGENTMAIL_INBOX_ID to python) — notification only.
- [ ] 0.48 **cblass (24) build 60 failed Tue 15:58** — declared style named `center` → `#let center(content)` shadows Typst's `center` → “expected content, found function” at series-template.typ:381. Same class as Fotis's `break`. Hotfix: reserved-word list in `typstStyleIdent` (bookspecs.go:790) → prefix `cs-`. Not retried since; email him once fixed.

## 1 · Today — BOTH
- [ ] 4.3 Floor watch: `build failed`, 401 bursts, queued builds; Sam 31 · Toby 29 · Fotis 30 (can retry; prune declarations) · Ellen 33 · new participant

  > **shelley** · 2026-09-21 16:37 UTC  
  > **build failed — fixed, no restart.** 16:10 UTC, skhoo / *Transcendence and Immanence* (project 31, book 45, proof): “pagebreaks are not allowed inside of containers”. Cause: his front matter is Half Title → Title → Subtitle ×2 → Copyright ×3 → Dedication → Epigraph. Pandoc only lifts *Title* into metadata when it's the very first paragraph; here it wasn't, so Title + Subtitle reached the Lua filter as paragraphs that the book map had already dropped and not counted — every front-matter piece boundary shifted by 3 and the Copyright paragraphs landed inside the dedication/epigraph containers, where the copyright page's pagebreak is illegal. Fix `26ff8e9`: the filter drops Title/Subtitle exactly like the map does (regression test mirrors his file). The filter is read from disk per build, so it's live already — **no restart, nothing in flight touched.** Verified: his own failed book.typ now compiles (124 pp) with the fixed filter; mcheck proof smoke green (0.4 s). **Ask Sam to press Build again** — his upload is still there, no re-upload needed.

  > **jenna** · 2026-09-21 17:10 UTC  
  > Is this resolved?

  > **shelley** · 2026-09-21 17:10 UTC  
  > **Yes, resolved on our side** — fix live since 16:37 UTC, and I've since run Sam's own file (the newer .docx you sent me) through the live pipeline on mcheck: proof built green, 224 pp, front matter as designed (his typed ©/CIP dropped, dedication and epigraphs on their own pages). **Still open on his side:** he hasn't pressed Build again yet (no build on project 31 since the 16:10 failure). Tell him to hit **Build → Proof** on his factory page; upload is intact, nothing to redo. If it fails again I'll see it here within seconds.

  > **shelley** · 2026-09-21 17:35 UTC  
  > **Second one, from chat — Toby / *Devotion* (project 29), 'weird output'. Fixed live (`0a3d…` see git), no restart.** Ours: (1) each `[[verse]]` line became its own poem block with padding above and below → double-spaced look; consecutive Verse lines now merge into one block per stanza. (2) A bare `[[break]]` paragraph was empty after the marker came off and pandoc drops empty paragraphs → every stanza break vanished; the break now survives. Verified on mcheck with his file (book 54): lines stack, stanza gaps show. His file: poem titles (REMINDER, ONE SEASON…) are plain Body A caps, not Heading 1 — so no poem-per-page, no contents; fix = Heading style in Pages, or `[[h1]]` at the start of the title line. And `[[verse2]]`/`[[verse3]]`/`[[break2]]` aren't styles (his invention for indented lines) — they print literally in 28 places; remove or use plain `[[verse]]`. Verse is still centred italic — that's our design for a poem quoted in prose; a poetry-collection shape is **5.23** (after workshop). He should press Build → Proof again.

  > **shelley** · 2026-09-21 17:35 UTC  
  > (commit for the Devotion fix is `3411728`, pushed.)

- [ ] 4.4 Tue night: `prodcal-store-live.timer` armed; Wed 00:00 HKT go-live email + `/api/public/store/config` live
- [ ] 3.2 YOU — talk deck: cut / reorder / add beats · 3.4 BOTH run-through · 3.5 ME push + link

  > **shelley** · 2026-09-20 18:21 UTC  
  > Your Mac/Claude deck is up, untouched: **https://jdbbs.exe.xyz/2026-pi-symposium/og-protocols** — “Books are OG protocols”, 26 slides. The saved HTML had everything inlined (46 webp images + the two title videos, 11 MB) except the Google-fonts CSS, which pointed at a local `_files/` folder; I inlined that one file into the head and changed nothing else. Loads clean here: title video plays, JetBrains Mono renders, arrow keys step, Notes (N) toggles, no console errors. The existing deck at /2026-pi-symposium/talk is still there for comparison. Tell me what you want changed and I’ll work on this copy.


## 2 · Thu 24 — ME (after the freeze)
- [~] 5.25 **Custom styles slices 4–7** — EPUB classes + Word template basedOn/indent; verse-family coalesce into one poem block; admin panel Based on/Indent; docs (#bring, WORD-FREE note). [design note](https://github.com/djinna/jdbbs/blob/main/docs/reviews/CUSTOM-STYLES-MARKERS-2026-09-22.md)

  > **shelley** · 2026-09-22T14:24:48Z  
  > Slices 1–3 verified on mcheck book 58 after the 0.39 hotfix: 457/457 markers resolve; verse → verse2 → verse3 step in visibly (p.12, Four Haiku). Open: verse/verse2/verse3 lines do not coalesce into one poem block (each level is its own block with a stanza gap between) — fix in slice 4 with the EPUB classes. Slices 4–7 remain for Thu 24.

- [ ] 5.26 **Inline character markers** `[[code]]…[[/code]]` (Fotis's ITAL/BOLD leak is the test case)
- [ ] 5.24 + 5.27 **Error / Inspect wording** — "our bug, email us" when we can't be sure; helpful not scolding

  > **shelley** · 2026-09-22 13:57 UTC  
  > Fotis review is in: https://github.com/djinna/jdbbs/blob/main/docs/reviews/FOTIS-HELO-WORD-REVIEW-2026-09-22.md — summary, timeline, census of his file, 15-row factory-vs-author table, rewrites of every error string he saw (incl. the two build-failed texts: 'our generated config' vs 'your paragraph'), and a short note you can send him. Root cause was ours (style named `break` → `#let break(content)` in our config; fixed in 9258e56, he can retry). Handing off to a fresh session now.

  > **jenna** · 2026-09-22 14:18 UTC  
  > x https://jdbbs.exe.xyz/hermescorp/book-001/stylesheet/
  > It looks like Fotis's book has a whole bunch of custom styles. Can you check whether the factory added those somehow during our review just now, or did Fotis misunderstand and think that she was supposed to add all of the styles that are in her book, not just the ones that are custom? Pass this to a new sub-agent in our current conversation.?
  >
  > ![screenshot](img/20260922-141747-400.png)

  > **shelley** · 2026-09-22 14:19 UTC  
  > Passed to a subagent (fotis-styles-origin, read-only). It is checking transmittal_versions for project 30 over time + every code path that writes custom_styles. Report to scratch/fotis/STYLES-ORIGIN.md; I will post the verdict here. Prior review (FOTIS-HELO-WORD-REVIEW §3) found 21 declarations, 9 duplicating factory/Word built-ins — consistent with declaring every style, but the timeline check will settle it.

  > **shelley** · 2026-09-22 14:22 UTC  
  > Verdict (b): Fotis typed them; the factory never adds declarations (only the transmittal PUT and the admin spec editor write custom_styles — upload/Inspect/build only copy a final transmittal into the spec). Her transmittal history grows 0→4→11→14→18→19→21 across client saves on Mon 21 Sep, starting 12:12 UTC right after Inspect flagged 2 Highs; 12:12:32 Inspect showed 0. So she read "declare your styles" as "declare every Word style". For Fotis: "You entered the full Word style inventory — the factory didn't add that. Declare only the ones that aren't factory styles: keep narration + in1–in6, drop verse/quote/break/part/chapter/preface/BOLD/ITAL/footnote*." Feeds 0.34 (onboarding: styles are names). Report: scratch/fotis/STYLES-ORIGIN.md

- [ ] 0.34 **Onboarding: styles are names** — welcome email, /factory intro, #bring (Fotis declared all 21 Word styles: evidence for this)

  > **shelley** · 2026-09-22 13:43 UTC  
  > Triage: After-workshop → §5 as 0.34 (parked, per your note). Scope I'll take: the welcome email, /factory intro + #bring, the transmittal's Custom Styles help text, and the Word template's own Template Guide page — one message repeated in the same words everywhere: “a style is a *name* you give a paragraph; the factory reads the names, not the look.” Also worth a 90-second screencast/GIF of applying a style in Word / Pages / Docs, since the blind spot is about the mechanics, not the concept. I'll fold in what Fotis's and Toby's files teach (both marked every paragraph by hand rather than using the style sheet). Nothing ships during the freeze.

- [ ] 0.40 **`Chapter Title` factory style** — H1/H2 become real 1st/2nd-level heads; normalisation pre-pass (Chapter Title→H1, H1/2/3→H2/3/4 only when the file uses Chapter Title, so existing manuscripts unchanged); mirrored in bookmap + chapter_detect; template/emails/#bring copy; then retire the title-drop rule that emptied Seapunk's book. Design: [Seapunk review §6.6](/admin/runs/) · ~1 day (Thu)

  > **shelley** · 2026-09-22 14:43 UTC  
  > Triaged from inbox (was auto-numbered 0.37, already taken by the style-sheet link). This is item 6 of the Seapunk review — their whole book was dropped because the title-as-H1 rule swallowed everything under it. Proposal: new factory style Chapter Title; a pre-pass shifts Chapter Title→H1 and H1/2/3→H2/3/4 only when a file uses Chapter Title, so nothing already on the floor changes. Thursday, ~1 day. Today I only hotfix the symptom (title rule must never drop content; finals refuse an empty body).

- [ ] 0.31 Drop `PRODCAL_BUILD_QC_EMAIL` from .env after the workshop

  > **shelley** · 2026-09-22 13:17 UTC  
  > Live (2b73cc8, restarted 13:14 UTC, mcheck smoke green). You got the first one: “QC · mcheck · … (proof both, book 54)”. Links, not attachments — before = the uploaded .docx (new /download/source route), after = newest PDF/EPUB of that kind, plus Inspect report, factory page, Floor. Failed builds mail too (FAILED in the subject). On the Floor every “book N” now has [docx · pdf · epub]. Remind me to switch it off after the workshop.

