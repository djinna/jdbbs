# Handler Summary: books.go, bookspecs.go, epub.go

## Auth Pattern

`requireExeDevAdminAPI` checks for `X-ExeDev-UserID` header. Returns 401 JSON error if missing.
`projectIDFromPath` parses `{id}` path param as int64.

---

## books.go

### 1. `handleListBooks`
- **Route:** `GET /api/books`
- **Auth:** requireExeDevAdminAPI (X-ExeDev-UserID header)
- **Request:** No params
- **Response:** 200 JSON array of `dbgen.ListBooksRow` (excludes blob data). Empty array `[]` if none.
- **Errors:** 500 on DB error

### 2. `handleUploadBook`
- **Route:** `POST /api/books/upload`
- **Auth:** requireExeDevAdminAPI
- **Request:** Multipart form (max 50MB)
  - `file` (required) — the .docx file
  - `title` (required, string)
  - `author` (required, string)
  - `series` (optional, string)
  - `project_id` (optional, string parseable as int64)
- **Response:** 201 JSON `{"id", "title", "author", "status"}`
- **Errors:** 400 if file missing, title/author empty, or bad form; 500 on read/DB error

### 3. `handleConvertBook`
- **Route:** `POST /api/books/{id}/convert`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (book ID as int64)
- **Response:** 200 JSON `{"status": "converting"}`
- **Behavior:** Sets book status to "converting", then runs conversion **asynchronously** (goroutine):
  1. Writes source .docx to temp dir
  2. Runs pandoc docx→typst with custom Lua filter + media extraction
  3. Replaces generated Typst header with real metadata + spec-driven config overrides
  4. Applies text fixups (literal @mentions, image clamping, spacing fixes, poem markers)
  5. Compiles Typst→PDF with `typst compile`
  6. Stores PDF as a `book_outputs` history row AND updates `books.pdf_data`
  7. On failure: sets status to "error" with error message via `failConversion`
- **Errors:** 400 bad id or no source file; 404 not found

### 4. `handleDownloadBook`
- **Route:** `GET /api/books/{id}/download/{format}`
- **Auth:** **NONE** (public endpoint)
- **Request:** Path params `{id}` (book ID), `{format}` ("pdf" or "epub")
- **Response:**
  - PDF: `Content-Type: application/pdf`, attachment disposition with sanitized title filename
  - EPUB: `Content-Type: application/epub+zip`, attachment disposition with sanitized title filename
- **Errors:** 400 bad id or unknown format; 404 not found or not generated yet

### 5. `handleDeleteBook`
- **Route:** `DELETE /api/books/{id}`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (book ID)
- **Response:** 200 JSON `{"ok": "true"}`
- **Errors:** 400 bad id; 500 on DB error

### 6. `handleLinkBookProject`
- **Route:** `PUT /api/books/{id}/project`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (book ID). JSON body: `{"project_id": <int64|null>}`
  - `null` unlinks the book from any project
- **Response:** 200 JSON `{"ok": "true"}`
- **Errors:** 400 bad id or bad JSON; 500 on DB error

---

## bookspecs.go

### 7. `handleGetBookSpec`
- **Route:** `GET /api/projects/{id}/book-spec`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (project ID)
- **Behavior:** Auto-creates spec with defaults if none exists (upsert)
- **Response:** 200 JSON:
  ```json
  {
    "id": <int64>,
    "project_id": <int64>,
    "data": <spec JSON object>,
    "has_cover": <bool>,
    "created_at": <timestamp>,
    "updated_at": <timestamp>
  }
  ```
- **Errors:** 400 bad id; 500 on DB error

### 8. `handleUpdateBookSpec`
- **Route:** `PUT /api/projects/{id}/book-spec`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (project ID). JSON body: `{"data": <JSON object>}`
  - If data is empty/null, defaults are used
- **Response:** 200 JSON `{"ok": true, "updated_at": <timestamp>}`
- **Errors:** 400 bad id or bad JSON; 500 on DB error

### 9. `handlePullTransmittalToSpec`
- **Route:** `POST /api/projects/{id}/book-spec/pull-transmittal`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (project ID). No body.
- **Behavior:** Reads the project's transmittal, maps its fields into the spec:
  - `tx.book` → `spec.metadata` (title, subtitle, author, series, publisher, ISBNs)
  - `tx.page_iv` → `spec.metadata` (copyright_year, copyright_holder, credit_lines)
  - `tx.editing` → `spec.typesetting` (developmental_instructions, copyeditor_instructions)
  - `tx.design` → `spec.typesetting` (trim_guidance, trim_size, est_pages, ppi, spine_width, complexity, outside_designer, reuse_previous, design_notes) + `spec.page` (trim, width_in, height_in via parseTrim)
  - `tx.checklist` → `spec.front_matter` (half_title, series_title, title_page, copyright_page, dedication, epigraph, toc, foreword, preface, acknowledgments, introduction)
  - `tx.backmatter` → `spec.back_matter` (notes, appendix, glossary, bibliography, index)
  - `tx.custom_styles` → `spec.custom_styles` (name, word_style, type, description, preset, typst code — with auto-generated defaults)
- **Response:** 200 JSON `{"ok": true, "data": <merged spec JSON>}`
- **Errors:** 400 bad id; 404 no transmittal; 500 on DB/parse error

### 10. `handleListFonts`
- **Route:** `GET /api/fonts`
- **Auth:** requireExeDevAdminAPI
- **Request:** No params
- **Behavior:** Runs `typst fonts --font-path <fontsDir>`, categorizes each font family as serif/sans-serif/monospace/other. Falls back to 3 hardcoded fonts on error.
- **Response:** 200 JSON array of `{"family": "...", "category": "..."}`

### 11. `handleGenerateConfig`
- **Route:** `POST /api/projects/{id}/book-spec/generate-config`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (project ID). No body.
- **Behavior:** Reads the spec, converts it to Typst config override code via `specToTypstConfig()`. Generates:
  - Page dimensions and margins
  - Typography settings (fonts, sizes, leading, indent)
  - Heading sizes/weights/alignment
  - Element styles (section break, blockquote, poem/code/footnote sizes)
  - Running head config
  - Custom style Typst function definitions
- **Response:** 200 JSON `{"config": "<typst config string>"}`
- **Errors:** 400 bad id; 404 no spec found

### 12. `handleGenerateWordTemplate`
- **Route:** `POST /api/projects/{id}/book-spec/word-template`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (project ID). No body.
- **Behavior:** Validates no duplicate custom style names, then runs Python script `scripts/generate-word-template.py` with spec JSON on stdin.
- **Response:** Binary .docx file download
  - `Content-Type: application/vnd.openxmlformats-officedocument.wordprocessingml.document`
  - `Content-Disposition: attachment; filename="<project-name>-template.docx"`
- **Errors:** 400 bad id or duplicate custom style names; 404 no spec; 500 on script failure

### 13. `handleUploadCover`
- **Route:** `POST /api/projects/{id}/book-spec/cover`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (project ID). Multipart form (max 10MB):
  - `cover` (required) — JPEG or PNG image file
- **Behavior:** Auto-creates spec row if missing, then stores cover data + content type in book_specs table.
- **Response:** 200 JSON `{"ok": true, "size": <bytes>, "type": "image/jpeg|image/png"}`
- **Errors:** 400 bad id, file too large, missing file, or wrong content type; 500 on read/DB error

### 14. `handleGetCover`
- **Route:** `GET /api/projects/{id}/book-spec/cover`
- **Auth:** **NONE** (public endpoint)
- **Request:** Path param `{id}` (project ID)
- **Response:** Binary image with proper Content-Type, Content-Length, and `Cache-Control: max-age=3600`
- **Errors:** 404 (http.NotFound) if no cover exists; 400 bad id

---

## epub.go

### 15. `handleGenerateEPUB`
- **Route:** `POST /api/books/{id}/generate-epub`
- **Auth:** requireExeDevAdminAPI
- **Request:** Path param `{id}` (book ID)
- **Response:** 200 JSON `{"status": "generating_epub"}`
- **Behavior:** Runs EPUB generation **asynchronously** (goroutine):
  1. Writes source .docx to temp dir
  2. If book is linked to a project, loads spec and extracts EPUB settings:
     - From `spec.metadata`: title, author
     - From `spec.epub`: language, subject, description, toc_depth, chapter_break, section_break, body_font_size, embed_fonts, custom_css
  3. If project has a cover image, writes it to temp dir for pandoc
  4. Builds custom CSS from spec (body font size, chapter break styling, section break ornaments, user custom CSS)
  5. Runs pandoc docx→epub3 with metadata, TOC, cover image, and CSS
  6. Stores EPUB as a `book_outputs` history row AND updates `books.epub_data`
  7. Errors are logged but NOT stored in book status (unlike PDF conversion)
- **Errors:** 400 bad id or no source file; 404 not found

---

## Default Spec Shape

The `defaultSpecData()` function defines the full spec schema with these sections:
- `metadata` — title, subtitle, author, series, publisher, ISBNs, copyright info
- `page` — trim preset, width/height in inches, margins (top/bottom/inside/outside)
- `typography` — body/heading/code/tweet fonts, base size, leading, paragraph indent, justify, hyphenate
- `headings` — h1/h2/h3 sizes and weights, alignment
- `elements` — section break style, blockquote style, poem/code/footnote sizes
- `running_heads` — enabled flag, size, verso/recto content
- `front_matter` — boolean flags for half title, series title, title page, copyright, dedication, epigraph, toc, foreword, preface, acknowledgments, introduction
- `back_matter` — boolean flags for notes, appendix, glossary, bibliography, index
- `typesetting` — developmental/copyeditor instructions, trim guidance, design notes, complexity, etc.
- `custom_styles` — array of custom style definitions
- `epub` — toc_depth, landmarks, chapter/section break, font size, embed fonts, custom CSS, language, subject, description
