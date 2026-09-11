# Page design, hosting & visibility standard (2026-09-11)

Source of truth for every jdbb web surface: how a page is shaped, where it
lives, who can see it, and how it is found. Supersedes the width/chrome notes
in `DESIGN-SYSTEM.md` (which keeps the visual canon) and `pi-public/README.md`
(which keeps the per-file table). Decisions were taken 2026-08-31 → 09-11 via
the sitemap-review tool (`~/sitemap-review`, port 8102).

---

## 1. Page anatomy

Every studio page — marketing, prose, ledger, admin tool, client tool — is
the same skeleton. Only the deck (`/exedeck`) is exempt (§5).

```html
<!doctype html><html lang="en"><head>
  <meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Page title · jdbb studio</title>
  <link rel="icon" href="/static/favicon.svg">
  <link rel="stylesheet" href="/static/theme.css">      <!-- tokens + shared chrome. ALWAYS first. -->
  <style>/* page-local rules only: no tokens, no chrome */</style>
</head><body>
<div class="jdbb-shell">

  <header class="jdbb-masthead">
    <a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span><span class="studio">studio</span></a>
    <nav>
      <a href="/workshop">Workshop</a>            <!-- 0–4 links, Sentence case -->
      <div id="theme-bar"></div>                  <!-- auto-mounted by theme.js -->
    </nav>
  </header>

  <div class="jdbb-statusline">                    <!-- optional context strip -->
    <span><span class="ok">●</span> Field notes · Language as protocol</span>
    <span>rev 5 · working</span>
  </div>

  <main>
    <section class="jdbb-prose"> … running copy at reading measure … </section>
    <table class="ledger"> … full shell width … </table>
  </main>

  <footer class="jdbb-footer">
    <a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span></a>
    <nav aria-label="Footer"><a href="/">Home</a><a href="/field-notes">Field notes</a><a href="/workshop">Workshop</a></nav>
    <span class="copy">&copy; 2026 Jenna Dixon</span>
  </footer>

</div>
<script src="/static/theme.js"></script>          <!-- last; applies stored theme, mounts #theme-bar -->
</body></html>
```

Landmarks: one `<header>`, one `<main>`, one `<footer>`, `<nav>` elements
labelled when there is more than one. Skip-links are not required at this
scale; focus rings (`:focus-visible`, accent outline) come from theme.css.

## 2. Width, gutters, reading measure

| token / class     | value                    | use |
|-------------------|--------------------------|-----|
| `--shell-max`     | `1240px`                 | the single outer shell width, all page types |
| `--shell-gutter`  | `24px` → 18 (≤768) → 14 (≤480) | horizontal padding of the shell |
| `--prose-measure` | `68ch`                   | running copy; applied via `.jdbb-prose` to a block or its direct `p/ul/ol/blockquote` |
| `.jdbb-shell`     | class                    | `width:100%; max-width:var(--shell-max); margin:0 auto; padding:0 var(--shell-gutter)` |

Rules:
- Ledgers, tables, grids, forms wider than a paragraph use the full shell.
- Prose never runs the full shell: wrap it in `.jdbb-prose` (or set
  `max-width: var(--prose-measure)` on the copy block). Ledes may go to 74ch.
- Forms cap at `640px` (`form.reg` convention) and left-align.
- **No other outer widths.** 860/920/960/1040/1100/1120/1180 are retired.
  Legacy wrappers (`.page`, `.wrap`, `.content`) alias to the vars while they
  are migrated; new pages use `.jdbb-shell` directly.
- Exception: `/exedeck` keeps its own 960px stage (slides are a fixed-aspect
  artifact, not a document).

## 3. Wordmark (decided 2026-09-11)

One markup (§1), one CSS block (theme.css "Canonical wordmark"). The wordmark
**wears the selected site face** — the whole page reflows together, which is
the point of the selector. The j/d collision is prevented centrally:

- `.kj { letter-spacing: 0 }` by default (proportional faces);
- `-0.06em` only under `html[data-font="jetbrains"|"martian"]` (monospace
  gap after the bracket);
- serif faces get `+0.04em` overall tracking.

**Never add page-local `.kj` rules.** If a face misbehaves, fix theme.css.

Font selection behaviour (theme.js):
- First visit: random pick from the **sans/mono group only** (JetBrains,
  Martian, Plex, Geist). Serifs are there to try, never auto-picked.
- The random pick is sticky for the browser session (`sessionStorage`) so
  navigation doesn't reflow.
- Choosing a face in the selector makes it permanent (`localStorage`
  `prodcal-theme-v1`). Dark toggle is always permanent.
- `theme.js` applies stored state before first paint and auto-mounts any
  `#theme-bar`.

## 4. Shared chrome

| piece | class | notes |
|---|---|---|
| masthead | `.jdbb-masthead` | wordmark left; `<nav>` right with ≤4 links + `#theme-bar` last. Links are **Sentence case**, 11px uppercase mono is applied by CSS — write them as words ("Client portal", not "CLIENT PORTAL"). Wraps under 640px. |
| statusline | `.jdbb-statusline` | optional. Left: `● context`; right: revision/date/meta. Not on admin tools that already have a hero. |
| theme bar | `#theme-bar` | present on every standard shell. Print views hide it (`@media print` in theme.css). Never on the deck stage itself (the deck has its own controls). |
| footer | `.jdbb-footer` | **minimum payload**: wordmark (no "studio"), a `<nav>` of 2–4 context links, `© 2026 Jenna Dixon`. Homepage adds the resources strip above. Admin pages may add `Admin` to the nav; client-facing pages never do. |
| vocabulary | | Home · Workshop · Field notes · Lit-mag stack · Client portal · Admin · Contact. |

## 5. Page-type variants

| type | shell | prose | statusline | theme bar | footer | examples |
|---|---|---|---|---|---|---|
| marketing | 1240 | lede 74ch | yes | yes | + resources strip | `/` |
| prose / reference | 1240 | 68ch | yes | yes | standard | `/field-notes`, `/workshop`, `/factory` |
| wide ledger | 1240 | — (table full width) | yes | yes | standard | `/litmags`, `/admin/registrations` |
| admin tool | 1240 | — | hero instead | yes | standard + Admin link | `/admin/`, `/admin/registrations` |
| client tool | 1240 | — | app chrome | yes | standard, no Admin | `/{client}/`, `/{client}/{project}/`, `…/factory/`, `…/transmittal/` |
| quoted artifact | 1240 outer | inner doc keeps own style | yes | yes | standard | `/field-guide`, `/work-notes-standard`, `/architecture-plan` — generated, see §6 |
| deck / special | own | own | own | own controls | identity only | `/exedeck` |
| cohort roster | 1240 | — | hero | yes | standard, no Admin | `/cohort/protocolize-your-book-2026-09` (client tier + cohort flag) |

## 6. Ownership: embedded vs on-disk

| belongs in `prodcal/srv/static` (embedded, rebuild to change) | belongs in `/home/exedev/pi-public` (on-disk, live on save) |
|---|---|
| anything with JS that talks to the API as an *application* (admin, client portal, SPA, factory, transmittal, roster) | documents: talks, handouts, references, offer pages |
| `theme.css`, `theme.js`, favicon — the design system itself | anonymized client artifacts (**generated** by `client-raw/anonymize.sh`; never hand-edit outputs) |
| login/gate screens | anything a non-engineer edits on a different clock than the code |

On-disk documents **link** `/static/theme.css` and `/static/theme.js`; they do
not inline copies (decision 2026-09-11: the open-from-disk requirement is
retired). A page-local `<style>` holds only page rules.

## 7. Visibility tiers and recipes

**Authorization ≠ discoverability. Unlisted is not private.**

### Public (no app auth)
Route → `servePublicDoc(w, "name.html")` from `pi-public/`, or an embedded
static file. Recipe: drop file in `pi-public/`, add
```go
mux.HandleFunc("GET /mypage", func(w http.ResponseWriter, r *http.Request) { s.servePublicDoc(w, "mypage.html") })
```
rebuild once. Content must be safe for the open web; no PII, no client names.

### Client-visible (client cookie / project token)
Concrete today: `/{client}/…` routes guarded by `checkAuth(r, projectID)` /
`checkClientAuth(r, slug)` (`srv/server.go`, `srv/client.go`). Scope is the
`client_slug` in the URL path, verified against the `prodcal_client_{slug}`
HMAC cookie. One client cannot see another's material because the cookie is
bound to the slug.

Recipe for a client-visible *page* (application):
```go
mux.HandleFunc("GET /{client}/thing/", func(w http.ResponseWriter, r *http.Request) {
    slug := r.PathValue("client")
    if !s.checkClientAuth(r, slug) && r.Header.Get("X-ExeDev-UserID") == "" {
        http.Redirect(w, r, "/"+slug+"/", http.StatusSeeOther) // client login screen
        return
    }
    serveStaticHTML(w, "thing.html")
})
```
Recipe for a client-visible *document*: **not built yet** (no queued need,
decision 2026-09-11). When one arrives: add `PRODCAL_CLIENT_DOCS` root
`pi-client/{slug}/…`, a `serveClientDoc(w, r, slug, name)` that calls
`checkClientAuth` first and never accepts `..`, and register per route.

Cohort flag: `clients.cohort_slug` (migration 024), set automatically when a
Factory Pass coupon bound to a workshop registration is redeemed (or by admin
SQL). `requireCohort(w, r, cohort)` (`srv/cohort.go`) accepts exe.dev admin
or any valid `prodcal_client_{slug}` cookie whose client is in that cohort.
Used by `/cohort/{cohort}` + `/api/cohort/{cohort}/roster`; the HTML shell is
public but carries no data — the JSON is what is gated.

Smoke account: **Mike Check** (`bookiq@gmail.com`, client `mike-check`,
registration #10) is a faux cohort member for exercising every flow end to
end. Keep it out of real announcements (uncheck it) and delete after the
workshop.

### Admin-only
`requireExeDevAdmin(w, r)` for pages, `requireExeDevAdminAPI` for JSON; both
key off the exe.dev `X-ExeDev-UserID` header injected by the proxy. Recipe:
```go
mux.HandleFunc("GET /admin/thing", func(w http.ResponseWriter, r *http.Request) {
    if !s.requireExeDevAdmin(w, r) { return }
    serveStaticHTML(w, "thing.html")
})
```

### Discoverability (orthogonal)
| level | how |
|---|---|
| listed | in the homepage resources strip and/or the standard footer nav |
| navigation-only | linked from admin/client mastheads, not the homepage |
| unlisted | routable, linked from a deck/talk/email only; recorded in the admin Pages registry |
| retired | route removed or `301` to successor (`/lg → /`) |

Every route, whatever its tier, is recorded in the **admin Pages registry**
(`/admin/` → Pages panel, table `site_pages`). That is the inventory of
record; §10 below is a snapshot.

## 8. Deploy / update behaviour

| hosting | change content | change route/auth |
|---|---|---|
| embedded (`srv/static`) | `make build && sudo systemctl restart prodcal` | same |
| on-disk (`pi-public`) | save file → live on reload | rebuild once |
| generated companions | edit `client-raw/anonymize.sh` (or the wrapper it emits) → rerun → live | rebuild once |

## 9. QA checklist (per page, per change)

Run 2026-09-11 (post-normalization): 1280-wide + 390 mobile emulation on
`/`, `/workshop`, `/field-notes`, `/litmags`, `/field-guide`, `/admin/`
(Pages tab), `/admin/registrations`, `/cohort/…` (gated + member views),
`/{client}/{project}/` SPA; light + dark; wordmark in all 8 faces (no j/d
collision); workshop form + config endpoint present; Factory redeem → cohort
flag → roster 200 as member / 401 anon + outsider (`TestCohortRosterGate`).


- Widths: 1440 / 768 / 390. No horizontal scroll; masthead wraps under 640.
- Themes: light + dark; body bg is `--bg`, no hard-coded colors leaking.
- Fonts: JetBrains, Martian, Plex, Newsreader at least — wordmark has no j/d collision in any.
- Print (where supported): theme bar/footer/statusline hidden, content flows.
- Keyboard: tab through masthead → theme bar → content → footer; rings visible.
- Landmarks: header/main/footer present; nav labelled if >1.
- No-JS: page readable in default face, light theme; no blank regions.
- Functional: workshop form POST, factory redeem, client login, transmittal, cohort edit/print/CSV, email-consent guard — untouched.

## 10. Inventory (snapshot 2026-09-11)

See the admin Pages registry for the live list. Snapshot:

| route | owner | visibility | listed | shell | exception |
|---|---|---|---|---|---|
| `/` | prodcal | public | — | 1240 | marketing variant |
| `/lg` | — | retired → 301 `/` | — | — | |
| `/admin/` | prodcal | admin | nav | 1240 | |
| `/admin/registrations` | prodcal | admin | admin nav | 1240 | print modes |
| `/cohort/protocolize-your-book-2026-09` | prodcal | client + cohort flag | admin/tracker nav | 1240 | shows names, material, goals, sessions only |
| `/{client}/` | prodcal | client | — | 1240 | |
| `/{client}/{project}/` | prodcal | client | — | 1240 | SPA |
| `/{client}/{project}/factory/` | prodcal | client | — | 1240 | |
| `/{client}/{project}/transmittal/` | prodcal | client | — | 1240 | print |
| `/stylesheet/`, `/stylesheet/authors` | prodcal | public | resources | 1240 | inner editorial UI; split pending (§11) |
| `/factory` | pi-public | public | unlisted | 1240 | |
| `/exedeck` | pi-public | public | unlisted | 960 | deck |
| `/litmags` | pi-public | public | resources | 1240 | |
| `/field-notes` | pi-public | public | resources | 1240 | |
| `/workshop` | pi-public | public | resources | 1240 | live form |
| `/field-guide`, `/work-notes-standard`, `/architecture-plan` | pi-public (generated) | public | unlisted | 1240 outer | quoted artifact |

## 11. Migration / deprecation

- Retired: `/lg`; per-page `.kj` kerning; inlined theme token copies in
  pi-public; outer widths other than 1240/960(deck); `footer.site` markup
  (→ `.jdbb-footer`); ad-hoc `.wrap/.page/.content` widths.
- Pending (post 2026-09-11): `/stylesheet` split — current tool moves to
  `/stylesheet-pi/`; a read-only anonymized universal stylesheet publishes at
  `/stylesheet/` with fiction/nonfiction tagging.
- Idea log (`docs/IDEAS.md`): per-client interactive stylesheet instances.
