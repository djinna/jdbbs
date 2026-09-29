# Brief for jd-blyg's Shelley: a publishing token for the blyg (2026-09-29)

Paste this whole file into a Shelley conversation **on the `jd-blyg` VM**.

## Why

Jenna's book-production app (ProdCal, on the `jdbbs` VM, served at
https://studio.jdbb.net) will post Book Factory news to this blyg
(https://blyg.jdbb.net) — e.g. "What's new" entries once she approves them.
Today the only way in is the studio password, which sets a browser cookie.
We don't want the password on another machine. We want **one revocable
publishing token** that Jenna stores in an exe.dev integration, so the token
never sits on the `jdbbs` VM either (the edge injects it into requests).

## What we found (from outside, 29 Sep)

- This blyg runs `blygger-studio/0.7.0` (from `/blyg.json`).
- Upstream `/api` is private and unversioned: ~30 endpoints behind a single
  owner cookie. A stable publishing API is an open upstream design question
  (roadmap item 1.8). `/api/*` without the cookie → `401 {"error":"unauthorized"}`.
- A downstream fork (aneeshsathe/blygger-desktop, upstream issue
  blygger/blygger-studio#11) solved the same problem with a single bearer
  token (`BLYG_OWNER_TOKEN`) accepted wherever the session cookie is. Follow
  that shape so this stays close to what upstream is likely to adopt.

## What to build

1. **Token check in the existing owner guard.** Wherever the studio accepts
   the owner session cookie for `/api/*`, also accept a token, read from
   **either** header:
   - `X-Blyg-Token: <token>` (preferred — see the note below), or
   - `Authorization: Bearer <token>`.
   Compare in constant time. Token lives in a secret/env var (e.g.
   `BLYG_OWNER_TOKEN`), never in the repo, never logged.
   Token-authenticated requests must not need CSRF/origin checks meant for
   the browser, but must still be rejected when the token is wrong or unset
   (unset = token auth off).
2. **Scope it if cheap:** ideally the token can create/update items and
   upload media, but not change settings, password or subscriptions. If
   that's a big change, full owner scope is acceptable for v1 — say which in
   your reply.
3. **Rotate/revoke:** document how Jenna changes the token (edit env + restart
   is fine). Tell her the exact commands.
4. **Minimal documentation** in the repo README: the header, which endpoints
   create a post (method, path, JSON body fields for a simple fragment:
   title, body/markdown, kind), and the response (id + public URL).
5. **Keep upgrades possible.** Put the change in its own file(s)/small patch
   so `npm run upgrade` to future blygger-studio releases stays easy; note it
   in whatever local-changes log the repo keeps.

Why `X-Blyg-Token` as well as `Authorization`: the exe.dev integration that
already fronts this VM for the `jdbbs` agent uses `Authorization` for the
exe.dev key. A separate header avoids the two colliding. The plan is a second
integration pointed at the public origin, e.g.
`integrations add http-proxy --name blyg-publish --target https://blyg.jdbb.net --header "X-Blyg-Token:<token>" --attach vm:jdbbs`.

## Also, while you're there (ask Jenna before changing — these alter public URLs)

- The blyg's **site origin** is still `https://jd-blyg.exe.xyz/` (in
  `/blyg.json` `site`, feed `<link>`, `og:url`, webmention link). Now that
  https://blyg.jdbb.net is live it should probably be the origin. There is
  only one post so far, so now is the cheapest moment. Check what the
  protocol/upstream says about changing origin (item URLs, subscribers,
  webmentions) before doing it.
- The author link points at `https://jdbbs.exe.xyz/factory`; the canonical
  address is now **https://studio.jdbb.net/factory** (old one still redirects).

## Test to report back

From any machine:

    curl -s -o /dev/null -w '%{http_code}\n' https://blyg.jdbb.net/api/items            # 401
    curl -s -o /dev/null -w '%{http_code}\n' -H "X-Blyg-Token: wrong" https://blyg.jdbb.net/api/items   # 401
    curl -s -H "X-Blyg-Token: $BLYG_OWNER_TOKEN" https://blyg.jdbb.net/api/<read or list endpoint>   # 200

Don't publish a test post to the live blyg without Jenna's OK (use a draft
endpoint if one exists, or delete it after). Reply with: what you built, the
scope, the post-creation endpoint + example body, and the rotation steps.
