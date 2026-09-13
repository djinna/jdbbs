package srv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ─── Admin content review: /admin/content-review/ ───
//
// A queue of proposed copy changes (from a better-documents review, see
// docs/reviews/CONTENT-REVIEW-*.md) that the admin accepts, edits, or
// rejects. The page proposes; applying accepted changes to the reviewed
// files is done by hand from the export. Nothing here writes to those files.

type contentReviewItem struct {
	ID         int64  `json:"id"`
	Batch      string `json:"batch"`
	Ord        int    `json:"ord"`
	Ref        string `json:"ref"`
	Page       string `json:"page"`
	File       string `json:"file"`
	Location   string `json:"location"`
	Current    string `json:"current"`
	Proposed   string `json:"proposed"`
	Reason     string `json:"reason"`
	Severity   string `json:"severity"`
	Pass       int    `json:"pass"`
	Decision   string `json:"decision"`
	EditedText string `json:"edited_text"`
	Note       string `json:"note"`
	DecidedAt  string `json:"decided_at,omitempty"`
	AppliedAt  string `json:"applied_at,omitempty"`
}

func (s *Server) handleAdminContentReviewPage(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdmin(w, r) {
		return
	}
	s.serveStaticHTML(w, "static/content-review.html")
}

func (s *Server) listContentReviewItems(batch string) ([]contentReviewItem, error) {
	q := `SELECT id, batch, ord, ref, page, file, location, current_text, proposed, reason, severity, pass,
	             decision, edited_text, note, COALESCE(decided_at,''), COALESCE(applied_at,'')
	      FROM content_review_items`
	args := []any{}
	if batch != "" {
		q += ` WHERE batch = ?`
		args = append(args, batch)
	}
	q += ` ORDER BY batch DESC, ord, id`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []contentReviewItem{}
	for rows.Next() {
		var it contentReviewItem
		if err := rows.Scan(&it.ID, &it.Batch, &it.Ord, &it.Ref, &it.Page, &it.File, &it.Location, &it.Current,
			&it.Proposed, &it.Reason, &it.Severity, &it.Pass, &it.Decision, &it.EditedText, &it.Note,
			&it.DecidedAt, &it.AppliedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// GET /api/admin/content-review[?batch=YYYY-MM-DD]
func (s *Server) handleAdminListContentReview(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	items, err := s.listContentReviewItems(r.URL.Query().Get("batch"))
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{"items": items})
}

// POST /api/admin/content-review — import a batch (array of items). Existing
// (batch, ref) pairs are left alone so re-importing never clobbers decisions.
func (s *Server) handleAdminImportContentReview(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	var in struct {
		Batch string              `json:"batch"`
		Items []contentReviewItem `json:"items"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		jsonErr(w, "invalid import", http.StatusBadRequest)
		return
	}
	in.Batch = strings.TrimSpace(in.Batch)
	if in.Batch == "" || len(in.Items) == 0 {
		jsonErr(w, "batch and items required", http.StatusBadRequest)
		return
	}
	added := 0
	for i, it := range in.Items {
		var exists int
		_ = s.DB.QueryRow(`SELECT COUNT(*) FROM content_review_items WHERE batch=? AND ref=?`, in.Batch, it.Ref).Scan(&exists)
		if exists > 0 {
			continue
		}
		if it.Severity == "" {
			it.Severity = "MINOR"
		}
		if it.Pass == 0 {
			it.Pass = 1
		}
		_, err := s.DB.Exec(`INSERT INTO content_review_items
			(batch, ord, ref, page, file, location, current_text, proposed, reason, severity, pass)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			in.Batch, i+1, it.Ref, it.Page, it.File, it.Location, it.Current, it.Proposed, it.Reason, it.Severity, it.Pass)
		if err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		added++
	}
	jsonOK(w, map[string]any{"added": added})
}

// PUT /api/admin/content-review/{id} — {decision, edited_text, note}
func (s *Server) handleAdminDecideContentReview(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "invalid id", http.StatusBadRequest)
		return
	}
	var in struct {
		Decision   string `json:"decision"`
		EditedText string `json:"edited_text"`
		Note       string `json:"note"`
		Applied    *bool  `json:"applied"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid update", http.StatusBadRequest)
		return
	}
	switch in.Decision {
	case "pending", "accepted", "edited", "rejected":
	default:
		jsonErr(w, "decision must be pending|accepted|edited|rejected", http.StatusBadRequest)
		return
	}
	if in.Decision == "edited" && strings.TrimSpace(in.EditedText) == "" {
		jsonErr(w, "edited decision needs edited_text", http.StatusBadRequest)
		return
	}
	var decided any
	if in.Decision != "pending" {
		decided = time.Now().UTC().Format(time.RFC3339)
	}
	res, err := s.DB.Exec(`UPDATE content_review_items
		SET decision=?, edited_text=?, note=?, decided_at=? WHERE id=?`,
		in.Decision, in.EditedText, in.Note, decided, id)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		jsonErr(w, "not found", http.StatusNotFound)
		return
	}
	if in.Applied != nil {
		var applied any
		if *in.Applied {
			applied = time.Now().UTC().Format(time.RFC3339)
		}
		_, _ = s.DB.Exec(`UPDATE content_review_items SET applied_at=? WHERE id=?`, applied, id)
	}
	items, err := s.listContentReviewItems("")
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	for _, it := range items {
		if it.ID == id {
			jsonOK(w, it)
			return
		}
	}
	jsonErr(w, "not found", http.StatusNotFound)
}

// GET /api/admin/content-review/export[?batch=] — Markdown patch list of
// accepted + edited items, for applying by hand.
func (s *Server) handleAdminExportContentReview(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	items, err := s.listContentReviewItems(r.URL.Query().Get("batch"))
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Content review — accepted changes\n\n")
	n := 0
	for _, it := range items {
		if it.Decision != "accepted" && it.Decision != "edited" {
			continue
		}
		n++
		text := it.Proposed
		if it.Decision == "edited" {
			text = it.EditedText
		}
		fmt.Fprintf(&b, "## %s · %s · %s\n\n", it.Ref, it.Page, it.Severity)
		fmt.Fprintf(&b, "**File:** %s  \n**Location:** %s  \n**Decision:** %s", it.File, it.Location, it.Decision)
		if it.AppliedAt != "" {
			fmt.Fprintf(&b, " · applied %s", it.AppliedAt)
		}
		fmt.Fprintf(&b, "\n\n**Current:**\n\n```\n%s\n```\n\n**Change to:**\n\n```\n%s\n```\n", it.Current, text)
		if strings.TrimSpace(it.Note) != "" {
			fmt.Fprintf(&b, "\n**Note:** %s\n", it.Note)
		}
		b.WriteString("\n")
	}
	if n == 0 {
		b.WriteString("_No accepted or edited items yet._\n")
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Write([]byte(b.String()))
}
