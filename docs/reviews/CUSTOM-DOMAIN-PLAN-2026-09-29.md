# Custom domain — when and which (2026-09-29)

Discussion note; nothing changed yet.

## What's there today (checked 29 Sep)

- **jdbb.net** — DNS at Cloudflare (proxied, orange cloud). Apex serves the
  Blot blog *Jenna Dixon, Bookbuilder* (posts back to 2021). MX → Google
  (`aspmx.l.google.com`), so mail may be live there. Don't touch apex or MX.
- **jdbb.studio** — Porkbun, 3-year registration. Carries our *sending* mail:
  `factory@mail.jdbb.studio` via Resend (SPF/DKIM/DMARC). Apex points
  elsewhere (Porkbun parking).
- **App** — `jdbbs.exe.xyz`. `PRODCAL_BASE_URL` unset, so BaseURL defaults to
  `https://jdbbs.exe.xyz`; all email links, Stripe return URLs and cookie
  `Secure` flags derive from it.

## What gets harder the longer we wait

Not the code. BaseURL is one env var; ~15 hard-coded strings remain
(`srv/email_shell.go`, `settings.go`, `registration.go`, `housestyle.go`,
`books.go`, `runs.go`, `scripts/factory*.py`, `factory-demo.sh`, 8 in
`jdbbs-public`). Half a day including tests.

What grows is **URLs we can't edit once they're out**: sent emails (magic
links, digests, workshop mail), talk decks and handouts, API users' scripts,
the Stripe account's business website and return URLs, search-engine
indexing, `llms.txt` that agents cache, and — biggest — URLs printed in the book.

**Timing: soon, and before any of these:** the help system goes public, the
Stripe store goes live, API-workshop invites go out, the book's URLs are
fixed. Not in a send or workshop week.

`jdbbs.exe.xyz` keeps working forever as an alias (old links, admin fallback,
the :8766 punch-list page, which only works on exe.xyz ports anyway).

## Recommendation

**App at `studio.jdbb.net`** (matches the `[jdbb] studio` wordmark). One
Cloudflare CNAME `studio → jdbbs.exe.xyz`, **DNS only (grey cloud)**. Blot
keeps the apex; Google mail untouched; no apex ALIAS/IP issue. Later, if
the studio should *be* jdbb.net, move Blot to `notes.jdbb.net` and swap — a
separate decision (old blog URLs need redirects).

**Keep jdbb.studio for sending mail** for now — it's warmed and passing
DMARC. Long term: if jdbb.studio won't be renewed, move sending to
`mail.jdbb.net` in a quiet month (new domain needs warm-up). Optionally
forward jdbb.studio web → studio.jdbb.net.

## Steps when we do it

1. Jenna, Cloudflare: CNAME `studio` → `jdbbs.exe.xyz`, grey cloud.
2. Jenna, **from her own machine**: `ssh exe.dev domain add jdbbs studio.jdbb.net`.
3. Test before switching anything: public pages, client magic-link sign-in,
   factory upload, and whether exe.dev admin login (`/__exe.dev/login`) works on
   the custom host. Not documented either way. If it doesn't, admin stays
   on `jdbbs.exe.xyz` and admin links in emails use a separate admin base URL.
4. Set `PRODCAL_BASE_URL=https://studio.jdbb.net`; replace hard-coded hosts;
   add `rel=canonical`.
5. On `jdbbs.exe.xyz`: 301 public GET pages to the new host (path + query
   kept, so old magic links still redeem); never redirect `/api/` POSTs or
   admin. Clients sign in once more (cookies are per host).
6. Update footer/signature setting, `/factory/api`, CLI defaults, DEPLOY.md.

## Open questions

- `studio.` vs `books.` / `factory.` / apex?
- Is Google mail on jdbb.net live (j@jdbb.net)? Affects the contact address, not the site.
- Renew jdbb.studio when it's up, or plan to move mail off it?
