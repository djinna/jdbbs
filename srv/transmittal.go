package srv

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func defaultTransmittalData() string {
	return `{
  "book": {"author":"","title":"","subtitle":"","title_status":"tentative","series":"","publisher":"","editor":"","transmittal_date":"","isbn_paper":"","isbn_cloth":""},
  "production": {"transmittal_date":"","mechs_delivery":"","weeks_in_production":"","bound_book_date":"","print_run":""},
  "checklist": [
    {"component":"Half title pg","here_now":false,"to_come_when":"","indent":false},
    {"component":"Series title/Frontis.","here_now":false,"to_come_when":"","indent":false},
    {"component":"Title pg","here_now":false,"to_come_when":"","indent":false},
    {"component":"Copyright pg","here_now":false,"to_come_when":"","indent":false},
    {"component":"CIP","here_now":false,"to_come_when":"","indent":false},
    {"component":"Dedication","here_now":false,"to_come_when":"","indent":false},
    {"component":"Epigraph","here_now":false,"to_come_when":"","indent":true},
    {"component":"Contents","here_now":false,"to_come_when":"","indent":false},
    {"component":"List of Figures","here_now":false,"to_come_when":"","indent":true},
    {"component":"List of Tables","here_now":false,"to_come_when":"","indent":true},
    {"component":"Foreword","here_now":false,"to_come_when":"","indent":true},
    {"component":"Preface","here_now":false,"to_come_when":"","indent":true},
    {"component":"Acknowledgments","here_now":false,"to_come_when":"","indent":true},
    {"component":"Introduction","here_now":false,"to_come_when":"","indent":false},
    {"component":"Text","here_now":false,"to_come_when":"","indent":false}
  ],
  "checklist_stats": {"parts":"","chapters":"","words_chars":"","ms_pp":"","est_book_pp":""},
  "backmatter": [
    {"component":"Notes","here_now":false,"to_come_when":"","subtype":"ft/end"},
    {"component":"Appendix(es)","here_now":false,"to_come_when":""},
    {"component":"Glossary","here_now":false,"to_come_when":""},
    {"component":"Bibliography","here_now":false,"to_come_when":""},
    {"component":"Index","here_now":false,"to_come_when":""},
    {"component":"Other BM","here_now":false,"to_come_when":""}
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
		// Auto-create a blank transmittal
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

		// Re-read to get the timestamps
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
		// No existing row — INSERT
		_, err = s.DB.ExecContext(r.Context(),
			`INSERT INTO transmittals (project_id, status, data) VALUES (?, ?, ?)`,
			pid, body.Status, dataStr,
		)
		if err != nil {
			jsonErr(w, "insert failed: "+err.Error(), 500)
			return
		}
	}

	jsonOK(w, map[string]any{"ok": true})
}
