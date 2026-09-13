package srv

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// ─── Slug scheme ───
//
// Client slug:  first initial + last name → "mcasey"; collision → "mcasey2".
// Project slug: per-client sequence "book-001", "book-002", … independent of
// the manuscript title, which changes during production and is stored as the
// renamable projects.name. Both are editable in admin; renames record an
// alias so emailed URLs keep working (see slug_aliases).

// foldASCII strips diacritics ("José" → "Jose") so names slug cleanly.
func foldASCII(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// deriveClientSlug turns a person's name into the studio's login-style slug:
// first initial of the first word plus the whole last word ("Mike Casey" →
// "mcasey", "Jean-Luc Picard" → "jpicard"). A single word is used as-is;
// an empty name yields "author".
func deriveClientSlug(name string) string {
	words := strings.Fields(foldASCII(name))
	if len(words) == 0 {
		return "author"
	}
	clean := func(w string) string { return strings.ReplaceAll(normalizeProjectSlug(w), "-", "") }
	last := clean(words[len(words)-1])
	if len(words) == 1 {
		if last == "" {
			return "author"
		}
		return truncateSlug(last)
	}
	first := clean(words[0])
	if last == "" {
		last = first
		first = ""
	}
	if first != "" {
		first = first[:1]
	}
	out := truncateSlug(first + last)
	if out == "" {
		return "author"
	}
	return out
}

// uniqueClientSlug derives a client slug from a customer name and appends
// 2, 3, … until it is free in clients and not a live alias.
func uniqueClientSlug(ctx context.Context, db slugQuerier, name string) (string, error) {
	base := deriveClientSlug(name)
	for i := 1; i < 1000; i++ {
		candidate := base
		if i > 1 {
			candidate = fmt.Sprintf("%s%d", base, i)
		}
		var n int
		if err := db.QueryRowContext(ctx, `
			SELECT (SELECT COUNT(*) FROM clients WHERE slug = ?)
			     + (SELECT COUNT(*) FROM slug_aliases WHERE old_client_slug = ?)`,
			candidate, candidate).Scan(&n); err != nil {
			return "", fmt.Errorf("check client slug: %w", err)
		}
		if n == 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no free client slug for %q", base)
}

var bookSeqRe = regexp.MustCompile(`^book-(\d{3,})$`)

// nextProjectSlug returns the next "book-NNN" for a client: one past the
// highest existing sequence number (archived projects included, so numbers
// are never reused), then bumped until unique.
func nextProjectSlug(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}, clientSlug string) (string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT project_slug FROM projects WHERE client_slug = ?
		UNION SELECT old_project_slug FROM slug_aliases WHERE old_client_slug = ?`, clientSlug, clientSlug)
	if err != nil {
		return "", fmt.Errorf("list project slugs: %w", err)
	}
	defer rows.Close()
	maxN := 0
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return "", err
		}
		if m := bookSeqRe.FindStringSubmatch(slug); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil && n > maxN {
				maxN = n
			}
		}
	}
	for i := maxN + 1; i < maxN+1000; i++ {
		candidate := fmt.Sprintf("book-%03d", i)
		var n int
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM projects WHERE client_slug = ? AND project_slug = ?`,
			clientSlug, candidate).Scan(&n); err != nil {
			return "", fmt.Errorf("check project slug: %w", err)
		}
		if n == 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no free project slug for %q", clientSlug)
}

// handleAdminSlugSuggest — GET /api/admin/slug-suggest?name=Mike+Casey&client=mcasey
// Returns the slugs the admin New Client / New Project modals should prefill:
// client_slug for a name, project_slug (next book-NNN) for a client.
func (s *Server) handleAdminSlugSuggest(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	out := map[string]string{}
	if name := strings.TrimSpace(r.URL.Query().Get("name")); name != "" {
		slug, err := uniqueClientSlug(r.Context(), s.DB, name)
		if err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		out["client_slug"] = slug
	}
	if client := normalizeProjectSlug(r.URL.Query().Get("client")); client != "" {
		slug, err := nextProjectSlug(r.Context(), s.DB, client)
		if err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		out["project_slug"] = slug
	}
	jsonOK(w, out)
}

// ─── Aliases: old URLs keep working after a rename ───

type slugAlias struct{ client, project string }

func (s *Server) loadSlugAliases() {
	rows, err := s.DB.Query(`SELECT old_client_slug, old_project_slug, new_client_slug, new_project_slug FROM slug_aliases`)
	if err != nil {
		slog.Warn("load slug aliases", "err", err)
		return
	}
	defer rows.Close()
	m := map[slugAlias]slugAlias{}
	for rows.Next() {
		var oc, op, nc, np string
		if err := rows.Scan(&oc, &op, &nc, &np); err == nil {
			m[slugAlias{oc, op}] = slugAlias{nc, np}
		}
	}
	s.aliasMu.Lock()
	s.aliases = m
	s.aliasMu.Unlock()
}

// resolveSlugAlias maps an old /{client}/{project} pair to its current one.
// project may be "" for a client-portal path. Returns ok=false when nothing
// is aliased (the common case: one map lookup, no DB).
func (s *Server) resolveSlugAlias(client, project string) (string, string, bool) {
	s.aliasMu.RLock()
	defer s.aliasMu.RUnlock()
	if len(s.aliases) == 0 {
		return "", "", false
	}
	if project != "" {
		if t, ok := s.aliases[slugAlias{client, project}]; ok {
			return t.client, t.project, true
		}
	}
	if t, ok := s.aliases[slugAlias{client, ""}]; ok {
		return t.client, project, true
	}
	return "", "", false
}

// redirectAliasedPath issues a 301 when the request path starts with an
// aliased client/project pair. Reports whether it handled the request.
func (s *Server) redirectAliasedPath(w http.ResponseWriter, r *http.Request, parts []string) bool {
	if len(parts) == 0 {
		return false
	}
	project := ""
	if len(parts) >= 2 {
		project = parts[1]
	}
	nc, np, ok := s.resolveSlugAlias(parts[0], project)
	if !ok {
		return false
	}
	newParts := append([]string{nc}, parts[1:]...)
	if project != "" {
		newParts[1] = np
	}
	target := "/" + strings.Join(newParts, "/")
	if strings.HasSuffix(r.URL.Path, "/") || len(parts) <= 2 {
		target += "/"
	}
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
	return true
}

// handleAdminRenameClient — POST /api/admin/clients/{slug}/rename {"slug":"mcasey"}
// Renames the client slug, moves every project under it, and records an alias
// so /{old}/... redirects. Client cookies are keyed by slug, so the client
// signs in again at the new address.
func (s *Server) handleAdminRenameClient(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	oldSlug := r.PathValue("slug")
	var body struct {
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	newSlug := normalizeProjectSlug(body.Slug)
	if newSlug == "" {
		jsonErr(w, "slug required", 400)
		return
	}
	if newSlug == oldSlug {
		jsonErr(w, "that is already the client's slug", 400)
		return
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM clients WHERE slug = ?`, oldSlug).Scan(&n); err != nil || n == 0 {
		jsonErr(w, "client not found", 404)
		return
	}
	if err := tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM clients WHERE slug = ?`, newSlug).Scan(&n); err != nil || n > 0 {
		jsonErr(w, "client slug already exists", 409)
		return
	}
	steps := []struct {
		q    string
		args []any
	}{
		{`UPDATE clients SET slug = ? WHERE slug = ?`, []any{newSlug, oldSlug}},
		{`UPDATE projects SET client_slug = ? WHERE client_slug = ?`, []any{newSlug, oldSlug}},
		// Keep earlier aliases pointing at the live slug, and free the new
		// slug from any alias that used to claim it.
		{`UPDATE slug_aliases SET new_client_slug = ? WHERE new_client_slug = ?`, []any{newSlug, oldSlug}},
		{`DELETE FROM slug_aliases WHERE old_client_slug = ?`, []any{newSlug}},
		{`INSERT OR REPLACE INTO slug_aliases (old_client_slug, old_project_slug, new_client_slug, new_project_slug) VALUES (?, '', ?, '')`, []any{oldSlug, newSlug}},
	}
	for _, st := range steps {
		if _, err := tx.ExecContext(r.Context(), st.q, st.args...); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	s.loadSlugAliases()
	slog.Info("client renamed", "from", oldSlug, "to", newSlug, "by", triggeredBy(r, "admin"))
	jsonOK(w, map[string]string{"old_slug": oldSlug, "slug": newSlug})
}

// handleAdminRenameProjectSlug — POST /api/admin/projects/{id}/slug {"project_slug":"book-001"}
// Changes only the URL segment; projects.name (the title) is untouched.
func (s *Server) handleAdminRenameProjectSlug(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	var body struct {
		ProjectSlug string `json:"project_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	newSlug := normalizeProjectSlug(body.ProjectSlug)
	if newSlug == "" {
		jsonErr(w, "project_slug required", 400)
		return
	}
	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer tx.Rollback()
	var client, oldSlug string
	if err := tx.QueryRowContext(r.Context(), `SELECT client_slug, project_slug FROM projects WHERE id = ?`, pid).Scan(&client, &oldSlug); err != nil {
		jsonErr(w, "project not found", 404)
		return
	}
	if newSlug == oldSlug {
		jsonErr(w, "that is already the project's slug", 400)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `UPDATE projects SET project_slug = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, newSlug, pid); err != nil {
		if isUniqueErr(err) {
			jsonErr(w, "a project with that slug already exists for this client", 409)
			return
		}
		jsonErr(w, err.Error(), 500)
		return
	}
	steps := []struct {
		q    string
		args []any
	}{
		{`UPDATE slug_aliases SET new_project_slug = ? WHERE new_client_slug = ? AND new_project_slug = ?`, []any{newSlug, client, oldSlug}},
		{`DELETE FROM slug_aliases WHERE old_client_slug = ? AND old_project_slug = ?`, []any{client, newSlug}},
		{`INSERT OR REPLACE INTO slug_aliases (old_client_slug, old_project_slug, new_client_slug, new_project_slug) VALUES (?, ?, ?, ?)`, []any{client, oldSlug, client, newSlug}},
	}
	for _, st := range steps {
		if _, err := tx.ExecContext(r.Context(), st.q, st.args...); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	s.loadSlugAliases()
	slog.Info("project slug renamed", "id", pid, "client", client, "from", oldSlug, "to", newSlug, "by", triggeredBy(r, "admin"))
	jsonOK(w, map[string]any{"id": pid, "client_slug": client, "project_slug": newSlug, "path": "/" + client + "/" + newSlug + "/"})
}
