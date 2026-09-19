# Driving the book factory from a terminal — CLI recipe

*2026-09-19. Punch-list 0.9B: "hit the book factory from your own factory".
One page, no browser. Every path below was run against the live server the
day this was written; the full reference is `docs/API.md` and the reasoning
is `docs/reviews/MACHINE-FACTORY-SPEC-2026-09-18.md`. A copy-paste Python
version (stdlib only) is `scripts/factory-cli.py`.*

## 0. What you need

| | Where it comes from |
|---|---|
| **Base URL** | `https://jdbbs.exe.xyz` |
| **Project id** (a number) | The Factory Pass email links to your factory page, `https://jdbbs.exe.xyz/{client}/{project}/factory/`. Take the two slugs from that URL and ask `GET /api/clients/{client}/projects` — the row whose `project_slug` matches carries `"id"`: `curl -s $B/api/clients/prot/projects \| jq '.[] \| select(.project_slug=="zoo") \| .id'`. |
| **Token** | The **project password** the studio set for your project (the same thing the factory page asks for when it says *Password*). The studio mints it with `POST /api/projects/{id}/auth`. |
| **A live Factory Pass** on the project | Upload, inspect and build all return `403 {"error":"this project has no factory pass"}` without one. |

Send the token on every call in any of these three equivalent ways:

```
Authorization: Bearer <token>      # new 2026-09-19; what most HTTP tooling has a switch for
X-Auth-Token: <token>
Cookie: prodcal_auth_<project id>=<token>   # what the browser uses
```

A wrong or missing token is `401 {"error":"unauthorized"}` on every route.

```sh
B=https://jdbbs.exe.xyz
P=14                                # your project id
H="Authorization: Bearer $TOKEN"
```

## 1. Check the pass (free)

```sh
curl -s -H "$H" $B/api/projects/$P/pass
# {"exists":true,"status":"active","live":true,"builds_included":3,"builds_extra":0,
#  "builds_used":1,"credits_remaining":2,"expires_at":"2027-03-19T12:41:46Z","customer_name":"…"}
```

## 2. Upload the manuscript (free)

Multipart; fields `file`, `title`, `author`, `project_id` — all four required
(`400 {"error":"title and author required"}` otherwise). Each upload is a new
book row on the project — it does not overwrite the last one.

```sh
BOOK=$(curl -s -H "$H" -F file=@manuscript.docx -F title='My Book' -F author='Me' \
            -F project_id=$P $B/api/books/upload | jq .id)
# 201 {"author":"Me","id":26,"status":"uploaded","title":"My Book"}
```

## 3. Inspect / preflight (free)

`POST` runs it and returns the report; `GET …?book_id=` re-reads the latest
(the `book_id` query parameter is required on the GETs). The HTML version is
at `report_url`.

```sh
curl -s -H "$H" -X POST $B/api/projects/$P/preflight \
     -H 'Content-Type: application/json' -d "{\"book_id\":$BOOK}" | jq .
curl -s -H "$H" "$B/api/projects/$P/preflight?book_id=$BOOK" | jq .summary
curl -s -H "$H" "$B/api/projects/$P/preflight/report?book_id=$BOOK" -o report.html
```

The JSON is `{exists, project_id, book_id, status: ready|error, source_filename,
updated_at, summary:{total, high, medium, low, preserved, by_type:{…}},
book_map:{sections:[{title, kind, paras, words}], warnings, notes, summary},
report_url, history:[…]}`. The per-finding detail (what to fix, and where) is
only in the HTML report today — the JSON carries counts and the book map.

## 4. Build (one credit)

`format` is `"both"` (default: print PDF + EPUB, one credit), `"pdf"` or `"epub"`.
Optional `callback_url` (public https; the server POSTs the status JSON from
step 5 there once, when the build ends — loopback/private targets are refused).

```sh
curl -s -H "$H" -X POST $B/api/books/$BOOK/convert \
     -H 'Content-Type: application/json' -d '{"format":"both"}'
# 200 {"book_id":26,"format":"both","status":"converting","status_url":"/api/books/26"}
# 402 {"error":"no builds remaining","credits_remaining":0}
# 409 {"error":"a build is already running for this project; wait for it to finish"}
```

Credits: the debit happens when `convert` is accepted. A build that **fails**
refunds the credit automatically (ledger rows `build` / `build_failed_refund`).
Inspect, upload, status and downloads are free. A final transmittal saved
after the spec was last written is pulled into the spec at convert time, so
you do not need to fetch the Word template first.

## 5. Poll until done

```sh
until curl -s -H "$H" $B/api/books/$BOOK | jq -e '.status=="ready" or .status=="error"' >/dev/null; do sleep 3; done
curl -s -H "$H" $B/api/books/$BOOK | jq .
```

Success (a ~10-page sample took ~10 s; a 110-page book ~30 s):

```json
{"book_id":26,"project_id":14,"title":"My Book","author":"Me","source_filename":"manuscript.docx",
 "status":"ready","updated_at":"2026-09-19T12:42:33Z",
 "outputs":[{"id":73,"format":"epub","size_bytes":5927,"created_at":"…","download_url":"/api/books/26/outputs/73/download"},
            {"id":72,"format":"pdf","size_bytes":38954,"created_at":"…","download_url":"/api/books/26/outputs/72/download"}]}
```

Failure — `status` is `"error"`, `outputs` is empty, and `error` is the same
text the factory page shows, up to three newline-separated parts:
plain explanation + what to do, optional `Near: “…”` (text from the manuscript
next to the fault, so the author can search for it in Word), optional
`Technical detail: …` (first meaningful line of pipeline stderr):

```json
{"book_id":27,"project_id":14,"title":"CLI fail test","author":"Me","source_filename":"bad.docx",
 "status":"error",
 "error":"We couldn't read this Word file. Re-save it as .docx from Word (File → Save As, Word Document) and try again. If it came from Pages or a converter, open and re-save it in Word or LibreOffice first.\nTechnical detail: pandoc typst: exit status 63",
 "updated_at":"2026-09-19T12:43:02Z","outputs":[]}
```

Statuses you will see: `uploaded` → `converting` → `ready` | `error`.

## 6. Download

Either by format (latest output of that format) or by the `download_url`
from the status JSON (a specific output row). Both set `Content-Disposition`
with a dated filename.

```sh
curl -s -H "$H" -o book.pdf  $B/api/books/$BOOK/download/pdf     # application/pdf
curl -s -H "$H" -o book.epub $B/api/books/$BOOK/download/epub    # application/epub+zip
curl -s -H "$H" -o out.pdf   $B/api/books/$BOOK/outputs/72/download
```

## Whole thing, one screen

```sh
B=https://jdbbs.exe.xyz; P=14; H="Authorization: Bearer $TOKEN"
BOOK=$(curl -s -H "$H" -F file=@ms.docx -F title='My Book' -F author='Me' -F project_id=$P $B/api/books/upload | jq .id)
curl -s -H "$H" -X POST $B/api/projects/$P/preflight -H 'Content-Type: application/json' -d "{\"book_id\":$BOOK}" | jq .summary
curl -s -H "$H" -X POST $B/api/books/$BOOK/convert -H 'Content-Type: application/json' -d '{"format":"both"}'
until curl -s -H "$H" $B/api/books/$BOOK | jq -e '.status!="converting" and .status!="uploaded"' >/dev/null; do sleep 3; done
curl -s -H "$H" $B/api/books/$BOOK | jq -r '.error // "ok"'
curl -s -H "$H" -o book.pdf $B/api/books/$BOOK/download/pdf; curl -s -H "$H" -o book.epub $B/api/books/$BOOK/download/epub
```

Or: `python3 scripts/factory-cli.py --base $B --project $P --token $TOKEN build ms.docx --out ./out`
and `… inspect ms.docx`.

## Also callable (not needed for a build)

- `GET/PUT /api/projects/{id}/transmittal` — read/fill the transmittal as JSON,
  `{"status":"final","data":{…}}`; a final one is applied to the spec at the
  next convert.
- `GET /api/projects/{id}/word-template` — the styled .docx template
  (`409 {"error":"fill in the transmittal and mark it final first"}` until the
  transmittal is final).
- `GET /api/projects/{id}/books` — every book on the project (ids, statuses).
- `GET /api/books/{id}/outputs` — all output rows for a book.

## Not available to a token caller (ask the studio)

- Minting or rotating the token (`POST /api/projects/{id}/auth` is admin-only).
- Editing the **book spec** directly (`/book-spec` routes are admin) — the
  transmittal is the customer-side lever.
- Buying credits; deleting books.
- Per-finding preflight detail as JSON (HTML report only).
