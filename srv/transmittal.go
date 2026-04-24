package srv

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

func defaultTransmittalData() string {
	return `{
  "book": {"author":"","title":"","subtitle":"","title_status":"tentative","series":"","publisher":"","editor":"","transmittal_date":"","isbn_paper":"","isbn_epub":"","isbn_cloth":""},
  "production": {"transmittal_date":"","mechs_delivery":"","weeks_in_production":"","bound_book_date":"","print_run":""},
  "checklist": [
    {"component":"Half title pg","status":"","here_now":false,"to_come_when":"","indent":false},
    {"component":"Series title/Frontis.","status":"","here_now":false,"to_come_when":"","indent":false},
    {"component":"Title pg","status":"","here_now":false,"to_come_when":"","indent":false},
    {"component":"Copyright pg","status":"","here_now":false,"to_come_when":"","indent":false},
    {"component":"CIP","status":"","here_now":false,"to_come_when":"","indent":false},
    {"component":"Dedication","status":"","here_now":false,"to_come_when":"","indent":false},
    {"component":"Epigraph","status":"","here_now":false,"to_come_when":"","indent":true},
    {"component":"Contents","status":"","here_now":false,"to_come_when":"","indent":false},
    {"component":"List of Figures","status":"","here_now":false,"to_come_when":"","indent":true},
    {"component":"List of Tables","status":"","here_now":false,"to_come_when":"","indent":true},
    {"component":"Foreword","status":"","here_now":false,"to_come_when":"","indent":true},
    {"component":"Preface","status":"","here_now":false,"to_come_when":"","indent":true},
    {"component":"Acknowledgments","status":"","here_now":false,"to_come_when":"","indent":true},
    {"component":"Introduction","status":"","here_now":false,"to_come_when":"","indent":false},
    {"component":"Text","status":"","here_now":false,"to_come_when":"","indent":false}
  ],
  "checklist_stats": {"parts":"","chapters":"","words_chars":"","ms_pp":"","est_book_pp":""},
  "backmatter": [
    {"component":"Notes","status":"","here_now":false,"to_come_when":"","subtype":"ft/end"},
    {"component":"Appendix(es)","status":"","here_now":false,"to_come_when":""},
    {"component":"Glossary","status":"","here_now":false,"to_come_when":""},
    {"component":"Bibliography","status":"","here_now":false,"to_come_when":""},
    {"component":"Index","status":"","here_now":false,"to_come_when":""},
    {"component":"Other BM","status":"","here_now":false,"to_come_when":""}
  ],
  "illustrations": {"figures_no":0,"figures_here":false,"figures_to_come":"","tables_no":0,"tables_here":false,"tables_to_come":"","photos_no":0,"photos_here":false,"photos_to_come":"","other_no":0,"other_here":false,"other_to_come":"","art_plan":""},
  "permissions": {"reprint_status":"","reprint_when":"","consents_status":"","consents_when":""},
  "page_iv": {"copyright_year":"","held_by":"","credit":"","other_credit":"","photo_credit":""},
  "subrights": {"copub":"na","title_page":"na","page_iv":"na","cover":"na","remove_mktg":"na"},
  "editing": {"copyediting_level":"","special_characters":"","math_formulas":"","instructions":""},
  "design": {"trim":"","est_pages":"","ppi":"","spine_width":"","complexity":"","outside_designer":"","reuse_previous":""},
  "cover": {"paper":"","colors":"","jdbb_front":false,"jdbb_spine":false,"jdbb_back":false,"pub_front":false,"pub_spine":false,"pub_back":false,"credit":""},
  "files": {"printer_format":"","archives":[]},
  "proofs": {"reviewers":[]},
  "custom_styles": [],
  "other_instructions": ""
}`
}

func (s *Server) handleGetTransmittal(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}

	var id int64
	var projectID int64
	var status string
	var dataStr string
	var createdAt string
	var updatedAt string

	err = s.DB.QueryRowContext(r.Context(),
		`SELECT id, project_id, status, data, created_at, updated_at
		 FROM transmittals WHERE project_id = ?`, pid,
	).Scan(&id, &projectID, &status, &dataStr, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		defData := defaultTransmittalData()
		result, insertErr := s.DB.ExecContext(r.Context(),
			`INSERT INTO transmittals (project_id, status, data) VALUES (?, 'draft', ?)`,
			pid, defData,
		)
		if insertErr != nil {
			jsonErr(w, "failed to create transmittal: "+insertErr.Error(), 500)
			return
		}
		id, _ = result.LastInsertId()
		err = s.DB.QueryRowContext(r.Context(),
			`SELECT id, project_id, status, data, created_at, updated_at
			 FROM transmittals WHERE id = ?`, id,
		).Scan(&id, &projectID, &status, &dataStr, &createdAt, &updatedAt)
		if err != nil {
			jsonErr(w, "failed to read transmittal: "+err.Error(), 500)
			return
		}
	} else if err != nil {
		jsonErr(w, "database error: "+err.Error(), 500)
		return
	}

	var raw json.RawMessage
	if jsonErr := json.Unmarshal([]byte(dataStr), &raw); jsonErr != nil {
		raw = json.RawMessage(dataStr)
	}

	jsonOK(w, map[string]any{
		"id":         id,
		"project_id": projectID,
		"status":     status,
		"data":       raw,
		"created_at": createdAt,
		"updated_at": updatedAt,
	})
}

// maybeSnapshotVersion saves the current transmittal state as a version
// if no version was saved in the last 5 minutes (throttle auto-save noise).
func (s *Server) maybeSnapshotVersion(r *http.Request, transmittalID int64, data, status string) {
	var count int
	_ = s.DB.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM transmittal_versions
		 WHERE transmittal_id = ? AND saved_at > datetime('now', '-5 minutes')`,
		transmittalID,
	).Scan(&count)
	if count > 0 {
		return // too recent, skip
	}
	s.DB.ExecContext(r.Context(),
		`INSERT INTO transmittal_versions (transmittal_id, data, status) VALUES (?, ?, ?)`,
		transmittalID, data, status,
	)
}

func (s *Server) handleUpdateTransmittal(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}

	var body struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}

	dataStr := string(body.Data)
	if dataStr == "" || dataStr == "null" {
		dataStr = defaultTransmittalData()
	}
	if body.Status == "" {
		body.Status = "draft"
	}

	// Snapshot current state as a version before updating
	var txID int64
	var oldData, oldStatus string
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT id, data, status FROM transmittals WHERE project_id = ?`, pid,
	).Scan(&txID, &oldData, &oldStatus)
	if err == nil {
		s.maybeSnapshotVersion(r, txID, oldData, oldStatus)
	}

	// Try UPDATE first
	result, err := s.DB.ExecContext(r.Context(),
		`UPDATE transmittals SET data = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE project_id = ?`,
		dataStr, body.Status, pid,
	)
	if err != nil {
		jsonErr(w, "update failed: "+err.Error(), 500)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		_, err = s.DB.ExecContext(r.Context(),
			`INSERT INTO transmittals (project_id, status, data) VALUES (?, ?, ?)`,
			pid, body.Status, dataStr,
		)
		if err != nil {
			jsonErr(w, "insert failed: "+err.Error(), 500)
			return
		}
	}

	// Notify admin of client transmittal update (throttled, background)
	// Skip notification if this is an admin/exe.dev user editing
	if r.Header.Get("X-ExeDev-UserID") == "" {
		txNotifier.maybeNotify(s, pid)
	}

	jsonOK(w, map[string]any{"ok": true})
}

// ─── Version history handlers ───

func (s *Server) handleListTransmittalVersions(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}

	rows, err := s.DB.QueryContext(r.Context(),
		`SELECT v.id, v.status, v.data, v.saved_at
		 FROM transmittal_versions v
		 JOIN transmittals t ON t.id = v.transmittal_id
		 WHERE t.project_id = ?
		 ORDER BY v.saved_at DESC
		 LIMIT 100`, pid,
	)
	if err != nil {
		jsonErr(w, "query failed: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	type versionSummary struct {
		ID      int64  `json:"id"`
		Status  string `json:"status"`
		Title   string `json:"title"`
		SavedAt string `json:"saved_at"`
	}
	var versions []versionSummary
	for rows.Next() {
		var v versionSummary
		var dataStr string
		if err := rows.Scan(&v.ID, &v.Status, &dataStr, &v.SavedAt); err != nil {
			continue
		}
		// Extract book title from JSON for the summary
		var parsed struct {
			Book struct {
				Title string `json:"title"`
			} `json:"book"`
		}
		json.Unmarshal([]byte(dataStr), &parsed)
		v.Title = parsed.Book.Title
		versions = append(versions, v)
	}
	if versions == nil {
		versions = []versionSummary{}
	}
	jsonOK(w, versions)
}

func (s *Server) handleGetTransmittalVersion(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	vid, err := strconv.ParseInt(r.PathValue("vid"), 10, 64)
	if err != nil {
		jsonErr(w, "bad version id", 400)
		return
	}

	var dataStr, status, savedAt string
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT v.data, v.status, v.saved_at
		 FROM transmittal_versions v
		 JOIN transmittals t ON t.id = v.transmittal_id
		 WHERE v.id = ? AND t.project_id = ?`, vid, pid,
	).Scan(&dataStr, &status, &savedAt)
	if err != nil {
		jsonErr(w, "version not found", 404)
		return
	}

	var raw json.RawMessage
	json.Unmarshal([]byte(dataStr), &raw)

	jsonOK(w, map[string]any{
		"id":       vid,
		"status":   status,
		"data":     raw,
		"saved_at": savedAt,
	})
}

func (s *Server) handleRestoreTransmittalVersion(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	vid, err := strconv.ParseInt(r.PathValue("vid"), 10, 64)
	if err != nil {
		jsonErr(w, "bad version id", 400)
		return
	}

	// Get the version data
	var vData, vStatus string
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT v.data, v.status
		 FROM transmittal_versions v
		 JOIN transmittals t ON t.id = v.transmittal_id
		 WHERE v.id = ? AND t.project_id = ?`, vid, pid,
	).Scan(&vData, &vStatus)
	if err != nil {
		jsonErr(w, "version not found", 404)
		return
	}

	// Always snapshot current state before restoring (bypass throttle)
	var txID int64
	var oldData, oldStatus string
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT id, data, status FROM transmittals WHERE project_id = ?`, pid,
	).Scan(&txID, &oldData, &oldStatus)
	if err == nil {
		s.DB.ExecContext(r.Context(),
			`INSERT INTO transmittal_versions (transmittal_id, data, status) VALUES (?, ?, ?)`,
			txID, oldData, oldStatus,
		)
	}

	// Restore
	_, err = s.DB.ExecContext(r.Context(),
		`UPDATE transmittals SET data = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE project_id = ?`, vData, vStatus, pid,
	)
	if err != nil {
		jsonErr(w, "restore failed: "+err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{"ok": true})
}

// ─── Duplicate transmittal ───

func (s *Server) handleDuplicateTransmittal(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}

	var body struct {
		TargetProjectID int64 `json:"target_project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TargetProjectID == 0 {
		jsonErr(w, "target_project_id required", 400)
		return
	}

	// Check auth on target too
	if !s.requireAuth(w, r, body.TargetProjectID) {
		return
	}

	// Check target doesn't already have a transmittal
	var existingCount int
	s.DB.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM transmittals WHERE project_id = ?`, body.TargetProjectID,
	).Scan(&existingCount)
	if existingCount > 0 {
		jsonErr(w, "target project already has a transmittal", 409)
		return
	}

	// Get source transmittal data
	var srcData string
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT data FROM transmittals WHERE project_id = ?`, pid,
	).Scan(&srcData)
	if err != nil {
		jsonErr(w, "source transmittal not found", 404)
		return
	}

	// Parse, clear book-specific fields, keep publisher/house defaults
	var d map[string]any
	json.Unmarshal([]byte(srcData), &d)

	// Clear book-specific fields
	if book, ok := d["book"].(map[string]any); ok {
		book["title"] = ""
		book["subtitle"] = ""
		book["isbn_paper"] = ""
		book["isbn_epub"] = ""
		book["isbn_cloth"] = ""
		book["transmittal_date"] = ""
		book["series"] = ""
		// Keep: author, publisher, editor, title_status
	}
	if prod, ok := d["production"].(map[string]any); ok {
		prod["transmittal_date"] = ""
		prod["mechs_delivery"] = ""
		prod["bound_book_date"] = ""
		prod["weeks_in_production"] = ""
		// Keep: print_run
	}
	if stats, ok := d["checklist_stats"].(map[string]any); ok {
		for k := range stats {
			stats[k] = ""
		}
	}
	// Reset checklist items
	if cl, ok := d["checklist"].([]any); ok {
		for _, item := range cl {
			if m, ok := item.(map[string]any); ok {
				m["status"] = ""
				m["here_now"] = false
				m["to_come_when"] = ""
			}
		}
	}
	if bm, ok := d["backmatter"].([]any); ok {
		for _, item := range bm {
			if m, ok := item.(map[string]any); ok {
				m["status"] = ""
				m["here_now"] = false
				m["to_come_when"] = ""
			}
		}
	}
	// Reset illustrations
	if ill, ok := d["illustrations"].(map[string]any); ok {
		for k := range ill {
			if k == "art_plan" {
				ill[k] = ""
			} else if k[len(k)-3:] == "_no" {
				ill[k] = float64(0)
			} else if k[len(k)-5:] == "_here" {
				ill[k] = false
			} else {
				ill[k] = ""
			}
		}
	}
	// Reset proofs and other
	if proofs, ok := d["proofs"].(map[string]any); ok {
		proofs["reviewers"] = []any{}
	}
	d["other_instructions"] = ""
	// Keep: permissions, page_iv, subrights, editing, design, cover, files

	newData, _ := json.Marshal(d)
	_, err = s.DB.ExecContext(r.Context(),
		`INSERT INTO transmittals (project_id, status, data) VALUES (?, 'draft', ?)`,
		body.TargetProjectID, string(newData),
	)
	if err != nil {
		jsonErr(w, "duplicate failed: "+err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{"ok": true, "target_project_id": body.TargetProjectID})
}
