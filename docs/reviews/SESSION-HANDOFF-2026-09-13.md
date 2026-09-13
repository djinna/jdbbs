# Session handoff — 2026-09-13 (workshop materials block)

Workshop: **Protocolize Your Book**, Protocol Symposium 2026, Sept 21–22
(four sessions; talk slot 25+5 somewhere Sept 21–25 — confirm day/time).
Deploy freeze Sep 19–22. Today is Sep 13.

## State at handoff

Shipped this session (all on GitHub + deployed): outbound-email log + Mail
tab; email restyle + `/admin/email-preview/`; BCC audit copy + permanent
smoke persona; stable slugs (`mcasey` / `book-001`) + admin Rename URL +
old-URL redirects; `pi-public` → **`jdbbs-public`** with its own GitHub repo
(`git@github.com:djinna/jdbbs-public.git`), planning material folded in at
`jdbbs-public/2026-pi-symposium/source/` (old `~/pi-book-workshop` renamed
`…MERGED-into-jdbbs-public`, safe to delete).

**Push both repos at the end of every work block:** prodcal →
`git push git@github.com:djinna/jdbbs.git main`; jdbbs-public →
`git -C ~/jdbbs-public push`.

## Standing reminders
- Mike Check (`bookiq@gmail.com`, reg #10, `/mcheck/book-001/`) is a
  permanent fixture: consent ON, auto-added to every announcement, never
  delete. j@djinna.com is BCC'd on every send.
- Tell Mike Casey his sign-in is now **jdbbs.exe.xyz/mcasey/** (same
  password); old links redirect.
- Toby Shorin (#9) has no Factory code. Andrea Leiter (#2) has a code but
  consent=false, so never emailed.

## URL plan for event pages (decided in principle, not built)

Namespace = the cohort slug that already exists: **`/2026-pi-symposium/`**
(roster lives at the root of it today, `handleCohortPage`). Add a directory
mount `GET /2026-pi-symposium/{page}` → `jdbbs-public/2026-pi-symposium/{page}.html`
so handouts publish by saving a file (roster route must keep precedence for
the bare path). Evergreen pages stay flat (`/field-notes`, `/factory`,
`/stylesheet/`). `/workshop` stays where it is (in nine emails); optionally
301 it later. Reserve `2026-pi-symposium` as a client slug.

Register each new page in `site_pages` (admin Pages registry) and follow
`docs/PAGE-DESIGN-HOSTING-VISIBILITY-2026-09-11.md` (1240 shell, theme.css,
canonical masthead/footer). `exedeck.html` is the documented deck exception
(own 960px stage) — reuse its stage for the talk deck.

## Work items — first-order instructions

### 1. Talk deck + handout → `/2026-pi-symposium/talk`
Source: `2026-pi-symposium/source/talk-proposal-2026-07-27.md` (locked title,
25-min outline in six beats, artifact list, "New Nature" hook) and
`/field-notes` for the examples. Build as ONE HTML file with two renderings:
slide stage (clone `exedeck.html` mechanics: keyboard nav, 960 stage) and a
print/handout stylesheet that lays the same sections out as prose. ~12
slides: problem → transmittal handshake → styles as controlled vocabulary +
preflight (the ABS "hard mode" analog from field-notes) → deterministic stack
(Ghosts as the worked example) → calendar + LLM agents as participants →
paid off / leaked → New Nature close. Open decision from the proposal still
open: live demo vs. 60–90 s pre-recorded clips — recommend clips (Factory
build of the smoke manuscript, preflight report). Numbers on slides come from
`jdbbs-public/notes/research-prodcal.md`; re-check before presenting.

### 2. Participant handout: session structure → `/2026-pi-symposium/workshop`
Sources: `workshop-proposal-2026-07-27.md` §"Session plan" (4 sessions:
handshake / conformance & negotiation / deterministic stack + LLM / payoff &
leaks) and `RUN-OF-SHOW-2026-07-27.md` (beat sheets, homework between
sessions, Discord coworking model). **Conflict to resolve first:** proposal
says 4 × 75–90 min, run-of-show plans ±3 h — get the confirmed schedule and
Zoom/Discord links from PI, then write. One screen / one printed page: dates
+ times per session, what each session does, homework after S1–S3, what to
bring, what you leave with, Factory link (`/factory` + their code), where to
get help. Same tone as `/workshop` registration page. Link it from the
cohort roster page.

### 3. `EXAMPLES-IN-THE-WILD-2026-07-27.html` — already live
It is the July 27 version of **`/field-notes`** (same four starred examples +
two shelves; field-notes.html is the maintained copy, restyled to the shared
theme on 09-11). Nothing to remount. If a side-by-side is still wanted, the
old file is at `jdbbs-public/2026-pi-symposium/source/EXAMPLES-IN-THE-WILD-2026-07-27.html`
and can be served temporarily by dropping a one-line `servePublicDoc` route.
The `.md` twin has the NDA-safe notes and verification statuses (✅/🟡/💡)
that the HTML dropped — useful when scripting the talk.

### 4. Other work products worth revisiting (quick sweep)
- `/exedeck` — Aug 24 SIGPfB deck about exe.dev; the *mechanics* and the
  studio voice are the template for item 1. `jdbbs-public/notes/research-*.md`
  = sourced facts behind it.
- `RUN-OF-SHOW-2026-07-27.md` — beat sheets with a tech-readiness checklist
  (hero manuscript timed run, messy sample manuscript, Zen Garden flip
  214↔215). The "messy sample" manuscript was never made; Session 2 needs it.
- `SUBMIT-paste-ready-2026-07-27.md` — final abstract text as accepted; the
  handout should not contradict it.
- `/stylesheet/` — house editorial stylesheet (62 items); cite it in the prep
  guidance (item 5) as "what the copyedit checks", distinct from Word styles.
- `docs/reviews/FACTORY-PASS-SESSION-A-DRY-RUN-2026-09-04.md` — the timed
  dry run of the participant surface; source for realistic timings in item 2.
- `typesetting/scripts/generate-word-template.py` and
  `detect-edge-cases.py` — the authoritative list of style names the Factory
  understands (Normal, Heading 1–3, Block Quote, Section Break, custom
  `tweet-p`/`metadata-p`, …). Item 5 must quote these, not invent them.
- `/2026-pi-symposium` roster page — the natural home for links to items 1, 2
  and the prep page.
- `docs/book-production-deepdive.html` — older architecture explainer;
  probably superseded by field-notes, skim once and retire or link.

### 5. Announcement email TODAY: Word-doc prep guidance
Send from the tracker outbox (`/admin/registrations`) to all consenting
registrants (Mike Check auto-added; you're BCC'd). Content, in this order:
dates/times reminder (once confirmed); "have your Factory Pass redeemed
before Session 1" with `/factory` + note that codes are on the card; the
prep rules — use **paragraph styles, not manual formatting** (Normal /
Heading 1–2 / Block Quote…), no faked section breaks, no pasted colors;
**Google Docs is fine** if they File → Download → .docx and the Word styles
pane shows only the styles they meant to use (no `Normal (Web)`, `Heading 1
Char`, `Style1` outliers — that is the tell); rough content is welcome,
rough *formatting* is what Session 2 fixes; upload to their Factory and run
Inspect before the session if they want a head start. Keep to ~250 words;
plain text is primary. Ideally the same text becomes
`/2026-pi-symposium/prep` so the email can link it instead of carrying it.
Before sending: issue Toby Shorin a code; decide about Andrea Leiter.

## Suggested order
5 (today, time-sensitive) → 2 (needs PI schedule; ask now) → 1 (largest) →
the directory mount + `site_pages` rows when the first page exists → 4 as
you go. All before Sep 19.
