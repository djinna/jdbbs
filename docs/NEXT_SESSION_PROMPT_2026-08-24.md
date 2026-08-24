# Next session: add SFF prestige-ranked magazines to /litmags

## Goal
Extend the existing lit-mag tool-stack reference at **https://jdbbs.exe.xyz/litmags**
with the magazines from Eric Schwitzgebel's 2026 SFF ranking, researching each
one's **publishing tool stack** the same way we did the first 74.

Source list:
https://schwitzsplinters.blogspot.com/2026/08/top-science-fiction-and-fantasy.html
("Top Science Fiction and Fantasy Magazines 2026" — a prestige ranking by award
nominations + "best of" anthology placements over 10 years. Prose only, not
poetry; horror/dark fantasy only incidental.)

## What already exists (read first)
- Page source: `srv/static/litmags.html` (self-contained; ~440 lines).
  - Working copy mirror at repo root `/home/exedev/prodcal/litmags.html`
    (gitignored — keep it in sync for local file:// preview but the served
    file is the one under `srv/static/`).
- Route: `GET /litmags` in `srv/server.go`, served via `s.serveStaticHTML`
  (public, no login — same pattern as `/workshop` and `/field-notes`).
- **Design language: "Terminal Folio"** — monospace masthead/wordmark, `//` kicker,
  statusline, hairline stat ledger, accent-topped chart cards, zero-radius ledger
  table with mono chips, dark-mode via theme tokens + localStorage bootstrap.
  Matches `srv/static/workshop.html`. DO NOT redesign; just add data rows.
- The data model is a single JS array `const mags = [...]` inside the page. Each
  entry has this shape (keep it EXACTLY the same so render/filter/search keep working):

  ```js
  {name:'', url:'', desc:'', platform:'', platformDetail:'',
   formats:['web','print','audio','epub'], pricing:'', cadence:'',
   editorial:'', outlier:false, note:'', tags:['sf','fantasy','horror','fiction','poetry',...]}
  ```

  Field meanings:
  - `platform`: coarse bucket the charts group on — one of `WordPress`, `Substack`,
    `Squarespace`, `Shopify`, `Wix`, `Ghost`, `Webflow`, `Custom`, `Unknown`.
    (The chart normalizes anything not matching those to "Custom / other".)
  - `platformDetail`: the human-readable chip shown in the table (e.g.
    "WordPress + Divi", "Custom (Drupal)", "Squarespace", "Custom (Movable Type)").
    If unknown, reuse the coarse `platform` value.
  - `formats`: subset of exactly `web`,`print`,`audio`,`epub`.
  - `pricing`: free text but MUST contain the substrings the pricing chart keys on
    where applicable: "free", "sub"(scription), "pay"/"per" (pay-per-issue).
    Examples: `Free`, `Free / Subscription`, `Subscription / Pay-per-issue`.
  - `cadence`: `Weekly`/`Monthly`/`Bimonthly`/`Quarterly`/`Biannual`/`Annual`/`Varies`.
  - `editorial`: 1–2 sentences on editorial voice / house style / notable stack facts.
  - `outlier:true` + `note`: use for entries that are NOT small literary magazines
    (big-3 corporate SFF like Reactor/Tor, general-interest mags that publish SFF
    on the side — New Yorker, Wired, Slate, Boston Review, Paris Review, MIT Tech
    Review, NYT, Buzzfeed, McSweeney's, Tin House, Conjunctions — and defunct ones).
    Keep them (source list value) but flag with `note` explaining why.
  - `tags`: for filtering. SFF entries should include `sf` and/or `fantasy`; add
    `horror` where apt, `fiction` generally, `audio` for podcast-format markets.

## De-dupe against the existing 74
These are ALREADY in the page — do NOT add duplicates; instead, if the new source
has better SFF-specific info, you may lightly enrich the existing entry's
`editorial`/`tags`:
- **Pseudopod** (source lists as "PseudoPod")
- **The Dark**
- **Reckoning** (source lists as "Reckoning")
Also note "Uncharted" is in our set but is a *different* magazine from anything on
this ranking — leave it.

## Magazines to research and add (from the ranking; skip pure dupes above)
Tier by prestige is irrelevant to us — just capture the stack for each. Group them
and fan out with parallel subagents (see method below):

Reactor/Tor.com, Clarkesworld, Uncanny, Lightspeed, Asimov's, Fantasy & Science
Fiction (F&SF), Beneath Ceaseless Skies, Strange Horizons (incl. Samovar), Analog,
Apex, Nightmare, FIYAH, Slate/Future Tense (ceased 2024), Fireside (ceased 2022),
Fantasy Magazine, The Sunday Morning Transport, Interzone, Diabolical Plots,
The Deadlands, khōréō, Future Science Fiction Digest (ran 2018–2023), PodCastle,
Conjunctions, GigaNotoSaurus, Lady Churchill's Rosebud Wristlet, The New Yorker,
Omni, Psychopomp (started 2023), Kaleidotrope, Anathema (ran 2017–2022), Baffling,
Boston Review, Omenana, Terraform/Vice (ceased 2023), Wired, B&N Sci-Fi & Fantasy
Blog (ceased 2019), Paris Review, Shimmer (ceased 2018), Sirenia Digest, Science
Fiction World (China), Galaxy's Edge (ceased 2023), Augur, Beloit Fiction Journal,
Bourbon Penn, Buzzfeed, Escape Pod, Flash Fiction Online, McSweeney's, Tin House,
Abyss & Apex, Black Static (ceased 2023), Fusion Fragment, Mothership Zeta (ceased
2017), Shortwave, Weird Horror, Constelación (ran 2021–2022), MIT Technology Review,
New York Times, Translunar Travelers Lounge.

(~57 net-new after de-dupe. Many are defunct — still add with a `note`; a dead
site's Wayback capture or masthead is fine for the stack.)

## Research method (how we did the first batch — replicate it)
1. **Fan out with `subagent` tool**, ~10–12 magazines per subagent, model
   `gpt-5.6-sol`, `wait:false`, then collect. Give each subagent the exact JSON
   shape above and tell it to visit each site with the `browser` tool and return
   one JSON object per line. Have subagents WRITE results to a file
   (e.g. `/tmp/sff_group_X.jsonl`) AND print them — retrieving via file is more
   reliable than scraping the subagent transcript (that bit us last time).
2. For the stack, look for tells: `<meta name="generator">`, `/wp-content/` or
   `/wp-json/` (WordPress), `*.substack.com` or Substack CDN, Squarespace/Wix/Ghost/
   Shopify/Webflow footprints, or hand-rolled/custom. Submissions/about/masthead
   pages usually reveal cadence, formats (print/ebook/audio), and paywall model.
   For defunct sites use web.archive.org.
3. Corporate/general-interest outliers (Reactor, New Yorker, Wired, etc.): a quick
   platform sniff is enough; mark `outlier:true` with a `note`.
4. Cross-check names against the existing `mags` array before inserting.

## How to add them to the page
- Insert the new objects into the `const mags = [...]` array in
  `srv/static/litmags.html` (append a new `// ===== SFF ranking (Schwitzgebel 2026) =====`
  section for provenance). Keep the object shape identical.
- The header currently says "74 Small Literary Magazines" and the statusline/lede
  reference 74. UPDATE those counts/copy to the new total, and consider tweaking
  the lede to note the page now merges two source lists (the @subclub reprint list
  + the Schwitzgebel SFF prestige ranking). The summary stats and charts are all
  computed from the array, so they auto-update.
- Optionally add an `sf`/`fantasy` genre filter is ALREADY present ("SF / fantasy"
  button) — verify it still behaves with the larger set.

## Verify + ship (same as last time)
1. Preview: open `file:///home/exedev/prodcal/srv/static/litmags.html` in the
   `browser` tool; confirm row count, charts, filters, and dark mode
   (`document.documentElement.classList.add('dark')`).
2. `cp srv/static/litmags.html litmags.html` (keep root working copy in sync).
3. `make build` (from repo root) — the route is unchanged so this is just a sanity build.
4. `sudo systemctl restart prodcal` then
   `curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8000/litmags` (expect 200).
5. Commit with a descriptive message, co-authored by Shelley.
6. **Push (hub-and-spoke):** `origin` is HTTPS and has NO stored creds on the VM,
   so `git push origin main` FAILS with "could not read Username". Push over SSH
   instead: `git push git@github.com:djinna/jdbbs.git main`. (This is the
   documented fallback in AGENTS.md.)

## Gotchas learned last session
- Don't chain `cd` in bash; use the `change_dir` tool (working dir is
  `/home/exedev/prodcal`).
- systemd unit is `prodcal.service` (older docs say `srv` — wrong).
- Port 8000 is the app behind the PUBLIC exe.dev proxy → that's why /litmags is
  gate-free. The busybox on :8001 is the private/gated copy; ignore it.
- Retrieve subagent output via files, not by re-reading the transcript.
- There are 2 open Dependabot alerts on the repo (1 high, 1 moderate) — unrelated;
  a separate handoff already tracks them. Leave alone unless asked.
