package srv

import (
	"encoding/json"
	"net/http"
	"strings"
)

type projectSummary struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	ClientSlug        string `json:"client_slug"`
	ProjectSlug       string `json:"project_slug"`
	StartDate         string `json:"start_date"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	ArchivedAt        string `json:"archived_at"`
	TaskCount         int    `json:"task_count"`
	DoneCount         int    `json:"done_count"`
	ActiveCount       int    `json:"active_count"`
	HasAuth           bool   `json:"has_auth"`
	HasTransmittal    bool   `json:"has_transmittal"`
	TransmittalStatus string `json:"transmittal_status"`
	Path              string `json:"path"`
}

func (s *Server) requireExeDevAdmin(w http.ResponseWriter, r *http.Request) bool {
	userID := r.Header.Get("X-ExeDev-UserID")
	if userID == "" {
		// Redirect to exe.dev login flow, which will bounce back after auth
		redirect := r.URL.Path
		if r.URL.RawQuery != "" {
			redirect += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, "/__exe.dev/login?redirect="+redirect, http.StatusFound)
		return false
	}
	return true
}

func (s *Server) requireExeDevAdminAPI(w http.ResponseWriter, r *http.Request) bool {
	userID := r.Header.Get("X-ExeDev-UserID")
	if userID == "" {
		jsonErr(w, "exe.dev login required", http.StatusUnauthorized)
		return false
	}
	return true
}

func (s *Server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdmin(w, r) {
		return
	}
	data, err := staticFS.ReadFile("static/admin.html")
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *Server) handleAdminProjectList(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}

	archivedOnly := r.URL.Query().Get("archived") == "1"
	whereClause := "p.archived_at IS NULL"
	orderClause := "p.updated_at DESC"
	if archivedOnly {
		whereClause = "p.archived_at IS NOT NULL"
		orderClause = "p.archived_at DESC, p.updated_at DESC"
	}

	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT
			p.id, p.name, p.client_slug, p.project_slug,
			p.start_date, p.created_at, p.updated_at, COALESCE(p.archived_at, ''),
			COUNT(t.id) as task_count,
			SUM(CASE WHEN t.status = 'done' THEN 1 ELSE 0 END) as done_count,
			SUM(CASE WHEN t.status = 'active' OR t.status = 'in_progress' THEN 1 ELSE 0 END) as active_count,
			EXISTS(SELECT 1 FROM auth_tokens WHERE project_id = p.id) as has_auth,
			COALESCE(tr.id, 0) as has_transmittal,
			COALESCE(tr.status, '') as transmittal_status
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		LEFT JOIN transmittals tr ON tr.project_id = p.id
		WHERE `+whereClause+`
		GROUP BY p.id
		ORDER BY `+orderClause)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var projects []projectSummary
	for rows.Next() {
		var ps projectSummary
		var hasTransmittalID int64
		var hasAuthInt int64
		err := rows.Scan(
			&ps.ID, &ps.Name, &ps.ClientSlug, &ps.ProjectSlug,
			&ps.StartDate, &ps.CreatedAt, &ps.UpdatedAt, &ps.ArchivedAt,
			&ps.TaskCount, &ps.DoneCount, &ps.ActiveCount,
			&hasAuthInt, &hasTransmittalID, &ps.TransmittalStatus,
		)
		if err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		ps.HasAuth = hasAuthInt != 0
		ps.HasTransmittal = hasTransmittalID != 0
		ps.Path = "/" + ps.ClientSlug + "/" + ps.ProjectSlug + "/"
		projects = append(projects, ps)
	}

	if projects == nil {
		projects = []projectSummary{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

type adminClientSummary struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	HasAuth      bool   `json:"has_auth"`
	ProjectCount int    `json:"project_count"`
	CreatedAt    string `json:"created_at"`
}

func (s *Server) handleAdminClientList(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}

	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT c.slug, c.name, c.password_hash, c.created_at, COUNT(p.id) as project_count
		FROM clients c
		LEFT JOIN projects p ON p.client_slug = c.slug
		GROUP BY c.slug, c.name, c.password_hash, c.created_at
		ORDER BY c.created_at DESC, c.slug ASC
	`)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var clients []adminClientSummary
	for rows.Next() {
		var c adminClientSummary
		var passwordHash string
		if err := rows.Scan(&c.Slug, &c.Name, &passwordHash, &c.CreatedAt, &c.ProjectCount); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		c.HasAuth = strings.TrimSpace(passwordHash) != ""
		clients = append(clients, c)
	}
	if clients == nil {
		clients = []adminClientSummary{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func (s *Server) handleAdminCreateClient(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}

	var body struct {
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.Slug = normalizeProjectSlug(body.Slug)
	if body.Name == "" {
		jsonErr(w, "name required", 400)
		return
	}
	if body.Slug == "" {
		jsonErr(w, "slug required", 400)
		return
	}
	passwordHash := ""
	if strings.TrimSpace(body.Password) != "" {
		passwordHash = hashToken(body.Password)
	}
	_, err := s.DB.ExecContext(r.Context(), `INSERT INTO clients (slug, name, password_hash) VALUES (?, ?, ?)`, body.Slug, body.Name, passwordHash)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			jsonErr(w, "client slug already exists", 409)
			return
		}
		jsonErr(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(map[string]any{
		"slug":      body.Slug,
		"name":      body.Name,
		"has_auth":  passwordHash != "",
		"created_ok": true,
	})
}
