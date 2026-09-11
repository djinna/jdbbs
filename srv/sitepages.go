package srv

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Site pages registry — the admin's inventory of every route we spin up.
// See docs/PAGE-DESIGN-HOSTING-VISIBILITY-2026-09-11.md §7 "Discoverability".
// Authorization is decided by each route's Go handler; this table only
// records what exists and what we intend to do with it.

type sitePage struct {
	ID         int64  `json:"id"`
	Route      string `json:"route"`
	Title      string `json:"title"`
	Owner      string `json:"owner"`
	Source     string `json:"source"`
	Visibility string `json:"visibility"`
	Listed     string `json:"listed"`
	PageType   string `json:"page_type"`
	Status     string `json:"status"`
	Note       string `json:"note"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

var (
	validPageVisibility = map[string]bool{"public": true, "client": true, "cohort": true, "admin": true}
	validPageListed     = map[string]bool{"listed": true, "nav": true, "unlisted": true, "retired": true}
	validPageStatus     = map[string]bool{"live": true, "keep": true, "review": true, "retire": true}
	validPageOwner      = map[string]bool{"prodcal": true, "pi-public": true, "generated": true, "external": true}
)

func (s *Server) handleAdminListSitePages(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, route, title, owner, source, visibility, listed, page_type, status, note, created_at, updated_at
		FROM site_pages ORDER BY visibility, route`)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	out := []sitePage{}
	for rows.Next() {
		var p sitePage
		if err := rows.Scan(&p.ID, &p.Route, &p.Title, &p.Owner, &p.Source, &p.Visibility,
			&p.Listed, &p.PageType, &p.Status, &p.Note, &p.CreatedAt, &p.UpdatedAt); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		out = append(out, p)
	}
	jsonOK(w, out)
}

type sitePageInput struct {
	Route      string `json:"route"`
	Title      string `json:"title"`
	Owner      string `json:"owner"`
	Source     string `json:"source"`
	Visibility string `json:"visibility"`
	Listed     string `json:"listed"`
	PageType   string `json:"page_type"`
	Status     string `json:"status"`
	Note       string `json:"note"`
}

func (in *sitePageInput) normalize() string {
	in.Route = strings.TrimSpace(in.Route)
	if in.Route == "" || !strings.HasPrefix(in.Route, "/") {
		return "route must start with /"
	}
	in.Title = clip(strings.TrimSpace(in.Title), 200)
	in.Source = clip(strings.TrimSpace(in.Source), 300)
	in.Note = clip(strings.TrimSpace(in.Note), 1000)
	in.PageType = clip(strings.TrimSpace(in.PageType), 40)
	if in.Owner == "" {
		in.Owner = "prodcal"
	}
	if in.Visibility == "" {
		in.Visibility = "public"
	}
	if in.Listed == "" {
		in.Listed = "unlisted"
	}
	if in.Status == "" {
		in.Status = "live"
	}
	if in.PageType == "" {
		in.PageType = "prose"
	}
	switch {
	case !validPageOwner[in.Owner]:
		return "invalid owner"
	case !validPageVisibility[in.Visibility]:
		return "invalid visibility"
	case !validPageListed[in.Listed]:
		return "invalid listed value"
	case !validPageStatus[in.Status]:
		return "invalid status"
	}
	return ""
}

// handleAdminUpsertSitePage creates or updates by route (PUT /api/admin/pages).
func (s *Server) handleAdminUpsertSitePage(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	var in sitePageInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid page", http.StatusBadRequest)
		return
	}
	if msg := in.normalize(); msg != "" {
		jsonErr(w, msg, http.StatusBadRequest)
		return
	}
	_, err := s.DB.ExecContext(r.Context(), `
		INSERT INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(route) DO UPDATE SET
		  title=excluded.title, owner=excluded.owner, source=excluded.source,
		  visibility=excluded.visibility, listed=excluded.listed, page_type=excluded.page_type,
		  status=excluded.status, note=excluded.note, updated_at=CURRENT_TIMESTAMP`,
		in.Route, in.Title, in.Owner, in.Source, in.Visibility, in.Listed, in.PageType, in.Status, in.Note)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{"ok": true})
}

func (s *Server) handleAdminDeleteSitePage(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		jsonErr(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := s.DB.ExecContext(r.Context(), `DELETE FROM site_pages WHERE id = ?`, id); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{"ok": true})
}
