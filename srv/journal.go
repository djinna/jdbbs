package srv

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type journalEntry struct {
	ID        int64  `json:"id"`
	ProjectID int64  `json:"project_id"`
	EntryType string `json:"entry_type"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

func nowDate() string {
	return time.Now().Format("2006-01-02")
}

func (s *Server) handleListJournal(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, project_id, entry_type, content, created_at
		FROM journal WHERE project_id = ? ORDER BY created_at DESC
	`, pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	entries := []journalEntry{}
	for rows.Next() {
		var e journalEntry
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.EntryType, &e.Content, &e.CreatedAt); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		entries = append(entries, e)
	}
	jsonOK(w, entries)
}

func (s *Server) handleCreateJournal(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	var body struct {
		EntryType string `json:"entry_type"`
		Content   string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.EntryType == "" {
		body.EntryType = "note"
	}
	if body.Content == "" {
		jsonErr(w, "content required", 400)
		return
	}
	var e journalEntry
	err = s.DB.QueryRowContext(r.Context(), `
		INSERT INTO journal (project_id, entry_type, content)
		VALUES (?, ?, ?)
		RETURNING id, project_id, entry_type, content, created_at
	`, pid, body.EntryType, body.Content,
	).Scan(&e.ID, &e.ProjectID, &e.EntryType, &e.Content, &e.CreatedAt)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	w.WriteHeader(201)
	jsonOK(w, e)
}

func (s *Server) handleDeleteJournal(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	entryID, err := strconv.ParseInt(r.PathValue("entry"), 10, 64)
	if err != nil {
		jsonErr(w, "bad entry id", 400)
		return
	}
	_, err = s.DB.ExecContext(r.Context(), `DELETE FROM journal WHERE id = ? AND project_id = ?`, entryID, pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"ok": "true"})
}

// --- Client-level journal (recent across all projects) ---

func (s *Server) handleClientJournal(w http.ResponseWriter, r *http.Request) {
	clientSlug := r.PathValue("client")
	if !s.checkClientAuthOrProjectAuth(w, r, clientSlug) {
		return
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
		limit = n
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT j.id, j.project_id, j.entry_type, j.content, j.created_at,
		       p.name as project_name
		FROM journal j
		JOIN projects p ON p.id = j.project_id
		WHERE p.client_slug = ?
		ORDER BY j.created_at DESC
		LIMIT ?
	`, clientSlug, limit)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	type clientJournalEntry struct {
		journalEntry
		ProjectName string `json:"project_name"`
	}
	entries := []clientJournalEntry{}
	for rows.Next() {
		var e clientJournalEntry
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.EntryType, &e.Content, &e.CreatedAt, &e.ProjectName); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		entries = append(entries, e)
	}
	jsonOK(w, entries)
}

// checkClientAuthOrProjectAuth verifies client-level or project-level auth,
// and sends 401 if neither is present.
func (s *Server) checkClientAuthOrProjectAuth(w http.ResponseWriter, r *http.Request, clientSlug string) bool {
	var passwordHash string
	err := s.DB.QueryRowContext(r.Context(),
		`SELECT password_hash FROM clients WHERE slug = ?`, clientSlug,
	).Scan(&passwordHash)
	if err != nil {
		jsonErr(w, "client not found", 404)
		return false
	}
	if passwordHash == "" {
		return true // no auth required
	}
	if s.checkClientAuth(r, clientSlug) {
		return true
	}
	if s.hasAnyProjectAuthForClient(r, clientSlug) {
		return true
	}
	if r.Header.Get("X-ExeDev-UserID") != "" {
		return true
	}
	jsonErr(w, "unauthorized", http.StatusUnauthorized)
	return false
}
