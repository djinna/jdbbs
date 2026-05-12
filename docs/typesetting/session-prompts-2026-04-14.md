# Session Prompts — ProdCal Improvements (2026-04-14)

Generated from deep technical analysis. Each prompt is self-contained for a new session.

---

## Prompt 1: Security Hardening

```
## Project: ProdCal Security Hardening

ProdCal is a Go web app at ~/prodcal/ — book production management tool running as systemd service `srv` on port 8000. SQLite DB at ~/db.sqlite3.

A security audit found 4 HIGH-severity issues. Fix all of them in this session.

### Issue 1: SHA-256 password hashing → bcrypt
- `srv/server.go` `hashToken()` uses `crypto/sha256` for auth tokens
- `srv/admin.go` `handleAdminCreateClient()` uses SHA-256 for client passwords
- Replace with `golang.org/x/crypto/bcrypt`. Use cost 12.
- Must migrate existing hashed passwords (3 clients, 3 auth tokens in DB).
- Update `checkAuth()` in server.go and `checkClientAuth()` in client.go to use bcrypt.CompareHashAndPassword.
- Update all tests in auth_test.go, client_test.go, admin_test.go.

### Issue 2: Missing auth on book download
- `srv/books.go` `handleDownloadBook()` has NO auth check — anyone with a book ID can download PDFs/EPUBs.
- Add `requireAuth` or at minimum check that the book's project_id matches an authenticated session.
- Same for `handleGetCover` in bookspecs.go — cover images are public.

### Issue 3: Missing auth on project list
- `srv/server.go` `handleListProjects()` returns ALL non-archived projects to any unauthenticated request.
- This endpoint should require exe.dev admin auth (X-ExeDev-UserID header).
- The client portal has its own filtered list via `handleClientProjects()` which is already auth-gated.

### Issue 4: API key exposed in git history
- ~/prodcal/SESSION-SUMMARY.txt contains the AgentMail API key (`am_us_8de7...`) and client passwords in plaintext.
- Rotate the AgentMail key: update ~/.env (AGENTMAIL_API_KEY) with a new key from the AgentMail dashboard, or note that it needs manual rotation.
- Add SESSION-SUMMARY.txt to .gitignore.
- Consider `git filter-repo` to scrub the key from history (ask before doing this — it rewrites history).

### Build & test
- `cd ~/prodcal && go test ./...` must pass
- `make build && sudo systemctl restart srv`
- Verify: `curl -s http://localhost:8000/api/projects` should return 401/403 without auth
- Verify: `curl -s http://localhost:8000/api/books/4/download/pdf` should return 401/403 without auth
```

---

## Prompt 2: Code Quality — Fix Weaknesses

```
## Project: ProdCal Code Quality Improvements

ProdCal is a Go web app at ~/prodcal/. See ~/prodcal/AGENTS.md for basics.

A code review identified structural weaknesses. Address these in priority order — get as far as you can.

### 1. Split server.go (1,059 lines)
Currently a god file containing: route registration, project CRUD, task CRUD, auth handlers, project seeding, project duplication.
- Extract to: `routes.go` (just Handler() with route registration), `projects.go` (project CRUD + seed + duplicate), `tasks.go` (task CRUD), keep auth in server.go.
- Do NOT change any behavior — pure refactor. All tests must pass.

### 2. Eliminate CSS duplication
Design tokens and base styles are duplicated across 4 files:
- `srv/static/style.css` (canonical)
- `srv/static/admin.html` (~400 lines inline CSS)
- `srv/static/landing.html` (inline CSS)
- `srv/static/client.html` (inline CSS)
Extract shared CSS into style.css. Have admin.html, landing.html, and client.html reference `<link rel="stylesheet" href="/static/style.css">` plus only their page-specific overrides inline.

### 3. Unify data access — add sqlc queries for raw-SQL tables
These 7 tables use raw SQL in handlers instead of sqlc-generated code:
- file_log (in filelog.go)
- journal (in journal.go)
- corrections (in corrections.go)
- transmittals (in transmittal.go)
- transmittal_versions (in transmittal.go)
- clients (in admin.go, client.go)
- admin aggregation queries (in admin.go)
Add query files in `db/queries/` and regenerate with `sqlc generate` (config at db/sqlc.yaml). Then update handlers to use the generated code.

### 4. Add graceful shutdown
- server.go `Serve()` should handle SIGINT/SIGTERM
- Use `http.Server.Shutdown(ctx)` with a 10-second timeout
- Close the database connection on shutdown
- Fix main.go to `os.Exit(1)` on error

### Build & test
- `cd ~/prodcal && go test ./...` — all tests must pass after each change
- `make build && sudo systemctl restart srv`
- Verify the admin dashboard still works at http://localhost:8000/admin/
```

---

## Prompt 3: API Reference Documentation

```
## Project: ProdCal API Reference Documentation

ProdCal is a Go web app at ~/prodcal/ with 74 registered HTTP routes. There is currently NO API documentation — endpoints are scattered across session notes.

Generate a comprehensive API reference doc at `~/prodcal/docs/API.md`.

### How to find all routes
Read `~/prodcal/srv/server.go` — the `Handler()` method registers all routes on an `http.ServeMux`. Each route uses Go 1.22+ method-pattern syntax like `"GET /api/projects"` or `"POST /api/books/{id}/convert"`.

Also read every handler file in `~/prodcal/srv/` to understand request/response formats.

### Document structure
For each endpoint, document:
- **Method + Path** (e.g., `POST /api/books/upload`)
- **Auth required** (none / project-level / client-level / admin)
- **Request**: content-type, body format (JSON fields or multipart), path params, query params
- **Response**: status codes, JSON shape (with example), error format
- **Notes**: any side effects, async behavior, etc.

### Group by domain
1. **Health** — /healthz
2. **Projects** — /api/projects/*
3. **Tasks** — /api/projects/{id}/tasks/*
4. **Auth** — /api/projects/{id}/auth/*
5. **Transmittals** — /api/projects/{id}/transmittal/*
6. **File Log** — /api/projects/{id}/file-log/*
7. **Journal** — /api/projects/{id}/journal/*
8. **Books** — /api/books/*
9. **Book Specs** — /api/projects/{id}/book-spec/*
10. **Preflight** — /api/projects/{id}/preflight/*
11. **Corrections** — /api/projects/{id}/corrections/*
12. **Email** — various /api/projects/{id}/*/email endpoints
13. **Clients** — /api/clients/*
14. **Admin** — /api/admin/*
15. **Pages** — HTML page routes (/, /admin/, /{client}/, etc.)

### Also
- Add a "Quick Reference" table at the top: Method | Path | Auth | Description (one line per route)
- Note which endpoints are admin-only vs client-accessible
- Flag any endpoints with known issues (e.g., missing auth on book download)
- Reference EMAIL_SYSTEM.md for the email endpoints (don't duplicate, just link)

### Verify accuracy
After writing the doc, grep server.go for all `mux.HandleFunc` calls and confirm every route is documented.

Commit as: `docs: add comprehensive API reference (74 routes)`
```

---

## Prompt 4: Push to GitHub

```
## Project: Push ProdCal to GitHub

ProdCal is a Go web app at ~/prodcal/ that has never been pushed to GitHub. It's currently a local git repo on an exe.dev VM.

### Pre-push cleanup
1. **Audit .gitignore** — ensure these are ignored:
   - *.sqlite3, *.db, *.sqlite3-shm, *.sqlite3-wal
   - .env (contains AGENTMAIL_API_KEY)
   - prodcal (the binary)
   - SESSION-SUMMARY.txt (contains API keys and passwords)
   - TEAM-UPDATE.txt (internal)
   - seed_data.json (contains real client data)
   - .hermes/ directory

2. **Check for secrets in git history**:
   - `git log --all -p -- SESSION-SUMMARY.txt .env` — if secrets were ever committed, note this.
   - `git log --all -p | grep -i "am_us_\|api_key\|password_hash\|artofgig\|willwrite"` — scan for leaked secrets.
   - If secrets are in history: create a fresh repo from the current working tree (don't rewrite history on a shared repo — just start clean).

3. **Write a proper README.md** replacing the current generic template text. Include:
   - What ProdCal is (book production calendar & manuscript transmittal tool)
   - Tech stack (Go, SQLite, Pandoc, Typst)
   - How to build (`make build`)
   - How to run (`./prodcal -listen :8000`)
   - How to deploy (reference DEPLOY.md)
   - Environment variables needed (AGENTMAIL_API_KEY, AGENTMAIL_INBOX_ID, BASE_URL)
   - Link to API docs if they exist (docs/API.md)
   - License (ask user what license to use — default to MIT)

4. **Create the GitHub repo**:
   - The user's GitHub username is likely `djinnadjinn` or similar — ASK before creating.
   - Instruct the user to create the repo on GitHub (you can't do this from the VM).
   - Set up the remote: `git remote add origin git@github.com:USER/prodcal.git`
   - Push: `git push -u origin main`

5. **If secrets are in history**, the cleanest approach is:
   ```bash
   cd ~/prodcal
   # Create a fresh repo from current state
   rm -rf .git
   git init
   git add .
   git commit -m "Initial commit: ProdCal book production management"
   git remote add origin git@github.com:USER/prodcal.git
   git push -u origin main
   ```
   Ask the user before doing this — they may want to preserve history.

### SSH key
Check if an SSH key exists: `cat ~/.ssh/id_ed25519.pub 2>/dev/null || cat ~/.ssh/id_rsa.pub 2>/dev/null`
If not, the user needs to add the VM's SSH key to their GitHub account. Instruct them to run `ssh exe.dev ssh-key` from their local machine.
```
