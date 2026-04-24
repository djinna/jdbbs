package srv

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"srv.exe.dev/db/dbgen"
)

func normalizeProjectSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastHyphen := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		case r == '-', r == '_', r == ' ', r == '.', r == '/':
			if b.Len() > 0 && !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// hasAnyProjectAuthForClient checks if the user has a valid project-level auth cookie
// for any project belonging to this client. This allows users who authenticated
// to a specific project to also see sibling projects in the same client.
func (s *Server) hasAnyProjectAuthForClient(r *http.Request, clientSlug string) bool {
	rows, err := s.DB.QueryContext(r.Context(),
		`SELECT p.id FROM projects p WHERE p.client_slug = ?`, clientSlug)
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			continue
		}
		if s.checkAuth(r, pid) {
			return true
		}
	}
	return false
}

// checkClientAuth checks if the request has a valid client-level auth cookie.
func (s *Server) checkClientAuth(r *http.Request, clientSlug string) bool {
	if clientSlug == "" {
		return false
	}

	// Look up client password
	var passwordHash string
	err := s.DB.QueryRowContext(r.Context(),
		`SELECT password_hash FROM clients WHERE slug = ?`, clientSlug,
	).Scan(&passwordHash)
	if err != nil || passwordHash == "" {
		return false // no client or no password set
	}

	// Check cookie
	c, err := r.Cookie("prodcal_client_" + clientSlug)
	if err != nil || c.Value == "" {
		return false
	}

	return hashToken(c.Value) == passwordHash
}

// getClientSlugForProject returns the client_slug for a given project ID.
func (s *Server) getClientSlugForProject(r *http.Request, projectID int64) string {
	var slug string
	s.DB.QueryRowContext(r.Context(),
		`SELECT client_slug FROM projects WHERE id = ?`, projectID,
	).Scan(&slug)
	return slug
}

// handleClientVerify verifies a client password and sets a cookie.
func (s *Server) handleClientVerify(w http.ResponseWriter, r *http.Request) {
	clientSlug := r.PathValue("client")
	if clientSlug == "" {
		jsonErr(w, "client slug required", 400)
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}

	// Look up client
	var passwordHash string
	err := s.DB.QueryRowContext(r.Context(),
		`SELECT password_hash FROM clients WHERE slug = ?`, clientSlug,
	).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		jsonErr(w, "client not found", 404)
		return
	}
	if err != nil {
		jsonErr(w, "server error", 500)
		return
	}
	if passwordHash == "" {
		// No password = open access
		jsonOK(w, map[string]any{"ok": true})
		return
	}

	if hashToken(body.Password) != passwordHash {
		jsonErr(w, "invalid password", http.StatusUnauthorized)
		return
	}

	// Set client cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "prodcal_client_" + clientSlug,
		Value:    body.Password,
		Path:     "/",
		MaxAge:   86400 * 90, // 90 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	jsonOK(w, map[string]any{"ok": true})
}

// handleClientInfo returns client info and auth status.
func (s *Server) handleClientInfo(w http.ResponseWriter, r *http.Request) {
	clientSlug := r.PathValue("client")

	var name, passwordHash string
	err := s.DB.QueryRowContext(r.Context(),
		`SELECT name, password_hash FROM clients WHERE slug = ?`, clientSlug,
	).Scan(&name, &passwordHash)
	if err == sql.ErrNoRows {
		jsonErr(w, "client not found", 404)
		return
	}
	if err != nil {
		jsonErr(w, "server error", 500)
		return
	}

	hasAuth := passwordHash != ""
	authed := !hasAuth || s.checkClientAuth(r, clientSlug)

	jsonOK(w, map[string]any{
		"slug":          clientSlug,
		"name":          name,
		"has_auth":      hasAuth,
		"authenticated": authed,
	})
}

// handleClientProjects returns projects for a client.
func (s *Server) handleClientProjects(w http.ResponseWriter, r *http.Request) {
	clientSlug := r.PathValue("client")

	// Check client auth
	var passwordHash string
	err := s.DB.QueryRowContext(r.Context(),
		`SELECT password_hash FROM clients WHERE slug = ?`, clientSlug,
	).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		jsonErr(w, "client not found", 404)
		return
	}
	if passwordHash != "" && !s.checkClientAuth(r, clientSlug) {
		// Also allow if the user has project-level auth for any project in this client
		if !s.hasAnyProjectAuthForClient(r, clientSlug) {
			jsonErr(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT
			p.id, p.name, p.client_slug, p.project_slug,
			p.start_date, p.updated_at,
			COUNT(t.id) as task_count,
			SUM(CASE WHEN t.status = 'done' THEN 1 ELSE 0 END) as done_count,
			SUM(CASE WHEN t.status = 'active' OR t.status = 'in_progress' THEN 1 ELSE 0 END) as active_count,
			COALESCE(tr.id, 0) as has_transmittal,
			COALESCE(tr.status, '') as transmittal_status
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		LEFT JOIN transmittals tr ON tr.project_id = p.id
		WHERE p.client_slug = ? AND p.archived_at IS NULL
		GROUP BY p.id
		ORDER BY p.name
	`, clientSlug)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	type clientProject struct {
		ID                int64  `json:"id"`
		Name              string `json:"name"`
		ClientSlug        string `json:"client_slug"`
		ProjectSlug       string `json:"project_slug"`
		StartDate         string `json:"start_date"`
		UpdatedAt         string `json:"updated_at"`
		TaskCount         int    `json:"task_count"`
		DoneCount         int    `json:"done_count"`
		ActiveCount       int    `json:"active_count"`
		HasTransmittal    bool   `json:"has_transmittal"`
		TransmittalStatus string `json:"transmittal_status"`
		Path              string `json:"path"`
	}
	var projects []clientProject
	for rows.Next() {
		var p clientProject
		var hasTransmittalID int64
		if err := rows.Scan(&p.ID, &p.Name, &p.ClientSlug, &p.ProjectSlug,
			&p.StartDate, &p.UpdatedAt, &p.TaskCount, &p.DoneCount,
			&p.ActiveCount, &hasTransmittalID, &p.TransmittalStatus); err != nil {
			continue
		}
		p.HasTransmittal = hasTransmittalID != 0
		p.Path = "/" + p.ClientSlug + "/" + p.ProjectSlug + "/"
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []clientProject{}
	}

	jsonOK(w, projects)
}

// handleClientCreateProject creates a new project from the client portal.
// Requires valid client-level auth when the client is password-protected.
func (s *Server) handleClientCreateProject(w http.ResponseWriter, r *http.Request) {
	clientSlug := r.PathValue("client")
	if clientSlug == "" {
		jsonErr(w, "client slug required", 400)
		return
	}

	var passwordHash string
	err := s.DB.QueryRowContext(r.Context(),
		`SELECT password_hash FROM clients WHERE slug = ?`, clientSlug,
	).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		jsonErr(w, "client not found", 404)
		return
	}
	if err != nil {
		jsonErr(w, "server error", 500)
		return
	}
	if passwordHash != "" && !s.checkClientAuth(r, clientSlug) {
		jsonErr(w, "client login required", http.StatusUnauthorized)
		return
	}

	var body struct {
		Name        string `json:"name"`
		StartDate   string `json:"start_date"`
		ProjectSlug string `json:"project_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		jsonErr(w, "name required", 400)
		return
	}

	body.ProjectSlug = normalizeProjectSlug(body.ProjectSlug)
	if body.ProjectSlug == "" {
		body.ProjectSlug = normalizeProjectSlug(body.Name)
	}
	if body.ProjectSlug == "" {
		jsonErr(w, "project slug required", 400)
		return
	}

	tx, err := s.DB.BeginTx(r.Context(), nil)
	if err != nil {
		jsonErr(w, "begin tx failed: "+err.Error(), 500)
		return
	}
	defer tx.Rollback()

	q := dbgen.New(tx)
	p, err := q.CreateProject(r.Context(), dbgen.CreateProjectParams{
		Name:        body.Name,
		StartDate:   strings.TrimSpace(body.StartDate),
		ClientSlug:  clientSlug,
		ProjectSlug: body.ProjectSlug,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			jsonErr(w, "project slug already exists for this client", 409)
			return
		}
		jsonErr(w, err.Error(), 500)
		return
	}
	if err := seedProjectWithStandardWorkflow(r.Context(), q, p.ID); err != nil {
		jsonErr(w, "seed workflow failed: "+err.Error(), 500)
		return
	}
	if err := tx.Commit(); err != nil {
		jsonErr(w, "commit failed: "+err.Error(), 500)
		return
	}

	w.WriteHeader(201)
	jsonOK(w, p)
}
