# Factory Pass — TAM and price-point math (2026-09-03)

Companion to `STOREFRONT-FACTORY-PASS-DESIGN-2026-09-03.md`. Decisions taken
today: **3 included builds, unlimited preflights, 6-month storage, no Typst
source, no human time after the workshop, live support USD 100/hr.**

## 1. Market anchors (Sept 2026, US)

| Anchor | Number | Note |
|---|---|---|
| Self-published titles w/ ISBN, US 2025 (Bowker) | ~3.5M, +38.7% YoY | Inflated by AI titles and by per-format ISBNs (print + ebook = 2). Distinct human-authored *works*: guess 1–1.5M. |
| Traditionally published, US 2025 | 642k | Not our buyer. |
| Self-publishing *services* market (cover, editing, formatting, distribution), 2025 | ~$269M, 6.4% CAGR | Formatting is a minority slice; assume 15–20% ⇒ **$40–55M/yr US formatting spend**. |
| Global self-publishing market | ~$2.2B | Context only. |
| **Free formatters** | Reedsy Studio, KDP Kindle Create, Draft2Digital | Price floor = $0 for "acceptable." |
| **One-time software** | Atticus $147 · Cambric $199 · Vellum $199.99 (ebook) / $249.99 (ebook+print, Mac-only) | Unlimited books, *you* do the work. Reviews call Vellum "too much for one book." |
| **Freelance formatting** | simple text-only: $50–$300 · Reedsy pros: 50% of jobs $475–$1,275, mean ~$840 | Human does the work; typically 1–2 revision rounds included. |
| **Freelance revision round** | $25–$75 per round | The market's existing price for "an additional build." |
| Author income | 75% of self-pub authors earn <$1k/yr | Price sensitivity is real; the buyer is spending savings, not revenue. |

## 2. Where the pass sits

The pass is **per-manuscript, done-for-you-by-machine, house-typeset**. That is
a slot nobody occupies: software is unlimited-but-DIY; freelancers are
done-for-you-but-human-priced. For the modal buyer — an author with *one* book
written in Word — "unlimited books" is worthless and "$840 for interior" is
out of reach. Also unique: the preflight (tells you *what to fix*) and licensed
print fonts (Plantin etc.), which neither free tools nor Atticus provide.

Price ladder and what each signals:

| Pass price | Reads as | Support expectation it creates | Risk |
|---|---|---|---|
| $99 | "a tool, per book" | low | undercuts the studio brand; attracts price-shoppers who are the neediest |
| **$149** | "Atticus, but for this one book, done for you" | low–medium; self-serve plausible | — |
| $199 | "Vellum-tier, no Mac needed" | medium | buyers start expecting a reply to email |
| $249–299 | freelancer floor | high — buyers assume a human | conflicts with "no human time" |

**Recommendation: $149** for the pass. Rationale: it equals the most popular
software's *unlimited* price, so "one book, but I don't touch software and the
output is press-grade" is an easy trade for the one-book author; it stays
under the threshold where people assume a human is included; and it leaves
room for the studio's actual human service (bespoke typesetting at $800+) to
sit clearly above it. The pass is the **top of the funnel for the studio**,
not a business on its own — it qualifies leads for the hand-typesetting tier.

Sanity check on 3 builds: the modal loop is build 1 (see problems) → fix Word
→ build 2 → tweak → build 3 final. Freelancers include 1–2 revisions; we
include 2 plus an unmetered preflight, which is the more generous offer.

## 3. Add-on SKUs (anchored to the $25–$75 revision-round price)

| SKU | Price | Per unit | Why |
|---|---|---|---|
| **+3 builds** | **$49** | ~$16/build | Well under a freelancer revision round, so it never feels punitive; ~⅓ of pass price so it can't be used as a cheap second manuscript. Single SKU — no 1-build option (fewer decisions, and the 3-pack matches the mental model of the pass). |
| +6 months storage | $29 | — | Low-stakes; mostly exists so expiry has a graceful exit. |
| **Preflight review** (human) | $75 | one written reply | *The* support product with hard edges: one Word file + its preflight report → one email listing what to fix. No back-and-forth. Converts "needy" into "paid" and pre-empts the $100/hr channel. |
| Live support | $100/hr, 30-min minimum ($50), booked in advance | — | Per your call. Booking link, not email. |
| Studio typesetting (bespoke) | from $800 | — | Existing service; the pass page should point at it as "when you need a human." |
| Press bundle (5 passes) | $599 | $120/pass | For small presses / litmags (`litmags.html` has 132 rows — that's a list). |

**Support edges (copy):** "Email support is not included in a Factory Pass.
Your preflight report is the support. If you want a human: Preflight Review
($75, one written reply) or live help ($100/hr, 30-min minimum, booked
here)." Transactional mail's Reply-To goes to an address whose autoresponder
says the same thing.

## 4. Unit economics

Per pass: compute ≈ seconds of Typst/pandoc (≈ $0); storage ≈ tens of MB for 6
months (≈ $0); email ≈ cents; Stripe 2.9% + $0.30 ≈ **$4.62** on $149.
Contribution ≈ $144 before support. The only real cost is **support minutes**.
At an internal $100/hr, break-even against the pass is ~85 minutes of unpaid
help per customer — so the whole design problem is keeping average unpaid
support under ~20 min (≈ $33), i.e. the edges above are not optional.

## 5. TAM / SAM / SOM — honestly

- **TAM** (US paid interior formatting): **$40–55M/yr** (§1). ~300k paid jobs
  at ~$150 avg.
- **SAM** (Word-authored, text-first books whose authors want done-for-you
  output but won't pay freelancer rates; English; reachable online): ~10% of
  TAM ⇒ **$4–5M/yr**, ~30k passes.
- **SOM, year 1** (a solo studio with no ad budget, distribution = workshops,
  Protocol Institute network, litmag list, word of mouth): **50–200 passes**.

| Passes/yr | @ $149 | + add-ons (~15% attach) | ≈ Revenue |
|---|---|---|---|
| 50 | $7.5k | $1.1k | **~$9k** |
| 200 | $30k | $4.5k | **~$35k** |
| 500 | $75k | $11k | **~$86k** |

Read: the pass alone is a nice sideline, not a company. Its strategic value is
(a) a *product* the workshop hands attendees ("$149 value, on us"), (b) a
lead qualifier for bespoke typesetting where the margin actually is, and (c)
the litmag/small-press B2B bundle, which is where volume could come from
without marketing spend. Revisit after the cohort: if attendees' median build
count is ≤3 and support minutes stay <20, the price can move to $179–199
without changing the promise.

## 6. Workshop framing

Present the coupon as **"Factory Pass — $149 value, included with your seat."**
Anchoring a price now makes the post-workshop sale (extra builds, review, a
second manuscript) legible, and the cohort's behaviour gives real data on N
and support load before Stripe is wired up.
