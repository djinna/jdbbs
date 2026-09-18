<!-- exported 2026-09-18 17:14 UTC by scripts/punchlist-export.py; source scratch/run/CHECKLIST.md + notes.json -->

# Pre-workshop punch list + parked work · Thu 17 → Sat 20 Sep

Legend: **YOU** = Jenna's turn · **ME** = Shelley's turn · **BOTH** = look together.
Ticked = done · ◐ = in progress. Click a box to cycle ☐ → ☑ → ◐. **note** opens a reply box under any item (⌘↵ saves) — I read those back. Bottom bar adds a new item to an Inbox I triage. Refreshes every 15 s (pauses while you type). The previous run checklist (Protocol Institute Test 2) is committed at `docs/runs/RUN-2026-09-17-protocol-institute.md`.

**Freeze:** Fri 19 – Sat 20 fix / test / clean allowed (build → `go test ./srv/` → restart → one factory smoke as `pinstitute`). Sun 21 – Tue 23 hotfix-only. Store flips to live Stripe by timer **Wed 23 Sep 00:00 HKT**.

## 1 · Fix window — ME

- [x] 1.1 **Word template**: 11 factory styles get `<w:qFormat/>` + `<w:uiPriority>` so Word's Styles pane (Recommended view) isn't empty. Verify with `scripts/docx-preview.sh` + unzip `styles.xml`

  > **shelley** · 2026-09-18 00:12 UTC  
  > Verified: the deployed generator already emits qFormat + uiPriority on all 11 styles (everything else semiHidden). Only Word itself can confirm the pane — that is 2.7.

- [x] 1.2 Sender address `studio@` → `factory@mail.jdbb.studio` (display name stays "jdbb studio"); one test send — approved
- [x] 1.3 Confirm nothing else hardcodes AgentMail as *sender* (archive CC rows `jdbb@agentmail.to` are fine); fix `snapshot_email.go` error text
- [x] 1.4 Strike the done "Admin doc editor" row in `docs/IDEAS.md`
- [x] 1.5 Checkpoint tag per `CHECKPOINTS.md`; full `go test ./srv/`; smoke `/factory`, `/admin/factory/`, `/admin/store/`, `/pinstitute/book-001/factory/` — tonight after 1.1–1.4 (re-tag after any later fix) — tag `checkpoint-2026-09-18-pre-workshop` at `0cd67fd`
- [x] 1.6 Dependabot: 5 vulns on djinna/jdbbs (3 high, 2 moderate) — check whether these were already fixed a day or two ago and just not marked; if new, fix now (rule: always address vulns)

  > **shelley** · 2026-09-18 00:13 UTC  
  > govulncheck on the VM: our code affected by 0 vulns (one advisory, x/crypto/openpgp, which we never call). Last dep bump was Sep 16 (bb4464f). I cannot see GitHub alerts from the VM — paste the 5 alert names (package + severity) here and I will fix the real ones.

  > **shelley** · 2026-09-18 00:20 UTC  
  > Fixed 894196d: grpc 1.79.3→1.83.2 (the 4 grpc alerts; 1.84.0 is still affected by one, 1.83.2 is clean per OSV), cel-go 0.26.1→0.31.0 (0.32 moved module path). Both indirect via the sqlc tool. Tests green, restarted, checkpoint tag moved. GitHub re-scans within ~an hour; the push banner still says 5 until then — refresh the alerts page later and tell me if any remain.

  > **jenna** · 2026-09-18 00:27 UTC  
  > all vulns now cleared on https://github.com/djinna/jdbbs/security/dependabot 👍


## 2 · Your side — YOU

- [x] 2.1 **C4** email footer string — done (Jenna, earlier; was carried forward by mistake — lesson: work from this shared list, not from memory)
- [x] 2.2 Sender address: `factory@` — yes
- [x] 2.3 Cohort resend from your own address
- [x] 2.4 Cal.com timezone — America/New_York confirmed correct. (Note to self: Jenna is **not** in HKT; HKT was chosen only for the store changeover / promo deadlines because of Asia participants.)
- [ ] 2.5 Try the **New pass form** on `/admin/store/` with a throwaway address (this is the attendee-shows-up path Mon/Tue)
- [ ] 2.6 Unlock the `pinstitute` portal and eyeball the new cards (Transmittal · Factory → · Calendar, `[FACTORY PASS]`)
- [ ] 2.7 After 1.1: re-download the Word template, confirm in Word — Styles pane shows the 11 factory styles, fonts Georgia/Arial/Courier New, page = trim

## 3 · P2 — talk deck from Venkat's "Have your factory call my factory" — BOTH

Source: https://protocolized.summerofprotocols.com/p/have-your-factory-call-my-factory · template only from `/2026-pi-symposium/talk` (never edit that file). Talk: Wed 23 Sep 17:30 UTC, SIGP4B.

- [x] 3.1 ME — read the post; draft **≤ 10 beats** (one line each) → paste here for you

  > **jenna** · 2026-09-18 01:12 UTC  
  > I don't see where you pasted this text for me :/

  > **jenna** · 2026-09-18 01:17 UTC  
  > Okay, asking this question again

  > **shelley** · 2026-09-18 01:17 UTC  
  > Beats (one line each):
  > 1. In March, Venkat wrote that the interesting thing in agentic AI is not the agents; it is two people's factories exchanging half-finished work.
  > 2. Venkat's factory makes manuscripts out of archives. Mine makes books out of manuscripts. The word "factory" turned out to be literal, not a metaphor.
  > 3. What we built: a Factory Pass, a form, a Word template with eleven named styles, an Inspect report, and a Build button. Money goes in one end, an EPUB and a print PDF come out the other.
  > 4. On 17 September the first real client went through end to end. The manuscript was a Google Docs export of Venkat's own writing, and it built in about four seconds.
  > 5. The first run produced twenty-seven catches, and most of them were not in the pipeline. They were in the paperwork around it.
  > 6. The transmittal form was borrowed from press-era production paperwork, and the first client showed us most of it is addressed to a factory that no longer exists.
  > 7. One ASCII-art bunny in a tweet made the EPUB twenty megabytes. Protocols fail at the edges of their vocabulary, and the edges are where the humans are.
  > 8. "Your factory calling my factory" for books means four concrete interfaces, and three of them already exist: the DOCX with named styles, the transmittal as JSON, the Inspect report, and the EPUB.
  > 9. The two factories are still linked by a shared folder and a person reading a report. That is fine for two friends. It is the thing to fix before there are twenty.
  > 10. The ask: if you run a factory that produces text, send me a manuscript in eleven styles and I will send you back a book. Then let us write down what passed between us.
  >
  > Open questions:
  > 1. Attribution on stage: do you want to open with Venkat's post by name and title (and say "I am the other factory in it"), or describe the idea and credit him once at the end? The first is stronger but puts his framing at the centre of your talk.
  > 2. Live demo or clip: the whole path (transmittal → template → upload → Inspect → build) runs in under two minutes wall clock. Live is convincing but the store flips to live Stripe at 00:00 HKT that same day; a recorded clip plus a live Inspect report might be the safer split.
  > 3. How much of the 17 September run to show: the log lines and the 27 catches are the most concrete material in the deck, but they also expose the rough edges of a product you are selling in the same room. Where is your line?
  > 4. Beat 8/9 make a half-promise about a machine-callable transmittal and Inspect endpoint. Is that something you want to commit to in public, or should those beats stay descriptive ("here is what would be needed")?
  > 5. Naming Venkat's manuscripts (Obliquities, The Twitter Years) and the Protocol Institute as the client: does he need to agree before you put the file names and page counts on a slide?
  >
  > Full draft with speaker notes: scratch/p2-beats.md (also pasted in chat).

- [ ] 3.2 YOU — cut / reorder / add beats
- [x] 3.3 ME — build the deck as a new page in `~/jdbbs-public` on the talk template

  > **shelley** · 2026-09-18 14:14 UTC  
  > Building the deck now from the 10 beats as drafted (all ten in; you cut afterwards, which is easier than cutting on paper). Will land at https://jdbbs.exe.xyz/2026-pi-symposium/factory-talk with speaker notes under each slide. Your 3.2 cuts can then be applied on the finished slides.

  > **shelley** · 2026-09-18 14:28 UTC  
  > Draft deck is up: https://jdbbs.exe.xyz/2026-pi-symposium/factory-talk — 12 slides (~14 min): title, the ten beats, close. Same template as the symposium talk (jump menu, arrow keys, `n` toggles speaker notes on every slide, print = handout). The swim-lane diagram is slide 4. Beats 8/9 are worded “what would be needed”, not a promise.
  >
  > When you cut (3.2), decide: (1) attribution — Venkat named on the title slide now; soften? (2) name Protocol Institute / Obliquities / The Twitter Years on 5, 6, 8, or anonymise? (3) how much rough edge on 6 and 10 (the 27 catches)? (4) flip slide 9 to a commitment if you say yes on 5.8; (5) it runs dense — drop slide 3 or 7 to land at ~12 min. Tell me cuts as slide numbers and I apply them; or edit the text yourself in /admin/docs/ (pick 2026-pi-symposium/factory-talk.html).

- [ ] 3.4 BOTH — run-through in the browser; timing
- [ ] 3.5 ME — push; link in the handoff

## 4 · Workshop watch (Mon 21 / Tue 22) — BOTH

- [ ] 4.1 `/admin/factory/` (Floor) on a second screen; Shelley session open with `factory-tail`
- [ ] 4.2 Attendees free via hand-granted passes (New pass form) — no coupons, no $0 checkouts
- [ ] 4.3 Watch for `build failed`, `401` bursts (password confusion), builds "queued" (> 2 at once)
- [ ] 4.4 Tue night: confirm `systemctl list-timers prodcal-store-live.timer` still armed; Wed 00:00 HKT check the go-live email arrived and `/api/public/store/config` says live

## 5 · Next session (fresh compaction) — queued, in order — ME unless marked

Not parked. Start after §1 is ticked. Each item: small commits, tests green, restart, smoke as `pinstitute`; anything touching the factory path during Sun–Tue waits for a hotfix window or lands Wed+.

- [x] 5.1 **Deploy-safe first (can land Fri/Sat):** C24 — one line on `/factory` that Kindle takes EPUB (no `.mobi`)

  > **shelley** · 2026-09-18 00:33 UTC  
  > Done, live on /factory (jdbbs-public 3 builds bullet): “The EPUB is what Kindle, Apple Books, Kobo and IngramSpark all take; there is no separate .mobi to make.”

- [x] 5.2 C27 CJK embed rule: fire only at ≥ 20 ideographs, and/or subset with `pyftsubset` (the 20 MB EPUB from one ASCII-art tweet)

  > **shelley** · 2026-09-18 00:33 UTC  
  > Done, deployed. Rule: Noto Serif TC only at ≥20 CJK runes; whatever is embedded (CJK or Thai) is subsetted to the glyphs used (fontTools). Smoke on real books: Ghosts test 002 EPUB 13.8 MB → 96 KB (5 ideographs dropped to reader fonts, 17 Thai chars keep an 18 KB subset face); Twitter Years now 7.3 MB of images, no fonts. Full fonts still embed if fontTools is ever missing.

- [x] 5.3 P4 **front matter ingestion** — decisions already taken 2026-09-17 (every section head = Heading 1; classify by heading text against a closed vocabulary + position; untitled pre-H1 blocks = dedication/epigraph in transmittal order; title + © pages generated from the transmittal; Parts opt-in shifts chapters to H2; arabic 1 on a recto). Inspect prints the resulting book map

  > **shelley** · 2026-09-18 01:19 UTC  
  > Pipeline mapped, nothing changed yet. Plan + code map in docs/reviews/P4-FRONT-MATTER-PLAN-2026-09-18.md (6 steps, commit order). Implementation starts next session.

  > **shelley** · 2026-09-18 01:59 UTC  
  > Steps 1–3 of 6 done and deployed (7c01341): book map in Inspect (factory page + report), print build generates i–iv from the transmittal, roman folios → arabic 1 on a recto, contents after dedication/epigraph, and — new for every book — running heads + drop folios (the factory PDFs had none before). Smoked on Twitter Years + Ghosts, no credits used. Left: EPUB side (drop title/byline/Contents, landmarks), Word template guide text, Parts opt-in.

  > **shelley** · 2026-09-18 12:19 UTC  
  > Step 4 of 6 done and deployed (a00ed84): the EPUB now reads the same book map — byline, typed copyright and typed Contents are dropped (the EPUB has its own title page + nav), dedication/epigraph get their own pages, and every section carries an epub:type so readers know front matter from body. epubcheck clean on Ghosts, Twitter Years and the test fixture. Next: step 5 Word template guide text, step 6 Parts opt-in.

  > **shelley** · 2026-09-18 12:49 UTC  
  > All six steps landed and deployed (7be8be5). Step 5: the Word template guide now has a “How the book is assembled” section — what is generated (do not type: half-title, title page, copyright, Contents), what to type before the first Heading 1 (dedication, epigraph, each on its own page), which heading names count as front/back matter. Step 6: parts. Opt-in = the Parts count on the transmittal (any number ≥ 1). Then Heading 1 = part opener (own right-hand page, blank verso), Heading 2 = chapter, Contents lists both. No live book has a Parts count, so nothing changes for existing books; Ghosts rebuilt identical. Found a latent template bug while in there (running_heads.enabled false is ignored) — logged in the P4 plan doc for after the workshop.

- [x] 5.4 **C13 transmittal rewrite for the factory** (post-workshop, the big one):

  > **shelley** · 2026-09-18 14:37 UTC  
  > All five done and deployed (one commit each, a2b4068 last). In plain terms, the transmittal lost: Production section (Mechs Delivery, Weeks in Prod., Bound Book Date, second Transmittal Date) — Print Run moved into Book, one optional Target date added · typed Chapters / Words / MS pp / Est. pp — now read from your last Inspect (Zoothesia shows 9 chapters, 23,981 words, 1 image); Parts stays typed and now actually drives the build · Developmental Edit + Level of Copyediting (section is now Typography notes) · Permissions status/date grid — replaced by one line + an attestation checkbox, with a new plain-English https://jdbbs.exe.xyz/factory/terms page · Pub Info & © is now the **Copyright page builder** (year, holder, publisher + city, edition line, ISBNs from Book, cover credit, interior credit, LoC line, printed-in, notices) and both the Word template and the Typst page iv print from it. Nothing typed before is lost: old keys still load and save. Please read the terms page — it is your voice on a legal-ish page, so edit freely in /admin/docs/. Details: docs/reviews/C13-TRANSMITTAL-2026-09-18.md.

    - [x] 5.4a C6 drop press-era Production section (Mechs Delivery, Weeks in Prod., Bound Book Date, dup Transmittal Date); Print Run → Book; one optional "Target date"; keep old JSON keys readable
    - [x] 5.4b C7 stop asking chapters / words / MS pp / est. book pp — Inspect counts them; fix missing input underline meanwhile

      > **shelley** · 2026-09-18 13:16 UTC  
      > Interim fix landed and deployed: the stats row (Parts / Chapters / Words / MS pp / Est. Book pp) has its underlines back — the checklist table’s inline-field rule was hiding them. The bigger C7 change (stop asking; Inspect counts) stays post-workshop with the rest of 5.4. Note: the Parts field is now also the parts opt-in for the build (any number ≥ 1 = book has parts; “none”/blank = no).

    - [x] 5.4c C10 drop Developmental Edit + Level of Copyediting; keep Special Characters etc.
    - [x] 5.4d C11 Permissions → one courtesy line + one attestation checkbox; terms text before Wed go-live if possible
    - [x] 5.4e C12 Pub Info & © → the **copyright-page builder** (credit fields actually used)
- [~] 5.5 P3 **H&J / composition quality** in the print PDF: typst `par(costs:)`, optimized linebreaks, a loose-lines / rivers / runts / widows scorer on the built PDF; compare against an InDesign-set page

  > **shelley** · 2026-09-18 14:11 UTC  
  > Scorer landed: `python3 typesetting/scripts/compscore.py book.pdf --list` counts loose lines, runts, widows/orphans, hyphen runs, stacks and rivers per book. Ghosts today: 27 loose, 42 runts, 0 widows/orphans, ~90 rivers over 2395 lines.
  >
  > The lever that actually fixes runts (Typst’s `costs`) needs Typst 0.13; we run 0.12. I tried 0.13 in scratch: runts 42 → 29 (−31%), and the template compiles after a one-line fix (landed, no output change on 0.12). But 0.13 also changes spacing around lists and on the Contents page, which shifts lines on ~30 pages of Ghosts — a page-by-page check on two books before it can go live. Recommendation: do the upgrade Wed 24 or after, not before attendees build. Full note: docs/reviews/P3-COMPOSITION-2026-09-18.md. Still want one InDesign-set page from you to run the same scorer on. Side-find: Ghosts ch.1 has 26 stray `>` from a quoted email in the .docx — Inspect should flag that (small, will add).

- [x] 5.6 P1 review docs.typesafe.ai/introduction — anything for the factory? Write a ½-page note, then decide

  > **shelley** · 2026-09-18 12:02 UTC  
  > Note written: docs/reviews/TYPESAFE-REVIEW-2026-09-18.md. Short version — Jev is a hosted classifier, not a document tool or an agent: you send text + narrow typed questions (pick-one / score / yes-no) and get answers with confidence; cents per book. It cannot make the pandoc↔typst glue less fragile or fix styles. Where it could plug in: heading front/body/back classification in the book map (replaces our vocab lists — best fit), heading-lookalike detection, pre-filling the transmittal from the manuscript, ranking Inspect findings. Risks: network call + API key inside a build path that is hermetic today, nondeterminism, manuscripts leave the building (zero-retention is enterprise-only). Recommendation: pilot on heading classification only, behind a flag with the vocab fallback and cached answers — after the workshop. Park the rest. Your call: pilot / park.

  > **shelley** · 2026-09-18 12:09 UTC  
  > Jenna agreed the recommendation (via 0.1). Pilot goes on docs/IDEAS.md for after the workshop. Closed.

  > **jenna** · 2026-09-18 12:09 UTC  
  > Just confirming that I see your note here in the 5.6 punchlist thread now

- [x] 5.7 **Magic-link client login** (email → one-time link → cookie) replacing emailed passwords

  > **shelley** · 2026-09-18 14:26 UTC  
  > Landed (commits cc94692, cd6ef7b, 59a1637). Portal and factory gates now lead with “enter your email → we send a sign-in link” (30 min, single use); the password box is behind “Have a password? Use it instead” so every existing password still works. Fulfilment email now says “Easiest way in: open your portal and enter this email address — we’ll send you a sign-in link” above the password. Verified end-to-end on the VM with a throwaway client, real send via Resend (delivered@resend.dev), redeem → signed in, second use → expired page. Details: srv/EMAIL_SYSTEM.md pathway 8.

  > **shelley** · 2026-09-18 14:27 UTC  
  > Done and deployed. Client portal and factory gates now lead with **Sign in by email**: type the address the pass was sent to → we email a link good for 30 minutes, one use → it sets the same cookie a password login did. Passwords still work behind “Have a password? Use it instead”, so nothing changes for people holding the old email. The fulfilment email keeps the password for the workshop week but now says the email route is the easy way in. Wrong/unknown address → same “check your email” reply (no hint whether the account exists); max 3 links per client per 15 min. Floor feed shows link sent / denied. Tests cover issue, redeem, second use, expiry. Try it: https://jdbbs.exe.xyz/pinstitute/ with your own address. EMAIL_SYSTEM.md has it as pathway 8.

- [x] 5.8 Machine-callable factory (from P2 beats 8/9): POST transmittal JSON + DOCX, read Inspect JSON, build — the "your factory calls my factory" endpoint. Spec first, ½ page, to Jenna

  > **shelley** · 2026-09-18 13:14 UTC  
  > Spec written (docs/reviews/MACHINE-FACTORY-SPEC-2026-09-18.md, committed). Pasted:
  >
  > interfaces, three exist. This is what a caller can do today, what is missing,
  > and the smallest change that closes it.*
  >
  > ## What exists today (no code needed)
  >
  > Auth: a **project token** (`X-Auth-Token` header; admin mints it with
  > `POST /api/projects/{id}/auth`). Every step below already accepts it — the
  > API doc's "Admin" column for upload/convert/preflight is stale; the code
  > gates on `requirePassAccess` (token + live Factory Pass + credits).
  >
  > | Step | Call | Notes |
  > |---|---|---|
  > | 1 Transmittal | `PUT /api/projects/{id}/transmittal` JSON | same shape the form saves |
  > | 2 Template | `GET /api/projects/{id}/word-template` → .docx | **side effect:** syncs transmittal → book spec (only place a non-admin can) |
  > | 3 Manuscript | `POST /api/books/upload` multipart `file,title,author,project_id` → `{book_id}` | |
  > | 4 Inspect | `POST /api/projects/{id}/preflight` `{book_id}` then `GET …/preflight` | JSON incl. `book_map` finding; HTML at `…/preflight/report` |
  > | 5 Build | `POST /api/books/{id}/convert` `{"format":"both"}` → `{status:"converting"}`; poll `GET /api/books/{id}/outputs`; fetch `…/outputs/{oid}/download` | 1 credit per build |
  >
  > ## What is missing
  >
  > 1. **Spec sync is a side effect of downloading the template.** A caller that
  >    already has the template must still GET it to refresh the spec before a
  >    build, or builds against stale spec. → make `convert` and `preflight`
  >    pull transmittal → spec themselves when the transmittal is newer (same
  >    rule the template route uses).
  > 2. **No completion signal.** Callers poll `outputs`. → `GET /api/books/{id}`
  >    returning `{status: converting|done|failed, error, outputs:[…]}`; optional
  >    `callback_url` on convert (POST the same JSON when done). Polling stays.
  > 3. **Token issue is admin-only and manual.** Fine for two friends; before
  >    twenty, the client portal shows the token (or a magic-link-scoped one,
  >    5.7). Not needed for the talk.
  > 4. **Docs.** `docs/API.md` rows 30/31/43–45 say Admin; fix to Project, and
  >    add a "Factory in five calls" walkthrough with `curl`.
  >
  > ## Proposal
  >
  > Post-workshop, ~½ day: (1) + (2) + (4). Ship as one commit with an
  > end-to-end test that runs the five calls against a temp DB. Then the talk's
  > beat 8 is literally true: four interfaces, all four callable, and beat 9's
  > "shared folder and a person reading a report" becomes "a URL and a JSON".
  >
  > **Decision needed:** ok to say in the talk that this exists as an API today
  > (it does, with the caveats above), and land the polish after Wednesday?

  > **jenna** · 2026-09-18 15:04 UTC  
  > lgtm -- add an item to our ideal list to remind me to see if there are any session/workshop participants who might have enough production going on that they could be invited to join an API testing workshop sometime after the symposium

  > **shelley** · 2026-09-18 15:04 UTC  
  > Noted, thanks. Parked in docs/IDEAS.md (2026-09-18 row): “API testing workshop, post-symposium — scout Protocolize + SIGP4B participants with enough production going on (a series, a press, a recurring publication) to invite to a small ‘your factory calls my factory’ session against real work; sources: registrations / Admin → Cohorts, plus anyone who asks about the API in the talk Q&A.” Taking your lgtm as the yes for 5.8: I will finish the endpoint polish today and flip deck slide 9 from “what would be needed” to a commitment.

  > **shelley** · 2026-09-18 16:45 UTC  
  > Landed and live. What a machine can do now, with one project token and no browser: PUT the transmittal (final) → upload the .docx → POST build → GET /api/books/{id} until it says ready (or give a callback_url and be POSTed the same JSON when it lands) → download. The build reads the transmittal it was just sent — before today the spec only refreshed when someone downloaded the Word template, so a robot would have built against stale settings. Guardrails: callback URLs can’t point at localhost/private networks, redirects aren’t followed, one attempt. There is an end-to-end test that runs the five calls against an empty database, and a curl walkthrough at the top of docs/API.md (also fixed that doc’s stale ‘Admin’ column and its 2026-04 Known Issues table — all five long since fixed). Deck: slide 9 now says all four interfaces exist (the call: ‘exists — as of this week’), slide 10’s machine column names the actual routes, slide 11 adds one line inviting the first second factory to ask you for a token. Still true and worth saying out loud: nobody outside has run it yet.

- [x] 5.10 Punch-list archive: export = checklist + every note thread → `docs/runs/PUNCHLIST-YYYY-MM-DD.md`; index in `docs/runs/README.md`; admin dashboard card “Runs” at `/admin/runs/` rendering those files from disk (read-only). Saturday, ~1 h

  > **shelley** · 2026-09-18 13:57 UTC  
  > Done and deployed (7e4e717). New nav entry **Runs** on every admin page → https://jdbbs.exe.xyz/admin/runs/ — an index of docs/runs/*.md; click one to read it rendered (check-boxes, sections, and each note thread as a shaded block under its item, who · when). Read-only; the live list stays here on :8766. Export = `python3 scripts/punchlist-export.py` (I run it whenever I commit punch-list state; today’s file already has every note so far). Also lists the 17 Sep run log.

- [x] 5.11 Swim-lane "How it runs" (Author · Studio · Machine · Out × 5 stages) on the public `/factory#how-it-runs`; linked from the client factory step strip and the admin Floor hero (from 0.4)
- [ ] 5.12 Jev (typesafe.ai) pilot — heading classification behind a flag, per `docs/reviews/TYPESAFE-REVIEW-2026-09-18.md` (agreed 2026-09-18; post-workshop)
- [ ] 5.13 Index as a factory add-on (LLM-drafted conceptual index, Typst-resolved locators) — per `docs/reviews/INDEX-ADDON-FEASIBILITY-2026-09-18.md`; ~5 days; post-workshop
- [ ] 5.9 4 vCPU bump — only if the workshop shows build queueing (load test: CPU-bound)

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

- [x] 6.4 LibreOffice preview of the uploaded DOCX inside the factory (see the page Inspect saw)

  > **jenna** · 2026-09-18 00:21 UTC  
  > If you think that this might really work (that we could show a LibreOffice preview before they do a build and that it would have good fidelity with the typeset output), it could be a neat addition? If the fidelity would be suspect, it would be more trouble than it's worth?

  > **shelley** · 2026-09-18 00:33 UTC  
  > Fidelity would be to Word (what Inspect sees), not to the typeset PDF — it shows whether your styles landed, not what your book will look like. By your test that makes it more trouble than it is worth. My call: kill; the free Inspect + EPUB-first already covers “did the machine understand my file”.

  > **jenna** · 2026-09-18 01:04 UTC  
  > Okay, okay, good, we're agreed. I think we install LibreOffice so that you could use it when you need to work on the Word template itself, though, right? We'll keep it installed, I don't think it takes up a lot of room?

  > **jenna** · 2026-09-18 12:12 UTC  
  > Did you confirm whether you thought it useful to keep LibreOffice installed? vs delete to save space or minimize clutter?

  > **shelley** · 2026-09-18 12:12 UTC  
  > Keep it. It is the no-GUI build (231 MB on a disk that is 26% used), not on any request path, and it is how I eyeball the Word template and uploaded DOCX files as Word would show them (scripts/docx-preview.sh) — that check matters for the template-guide work in 5.3 step 5 and for the workshop week if a manuscript looks wrong. Zero clutter cost: nothing in the app depends on it, so removing it later is one apt command (sudo apt remove libreoffice-writer-nogui). Decision on the in-app preview (this item) stays parked.

  > **jenna** · 2026-09-18 16:54 UTC  
  > Yes let's keep

  > **shelley** · 2026-09-18 16:55 UTC  
  > Done. LibreOffice stays installed (studio tool for checking the Word template); the in-app preview stays killed as agreed — the fidelity would be to Word, not to the book. Recorded in docs/IDEAS.md and ticked.

- [ ] 6.5 Print cover / spine calc — separate product question (cover is EPUB-only today)

  > **jenna** · 2026-09-18 00:22 UTC  
  > Yeah, let's not lose track of this, but the whole "how to get it printed" question is not built into the factory at all yet. I think this is a good one to keep on the idea list for now. I just did some good research, and the two print companies we could point people towards are Bookmobile and Accutrack. Add those to our research if we add this to the factory after the workshop.

  > **shelley** · 2026-09-18 00:33 UTC  
  > Added to docs/IDEAS.md with Bookmobile + Accutrack named. Stays on the ideas list.

- [~] 6.6 Strike the done rows (Admin doc editor) and move the kept ones to §5 with a date

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


## 0 · Inbox — new items, untriaged (Shelley moves them into a section)
- [x] 0.1 Find where I asked that we look into Jev and how it might help make factory more efficient: https://docs.typesafe.ai/introduction  ·  _added Fri 11:55 UTC_

  > **shelley** · 2026-09-18 11:56 UTC  
  > Found it. First asked Wed Sep 17 during the Protocol Institute run — logged as P1 in docs/runs/RUN-2026-09-17-protocol-institute.md (line 110): “Review https://docs.typesafe.ai/introduction — anything useful for the factory?”, parked to after current work. It is already on this list as 5.6 (“P1 review docs.typesafe.ai/introduction — anything for the factory? Write a ½-page note, then decide”), queued after 5.3–5.5. Nothing else has been written about it yet — no note exists. Say “do 5.6 next” if you want it pulled ahead of the transmittal rewrite (5.4).

  > **jenna** · 2026-09-18 11:57 UTC  
  > Can you do this in parallel with what you're doing now?

  > **shelley** · 2026-09-18 11:57 UTC  
  > Yes — a research helper is reading the typesafe docs now and drafting the ½-page note (docs/reviews/TYPESAFE-REVIEW-2026-09-18.md) while I carry on with P4. I will paste the note here when it lands and tick 5.6 with it.

  > **jenna** · 2026-09-18 12:08 UTC  
  > "I will paste the note here when it lands and tick 5.6 with it." I see the reply in the main session, but not here? Agree on the rec you posted: "Recommendation: pilot on heading classification only, behind a flag with the vocab fallback and cached answers — after the workshop. Park the rest."

  > **shelley** · 2026-09-18 12:09 UTC  
  > The note is on 5.6 (the row this item pointed to) — scroll to §5. Decision recorded: pilot heading classification behind a flag after the workshop, park the rest — logged in docs/IDEAS.md with the review link.

- [x] 0.2 Good copy  ·  _added Fri 11:58 UTC_
- [x] 0.3 Fix: looks like our global theme picker lost the system choice. Now has only light and dark. Let's put the system back. Do this in parallel if you can.  ·  _added Fri 11:59 UTC_

  > **shelley** · 2026-09-18 12:02 UTC  
  > Fixed and deployed. The moon/sun button is now a three-way: ◐ system (follows your OS, default for new visitors, switches live when the OS does) → ☀ light → ☾ dark → back to system. Hover shows the current mode. Existing saved choices are kept as explicit light/dark.

  > **jenna** · 2026-09-18 12:05 UTC  
  > oops meant to enter this as a separate quick to-do

  > **shelley** · 2026-09-18 12:05 UTC  
  > No harm — 0.3 landed as its own item and is done. If there was another quick to-do behind that “oops”, it did not come through: drop it in the inbox box and I will pick it up.

  > **jenna** · 2026-09-18 12:06 UTC  
  > ah got it, all good on this one, confirmed I see it now too, and just getting used to this punchlist pattern: I'm liking it a lot, do add it to our preferred pattern when doing app review work!

  > **shelley** · 2026-09-18 12:07 UTC  
  > Added to AGENTS.md as the preferred pattern for review and run work (inbox → triage, notes both ways, tick + copy to docs/runs, parallel quick fixes). Future sessions will start from it.

- [x] 0.4 Do we already, on both the client side and the admin side, have somewhere that I can go to show a pretty simple view of the tools involved in the factory? I'm thinking primarily of how Pandoc, Lua, and Typeset fit into the flow, but I might be missing something.  Also, I showed the map that we made to some work colleagues, and one of them specifically asked for it to be turned into the variant of that kind of a chart that he called "swim lanes." He suggested that it was pretty common, so maybe you could find that and see if it was at all suitable for us.  ·  _added Fri 13:08 UTC_

  > **shelley** · 2026-09-18 13:11 UTC  
  > Short answer: no. Today the tool flow exists only as (a) one line on the landing page — “docx → pandoc·Lua → Typst → PDF/X + EPUB” — and (b) the exedeck’s stage list (01 handshake … 04 transform) plus its small tools table. Nothing on the admin side (Floor, dashboard) and nothing on the client factory page explains what Pandoc, Lua and Typst each do. The public /factory doc only says “you get outputs, not the Typst source”.
  >
  > What the pieces do, in one breath: Word → **pandoc** reads the .docx into a neutral document tree → **Lua filters** (ours) rewrite that tree using the transmittal — book map, styles, front matter, parts → pandoc writes **Typst** source (print) or **EPUB** directly → **Typst** typesets against our series template and the spec-generated config → PDF. The pieces you might be missing: the *book spec* (transmittal → JSON) feeds every stage, the Word *template generator* (python-docx) is the handshake at the front, and *epubcheck* / *preflight* sit at the end.
  >
  > Swim lanes: yes, common (they are flowcharts with one horizontal band per actor, so hand-offs show as arrows crossing lanes). Suitable for us — the factory has natural lanes: **Author** (Word, template, transmittal) · **Studio** (Jenna: spec, review, corrections) · **Machine** (pandoc → Lua → Typst / EPUB) · **Out** (PDF/X, EPUB, printer). The current map is stage-ordered; a swim-lane version would show *who* does each step, which is exactly what colleagues outside publishing want to see.
  >
  > Proposal, if you want it before Wednesday: **5.11** (~1 h, Saturday): one static swim-lane diagram (HTML/CSS, prints clean, theme-aware) added to the public /factory doc as a “How it runs” section, linked from the client factory page footer and the admin Floor. Also usable as a slide in the talk. Say yes and I will slot it after 5.4/5.5.

  > **shelley** · 2026-09-18 14:03 UTC  
  > Built it (5.11): https://jdbbs.exe.xyz/factory#how-it-runs — four lanes (Author · Studio · Machine · Out) across the five steps; the Machine lane shows python-docx / pandoc + Lua → Typst → PDF / pandoc → EPUB3, with a paragraph above naming each tool’s job. Reachable as “How it runs →” at the end of the step strip on every client factory page, and from the Floor hero on the admin side. Inline SVG in theme tokens, so it follows the font selector. Tell me if a lane or a box is wrong and I’ll regenerate it (scratch/swim/gen.py).

  > **jenna** · 2026-09-18 14:15 UTC  
  > Looks good at first glance, but I note it's missing the full client-side top nav and the theme switcher, which we said were going to be default on every page. Let's clean that up here and in our work notes?

  > **shelley** · 2026-09-18 14:19 UTC  
  > Good catch — and it was bigger than that page. Two things:
  >
  > 1. **The theme switcher had been silently missing on every page** that relies on the auto-mount (public docs, cohort, store thanks — most of the site) since the admin-nav commit on the 15th: the one line that mounted it got dropped in that edit. Restored; the convergence test now fails if it ever goes again.
  >
  > 2. **Public pages each hand-wrote their own top links** (Factory had Workshop + Field notes; Litmags had a different pair; the deck companions had a third). Now there is one shared public strip in theme.js — Workshop · Field notes · Factory · Client portal — and every public doc opts in with `<nav data-public-nav>`, same as admin and client pages do. Applied to all 13 public pages (talk.html left alone as agreed; decks have their own controls). Test covers jdbbs-public too.
  >
  > Work notes: PAGE-DESIGN §4 masthead row + recipe updated (“never hand-write links”). Deployed — hard-refresh /factory to see it.

- [x] 0.5 new: https://jdbbs.exe.xyz/pinstitute/perception/transmittal/ says "Trim Size 5.5 X 8.5 6 X 9 8.5 X 11 DON'T CARE OTHER:" And the "don't care" is, of course, now out of date. This is another section that still has some holdover from when I was the production editor in a traditional production line.  Rethink what we should say about the trim size choice from the perspective of a civilian who doesn't maybe yet know anything about the finer points of small, medium, and large size books. Also consider whether we need to keep the spine allowance calculations in there, since we won't know very much about the paper stock. I suspect we should just eliminate it, but I suppose eventually we could offer to create a template for their cover production if they came back to us with their PPI.  No, I think I'm rambling now, but I think that information, if they're going to go to a printer, they're better off getting that information directly from the printer. We should say something to that effect for this section.  ·  _added Fri 14:50 UTC_

  > **shelley** · 2026-09-18 14:59 UTC  
  > Done and live — reload https://jdbbs.exe.xyz/pinstitute/perception/transmittal/ and scroll to FORMAT (was Book Design). What it says now: one line up top that the page size (the trim) is the one physical decision that is yours; margins/type/heads follow from the series. Then "What kind of book is it, as an object?" in words. Then a stacked choice, one plain sentence each: Small 5½×8½ (hand/coat pocket; fiction, essays, poetry; more pages) · Medium 6×9 (standard trade paperback; nonfiction, memoir, notes/figures; "if you are unsure, choose this") · Large 8½×11 (workbooks, manuals, wide tables, many images; heavy, not for bed) · Let the studio choose (replaces Don’t care — we pick from the manuscript and your note, and tell you) · Exact size W×H (when a printer/series/distributor already fixes it; now actually sets the page — "7 x 10" works). PPI + Spine width + Est. pp + Text Complexity + Outside Designer + Reuse Previous are gone, replaced by one paragraph saying exactly what you said: spine width depends on the printer’s paper (PPI), only the printer knows; when the PDF is final give them trim + page count, they send back a cover template; bring it back to us if you want help with the cover. Old saved values still load. Wording is yours to tune — tell me and I’ll change it, or say if I cut too much (Reuse previous is the one I’d bring back if you miss it).

- [x] 0.6 On Transmittal, we're still showing typed source file as a deliverable, but I'm pretty sure we decided not to do that. Can you find anything about that in our notes?  ·  _added Fri 15:02 UTC_

  > **shelley** · 2026-09-18 15:03 UTC  
  > Yes — it is in the Sep 17 run notes (docs/runs/RUN-2026-09-17-protocol-institute.md, item C13): “Deliverables → replace checkboxes with a fixed statement. Every pass yields: print-interior PDF + EPUB, the Word template, and the Inspect report. Typst source stays with the factory (house template + our filters; useless without the toolchain and giving it away invites ‘why doesn’t it compile’). Fonts never (licensed). Cover files never (theirs). Printer delivery radio → delete. Page Proofs → delete. Subrights → delete.” The Sep 3 storefront design doc had it as [DECIDE] not-by-default too. That was the unfinished tail of C13 — done now and live: the checkboxes are gone, the section is “What you get” with the four fixed things, one line on what is not included and why, and no proof-routing step. Page Proofs and Subrights sections are gone as well. Nothing previously saved is lost.

  > **jenna** · 2026-09-18 15:08 UTC  
  > Update to your new paragraph here:
  > > We no longer estimate the spine here. Its width depends on the paper your printer uses — the PPI, pages per inch — and only the printer knows that. When your interior PDF is final, give the printer the trim and the page count; they send back a cover template with the exact spine. If you would like help laying out the cover, bring that template (or the PPI) back to us.
  >
  > Mine:
  > A book's spine width depends on the PPI, pages per inch, of the paper from your printer that you choose. When your interior PDF is final, you give the printer the trim and the page count; they send back a cover template your cover designer will use with the front-spine-back set up correctly.

  > **shelley** · 2026-09-18 15:09 UTC  
  > Swapped in verbatim, live now (reload the transmittal, Format section).

- [x] 0.7 new: we'll need to catch forward authorship as its own style. Occasionally, I suppose someone could have an afterword. In both cases, it's sometimes a document that's added to the book that's been authored by someone other than the book's author. I'll drop a screenshot into our session thread since this tool doesn't yet let me add screenshots. If that's a quick fix, go ahead and do it?  ·  _added Fri 16:59 UTC_

  > **shelley** · 2026-09-18 17:04 UTC  
  > Done: a new house style, Signature. Each line of the sign-off (name / title / place) goes in Signature style in Word; the factory sets the block flush left, tight, with a line of air above, kept on one page, in both the PDF and the EPUB. Inspect knows it (won’t flag it as an undeclared style). It appears in the Word template and its style guide from the next template download — the template count is now twelve styles, not eleven; the deck says “eleven” in a few places, your call whether to change it or leave it. No restart needed (scripts are read at run time). Not a separate “Foreword author” style: the sign-off block is the thing that distinguishes another hand, and it works for afterwords and introductions too.

  > **jenna** · 2026-09-18 17:11 UTC  
  > > the template count is now twelve styles, not eleven; the deck says “eleven” in a few places, your call whether to change it or leave it
  > Guess we might as well fix it wherever it's referenced. Twelve now 👍

- [ ] 0.8 new: let's draft a short paragraph that we can offer to the books that run through the factory. It could be either on the copyright page in small type or as an item element of back matter that would be called "About the Book Factory." Pop this to the top of your list now if you can. Maybe 500 words.  ·  _added Fri 17:04 UTC_
- [x] 0.9 New: I'm considering that we could add index as an add-on in the factory, and maybe that is a good use of Jev. The main problem is that a word list is not actual indexing. An indexer does a lot of thinking about what to include, but maybe this is the time, and an LLM can actually produce a pretty good index that is meaningful, not just a sorted word list.  We do a first pass on whether it looks feasible or not today. Add that to your list, and then let's hold on. If it's at all complicated, let's hold on to that for after the workshop. Also, maybe that tool we looked at earlier today, Jev, could be ideal for and we can do it now.  ·  _added Fri 17:07 UTC_

  > **shelley** · 2026-09-18 17:09 UTC  
  > First pass done: docs/reviews/INDEX-ADDON-FEASIBILITY-2026-09-18.md. Verdict: feasible and worth doing, but ~5 days, so holding it — now 5.13, post-workshop. Three things worth knowing now. (1) Yes, an LLM can write a real conceptual index (topics, merged synonyms, subentries, see/see-also, skipping passing mentions) — indexers’ own societies have been testing this since 2024 and call it a strong first draft for a human to edit, which is exactly the tier we’d sell. (2) Jev is the wrong tool: it classifies (pick one / score / yes-no), it doesn’t generate; indexing is generation. Jev could at most be a cheap yes/no gate on candidate terms. Jev stays on 5.12. (3) The hard part isn’t the entries, it’s the page numbers: they only exist after Typst sets the book, and any correction reflows them. The right design is the LaTeX one — the model proposes terms and the sentences that mention them, we plant #index[] markers in the Typst, Typst resolves the locators at compile time, so every rebuild keeps the index true. That’s the new Typst subsystem that makes it a five-day job. Cost per book is well under a dollar; a human indexer is $3–5 a page, so the add-on prices itself.

- [x] 0.10 new: I don't think we have yet created a default style for glossary entries. Let's check.  ·  _added Fri 17:10 UTC_
