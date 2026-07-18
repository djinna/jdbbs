# PI Stylesheet feature — shared build spec

Host app: prodcal (Go, SQLite, sqlc), repo /home/exedev/prodcal, module `srv.exe.dev`.
Served on default port 8000 via exe.dev proxy. exe.dev injects `X-ExeDev-Email`
and `X-ExeDev-UserID` headers for authenticated users.

Original page (DO NOT TOUCH): /home/exedev/prodcal/srv/static/zoothesia/*
Canonical source stylesheet (markdown): /home/exedev/book-production/editorial/Zoothesia/STYLESHEET.md

## Concept
A working PI editorial stylesheet, seeded from Zoothesia. Team can accept/reject/edit
items and tag some as "author-facing" to derive a tighter authors' sheet.
DB is canonical going forward; we can regenerate the shared HTML/MD.

## Write allowlist (identity)
- j@djinna.com          -> display "JD (Publisher)"
- editor@protocolized.io -> display "James Langdon (PI Editor)"
Reads: any authenticated user (header present). Writes: allowlist only.

## Item model (JSON shape returned by API)
Item {
  id: int
  section_ord: int
  section: string            // e.g. "1. House Style Basics"
  item_ord: int
  kind: "rule" | "prose"
  col1: string               // rule: "Item" column
  col2: string               // rule: "Rule" column
  col3: string               // rule: "Example" column
  body: string               // prose: markdown text
  status: "proposed" | "accepted" | "rejected"
  author_facing: bool
  status_by: string          // display name
  status_at: string          // RFC3339 or ""
  updated_at: string
  pending_edit: Edit | null  // the current pending edit if any
}
Edit {
  id: int
  item_id: int
  col1: string col2: string col3: string body: string  // proposed new values
  note: string
  proposed_by: string        // display name
  proposed_at: string
  state: "pending" | "accepted" | "rejected"
}

## HTTP API (all JSON unless noted)
GET  /api/stylesheet/items
     -> { items: [Item...], me: { email, name, can_write: bool } }
PATCH /api/stylesheet/items/{id}      body {status?, author_facing?}  -> Item   (write)
POST  /api/stylesheet/items           body {section_ord,section,kind,col1,col2,col3,body} -> Item (write)
POST  /api/stylesheet/items/{id}/edits body {col1,col2,col3,body,note} -> Item (with pending_edit) (write)
POST  /api/stylesheet/edits/{id}/accept  -> Item (applies edit to item) (write)
POST  /api/stylesheet/edits/{id}/reject  -> Item (write)

## Page routes
GET /stylesheet/            -> editor SPA (static/stylesheet/index.html)
GET /stylesheet/authors     -> authors view (static/stylesheet/authors.html) [read-only render]
GET /stylesheet/app.js, /stylesheet/style.css -> static assets under static/stylesheet/
GET /stylesheet/export.html -> full working stylesheet, standalone styled HTML (accepted items), Content-Disposition attachment
GET /stylesheet/export.md   -> regenerated canonical markdown (accepted items)
GET /stylesheet/authors.html-> authors sheet standalone HTML (accepted AND author_facing)
GET /stylesheet/authors.md  -> authors sheet markdown

Authors' sheet filter: status == "accepted" AND author_facing == true.
Working export: status == "accepted".

## Notes
- Static files live in srv/static/stylesheet/ and are embedded via existing //go:embed static/* in srv/server.go.
- Routes registered EXPLICITLY in srv/server.go BEFORE the catch-all `GET /` so they win.
- Diff display is a frontend concern (word-level highlight of pending_edit vs current).
