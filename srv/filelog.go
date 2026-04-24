package srv

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type fileLogEntry struct {
	ID           int64  `json:"id"`
	ProjectID    int64  `json:"project_id"`
	Direction    string `json:"direction"`
	Filename     string `json:"filename"`
	FileType     string `json:"file_type"`
	SentBy       string `json:"sent_by"`
	ReceivedBy   string `json:"received_by"`
	Notes        string `json:"notes"`
	TransferDate string `json:"transfer_date"`
	CreatedAt    string `json:"created_at"`
}

func (s *Server) handleListFileLog(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, project_id, direction, filename, file_type, sent_by, received_by, notes, transfer_date, created_at
		FROM file_log WHERE project_id = ? ORDER BY transfer_date DESC, created_at DESC
	`, pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	entries := []fileLogEntry{}
	for rows.Next() {
		var e fileLogEntry
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Direction, &e.Filename, &e.FileType, &e.SentBy, &e.ReceivedBy, &e.Notes, &e.TransferDate, &e.CreatedAt); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		entries = append(entries, e)
	}
	jsonOK(w, entries)
}

func (s *Server) handleCreateFileLog(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	var body struct {
		Direction    string `json:"direction"`
		Filename     string `json:"filename"`
		FileType     string `json:"file_type"`
		SentBy       string `json:"sent_by"`
		ReceivedBy   string `json:"received_by"`
		Notes        string `json:"notes"`
		TransferDate string `json:"transfer_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.Direction == "" {
		body.Direction = "inbound"
	}
	if body.TransferDate == "" {
		body.TransferDate = nowDate()
	}
	var e fileLogEntry
	err = s.DB.QueryRowContext(r.Context(), `
		INSERT INTO file_log (project_id, direction, filename, file_type, sent_by, received_by, notes, transfer_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, project_id, direction, filename, file_type, sent_by, received_by, notes, transfer_date, created_at
	`, pid, body.Direction, body.Filename, body.FileType, body.SentBy, body.ReceivedBy, body.Notes, body.TransferDate,
	).Scan(&e.ID, &e.ProjectID, &e.Direction, &e.Filename, &e.FileType, &e.SentBy, &e.ReceivedBy, &e.Notes, &e.TransferDate, &e.CreatedAt)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	w.WriteHeader(201)
	jsonOK(w, e)
}

func (s *Server) handleDeleteFileLog(w http.ResponseWriter, r *http.Request) {
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
	_, err = s.DB.ExecContext(r.Context(), `DELETE FROM file_log WHERE id = ? AND project_id = ?`, entryID, pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"ok": "true"})
}

// --- Client-level file log (recent across all projects) ---

func (s *Server) handleClientFileLog(w http.ResponseWriter, r *http.Request) {
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
		SELECT f.id, f.project_id, f.direction, f.filename, f.file_type, f.sent_by, f.received_by, f.notes, f.transfer_date, f.created_at,
		       p.name as project_name
		FROM file_log f
		JOIN projects p ON p.id = f.project_id
		WHERE p.client_slug = ?
		ORDER BY f.transfer_date DESC, f.created_at DESC
		LIMIT ?
	`, clientSlug, limit)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	type clientFileEntry struct {
		fileLogEntry
		ProjectName string `json:"project_name"`
	}
	entries := []clientFileEntry{}
	for rows.Next() {
		var e clientFileEntry
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Direction, &e.Filename, &e.FileType, &e.SentBy, &e.ReceivedBy, &e.Notes, &e.TransferDate, &e.CreatedAt, &e.ProjectName); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		entries = append(entries, e)
	}
	jsonOK(w, entries)
}
