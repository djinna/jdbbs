# Fresh session prompt — normalize jdbb page chrome and visibility (2026-08-31)

Work across the two repos that make up the jdbb web surface:

- `/home/exedev/prodcal` — Go application, embedded admin/client/static pages, shared `theme.css` + `theme.js`, route/auth code. Read `AGENTS.md` and `CLAUDE.md` first.
- `/home/exedev/pi-public` — on-disk HTML documents served by ProdCal at request time. Read `README.md` first. It has its own git history.

The production app is `prodcal.service` on port 8000. Build/restart application changes with:

```bash
make build && sudo systemctl restart prodcal
```

Files in `pi-public` publish immediately on reload after editing; adding or changing routes still requires rebuilding ProdCal.

## Goal

Normalize the studio's page layouts and shared chrome across existing pages, then leave behind a durable design/hosting/visibility standard for future pages we bolt onto the book-factory UX.

Do not merely write recommendations: inventory, implement the normalization, QA it, deploy it, and finish with the reference document described below.

## Pages and routes to inventory

### Application-owned (`prodcal/srv/static`)

- `/` → `landing.html`
- `/admin/` → `admin.html` (exe.dev-admin only)
- `/admin/registrations` → `registrations.html` (exe.dev-admin only; new cohort tracker)
- `/{client}/` → `client.html`
- `/{client}/{project}/` → `index.html`
- `/{client}/{project}/transmittal/` → `transmittal.html`
- `/lg` → `landing-lg.html` (review-only variant; decide whether to retire)
- The stylesheet/review surfaces registered by `registerStylesheetRoutes`; include them in the inventory even if their inner document design remains specialized.

### On-disk documents (`/home/exedev/pi-public`)

Studio/Terminal Folio pages:

- `/exedeck` → `exedeck.html`
- `/litmags` → `litmags.html`
- `/field-notes` → `field-notes.html`
- `/workshop` → `workshop.html` (live registration form; preserve its API contract)

Anonymized client companion pages:

- `/field-guide`
- `/work-notes-standard`
- `/architecture-plan`

The companion pages intentionally preserve a distinct quoted-artifact visual style. Normalize their **outer studio chrome, width contract, navigation, theme controls, and footer** where sensible without erasing the identity of the embedded client documents. Never hand-edit generated client content contrary to `pi-public/README.md`; use the anonymization pipeline.

## Current variance already observed

Confirm this inventory yourself rather than trusting it blindly:

- Outer widths currently include approximately 860, 920, 960, 1040, 1100, 1120, 1180, and 1240px.
- Homepage outer width is `1120px`; use that as the leading candidate for the standard shell.
- `/litmags` is intentionally data-wide at `1240px`; it may warrant an explicit wide-page modifier rather than setting the global default that wide.
- Reading copy generally wants an independent `66–70ch` measure inside the outer shell.
- `theme.js`/font switching exists on many pages but is absent or incompletely mounted on others, notably the generated companion pages, transmittal, and some newer admin surfaces.
- Footer markup and content vary substantially; some pages have none.
- Public documents duplicate/in-line portions of shared Terminal Folio CSS, which has already drifted from `srv/static/theme.css`.
- The current wordmark markup usually wraps only the `j` in `<span class="kj">`; historical negative tracking on that span creates the visible **j/d smash** in some fonts. `exedeck.html` and `litmags.html` have newer per-font workarounds, while older pages still use `letter-spacing:-.08em` unconditionally.
- `srv/static/theme.css` itself currently appears to contain a duplicated `.theme-bar {` opener; inspect and correct if confirmed.
- Navigation vocabulary and capitalization differ page to page.
- Header/statusline combinations vary: some pages have both, some only a masthead, and the deck has specialized controls.
- The homepage has the strongest overall shell and should be the visual baseline, not necessarily a pixel-for-pixel template for every document type.

## Required design decisions and implementation

### 1. Establish a page-width contract

Prefer a standard outer shell around the homepage's `1120px`, with shared responsive gutters. Define explicit modifiers rather than ad hoc widths, for example:

- standard application/document shell: ~1120px
- readable prose measure: ~66–70ch within that shell
- wide ledger/data surface: ~1240px, opt-in
- presentation/deck surface: documented specialized exception

Use shared variables/classes in `theme.css` where possible. Public pages should consume the same contract without losing reasonable standalone-file fallback behavior.

### 2. Fix the wordmark once

Create one canonical `[jdbb] studio` markup and CSS treatment. Eliminate the j/d collision across **all eight selectable fonts**, light/dark modes, and mobile widths. Do not keep a growing list of page-local kerning patches.

Decide whether the wordmark should:

- remain in the selected site face with safe neutral tracking, or
- use a fixed brand face independent of the body/theme selection.

Choose deliberately, document the decision, and update every page to match.

### 3. Standardize shared chrome

Define and apply canonical versions of:

- masthead
- optional statusline/context strip
- navigation link treatment and capitalization
- footer
- theme + font switcher placement
- responsive collapse behavior

The footer should have a stable minimum payload—studio wordmark/identity, useful navigation or context, and copyright/contact as appropriate—while allowing documented page-type additions.

Add the theme/font switcher everywhere a normal studio shell appears. Specialized print views can suppress it. Ensure stored theme applies before paint to avoid flashing.

### 4. Reduce future drift

Avoid copying another large private fork of tokens into each new page. Establish a maintainable pattern for application pages and on-disk documents—shared stylesheet/classes, a small fallback, reusable snippets, or a tiny shell helper. Preserve the ability to open public documents directly from disk if that remains a real requirement; otherwise explicitly retire that requirement.

### 5. QA the system, not one screenshot

At minimum test representative pages at:

- 1440px desktop
- ~768px tablet
- 390px mobile
- light and dark themes
- several contrasting fonts, including JetBrains, Martian, a proportional sans, and a serif
- print where the page supports printing

Check focus visibility, semantic landmarks, contrast, overflow, long navigation, and no-JS degradation. Preserve all functional behavior—especially workshop registration submission, cohort editing/printing/email selection, client auth, transmittals, and admin controls.

## Hosting and visibility model

Authorization and discoverability must be treated separately: **unlisted is not private**.

The final implementation/document must define supported patterns for three visibility levels:

1. **Public** — no application auth; suitable for marketing, talks, public references, and registration forms.
2. **Client-visible** — accessible only through valid client/project authentication and scoped so one client cannot see another client's material.
3. **Admin-only** — requires `requireExeDevAdmin` / `requireExeDevAdminAPI` and the exe.dev identity header.

Current facts to account for:

- `servePublicDoc` reads named files from `/home/exedev/pi-public` and performs no auth check.
- `/admin/` and `/admin/registrations` use exe.dev-admin guards.
- Client/project pages use the existing client cookie / project auth cascade (`checkAuth`, `requireAuth`, and client-auth helpers).
- Embedded static pages require a Go rebuild/restart to change.
- On-disk public docs can be edited live without rebuilding once their route exists.

Design a safe, boring way to add future documents at each level. Strongly consider separate on-disk roots and explicit serving helpers/route registries for public, client, and admin documents rather than a single ambiguously named directory. Document how client scope is supplied and checked. Do not place sensitive files in a public route and rely on obscurity.

Also define **visibility/discovery controls** separately:

- linked from homepage/resource index
- linked only from admin/client navigation
- routable but intentionally unlisted
- retired/redirected

## Required final document

End the session by creating and committing:

`prodcal/docs/PAGE-DESIGN-HOSTING-VISIBILITY-2026-08-31.md`

It should be usable without rereading the implementation session and include:

- canonical page anatomy with concise HTML examples
- width/gutter/reading-measure rules and allowed modifiers
- canonical wordmark markup
- canonical masthead, statusline, theme switcher, and footer patterns
- page-type variants: marketing, prose/reference, wide ledger, admin tool, client tool, deck/special artifact
- ownership rule: what belongs embedded in ProdCal vs. in an on-disk document repo
- exact recipes for adding public, client-visible, and admin-only pages
- route/auth helper examples
- discoverability/listing rules
- deploy/update behavior for each hosting mode
- accessibility, responsive, dark-mode, and print checklist
- an inventory table of every current page, its owner repo, route, visibility, width variant, and any deliberate exception
- a short migration/deprecation section for legacy page patterns

Update `pi-public/README.md`, `prodcal/docs/DESIGN-SYSTEM.md`, and other existing docs as needed so they point to the new source of truth rather than contradicting it.

## Current cohort-tracker context

A cohort tracker is being added at `/admin/registrations`. As of this handoff, its preview includes:

- explicit **Edit registration** per participant
- status, prep, attendance, private notes, and all original signup fields
- consent-aware one-to-one announcement email selection
- **Print verbose**
- **Print terse · 1 page** (landscape one-page roster)
- CSV export

Treat it as an admin-tool page in the normalization. Do not remove either print mode or weaken the server-side email-consent check.

## Working discipline

- Keep unrelated application behavior untouched.
- Make focused commits in each repo; do not combine both repositories into one commit.
- Run `gofmt`, `go test ./...`, `go vet ./...`, and `git diff --check` in ProdCal.
- Browser-smoke representative routes after deployment and capture screenshots for comparison.
- Commit the final dated reference document before ending.
