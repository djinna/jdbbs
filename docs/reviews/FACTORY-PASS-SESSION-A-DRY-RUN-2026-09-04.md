# Factory Pass — Session A production dry run (2026-09-04)

Session A blockers were deployed to `prodcal.service` on the canonical VM and
verified against the production database and toolchain.

## Code shipped

- Admin password rotation: `POST /api/admin/clients/{slug}/password`, with a
  tracker button after redemption. Old passwords and client cookies stop
  working. The replacement is returned once to the admin; AgentMail resend
  success is reported separately.
- EPUB fallback fonts are selected from text nodes in the corrected DOCX.
  Latin-only books embed no Noto fonts; CJK gets Noto Serif TC and Thai gets
  Noto Serif Thai. The `/licensed/` hard guard remains in place.
- Conversion failures store customer-readable messages. Raw pandoc/Typst
  stderr is retained in structured server logs as `raw_error`. A ready PDF with
  a failed EPUB is now a warning in the Factory UI rather than a failed build.

## Production exercise

1. Issued and redeemed an unbound self-test coupon using `j@djinna.com`.
2. Confirmed the fulfillment email returned AgentMail HTTP 200.
3. Reset the password through the new admin endpoint; confirmed the resend
   returned HTTP 200 and used the one-time admin copy to authenticate as the
   customer.
4. Uploaded `manuscripts/samples/sample-chapter.docx` (five Word styles), ran
   Inspect, and got 10 findings: 3 high, 4 medium, 3 low.
5. Built and downloaded both artifacts through customer authentication:
   - print PDF: 29,934 bytes, 2 pages
   - EPUB: 5,614 bytes, no embedded Noto files
   - EPUBCheck: 0 fatals, 0 errors, 0 warnings
6. Rotated the password again to simulate loss. The old password returned 401;
   the replacement returned 200; the resend again returned AgentMail HTTP 200.
7. Uploaded the real Twitter Years DOCX as an error smoke. Typst failed on
   `tweet-p`; the customer row contained only the Word-style guidance, while
   journal logs retained the full Typst source excerpt under `raw_error`. The
   failed build was refunded and the smoke book was deleted afterward.

The retained dry-run pass is
`/jenna-dixon-dry-run/session-a-dry-run-2026-0/factory/`; it has one successful
build used and two remaining.

## Workshop issuance

Production had **7 registrations**, not 8, on September 4. Issued one code for
each registration. Together with the redeemed self-test coupon, production now
has 8 Factory Pass coupons total.

Sent six individualized prep announcements (one per consenting registrant),
each containing that person's code and the instruction **“Redeem before session
1”**. All six AgentMail requests returned HTTP 200. Andrea Leiter's code is in
the tracker but was not emailed because her registration has `consent_email =
false`.
