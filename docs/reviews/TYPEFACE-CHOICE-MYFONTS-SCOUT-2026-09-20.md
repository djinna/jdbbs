# Typeface choice for customers — MyFonts / Monotype scout (5.22)

*Scout only, 2026-09-20. Nothing built. Question from Jenna (5.21): people should
be able to choose their own typefaces and buy the licences themselves — does
MyFonts have an API we could connect to?*

## Short answer

No usable storefront API. MyFonts once had a SOAP API for search + purchase
(third-party clients; c. 2012) and a `dev.myfonts.com` doc set (still linked
from a 2019 Google Fonts issue) — both are legacy from before the Monotype
consolidation and are not offered to new integrators. What Monotype sells today
is the **Monotype Fonts API** (`api.monotype.com`, `fonts-api.monotype.com`):
OAuth, font *search/inventory/similarity/pairing*, for **enterprise-plan
customers** of Monotype Fonts (a subscription), not a retail checkout we could
embed so a customer buys a licence on MyFonts and hands it to us.

The bigger obstacle is not the API, it is the **licence type**:

- A MyFonts retail purchase is a **desktop** licence (per user, per computer).
  Monotype's own developer guidance says retail desktop licences do not permit
  uploading the font to cloud platforms; cloud / SaaS use "usually requires
  server, app, or enterprise licenses".
- Our print PDF is rendered by Typst **on the VM**, not on the customer's
  machine. A customer's desktop licence therefore does not cleanly cover a
  build here, even though the resulting PDF (subset-embedded) is a permitted
  output of a desktop licence when *they* make it.
- Our own Plantin / Proxima are held under exactly that model
  (`typesetting/fonts/licensed/README.md`: desktop print-only, never in git,
  never in EPUB, `srv/epub.go` refuses `/licensed/` paths). That works because
  the studio is the licensee and the studio runs the build.

## What this leaves — four shapes

| | Shape | Licence story | Work | Fit |
|---|---|---|---|---|
| **1** | **OFL / Google Fonts menu** (what we have: EB Garamond, Libertinus, Source Sans, Noto, JetBrains Mono) | Clean for PDF *and* EPUB | none | free tier, today |
| **2** | **Curated studio menu** — 5–10 families the studio licenses with a **server/app** licence (Monotype, or independent foundries such as Klim, Commercial, Grilli, whose server licences are priced per CPU/domain) | Studio is the licensee; per-book **surcharge** (e.g. $50–150) recovers it. EPUB stays OFL/system fonts (as now) | font menu already reads `typesetting/fonts/` recursively; add a price flag per family + SKU | best next step; matches "keep page-design choices simple" |
| **3** | **Bring-your-own upload + attestation** — customer uploads OTFs for their book; we render, then delete | The customer attests they hold a licence that covers third-party rendering (Monotype "server/app", or a foundry that allows it). Retail desktop buyers will usually *not* qualify — say so plainly | upload endpoint per project, private font dir per build, `body-font`/`heading-font` already accept any family name; purge after build | possible; legal grey for most buyers; support cost |
| **4** | **Desktop app path** — `cmd/prodcal-app` (Mac) renders locally | A retail **desktop** licence *does* cover this: the customer's font, the customer's machine, PDF embedding permitted | the app prototype exists; the font menu would read `~/Library/Fonts` | the honest home for "buy it on MyFonts yourself"; post-workshop |

**Referral-out** (a MyFonts link from the factory page) is fine as a courtesy
link but earns nothing we can confirm — MyFonts has no public affiliate scheme
today — and only makes sense paired with 3 or 4.

## Recommendation

- Keep 1 as the free tier. Build **2** when a customer asks (a "Typeface" line
  on the pass, priced per book, families we already hold: Plantin MT Pro,
  Proxima Nova; add on demand). Say in `/factory` what is included and what a
  surcharge family is.
- Offer **4** to the person who has already bought a MyFonts family: the desktop
  app, not the web factory.
- Do **3** only with the attestation wording reviewed, and never for EPUB.
- Do not chase the Monotype API: enterprise subscription pricing, and it solves
  discovery, not licensing.

## Already in place

- `typography.body_font` / `typography.heading_font` → `body-font` /
  `heading-font` in the generated config (`srv/bookspecs.go` ~L637) → Typst
  only; EPUB never reads them.
- Font menu = `typst fonts --font-path typesetting/fonts` (recursive), so a new
  family is a directory drop + `scripts/sync-licensed-fonts.sh`.
- Plantin MT Pro **Bold** still missing (5.21) — Jenna supplies; never committed.
