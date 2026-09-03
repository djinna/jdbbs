# Factory Pass — first end-to-end smoke (2026-09-03)

Ran the built binary from `scratch/smoke/` (fresh DB, port 8101, no email env)
against the real pipeline on the VM. Prod untouched.

## What passed
- `POST /api/public/register` → admin `POST /api/admin/coupons` (registration
  1) → `PYB-CAU9-QY8X` → `POST /api/public/redeem` → client `ada-petrov`,
  project `the-weather-in-kansas`, pass 3/3, portal URL returned. Second
  redeem → 400 "already redeemed". Fulfilment email correctly *logged* as
  skipped (no mailer).
- `/ada-petrov/the-weather-in-kansas/factory/` → password gate → unlock with
  client password → pass badge from live API.
- Customer cookie can upload, preflight (990 findings on a real docx), and
  build. Non-admin without cookie: 401.
- Ledger: two failed builds each wrote `-1 build` / `+1 build_failed_refund`;
  credits stayed 3. Successful build → `builds_used=1`, credits 2, **PDF +
  EPUB both written in one `converting` → `ready` cycle**, delivery email
  attempted. UI showed both download buttons.
- `GET /factory` serves `pi-public/factory.html`.

## Found & fixed
- **VM pandoc was 3.1.3 — the pipeline has been broken on this VM since
  the `typst+smart` change (8b1c6c4, May 26).** Every convert failed with
  `exit 23: The extension smart is not supported for typst`. journalctl has
  no successful conversions in its retained window. Installed **pandoc 3.11**
  from the GitHub .deb (`sudo dpkg -i`); builds succeed. **prodcal.service
  does not need a restart for this** (pandoc is exec'd per build), but note
  it in DEPLOY.md and pin the version in the deploy checklist.

## Found, not fixed (follow-ups)
1. **EPUB is 14 MB for a 3-chapter text file** — it embeds
   `NotoSerifTC-{Regular,Bold}.otf` (8 MB each) and Noto Thai. The EPUB
   fallback-font set should be subset or made conditional on script detection.
   Customers will notice; KDP has a 650 MB cap but delivery fees scale with size.
2. **Real manuscripts hit template-specific Typst functions** (`#tweet-p` in
   the Twitter Years docx → `unknown variable`). A pass customer's docx with
   custom Word styles will produce a failed build with a raw Typst error. The
   preflight's "Word styles not in your transmittal" is the right guard; the
   build error surfaced to customers should be softened to "a style in your
   file (X) isn't in the template — see preflight" rather than a Typst trace.
3. Password reset: fulfilment mails the only copy of the generated password.
   The UI has a "Lost the password?" link — verify where it goes; admin has no
   "reset client password" endpoint (had to bcrypt by hand for the smoke).
   Needed before Sep 21: an admin action or a magic-link.
4. Delivery email links to `/download/epub` even for PDF-only builds
   (backend agent's note).
5. `error_msg` doubles as the EPUB-failed warning on `ready` books — UI
   should render it as a warning, not an error, when status is `ready`.
6. Redeem is not rate-limited per *code* — brute-forcing `PYB-XXXX-XXXX`
   (32^8) against the 5/10min/IP limiter is impractical; fine for v1.
