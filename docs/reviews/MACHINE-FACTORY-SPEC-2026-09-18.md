# Machine-callable factory — ½-page spec (5.8)

*Fri 2026-09-18. From P2 beats 8/9: "your factory calling my factory" = four
interfaces, three exist. This is what a caller can do today, what is missing,
and the smallest change that closes it.*

## What exists today (no code needed)

Auth: a **project token** (`X-Auth-Token` header; admin mints it with
`POST /api/projects/{id}/auth`). Every step below already accepts it — the
API doc's "Admin" column for upload/convert/preflight is stale; the code
gates on `requirePassAccess` (token + live Factory Pass + credits).

| Step | Call | Notes |
|---|---|---|
| 1 Transmittal | `PUT /api/projects/{id}/transmittal` JSON | same shape the form saves |
| 2 Template | `GET /api/projects/{id}/word-template` → .docx | **side effect:** syncs transmittal → book spec (only place a non-admin can) |
| 3 Manuscript | `POST /api/books/upload` multipart `file,title,author,project_id` → `{book_id}` | |
| 4 Inspect | `POST /api/projects/{id}/preflight` `{book_id}` then `GET …/preflight` | JSON incl. `book_map` finding; HTML at `…/preflight/report` |
| 5 Build | `POST /api/books/{id}/convert` `{"format":"both"}` → `{status:"converting"}`; poll `GET /api/books/{id}/outputs`; fetch `…/outputs/{oid}/download` | 1 credit per build |

## What is missing

1. **Spec sync is a side effect of downloading the template.** A caller that
   already has the template must still GET it to refresh the spec before a
   build, or builds against stale spec. → make `convert` and `preflight`
   pull transmittal → spec themselves when the transmittal is newer (same
   rule the template route uses).
2. **No completion signal.** Callers poll `outputs`. → `GET /api/books/{id}`
   returning `{status: converting|done|failed, error, outputs:[…]}`; optional
   `callback_url` on convert (POST the same JSON when done). Polling stays.
3. **Token issue is admin-only and manual.** Fine for two friends; before
   twenty, the client portal shows the token (or a magic-link-scoped one,
   5.7). Not needed for the talk.
4. **Docs.** `docs/API.md` rows 30/31/43–45 say Admin; fix to Project, and
   add a "Factory in five calls" walkthrough with `curl`.

## Proposal

Post-workshop, ~½ day: (1) + (2) + (4). Ship as one commit with an
end-to-end test that runs the five calls against a temp DB. Then the talk's
beat 8 is literally true: four interfaces, all four callable, and beat 9's
"shared folder and a person reading a report" becomes "a URL and a JSON".

**Decision needed:** ok to say in the talk that this exists as an API today
(it does, with the caveats above), and land the polish after Wednesday?
