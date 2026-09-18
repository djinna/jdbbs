# TypeSafe / Jev review — would it make the factory more efficient?

**Date:** 2026-09-18 · **Source:** docs.typesafe.ai (Introduction, Models, Confidence, State, cookbooks)

## What it is

TypeSafe sells one hosted model, Jev, behind a single HTTP endpoint (`POST /v1/systemone`). Their
pitch: "Send state and typed questions; get structured answers your code can use directly." It does
**not** generate text. You send a blob of state (string or JSON, ≤32k tokens) plus a bag of narrow
questions of three kinds — Choice (pick one of N), Score (rate against ordered levels), Noul (yes/no
probability) — and get back the answer, a probability distribution and a `confidence` number. Many
questions go in one request at almost no extra cost. Python and JS SDKs, no Go SDK (it's a plain JSON
POST, so fine). $0.042 per million input tokens, output free — cents per book at our volumes.

## Where it could plug into the factory

- **Front/back matter classification in the book map.** `srv/bookmap.go` classifies headings with
  hand-maintained vocab lists and prefix matching. One Choice per heading (front/body/back × section
  type) with confidence gating would replace the list-maintenance treadmill. Moderate effort (~a day),
  real value: this is the code most likely to be wrong on an unusual manuscript.
- **Heading lookalikes.** `detect_heading_lookalikes` in `detect-edge-cases.py` guesses whether a bold
  centred paragraph is a heading or just emphasis. A Noul per candidate is a near drop-in for that
  heuristic and gives us a number to threshold on. Low effort, decent value, easy to A/B against the
  existing rule.
- **Pre-filling the transmittal from the manuscript.** Regex out candidate titles, authors, copyright
  years, dedications; ask Jev to pick the right span (their "pre-parsed value extraction" pattern).
  Medium effort, medium value — saves Jenna typing, but it must stay a suggestion she confirms.
- **Triage of Inspect findings.** A Score per finding ("how likely is this to need a human fix?")
  would let the report lead with the three things that matter. Low effort, modest value, additive.
- **Clearer author-facing failures.** Little on offer. Jev can classify a failure but cannot write the
  explanation; our messages are bad because nobody has written them, not for want of a model.

## What it is not / risks

It is not a document toolchain, and it will not make the pandoc↔typst↔Python glue less fragile — that
glue is our problem, not a classification problem. It is not an agent runtime and cannot "fix" styles.
Adopting it puts a network call and an API key inside a build path that is currently hermetic and
unit-testable; adds nondeterminism to outputs we snapshot in tests (mitigable by caching answers per
manuscript hash and pinning `jev-1.13.0` rather than the moving `jev-latest` alias); and ships
unpublished manuscripts to a third party — they say they don't train on customer data, but zero
retention is enterprise-only. Their docs warn rate limits can change without notice, and the 32k state
cap means chunking anything novel-length.

## Recommendation

**Pilot on one thing only:** heading/section classification in the book map, behind a flag, with the
vocab-list heuristic kept as fallback and answers cached. Park everything else.
