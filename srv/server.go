package srv

import (
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"srv.exe.dev/db"
	"srv.exe.dev/db/dbgen"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	DB       *sql.DB
	Hostname string
	Email    *EmailConfig
}

func New(dbPath, hostname string) (*Server, error) {
	srv := &Server{Hostname: hostname}
	if err := srv.setUpDatabase(dbPath); err != nil {
		return nil, err
	}
	srv.Email = LoadEmailConfig()
	if srv.Email != nil {
		slog.Info("email configured", "inbox_id", srv.Email.InboxID)
	} else {
		slog.Warn("email not configured (set AGENTMAIL_API_KEY + AGENTMAIL_INBOX_ID)")
	}
	return srv, nil
}

func (s *Server) setUpDatabase(dbPath string) error {
	wdb, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	s.DB = wdb
	if err := db.RunMigrations(wdb); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func (s *Server) Serve(addr string) error {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /healthz", s.handleHealthz)

	// Admin dashboard
	mux.HandleFunc("GET /admin/", s.handleAdminDashboard)
	mux.HandleFunc("GET /api/admin/projects", s.handleAdminProjectList)

	// API routes (global, no path prefix)
	mux.HandleFunc("GET /api/projects", s.handleListProjects)
	mux.HandleFunc("POST /api/projects", s.handleCreateProject)
	mux.HandleFunc("GET /api/projects/{id}", s.handleGetProject)
	mux.HandleFunc("PUT /api/projects/{id}", s.handleUpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", s.handleDeleteProject)
	mux.HandleFunc("GET /api/projects/{id}/tasks", s.handleListTasks)
	mux.HandleFunc("POST /api/projects/{id}/tasks", s.handleCreateTask)
	mux.HandleFunc("PUT /api/tasks/{id}", s.handleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)
	mux.HandleFunc("POST /api/projects/{id}/auth", s.handleSetAuth)
	mux.HandleFunc("POST /api/projects/{id}/verify", s.handleVerifyAuth)
	mux.HandleFunc("POST /api/projects/{id}/seed", s.handleSeedProject)
	mux.HandleFunc("POST /api/projects/{id}/duplicate", s.handleDuplicateProject)
	mux.HandleFunc("GET /api/project-by-path/{client}/{project}", s.handleGetProjectByPath)

	// Transmittal API
	mux.HandleFunc("GET /api/projects/{id}/transmittal", s.handleGetTransmittal)
	mux.HandleFunc("PUT /api/projects/{id}/transmittal", s.handleUpdateTransmittal)
	mux.HandleFunc("GET /api/transmittals/{id}/versions", s.handleListTransmittalVersions)
	mux.HandleFunc("GET /api/transmittals/{id}/versions/{vid}", s.handleGetTransmittalVersion)
	mux.HandleFunc("POST /api/transmittals/{id}/versions/{vid}/restore", s.handleRestoreTransmittalVersion)
	mux.HandleFunc("POST /api/transmittals/{id}/duplicate", s.handleDuplicateTransmittal)

	// Email API
	mux.HandleFunc("POST /api/projects/{id}/transmittal/email", s.handleSendTransmittalEmail)
	mux.HandleFunc("POST /api/projects/{id}/snapshot/email", s.handleSendProjectSnapshot)
	mux.HandleFunc("GET /api/email/status", s.handleEmailStatus)

	// Client API
	mux.HandleFunc("GET /api/clients/{client}", s.handleClientInfo)
	mux.HandleFunc("POST /api/clients/{client}/verify", s.handleClientVerify)
	mux.HandleFunc("GET /api/clients/{client}/projects", s.handleClientProjects)

	// Static files (CSS, JS) at known paths
	static, _ := fs.Sub(staticFS, "static")
	staticServer := http.FileServer(http.FS(static))
	mux.Handle("GET /static/", http.StripPrefix("/static", staticServer))

	// Root: redirect admin users to /admin/, show landing page for everyone else
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-ExeDev-UserID")
		if userID != "" {
			http.Redirect(w, r, "/admin/", http.StatusFound)
			return
		}
		s.serveLanding(w)
	})

	// /{client}/{project}/ serves the SPA, /{client}/ serves client portal
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.NotFound(w, r)
			return
		}
		// /{client}/ -> client portal
		if len(parts) == 1 {
			if !strings.HasSuffix(path, "/") {
				http.Redirect(w, r, path+"/", http.StatusMovedPermanently)
				return
			}
			s.serveClientPortal(w)
			return
		}
		// /vgr/aog -> redirect to /vgr/aog/
		if len(parts) == 2 && !strings.HasSuffix(path, "/") {
			http.Redirect(w, r, path+"/", http.StatusMovedPermanently)
			return
		}
		// /vgr/aog/transmittal/ -> serve transmittal SPA
		if len(parts) == 3 && parts[2] == "transmittal" && !strings.HasSuffix(path, "/") {
			http.Redirect(w, r, path+"/", http.StatusMovedPermanently)
			return
		}
		if len(parts) >= 3 && parts[2] == "transmittal" {
			if len(parts) > 3 {
				// /vgr/aog/transmittal/style.css -> serve static
				assetPath := strings.Join(parts[3:], "/")
				r.URL.Path = "/" + assetPath
				staticServer.ServeHTTP(w, r)
				return
			}
			s.serveTransmittal(w)
			return
		}
		// /vgr/aog/style.css -> serve static asset
		if len(parts) > 2 {
			assetPath := strings.Join(parts[2:], "/")
			r.URL.Path = "/" + assetPath
			staticServer.ServeHTTP(w, r)
			return
		}
		// /vgr/aog/ -> serve the calendar SPA
		s.serveIndex(w)
	})

	slog.Info("starting server", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) serveTransmittal(w http.ResponseWriter) {
	data, err := staticFS.ReadFile("static/transmittal.html")
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *Server) serveClientPortal(w http.ResponseWriter) {
	data, err := staticFS.ReadFile("static/client.html")
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *Server) serveLanding(w http.ResponseWriter) {
	data, err := staticFS.ReadFile("static/landing.html")
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *Server) serveIndex(w http.ResponseWriter) {
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	var n int
	err := s.DB.QueryRowContext(r.Context(), "SELECT 1").Scan(&n)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "detail": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(h[:])
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *Server) projectIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// Auth middleware: checks project-level cookie, client-level cookie, or Authorization header
func (s *Server) checkAuth(r *http.Request, projectID int64) bool {
	q := dbgen.New(s.DB)
	// Check if project has any auth tokens at all
	tokens, err := q.ListAuthTokens(r.Context(), projectID)
	if err != nil || len(tokens) == 0 {
		// No auth configured = open access
		return true
	}

	// Check project-level cookie
	var raw string
	if c, err := r.Cookie(fmt.Sprintf("prodcal_auth_%d", projectID)); err == nil {
		raw = c.Value
	}
	if raw == "" {
		raw = r.Header.Get("X-Auth-Token")
	}
	if raw != "" {
		_, err = q.GetAuthToken(r.Context(), dbgen.GetAuthTokenParams{
			ProjectID: projectID,
			TokenHash: hashToken(raw),
		})
		if err == nil {
			return true
		}
	}

	// Check client-level cookie
	clientSlug := s.getClientSlugForProject(r, projectID)
	if clientSlug != "" && s.checkClientAuth(r, clientSlug) {
		return true
	}

	// Check exe.dev admin
	if r.Header.Get("X-ExeDev-UserID") != "" {
		return true
	}

	return false
}

func (s *Server) requireAuth(w http.ResponseWriter, r *http.Request, projectID int64) bool {
	if !s.checkAuth(r, projectID) {
		jsonErr(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

// --- Project handlers ---

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	q := dbgen.New(s.DB)
	projects, err := q.ListProjects(r.Context())
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if projects == nil {
		projects = []dbgen.Project{}
	}
	jsonOK(w, projects)
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		StartDate   string `json:"start_date"`
		ClientSlug  string `json:"client_slug"`
		ProjectSlug string `json:"project_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.ClientSlug == "" || body.ProjectSlug == "" {
		jsonErr(w, "client_slug and project_slug required", 400)
		return
	}
	q := dbgen.New(s.DB)
	p, err := q.CreateProject(r.Context(), dbgen.CreateProjectParams{
		Name:        body.Name,
		StartDate:   body.StartDate,
		ClientSlug:  body.ClientSlug,
		ProjectSlug: body.ProjectSlug,
	})
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	w.WriteHeader(201)
	jsonOK(w, p)
}

func (s *Server) handleGetProjectByPath(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	project := r.PathValue("project")
	q := dbgen.New(s.DB)
	p, err := q.GetProjectByPath(r.Context(), dbgen.GetProjectByPathParams{
		ClientSlug:  client,
		ProjectSlug: project,
	})
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	tokens, _ := q.ListAuthTokens(r.Context(), p.ID)
	hasAuth := len(tokens) > 0
	authed := s.checkAuth(r, p.ID)
	jsonOK(w, map[string]any{
		"project":       p,
		"has_auth":      hasAuth,
		"authenticated": authed,
	})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	q := dbgen.New(s.DB)
	p, err := q.GetProject(r.Context(), pid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}

	// Check if auth is required
	tokens, _ := q.ListAuthTokens(r.Context(), pid)
	hasAuth := len(tokens) > 0
	authed := s.checkAuth(r, pid)

	jsonOK(w, map[string]any{
		"project":       p,
		"has_auth":      hasAuth,
		"authenticated": authed,
	})
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	var body struct {
		Name        string `json:"name"`
		StartDate   string `json:"start_date"`
		ClientSlug  string `json:"client_slug"`
		ProjectSlug string `json:"project_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	q := dbgen.New(s.DB)
	// Preserve existing slugs if not provided
	if body.ClientSlug == "" || body.ProjectSlug == "" {
		existing, _ := q.GetProject(r.Context(), pid)
		if body.ClientSlug == "" {
			body.ClientSlug = existing.ClientSlug
		}
		if body.ProjectSlug == "" {
			body.ProjectSlug = existing.ProjectSlug
		}
	}
	if err := q.UpdateProject(r.Context(), dbgen.UpdateProjectParams{
		Name: body.Name, StartDate: body.StartDate,
		ClientSlug: body.ClientSlug, ProjectSlug: body.ProjectSlug,
		ID: pid,
	}); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"ok": "true"})
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	q := dbgen.New(s.DB)
	if err := q.DeleteProject(r.Context(), pid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"ok": "true"})
}

// --- Task handlers ---

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	q := dbgen.New(s.DB)
	tasks, err := q.ListTasks(r.Context(), pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if tasks == nil {
		tasks = []dbgen.Task{}
	}
	jsonOK(w, tasks)
}

type taskInput struct {
	SortOrder    int64   `json:"sort_order"`
	Assignee     string  `json:"assignee"`
	Title        string  `json:"title"`
	IsMilestone  int64   `json:"is_milestone"`
	OrigWeeks    float64 `json:"orig_weeks"`
	CurrWeeks    float64 `json:"curr_weeks"`
	OrigDue      string  `json:"orig_due"`
	CurrDue      string  `json:"curr_due"`
	ActualDone   string  `json:"actual_done"`
	Status       string  `json:"status"`
	Words        int64   `json:"words"`
	WordsPerHour int64   `json:"words_per_hour"`
	Hours        float64 `json:"hours"`
	Rate         float64 `json:"rate"`
	BudgetNotes  string  `json:"budget_notes"`
	OrigBudget   float64 `json:"orig_budget"`
	CurrBudget   float64 `json:"curr_budget"`
	ActualBudget float64 `json:"actual_budget"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	var body taskInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.Status == "" {
		body.Status = "pending"
	}
	q := dbgen.New(s.DB)
	t, err := q.CreateTask(r.Context(), dbgen.CreateTaskParams{
		ProjectID: pid, SortOrder: body.SortOrder,
		Assignee: body.Assignee, Title: body.Title,
		IsMilestone: body.IsMilestone,
		OrigWeeks: body.OrigWeeks, CurrWeeks: body.CurrWeeks,
		OrigDue: body.OrigDue, CurrDue: body.CurrDue,
		ActualDone: body.ActualDone, Status: body.Status,
		Words: body.Words, WordsPerHour: body.WordsPerHour,
		Hours: body.Hours, Rate: body.Rate,
		BudgetNotes: body.BudgetNotes,
		OrigBudget: body.OrigBudget, CurrBudget: body.CurrBudget,
		ActualBudget: body.ActualBudget,
	})
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	w.WriteHeader(201)
	jsonOK(w, t)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	q := dbgen.New(s.DB)
	existing, err := q.GetTask(r.Context(), tid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	if !s.requireAuth(w, r, existing.ProjectID) {
		return
	}
	var body taskInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.Status == "" {
		body.Status = "pending"
	}
	if err := q.UpdateTask(r.Context(), dbgen.UpdateTaskParams{
		Assignee: body.Assignee, Title: body.Title,
		IsMilestone: body.IsMilestone,
		OrigWeeks: body.OrigWeeks, CurrWeeks: body.CurrWeeks,
		OrigDue: body.OrigDue, CurrDue: body.CurrDue,
		ActualDone: body.ActualDone, Status: body.Status,
		Words: body.Words, WordsPerHour: body.WordsPerHour,
		Hours: body.Hours, Rate: body.Rate,
		BudgetNotes: body.BudgetNotes,
		OrigBudget: body.OrigBudget, CurrBudget: body.CurrBudget,
		ActualBudget: body.ActualBudget,
		SortOrder: body.SortOrder,
		ID: tid,
	}); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"ok": "true"})
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	q := dbgen.New(s.DB)
	existing, err := q.GetTask(r.Context(), tid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	if !s.requireAuth(w, r, existing.ProjectID) {
		return
	}
	if err := q.DeleteTask(r.Context(), tid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"ok": "true"})
}

// --- Auth handlers ---

func (s *Server) handleSetAuth(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	// Allow setting auth if no auth exists yet, otherwise require existing auth
	q := dbgen.New(s.DB)
	tokens, _ := q.ListAuthTokens(r.Context(), pid)
	if len(tokens) > 0 && !s.checkAuth(r, pid) {
		jsonErr(w, "unauthorized", 401)
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Password == "" {
		jsonErr(w, "password required", 400)
		return
	}
	if err := q.CreateAuthToken(r.Context(), dbgen.CreateAuthTokenParams{
		ProjectID: pid,
		TokenHash: hashToken(body.Password),
		Label:     "shared",
	}); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     fmt.Sprintf("prodcal_auth_%d", pid),
		Value:    body.Password,
		Path:     "/",
		MaxAge:   86400 * 365,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	jsonOK(w, map[string]string{"ok": "true"})
}

func (s *Server) handleVerifyAuth(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Password == "" {
		jsonErr(w, "password required", 400)
		return
	}
	q := dbgen.New(s.DB)
	_, err = q.GetAuthToken(r.Context(), dbgen.GetAuthTokenParams{
		ProjectID: pid,
		TokenHash: hashToken(body.Password),
	})
	if err != nil {
		jsonErr(w, "invalid password", 401)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     fmt.Sprintf("prodcal_auth_%d", pid),
		Value:    body.Password,
		Path:     "/",
		MaxAge:   86400 * 365,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	jsonOK(w, map[string]string{"ok": "true"})
}

// --- Seed handler ---

type seedData struct {
	ProjectName string     `json:"project_name"`
	StartDate   string     `json:"project_start"`
	Tasks       []seedTask `json:"tasks"`
}

type seedTask struct {
	SortOrder    int     `json:"sort_order"`
	Assignee     string  `json:"assignee"`
	Task         string  `json:"task"`
	IsMilestone  bool    `json:"is_milestone"`
	OrigWeeks    float64 `json:"orig_weeks"`
	CurrWeeks    float64 `json:"curr_weeks"`
	OrigDue      string  `json:"orig_due"`
	CurrDue      string  `json:"curr_due"`
	ActualDone   string  `json:"actual_done"`
	Words        int     `json:"words"`
	WordsPerHour int     `json:"words_per_hour"`
	Hours        float64 `json:"hours"`
	Rate         float64 `json:"rate"`
	BudgetNotes  string  `json:"budget_notes"`
	OrigBudget   float64 `json:"orig_budget"`
	CurrBudget   float64 `json:"curr_budget"`
	ActualBudget float64 `json:"actual_budget"`
}

func (s *Server) handleSeedProject(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	var body seedData
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad json: "+err.Error(), 400)
		return
	}
	q := dbgen.New(s.DB)
	for _, t := range body.Tasks {
		milestone := int64(0)
		if t.IsMilestone {
			milestone = 1
		}
		status := "pending"
		if t.ActualDone != "" {
			status = "done"
		}
		_, err := q.CreateTask(r.Context(), dbgen.CreateTaskParams{
			ProjectID: pid, SortOrder: int64(t.SortOrder),
			Assignee: t.Assignee, Title: t.Task,
			IsMilestone: milestone,
			OrigWeeks: t.OrigWeeks, CurrWeeks: t.CurrWeeks,
			OrigDue: t.OrigDue, CurrDue: t.CurrDue,
			ActualDone: t.ActualDone, Status: status,
			Words: int64(t.Words), WordsPerHour: int64(t.WordsPerHour),
			Hours: t.Hours, Rate: t.Rate,
			BudgetNotes: t.BudgetNotes,
			OrigBudget: t.OrigBudget, CurrBudget: t.CurrBudget,
			ActualBudget: t.ActualBudget,
		})
		if err != nil {
			slog.Warn("seed task", "error", err, "task", t.Task)
		}
	}
	// Update project dates if provided
	if body.StartDate != "" {
		p, _ := q.GetProject(r.Context(), pid)
		_ = q.UpdateProject(r.Context(), dbgen.UpdateProjectParams{
			Name: p.Name, StartDate: body.StartDate,
			ClientSlug: p.ClientSlug, ProjectSlug: p.ProjectSlug,
			ID: pid,
		})
	}
	jsonOK(w, map[string]any{"ok": true, "count": len(body.Tasks)})
}

func shiftDate(dateStr string, delta time.Duration) string {
	if dateStr == "" {
		return ""
	}
	// Try parsing with time component first (some dates have 12:00:00)
	for _, layout := range []string{"2006-01-02", "2006-01-02T15:04:05"} {
		if d, err := time.Parse(layout, dateStr); err == nil {
			return d.Add(delta).Format("2006-01-02")
		}
	}
	return dateStr
}

func (s *Server) handleDuplicateProject(w http.ResponseWriter, r *http.Request) {
	srcID, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, srcID) {
		return
	}

	var body struct {
		Name        string `json:"name"`
		StartDate   string `json:"start_date"`
		ClientSlug  string `json:"client_slug"`
		ProjectSlug string `json:"project_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.Name == "" {
		jsonErr(w, "name required", 400)
		return
	}
	if body.ClientSlug == "" || body.ProjectSlug == "" {
		jsonErr(w, "client_slug and project_slug required", 400)
		return
	}

	q := dbgen.New(s.DB)
	srcProject, err := q.GetProject(r.Context(), srcID)
	if err != nil {
		jsonErr(w, "source project not found", 404)
		return
	}

	// Calculate date delta
	var delta time.Duration
	if body.StartDate != "" && srcProject.StartDate != "" {
		oldStart, e1 := time.Parse("2006-01-02", srcProject.StartDate)
		newStart, e2 := time.Parse("2006-01-02", body.StartDate)
		if e1 == nil && e2 == nil {
			delta = newStart.Sub(oldStart)
		}
	}

	// Create new project
	startDate := body.StartDate
	if startDate == "" {
		startDate = srcProject.StartDate
	}
	newProject, err := q.CreateProject(r.Context(), dbgen.CreateProjectParams{
		Name:        body.Name,
		StartDate:   startDate,
		ClientSlug:  body.ClientSlug,
		ProjectSlug: body.ProjectSlug,
	})
	if err != nil {
		jsonErr(w, "create project: "+err.Error(), 500)
		return
	}

	// Copy tasks with shifted dates and zeroed amounts
	tasks, err := q.ListTasks(r.Context(), srcID)
	if err != nil {
		jsonErr(w, "list tasks: "+err.Error(), 500)
		return
	}

	for _, t := range tasks {
		_, err := q.CreateTask(r.Context(), dbgen.CreateTaskParams{
			ProjectID:    newProject.ID,
			SortOrder:    t.SortOrder,
			Assignee:     t.Assignee,
			Title:        t.Title,
			IsMilestone:  t.IsMilestone,
			OrigWeeks:    t.OrigWeeks,
			CurrWeeks:    t.OrigWeeks, // Reset curr to orig
			OrigDue:      shiftDate(t.OrigDue, delta),
			CurrDue:      shiftDate(t.OrigDue, delta), // Reset curr to shifted orig
			ActualDone:   "",       // Clear
			Status:       "pending", // Reset
			Words:        t.Words,
			WordsPerHour: t.WordsPerHour,
			Hours:        t.Hours,
			Rate:         t.Rate,
			BudgetNotes:  t.BudgetNotes,
			OrigBudget:   0, // Zero
			CurrBudget:   0, // Zero
			ActualBudget: 0, // Zero
		})
		if err != nil {
			slog.Warn("duplicate task", "error", err, "task", t.Title)
		}
	}

	jsonOK(w, map[string]any{
		"project": newProject,
		"tasks_copied": len(tasks),
	})
}
