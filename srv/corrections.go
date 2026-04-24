package srv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type correctionEntry struct {
	ID          int64  `json:"id"`
	ProjectID   int64  `json:"project_id"`
	FindText    string `json:"find_text"`
	ReplaceText string `json:"replace_text"`
	Chapter     string `json:"chapter"`
	Note        string `json:"note"`
	Status      string `json:"status"`
	AppliedAt   string `json:"applied_at"`
	CreatedAt   string `json:"created_at"`
}

func (s *Server) handleListCorrections(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, project_id, find_text, replace_text,
		       COALESCE(chapter, ''), COALESCE(note, ''),
		       status, COALESCE(applied_at, ''), created_at
		FROM corrections WHERE project_id = ? ORDER BY created_at DESC
	`, pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	entries := []correctionEntry{}
	for rows.Next() {
		var e correctionEntry
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.FindText, &e.ReplaceText, &e.Chapter, &e.Note, &e.Status, &e.AppliedAt, &e.CreatedAt); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		entries = append(entries, e)
	}
	jsonOK(w, entries)
}

func (s *Server) handleCreateCorrection(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	var body struct {
		FindText    string `json:"find_text"`
		ReplaceText string `json:"replace_text"`
		Chapter     string `json:"chapter"`
		Note        string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.FindText == "" {
		jsonErr(w, "find_text required", 400)
		return
	}
	var e correctionEntry
	err = s.DB.QueryRowContext(r.Context(), `
		INSERT INTO corrections (project_id, find_text, replace_text, chapter, note)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, project_id, find_text, replace_text,
		          COALESCE(chapter, ''), COALESCE(note, ''),
		          status, COALESCE(applied_at, ''), created_at
	`, pid, body.FindText, body.ReplaceText, body.Chapter, body.Note,
	).Scan(&e.ID, &e.ProjectID, &e.FindText, &e.ReplaceText, &e.Chapter, &e.Note, &e.Status, &e.AppliedAt, &e.CreatedAt)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	w.WriteHeader(201)
	jsonOK(w, e)
}

func (s *Server) handleDeleteCorrection(w http.ResponseWriter, r *http.Request) {
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
	_, err = s.DB.ExecContext(r.Context(), `DELETE FROM corrections WHERE id = ? AND project_id = ?`, entryID, pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"ok": "true"})
}

func (s *Server) handleUpdateCorrectionStatus(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	// Look up the correction to find its project for auth
	var pid int64
	err = s.DB.QueryRowContext(r.Context(), `SELECT project_id FROM corrections WHERE id = ?`, cid).Scan(&pid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.Status != "pending" && body.Status != "applied" && body.Status != "skipped" {
		jsonErr(w, "status must be pending, applied, or skipped", 400)
		return
	}
	var appliedAt interface{}
	if body.Status == "applied" {
		appliedAt = nowDate()
	}
	_, err = s.DB.ExecContext(r.Context(), `
		UPDATE corrections SET status = ?, applied_at = ? WHERE id = ?
	`, body.Status, appliedAt, cid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"ok": "true"})
}

func (s *Server) handleExportCorrections(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT find_text, replace_text, COALESCE(chapter, ''), COALESCE(note, ''), status
		FROM corrections WHERE project_id = ? ORDER BY created_at ASC
	`, pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("# Corrections export\n")
	sb.WriteString("# Generated from prodcal admin\n\n")
	sb.WriteString("corrections:\n")

	count := 0
	for rows.Next() {
		var find, replace, chapter, note, status string
		if err := rows.Scan(&find, &replace, &chapter, &note, &status); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		sb.WriteString(fmt.Sprintf("  - find: %q\n", find))
		sb.WriteString(fmt.Sprintf("    replace: %q\n", replace))
		if chapter != "" {
			sb.WriteString(fmt.Sprintf("    chapter: %q\n", chapter))
		}
		if note != "" {
			sb.WriteString(fmt.Sprintf("    # %s\n", note))
		}
		if status != "pending" {
			sb.WriteString(fmt.Sprintf("    # status: %s\n", status))
		}
		count++
	}

	if count == 0 {
		sb.WriteString("  [] # no corrections\n")
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="corrections-project-%d.yaml"`, pid))
	w.Write([]byte(sb.String()))
}
