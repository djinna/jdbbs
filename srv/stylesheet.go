package srv

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "embed"
)

//go:embed stylesheet_seed.json
var stylesheetSeedJSON []byte

// Write allowlist: exe.dev email -> display name.
var stylesheetEditors = map[string]string{
	"j@djinna.com":           "JD (Publisher)",
	"editor@protocolized.io": "James Langdon (PI Editor)",
}

type ssItem struct {
	ID           int64   `json:"id"`
	SectionOrd   int     `json:"section_ord"`
	Section      string  `json:"section"`
	ItemOrd      int     `json:"item_ord"`
	Kind         string  `json:"kind"`
	Col1         string  `json:"col1"`
	Col2         string  `json:"col2"`
	Col3         string  `json:"col3"`
	Body         string  `json:"body"`
	Status       string  `json:"status"`
	AuthorFacing bool    `json:"author_facing"`
	StatusBy     string  `json:"status_by"`
	StatusAt     string  `json:"status_at"`
	UpdatedAt    string  `json:"updated_at"`
	PendingEdit  *ssEdit `json:"pending_edit"`
}

type ssEdit struct {
	ID         int64  `json:"id"`
	ItemID     int64  `json:"item_id"`
	Col1       string `json:"col1"`
	Col2       string `json:"col2"`
	Col3       string `json:"col3"`
	Body       string `json:"body"`
	Note       string `json:"note"`
	ProposedBy string `json:"proposed_by"`
	ProposedAt string `json:"proposed_at"`
	State      string `json:"state"`
}

type ssSeedRow struct {
	SectionOrd int    `json:"section_ord"`
	Section    string `json:"section"`
	ItemOrd    int    `json:"item_ord"`
	Kind       string `json:"kind"`
	Col1       string `json:"col1"`
	Col2       string `json:"col2"`
	Col3       string `json:"col3"`
	Body       string `json:"body"`
}

// registerStylesheetRoutes wires the stylesheet feature. Called from routes()
// BEFORE the SPA catch-all so these explicit paths win.
func (s *Server) registerStylesheetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stylesheet/items", s.handleSSItems)
	mux.HandleFunc("PATCH /api/stylesheet/items/{id}", s.handleSSPatchItem)
	mux.HandleFunc("POST /api/stylesheet/items", s.handleSSCreateItem)
	mux.HandleFunc("POST /api/stylesheet/items/{id}/edits", s.handleSSCreateEdit)
	mux.HandleFunc("POST /api/stylesheet/edits/{id}/accept", s.handleSSAcceptEdit)
	mux.HandleFunc("POST /api/stylesheet/edits/{id}/reject", s.handleSSRejectEdit)

	// Pages (static shells)
	mux.HandleFunc("GET /stylesheet/{$}", s.ssServeStatic("stylesheet/index.html", "text/html; charset=utf-8"))
	mux.HandleFunc("GET /stylesheet/authors", s.ssServeStatic("stylesheet/authors.html", "text/html; charset=utf-8"))
	mux.HandleFunc("GET /stylesheet/app.js", s.ssServeStatic("stylesheet/app.js", "application/javascript; charset=utf-8"))
	mux.HandleFunc("GET /stylesheet/style.css", s.ssServeStatic("stylesheet/style.css", "text/css; charset=utf-8"))

	// Exports (regenerated from DB — canonical)
	mux.HandleFunc("GET /stylesheet/export.html", s.handleSSExportHTML(false))
	mux.HandleFunc("GET /stylesheet/authors.html", s.handleSSExportHTML(true))
	mux.HandleFunc("GET /stylesheet/export.md", s.handleSSExportMD(false))
	mux.HandleFunc("GET /stylesheet/authors.md", s.handleSSExportMD(true))
}

func (s *Server) ssServeStatic(path, ctype string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFS.ReadFile("static/" + path)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", ctype)
		_, _ = w.Write(data)
	}
}

// ssIdentity returns (email, displayName, canWrite) from exe.dev headers.
func ssIdentity(r *http.Request) (string, string, bool) {
	email := strings.ToLower(strings.TrimSpace(r.Header.Get("X-ExeDev-Email")))
	if name, ok := stylesheetEditors[email]; ok {
		return email, name, true
	}
	if email != "" {
		return email, email, false
	}
	return "", "", false
}

func (s *Server) ssEnsureSeeded(ctx context.Context) error {
	var n int
	if err := s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM stylesheet_items").Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	var rows []ssSeedRow
	if err := json.Unmarshal(stylesheetSeedJSON, &rows); err != nil {
		return fmt.Errorf("parse seed: %w", err)
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, rr := range rows {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO stylesheet_items (section_ord, section, item_ord, kind, col1, col2, col3, body, status)
			 VALUES (?,?,?,?,?,?,?,?, 'proposed')`,
			rr.SectionOrd, rr.Section, rr.ItemOrd, rr.Kind, rr.Col1, rr.Col2, rr.Col3, rr.Body)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Server) ssLoadItems(ctx context.Context) ([]ssItem, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, section_ord, section, item_ord, kind, col1, col2, col3, body,
		        status, author_facing, status_by, status_at, updated_at
		 FROM stylesheet_items ORDER BY section_ord, item_ord, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ssItem
	byID := map[int64]int{}
	for rows.Next() {
		var it ssItem
		var af int
		var statusAt, updatedAt sql.NullTime
		if err := rows.Scan(&it.ID, &it.SectionOrd, &it.Section, &it.ItemOrd, &it.Kind,
			&it.Col1, &it.Col2, &it.Col3, &it.Body, &it.Status, &af,
			&it.StatusBy, &statusAt, &updatedAt); err != nil {
			return nil, err
		}
		it.AuthorFacing = af != 0
		if statusAt.Valid {
			it.StatusAt = statusAt.Time.Format(time.RFC3339)
		}
		if updatedAt.Valid {
			it.UpdatedAt = updatedAt.Time.Format(time.RFC3339)
		}
		byID[it.ID] = len(items)
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Attach pending edits (latest pending per item).
	erows, err := s.DB.QueryContext(ctx,
		`SELECT id, item_id, col1, col2, col3, body, note, proposed_by, proposed_at, state
		 FROM stylesheet_edits WHERE state='pending' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer erows.Close()
	for erows.Next() {
		var e ssEdit
		var pat sql.NullTime
		if err := erows.Scan(&e.ID, &e.ItemID, &e.Col1, &e.Col2, &e.Col3, &e.Body,
			&e.Note, &e.ProposedBy, &pat, &e.State); err != nil {
			return nil, err
		}
		if pat.Valid {
			e.ProposedAt = pat.Time.Format(time.RFC3339)
		}
		if idx, ok := byID[e.ItemID]; ok {
			ec := e
			items[idx].PendingEdit = &ec
		}
	}
	return items, erows.Err()
}

func (s *Server) ssLoadItem(ctx context.Context, id int64) (*ssItem, error) {
	items, err := s.ssLoadItems(ctx)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == id {
			return &items[i], nil
		}
	}
	return nil, sql.ErrNoRows
}

func (s *Server) handleSSItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := s.ssEnsureSeeded(ctx); err != nil {
		jsonErr(w, "seed failed: "+err.Error(), 500)
		return
	}
	items, err := s.ssLoadItems(ctx)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	email, name, canWrite := ssIdentity(r)
	jsonOK(w, map[string]any{
		"items": items,
		"me":    map[string]any{"email": email, "name": name, "can_write": canWrite},
	})
}

func (s *Server) handleSSPatchItem(w http.ResponseWriter, r *http.Request) {
	_, name, canWrite := ssIdentity(r)
	if !canWrite {
		jsonErr(w, "editor sign-in required", http.StatusForbidden)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	var body struct {
		Status       *string `json:"status"`
		AuthorFacing *bool   `json:"author_facing"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad body", 400)
		return
	}
	ctx := r.Context()
	if body.Status != nil {
		st := *body.Status
		if st != "proposed" && st != "accepted" && st != "rejected" {
			jsonErr(w, "bad status", 400)
			return
		}
		if _, err := s.DB.ExecContext(ctx,
			`UPDATE stylesheet_items SET status=?, status_by=?, status_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
			st, name, id); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
	}
	if body.AuthorFacing != nil {
		af := 0
		if *body.AuthorFacing {
			af = 1
		}
		if _, err := s.DB.ExecContext(ctx,
			`UPDATE stylesheet_items SET author_facing=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
			af, id); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
	}
	it, err := s.ssLoadItem(ctx, id)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	jsonOK(w, it)
}

func (s *Server) handleSSCreateItem(w http.ResponseWriter, r *http.Request) {
	_, _, canWrite := ssIdentity(r)
	if !canWrite {
		jsonErr(w, "editor sign-in required", http.StatusForbidden)
		return
	}
	var body ssSeedRow
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad body", 400)
		return
	}
	if body.Kind != "rule" && body.Kind != "prose" {
		body.Kind = "prose"
	}
	ctx := r.Context()
	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO stylesheet_items (section_ord, section, item_ord, kind, col1, col2, col3, body, status)
		 VALUES (?,?,?,?,?,?,?,?, 'accepted')`,
		body.SectionOrd, body.Section, body.ItemOrd, body.Kind, body.Col1, body.Col2, body.Col3, body.Body)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	id, _ := res.LastInsertId()
	it, err := s.ssLoadItem(ctx, id)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, it)
}

func (s *Server) handleSSCreateEdit(w http.ResponseWriter, r *http.Request) {
	_, name, canWrite := ssIdentity(r)
	if !canWrite {
		jsonErr(w, "editor sign-in required", http.StatusForbidden)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	var body struct {
		Col1 string `json:"col1"`
		Col2 string `json:"col2"`
		Col3 string `json:"col3"`
		Body string `json:"body"`
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad body", 400)
		return
	}
	ctx := r.Context()
	// Supersede any existing pending edit for this item.
	if _, err := s.DB.ExecContext(ctx,
		`UPDATE stylesheet_edits SET state='rejected' WHERE item_id=? AND state='pending'`, id); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if _, err := s.DB.ExecContext(ctx,
		`INSERT INTO stylesheet_edits (item_id, col1, col2, col3, body, note, proposed_by, state)
		 VALUES (?,?,?,?,?,?,?, 'pending')`,
		id, body.Col1, body.Col2, body.Col3, body.Body, body.Note, name); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	it, err := s.ssLoadItem(ctx, id)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	jsonOK(w, it)
}

func (s *Server) handleSSAcceptEdit(w http.ResponseWriter, r *http.Request) {
	s.ssResolveEdit(w, r, true)
}

func (s *Server) handleSSRejectEdit(w http.ResponseWriter, r *http.Request) {
	s.ssResolveEdit(w, r, false)
}

func (s *Server) ssResolveEdit(w http.ResponseWriter, r *http.Request, accept bool) {
	_, name, canWrite := ssIdentity(r)
	if !canWrite {
		jsonErr(w, "editor sign-in required", http.StatusForbidden)
		return
	}
	editID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	ctx := r.Context()
	var e ssEdit
	err = s.DB.QueryRowContext(ctx,
		`SELECT id, item_id, col1, col2, col3, body FROM stylesheet_edits WHERE id=? AND state='pending'`,
		editID).Scan(&e.ID, &e.ItemID, &e.Col1, &e.Col2, &e.Col3, &e.Body)
	if err != nil {
		jsonErr(w, "pending edit not found", 404)
		return
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer tx.Rollback()
	if accept {
		if _, err := tx.ExecContext(ctx,
			`UPDATE stylesheet_items SET col1=?, col2=?, col3=?, body=?, status_by=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
			e.Col1, e.Col2, e.Col3, e.Body, name, e.ItemID); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		if _, err := tx.ExecContext(ctx, `UPDATE stylesheet_edits SET state='accepted' WHERE id=?`, editID); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE stylesheet_edits SET state='rejected' WHERE id=?`, editID); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	it, err := s.ssLoadItem(ctx, e.ItemID)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	jsonOK(w, it)
}

// ---- Exports ----

// filtered returns accepted items; if authorsOnly, also require author_facing.
func ssFilter(items []ssItem, authorsOnly bool) []ssItem {
	var out []ssItem
	for _, it := range items {
		if it.Status != "accepted" {
			continue
		}
		if authorsOnly && !it.AuthorFacing {
			continue
		}
		out = append(out, it)
	}
	return out
}

type ssSection struct {
	Ord   int
	Title string
	Items []ssItem
}

func ssGroup(items []ssItem) []ssSection {
	order := []int{}
	m := map[int]*ssSection{}
	for _, it := range items {
		s, ok := m[it.SectionOrd]
		if !ok {
			s = &ssSection{Ord: it.SectionOrd, Title: it.Section}
			m[it.SectionOrd] = s
			order = append(order, it.SectionOrd)
		}
		s.Items = append(s.Items, it)
	}
	sort.Ints(order)
	out := make([]ssSection, 0, len(order))
	for _, o := range order {
		out = append(out, *m[o])
	}
	return out
}

func (s *Server) handleSSExportMD(authorsOnly bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = s.ssEnsureSeeded(ctx)
		items, err := s.ssLoadItems(ctx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		secs := ssGroup(ssFilter(items, authorsOnly))
		var b strings.Builder
		title := "Protocol Institute — Editorial Stylesheet"
		if authorsOnly {
			title = "Protocol Institute — Author Stylesheet (essentials)"
		}
		fmt.Fprintf(&b, "# %s\n\n", title)
		fmt.Fprintf(&b, "_Generated %s from the working stylesheet database._\n\n", time.Now().Format("2006-01-02"))
		for _, sec := range secs {
			fmt.Fprintf(&b, "## %s\n\n", sec.Title)
			// rules grouped into a table if any
			var rules, prose []ssItem
			for _, it := range sec.Items {
				if it.Kind == "rule" {
					rules = append(rules, it)
				} else {
					prose = append(prose, it)
				}
			}
			for _, p := range prose {
				b.WriteString(p.Body)
				b.WriteString("\n\n")
			}
			if len(rules) > 0 {
				b.WriteString("| Item | Rule | Example |\n|---|---|---|\n")
				for _, rl := range rules {
					fmt.Fprintf(&b, "| %s | %s | %s |\n",
						mdCell(rl.Col1), mdCell(rl.Col2), mdCell(rl.Col3))
				}
				b.WriteString("\n")
			}
		}
		fname := "pi-stylesheet.md"
		if authorsOnly {
			fname = "pi-author-stylesheet.md"
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fname+"\"")
		_, _ = w.Write([]byte(b.String()))
	}
}

func mdCell(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "|", "\\|")
}

func (s *Server) handleSSExportHTML(authorsOnly bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = s.ssEnsureSeeded(ctx)
		items, err := s.ssLoadItems(ctx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		secs := ssGroup(ssFilter(items, authorsOnly))
		title := "Protocol Institute — Editorial Stylesheet"
		fname := "pi-stylesheet.html"
		if authorsOnly {
			title = "Protocol Institute — Author Stylesheet (essentials)"
			fname = "pi-author-stylesheet.html"
		}
		var b strings.Builder
		b.WriteString(ssExportHead(title))
		fmt.Fprintf(&b, "<h1>%s</h1>\n", html.EscapeString(title))
		fmt.Fprintf(&b, "<p class=\"gen\">Generated %s from the working stylesheet database.</p>\n", time.Now().Format("January 2, 2006"))
		for _, sec := range secs {
			fmt.Fprintf(&b, "<h2>%s</h2>\n", html.EscapeString(sec.Title))
			var rules, prose []ssItem
			for _, it := range sec.Items {
				if it.Kind == "rule" {
					rules = append(rules, it)
				} else {
					prose = append(prose, it)
				}
			}
			for _, p := range prose {
				b.WriteString("<div class=\"prose\">")
				b.WriteString(mdToHTML(p.Body))
				b.WriteString("</div>\n")
			}
			if len(rules) > 0 {
				b.WriteString("<table><thead><tr><th>Item</th><th>Rule</th><th>Example</th></tr></thead><tbody>\n")
				for _, rl := range rules {
					fmt.Fprintf(&b, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n",
						inlineMD(rl.Col1), inlineMD(rl.Col2), inlineMD(rl.Col3))
				}
				b.WriteString("</tbody></table>\n")
			}
		}
		b.WriteString("</main></body></html>\n")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.URL.Query().Get("download") != "" {
			w.Header().Set("Content-Disposition", "attachment; filename=\""+fname+"\"")
		}
		_, _ = w.Write([]byte(b.String()))
	}
}

func ssExportHead(title string) string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + html.EscapeString(title) + `</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;600&family=Literata:opsz,wght@7..72,400;7..72,600;7..72,700&display=swap" rel="stylesheet">
<style>
:root{--fg:#1a1a1a;--muted:#666;--line:#e2e2e2;--accent:#7a4de0;--bg:#fff;--zebra:#faf9fc}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font-family:"Literata",Georgia,serif;line-height:1.55}
main{max-width:820px;margin:0 auto;padding:3rem 1.5rem 6rem}
h1{font-size:2rem;line-height:1.2;margin:0 0 .25rem}
h2{font-family:"IBM Plex Sans",system-ui,sans-serif;font-size:1.15rem;margin:2.4rem 0 .8rem;padding-bottom:.3rem;border-bottom:2px solid var(--accent)}
.gen{color:var(--muted);font-family:"IBM Plex Sans",sans-serif;font-size:.85rem;margin:0 0 1.5rem}
table{border-collapse:collapse;width:100%;margin:.5rem 0 1rem;font-size:.95rem}
th{text-align:left;font-family:"IBM Plex Sans",sans-serif;font-size:.8rem;text-transform:uppercase;letter-spacing:.04em;color:var(--muted);border-bottom:2px solid var(--line);padding:.4rem .6rem}
td{border-bottom:1px solid var(--line);padding:.5rem .6rem;vertical-align:top}
tbody tr:nth-child(even){background:var(--zebra)}
td:first-child{font-weight:600;white-space:nowrap}
code{font-family:"IBM Plex Mono",ui-monospace,monospace;background:#f3f0fa;padding:.05em .35em;border-radius:4px;font-size:.9em}
.prose{margin:.4rem 0 1rem}
.prose ul{margin:.3rem 0 .3rem 1.2rem;padding:0}
.prose blockquote{margin:.6rem 0;padding:.4rem .9rem;border-left:3px solid var(--accent);background:#faf8ff;color:#333}
@media print{h2{break-after:avoid}tr{break-inside:avoid}}
</style></head><body><main>
`
}

// inlineMD renders a tiny subset of inline markdown to safe HTML.
func inlineMD(s string) string {
	out := html.EscapeString(s)
	out = replacePairs(out, "`", "<code>", "</code>")
	out = replacePairs(out, "**", "<strong>", "</strong>")
	out = replacePairs(out, "*", "<em>", "</em>")
	return out
}

// replacePairs turns paired delimiters into open/close tags, left to right.
func replacePairs(s, delim, open, close string) string {
	var b strings.Builder
	openNext := true
	for {
		i := strings.Index(s, delim)
		if i < 0 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:i])
		if openNext {
			b.WriteString(open)
		} else {
			b.WriteString(close)
		}
		openNext = !openNext
		s = s[i+len(delim):]
	}
	return b.String()
}

// mdToHTML renders block-level markdown (paragraphs, bullets, blockquotes) with
// inline formatting. Deliberately small; source is trusted (editor-authored).
func mdToHTML(md string) string {
	lines := strings.Split(md, "\n")
	var b strings.Builder
	inList := false
	inQuote := false
	flushList := func() {
		if inList {
			b.WriteString("</ul>")
			inList = false
		}
	}
	flushQuote := func() {
		if inQuote {
			b.WriteString("</blockquote>")
			inQuote = false
		}
	}
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		switch {
		case t == "":
			flushList()
			flushQuote()
		case strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* "):
			flushQuote()
			if !inList {
				b.WriteString("<ul>")
				inList = true
			}
			b.WriteString("<li>" + inlineMD(strings.TrimSpace(t[2:])) + "</li>")
		case strings.HasPrefix(t, "> "):
			flushList()
			if !inQuote {
				b.WriteString("<blockquote>")
				inQuote = true
			}
			b.WriteString(inlineMD(strings.TrimSpace(t[2:])) + " ")
		case strings.HasPrefix(t, "### "):
			flushList()
			flushQuote()
			b.WriteString("<h3>" + inlineMD(strings.TrimSpace(t[4:])) + "</h3>")
		default:
			flushList()
			flushQuote()
			b.WriteString("<p>" + inlineMD(t) + "</p>")
		}
	}
	flushList()
	flushQuote()
	return b.String()
}
