# Pre-workshop punch list + parked work · Thu 17 → Sat 20 Sep

Legend: **YOU** = Jenna's turn · **ME** = Shelley's turn · **BOTH** = look together.
Ticked = done · ◐ = in progress. Click a box to cycle ☐ → ☑ → ◐. **note** opens a reply box under any item (⌘↵ saves) — I read those back. Bottom bar adds a new item to an Inbox I triage. Refreshes every 15 s (pauses while you type). The previous run checklist (Protocol Institute Test 2) is committed at `docs/runs/RUN-2026-09-17-protocol-institute.md`.

**Freeze:** Fri 19 – Sat 20 fix / test / clean allowed (build → `go test ./srv/` → restart → one factory smoke as `pinstitute`). Sun 21 – Tue 23 hotfix-only. Store flips to live Stripe by timer **Wed 23 Sep 00:00 HKT**.

## 1 · Fix window — ME

- [x] 1.1 **Word template**: 11 factory styles get `<w:qFormat/>` + `<w:uiPriority>` so Word's Styles pane (Recommended view) isn't empty. Verify with `scripts/docx-preview.sh` + unzip `styles.xml`
- [x] 1.2 Sender address `studio@` → `factory@mail.jdbb.studio` (display name stays "jdbb studio"); one test send — approved
- [x] 1.3 Confirm nothing else hardcodes AgentMail as *sender* (archive CC rows `jdbb@agentmail.to` are fine); fix `snapshot_email.go` error text
- [x] 1.4 Strike the done "Admin doc editor" row in `docs/IDEAS.md`
- [x] 1.5 Checkpoint tag per `CHECKPOINTS.md`; full `go test ./srv/`; smoke `/factory`, `/admin/factory/`, `/admin/store/`, `/pinstitute/book-001/factory/` — tonight after 1.1–1.4 (re-tag after any later fix) — tag `checkpoint-2026-09-18-pre-workshop` at `0cd67fd`
- [x] 1.6 Dependabot: 5 vulns on djinna/jdbbs (3 high, 2 moderate) — check whether these were already fixed a day or two ago and just not marked; if new, fix now (rule: always address vulns)

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
- [ ] 3.2 YOU — cut / reorder / add beats
- [ ] 3.3 ME — build the deck as a new page in `~/jdbbs-public` on the talk template
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
- [x] 5.2 C27 CJK embed rule: fire only at ≥ 20 ideographs, and/or subset with `pyftsubset` (the 20 MB EPUB from one ASCII-art tweet)
- [~] 5.3 P4 **front matter ingestion** — decisions already taken 2026-09-17 (every section head = Heading 1; classify by heading text against a closed vocabulary + position; untitled pre-H1 blocks = dedication/epigraph in transmittal order; title + © pages generated from the transmittal; Parts opt-in shifts chapters to H2; arabic 1 on a recto). Inspect prints the resulting book map
- [ ] 5.4 **C13 transmittal rewrite for the factory** (post-workshop, the big one):
    - [ ] 5.4a C6 drop press-era Production section (Mechs Delivery, Weeks in Prod., Bound Book Date, dup Transmittal Date); Print Run → Book; one optional "Target date"; keep old JSON keys readable
    - [ ] 5.4b C7 stop asking chapters / words / MS pp / est. book pp — Inspect counts them; fix missing input underline meanwhile
    - [ ] 5.4c C10 drop Developmental Edit + Level of Copyediting; keep Special Characters etc.
    - [ ] 5.4d C11 Permissions → one courtesy line + one attestation checkbox; terms text before Wed go-live if possible
    - [ ] 5.4e C12 Pub Info & © → the **copyright-page builder** (credit fields actually used)
- [ ] 5.5 P3 **H&J / composition quality** in the print PDF: typst `par(costs:)`, optimized linebreaks, a loose-lines / rivers / runts / widows scorer on the built PDF; compare against an InDesign-set page
- [x] 5.6 P1 review docs.typesafe.ai/introduction — anything for the factory? Write a ½-page note, then decide
- [ ] 5.7 **Magic-link client login** (email → one-time link → cookie) replacing emailed passwords
- [ ] 5.8 Machine-callable factory (from P2 beats 8/9): POST transmittal JSON + DOCX, read Inspect JSON, build — the "your factory calls my factory" endpoint. Spec first, ½ page, to Jenna
- [ ] 5.9 4 vCPU bump — only if the workshop shows build queueing (load test: CPU-bound)

## 6 · Ideas log (`docs/IDEAS.md`) — BOTH decide, then queue or kill

- [ ] 6.1 Per-client interactive stylesheet instances (accept/reject seeded from the universal fiction/nonfiction stylesheet)
- [ ] 6.2 Reusable "review table" pattern from the sitemap-review tool (JSON table + per-row reply box)
- [ ] 6.3 Client-visible on-disk documents (`pi-client/{slug}/` + `serveClientDoc`) — trigger: first real client doc
- [ ] 6.4 LibreOffice preview of the uploaded DOCX inside the factory (see the page Inspect saw)
- [ ] 6.5 Print cover / spine calc — separate product question (cover is EPUB-only today)
- [ ] 6.6 Strike the done rows (Admin doc editor) and move the kept ones to §5 with a date

## 0 · Inbox — new items, untriaged (Shelley moves them into a section)
- [x] 0.1 Find where I asked that we look into Jev and how it might help make factory more efficient: https://docs.typesafe.ai/introduction  ·  _added Fri 11:55 UTC_
- [x] 0.2 Good copy  ·  _added Fri 11:58 UTC_
- [x] 0.3 Fix: looks like our global theme picker lost the system choice. Now has only light and dark. Let's put the system back. Do this in parallel if you can.  ·  _added Fri 11:59 UTC_
