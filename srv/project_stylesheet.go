package srv

// Per-project editorial style sheets (punch list 6.1 B).
//
// The house sheet (housestyle.go, table house_style) is the studio default.
// Every project gets its own *instance* of it: rows are copied in on first
// open, and the client accepts our defaults (the encouraged path — seeded
// rows are already 'accepted'), rejects or edits individual rules, and adds
// their own. The copyeditor reads the effective sheet as Markdown at
// /{client}/{project}/stylesheet/index.md.
//
// Every JSON handler here goes through s.requireAuth(w, r, projectID) — the
// same gate as the transmittal. The HTML shell carries no data.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	pssStatusAccepted = "accepted"
	pssStatusRejected = "rejected"
	pssStatusEdited   = "edited"
	pssStatusAdded    = "added"

	pssAddedSection    = "Our additions"
	pssAddedSectionOrd = 90
)

type pssText struct {
	Col1 string `json:"col1"`
	Col2 string `json:"col2"`
	Col3 string `json:"col3"`
	Body string `json:"body"`
}

type pssItem struct {
	ID         int64    `json:"id"`
	HouseID    *int64   `json:"house_id"`
	SectionOrd int      `json:"section_ord"`
	Section    string   `json:"section"`
	ItemOrd    int      `json:"item_ord"`
	Kind       string   `json:"kind"`
	Col1       string   `json:"col1"`
	Col2       string   `json:"col2"`
	Col3       string   `json:"col3"`
	Body       string   `json:"body"`
	Status     string   `json:"status"`
	Note       string   `json:"note"`
	House      *pssText `json:"house,omitempty"` // the studio original, when it differs
}

type pssSection struct {
	Ord   int       `json:"ord"`
	Title string    `json:"title"`
	Items []pssItem `json:"items"`
}

type pssCounts struct {
	House    int `json:"house"`
	Accepted int `json:"accepted"`
	Edited   int `json:"edited"`
	Rejected int `json:"rejected"`
	Added    int `json:"added"`
}

type pssSheet struct {
	ProjectID   int64        `json:"project_id"`
	ProjectName string       `json:"project_name"`
	ClientSlug  string       `json:"client_slug"`
	ProjectSlug string       `json:"project_slug"`
	BookKind    string       `json:"book_kind"`
	KindChosen  bool         `json:"kind_chosen"`
	Counts      pssCounts    `json:"counts"`
	Sections    []pssSection `json:"sections"`
	UpdatedAt   string       `json:"updated_at"`
}

func (s *Server) registerProjectStylesheetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/projects/{id}/stylesheet", s.handleGetProjectStylesheet)
	mux.HandleFunc("PATCH /api/projects/{id}/stylesheet", s.handlePatchProjectStylesheet)
	mux.HandleFunc("POST /api/projects/{id}/stylesheet/accept-all", s.handleProjectStylesheetAcceptAll)
	mux.HandleFunc("POST /api/projects/{id}/stylesheet/items", s.handleProjectStylesheetAddItem)
	mux.HandleFunc("PATCH /api/projects/{id}/stylesheet/items/{item}", s.handleProjectStylesheetPatchItem)
	mux.HandleFunc("DELETE /api/projects/{id}/stylesheet/items/{item}", s.handleProjectStylesheetDeleteItem)
	mux.HandleFunc("POST /api/projects/{id}/stylesheet/items/{item}/restore", s.handleProjectStylesheetRestoreItem)
}

// pssKind normalises a book kind. Anything unknown means "both" — the
// conservative seed (rules that apply to fiction and nonfiction alike).
func pssKind(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "fiction":
		return "fiction"
	case "nonfiction":
		return "nonfiction"
	}
	return "both"
}

// pssEnsure creates the project's sheet if it is absent and makes sure the
// rows for its current book kind are present. Safe to call on every read.
func (s *Server) pssEnsure(ctx context.Context, projectID int64) (string, error) {
	if err := s.hsEnsureSeeded(ctx); err != nil {
		return "", err
	}
	var kind string
	err := s.DB.QueryRowContext(ctx,
		`SELECT book_kind FROM project_stylesheets WHERE project_id = ?`, projectID).Scan(&kind)
	if err == sql.ErrNoRows {
		kind = "both"
		if _, err := s.DB.ExecContext(ctx,
			`INSERT INTO project_stylesheets (project_id, book_kind) VALUES (?, ?)`,
			projectID, kind); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	if err := s.pssSeedKind(ctx, projectID, kind); err != nil {
		return "", err
	}
	return kind, nil
}

// pssSeedKind copies in every house row for 'both' plus the chosen kind that
// the project does not already have. It never touches existing rows, so
// changing kind adds and never overwrites a client decision.
func (s *Server) pssSeedKind(ctx context.Context, projectID int64, kind string) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO project_stylesheet_items
		    (project_id, house_id, section_ord, section, item_ord, kind, col1, col2, col3, body, status)
		 SELECT ?, h.id, h.section_ord, h.section, h.item_ord, h.kind, h.col1, h.col2, h.col3, h.body, 'accepted'
		   FROM house_style h
		  WHERE (h.book_kind = 'both' OR h.book_kind = ?)
		    AND NOT EXISTS (SELECT 1 FROM project_stylesheet_items i
		                     WHERE i.project_id = ? AND i.house_id = h.id)`,
		projectID, kind, projectID)
	return err
}

// pssPruneOtherKinds removes rows belonging to the *other* book kind, but
// only while they are untouched (still 'accepted' with no note — status is
// what tracks an edit). A client who switches fiction → nonfiction loses the
// fiction rules they never touched and keeps everything they edited,
// rejected or annotated.
func (s *Server) pssPruneOtherKinds(ctx context.Context, projectID int64, kind string) error {
	_, err := s.DB.ExecContext(ctx,
		`DELETE FROM project_stylesheet_items
		  WHERE project_id = ? AND house_id IS NOT NULL
		    AND status = 'accepted' AND note = ''
		    AND house_id IN (SELECT h.id FROM house_style h
		                      WHERE h.book_kind <> 'both' AND h.book_kind <> ?)`,
		projectID, kind)
	return err
}

func (s *Server) pssLoad(ctx context.Context, projectID int64) (*pssSheet, error) {
	kind, err := s.pssEnsure(ctx, projectID)
	if err != nil {
		return nil, err
	}
	sheet := &pssSheet{ProjectID: projectID, BookKind: kind, KindChosen: kind != "both", Sections: []pssSection{}}

	_ = s.DB.QueryRowContext(ctx,
		`SELECT name, client_slug, project_slug FROM projects WHERE id = ?`, projectID).
		Scan(&sheet.ProjectName, &sheet.ClientSlug, &sheet.ProjectSlug)
	_ = s.DB.QueryRowContext(ctx,
		`SELECT updated_at FROM project_stylesheets WHERE project_id = ?`, projectID).Scan(&sheet.UpdatedAt)

	rows, err := s.DB.QueryContext(ctx,
		`SELECT i.id, i.house_id, i.section_ord, i.section, i.item_ord, i.kind,
		        i.col1, i.col2, i.col3, i.body, i.status, i.note,
		        h.col1, h.col2, h.col3, h.body
		   FROM project_stylesheet_items i
		   LEFT JOIN house_style h ON h.id = i.house_id
		  WHERE i.project_id = ?
		  ORDER BY i.section_ord, i.item_ord, i.id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it pssItem
		var hc1, hc2, hc3, hb sql.NullString
		if err := rows.Scan(&it.ID, &it.HouseID, &it.SectionOrd, &it.Section, &it.ItemOrd, &it.Kind,
			&it.Col1, &it.Col2, &it.Col3, &it.Body, &it.Status, &it.Note,
			&hc1, &hc2, &hc3, &hb); err != nil {
			return nil, err
		}
		if hc1.Valid && (hc1.String != it.Col1 || hc2.String != it.Col2 || hc3.String != it.Col3 || hb.String != it.Body) {
			it.House = &pssText{Col1: hc1.String, Col2: hc2.String, Col3: hc3.String, Body: hb.String}
		}
		switch it.Status {
		case pssStatusRejected:
			sheet.Counts.Rejected++
		case pssStatusEdited:
			sheet.Counts.Edited++
		case pssStatusAdded:
			sheet.Counts.Added++
		default:
			sheet.Counts.Accepted++
		}
		if it.HouseID != nil {
			sheet.Counts.House++
		}
		n := len(sheet.Sections)
		if n == 0 || sheet.Sections[n-1].Ord != it.SectionOrd {
			sheet.Sections = append(sheet.Sections, pssSection{Ord: it.SectionOrd, Title: it.Section, Items: []pssItem{}})
			n++
		}
		sheet.Sections[n-1].Items = append(sheet.Sections[n-1].Items, it)
	}
	return sheet, rows.Err()
}

func (s *Server) pssTouch(ctx context.Context, projectID int64) {
	_, _ = s.DB.ExecContext(ctx,
		`UPDATE project_stylesheets SET updated_at = CURRENT_TIMESTAMP WHERE project_id = ?`, projectID)
}

// pssAuth resolves the project id from the path and applies the project gate.
// Returns ok=false once the response has been written.
func (s *Server) pssAuth(w http.ResponseWriter, r *http.Request) (int64, bool) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return 0, false
	}
	if !s.requireAuth(w, r, pid) {
		return 0, false
	}
	return pid, true
}

// pssItemRow loads one item and confirms it belongs to this project, so an
// item id from another book cannot be reached through this project's gate.
func (s *Server) pssItemRow(ctx context.Context, projectID int64, r *http.Request) (pssItem, error) {
	var it pssItem
	itemID, err := strconv.ParseInt(r.PathValue("item"), 10, 64)
	if err != nil {
		return it, fmt.Errorf("bad item id")
	}
	err = s.DB.QueryRowContext(ctx,
		`SELECT id, house_id, section_ord, section, item_ord, kind, col1, col2, col3, body, status, note
		   FROM project_stylesheet_items WHERE id = ? AND project_id = ?`, itemID, projectID).
		Scan(&it.ID, &it.HouseID, &it.SectionOrd, &it.Section, &it.ItemOrd, &it.Kind,
			&it.Col1, &it.Col2, &it.Col3, &it.Body, &it.Status, &it.Note)
	if err == sql.ErrNoRows {
		return it, fmt.Errorf("not found")
	}
	return it, err
}

// --- handlers ---

func (s *Server) handleGetProjectStylesheet(w http.ResponseWriter, r *http.Request) {
	pid, ok := s.pssAuth(w, r)
	if !ok {
		return
	}
	sheet, err := s.pssLoad(r.Context(), pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, sheet)
}

// handlePatchProjectStylesheet sets the book kind (the Fiction / Nonfiction
// chooser at the top of the page).
func (s *Server) handlePatchProjectStylesheet(w http.ResponseWriter, r *http.Request) {
	pid, ok := s.pssAuth(w, r)
	if !ok {
		return
	}
	var body struct {
		BookKind string `json:"book_kind"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if _, err := s.pssEnsure(r.Context(), pid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	kind := pssKind(body.BookKind)
	if _, err := s.DB.ExecContext(r.Context(),
		`UPDATE project_stylesheets SET book_kind = ?, updated_at = CURRENT_TIMESTAMP WHERE project_id = ?`,
		kind, pid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if err := s.pssPruneOtherKinds(r.Context(), pid, kind); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if err := s.pssSeedKind(r.Context(), pid, kind); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	sheet, err := s.pssLoad(r.Context(), pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, sheet)
}

// handleProjectStylesheetAcceptAll puts every rejected house rule back in
// force. Edits and the client's own rules are left alone — "accept all" means
// "our defaults stand", not "throw my work away".
func (s *Server) handleProjectStylesheetAcceptAll(w http.ResponseWriter, r *http.Request) {
	pid, ok := s.pssAuth(w, r)
	if !ok {
		return
	}
	if _, err := s.pssEnsure(r.Context(), pid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if _, err := s.DB.ExecContext(r.Context(),
		`UPDATE project_stylesheet_items
		    SET status = CASE
		          WHEN house_id IS NULL THEN 'added'
		          WHEN col1 = (SELECT h.col1 FROM house_style h WHERE h.id = house_id)
		           AND col2 = (SELECT h.col2 FROM house_style h WHERE h.id = house_id)
		           AND col3 = (SELECT h.col3 FROM house_style h WHERE h.id = house_id)
		           AND body = (SELECT h.body FROM house_style h WHERE h.id = house_id)
		          THEN 'accepted' ELSE 'edited' END,
		        updated_at = CURRENT_TIMESTAMP
		  WHERE project_id = ? AND status = 'rejected'`, pid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	s.pssTouch(r.Context(), pid)
	sheet, err := s.pssLoad(r.Context(), pid)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, sheet)
}

func (s *Server) handleProjectStylesheetAddItem(w http.ResponseWriter, r *http.Request) {
	pid, ok := s.pssAuth(w, r)
	if !ok {
		return
	}
	if _, err := s.pssEnsure(r.Context(), pid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	var body struct {
		SectionOrd *int   `json:"section_ord"`
		Section    string `json:"section"`
		Kind       string `json:"kind"`
		Col1       string `json:"col1"`
		Col2       string `json:"col2"`
		Col3       string `json:"col3"`
		Body       string `json:"body"`
		Note       string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	kind := "rule"
	if strings.TrimSpace(body.Kind) == "prose" {
		kind = "prose"
	}
	if kind == "rule" && strings.TrimSpace(body.Col1) == "" && strings.TrimSpace(body.Col2) == "" {
		jsonErr(w, "a rule needs at least an item or a rule", 400)
		return
	}
	if kind == "prose" && strings.TrimSpace(body.Body) == "" {
		jsonErr(w, "a note needs some text", 400)
		return
	}

	section := strings.TrimSpace(body.Section)
	sectionOrd := 0
	if body.SectionOrd != nil {
		sectionOrd = *body.SectionOrd
		if section == "" {
			// Section named by ord only: take the title from a sibling row.
			_ = s.DB.QueryRowContext(r.Context(),
				`SELECT section FROM project_stylesheet_items WHERE project_id = ? AND section_ord = ? LIMIT 1`,
				pid, sectionOrd).Scan(&section)
		}
	}
	if section == "" {
		section = pssAddedSection
		sectionOrd = pssAddedSectionOrd
	}

	var nextOrd int
	_ = s.DB.QueryRowContext(r.Context(),
		`SELECT COALESCE(MAX(item_ord), 0) + 1 FROM project_stylesheet_items WHERE project_id = ? AND section_ord = ?`,
		pid, sectionOrd).Scan(&nextOrd)

	res, err := s.DB.ExecContext(r.Context(),
		`INSERT INTO project_stylesheet_items
		   (project_id, house_id, section_ord, section, item_ord, kind, col1, col2, col3, body, status, note)
		 VALUES (?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, 'added', ?)`,
		pid, sectionOrd, section, nextOrd, kind,
		strings.TrimSpace(body.Col1), strings.TrimSpace(body.Col2), strings.TrimSpace(body.Col3),
		strings.TrimSpace(body.Body), strings.TrimSpace(body.Note))
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	id, _ := res.LastInsertId()
	s.pssTouch(r.Context(), pid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "section_ord": sectionOrd, "section": section})
}

// handleProjectStylesheetPatchItem updates one row: status, text, note. Text
// that differs from the house original flips a house row to 'edited'; text
// typed back to ours flips it to 'accepted'. The client's own rows stay
// 'added' unless they reject them.
func (s *Server) handleProjectStylesheetPatchItem(w http.ResponseWriter, r *http.Request) {
	pid, ok := s.pssAuth(w, r)
	if !ok {
		return
	}
	it, err := s.pssItemRow(r.Context(), pid, r)
	if err != nil {
		jsonErr(w, err.Error(), 404)
		return
	}
	var body struct {
		Status *string `json:"status"`
		Col1   *string `json:"col1"`
		Col2   *string `json:"col2"`
		Col3   *string `json:"col3"`
		Body   *string `json:"body"`
		Note   *string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.Col1 != nil {
		it.Col1 = strings.TrimSpace(*body.Col1)
	}
	if body.Col2 != nil {
		it.Col2 = strings.TrimSpace(*body.Col2)
	}
	if body.Col3 != nil {
		it.Col3 = strings.TrimSpace(*body.Col3)
	}
	if body.Body != nil {
		it.Body = strings.TrimSpace(*body.Body)
	}
	if body.Note != nil {
		it.Note = strings.TrimSpace(*body.Note)
	}

	status := it.Status
	if body.Status != nil {
		switch want := strings.ToLower(strings.TrimSpace(*body.Status)); want {
		case pssStatusRejected:
			status = pssStatusRejected
		case pssStatusAccepted, pssStatusEdited, pssStatusAdded:
			status = pssStatusAccepted // re-derived from the text below
		default:
			jsonErr(w, "unknown status", 400)
			return
		}
	}
	if status != pssStatusRejected {
		if it.HouseID == nil {
			status = pssStatusAdded
		} else {
			same, err := s.pssMatchesHouse(r.Context(), it)
			if err != nil {
				jsonErr(w, err.Error(), 500)
				return
			}
			if same {
				status = pssStatusAccepted
			} else {
				status = pssStatusEdited
			}
		}
	}

	if _, err := s.DB.ExecContext(r.Context(),
		`UPDATE project_stylesheet_items
		    SET col1 = ?, col2 = ?, col3 = ?, body = ?, status = ?, note = ?, updated_at = CURRENT_TIMESTAMP
		  WHERE id = ? AND project_id = ?`,
		it.Col1, it.Col2, it.Col3, it.Body, status, it.Note, it.ID, pid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	s.pssTouch(r.Context(), pid)
	it.Status = status
	jsonOK(w, it)
}

func (s *Server) pssMatchesHouse(ctx context.Context, it pssItem) (bool, error) {
	if it.HouseID == nil {
		return false, nil
	}
	var h pssText
	err := s.DB.QueryRowContext(ctx,
		`SELECT col1, col2, col3, body FROM house_style WHERE id = ?`, *it.HouseID).
		Scan(&h.Col1, &h.Col2, &h.Col3, &h.Body)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return h.Col1 == it.Col1 && h.Col2 == it.Col2 && h.Col3 == it.Col3 && h.Body == it.Body, nil
}

// handleProjectStylesheetDeleteItem removes a client's own rule. House rules
// are never deleted — rejecting is not deleting, so the decision stays visible.
func (s *Server) handleProjectStylesheetDeleteItem(w http.ResponseWriter, r *http.Request) {
	pid, ok := s.pssAuth(w, r)
	if !ok {
		return
	}
	it, err := s.pssItemRow(r.Context(), pid, r)
	if err != nil {
		jsonErr(w, err.Error(), 404)
		return
	}
	if it.HouseID != nil {
		jsonErr(w, "house rules can be rejected, not deleted", 400)
		return
	}
	if _, err := s.DB.ExecContext(r.Context(),
		`DELETE FROM project_stylesheet_items WHERE id = ? AND project_id = ?`, it.ID, pid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	s.pssTouch(r.Context(), pid)
	jsonOK(w, map[string]any{"deleted": it.ID})
}

// handleProjectStylesheetRestoreItem puts the studio's words back and marks
// the rule accepted. The client's note is kept (it is their margin, not ours).
func (s *Server) handleProjectStylesheetRestoreItem(w http.ResponseWriter, r *http.Request) {
	pid, ok := s.pssAuth(w, r)
	if !ok {
		return
	}
	it, err := s.pssItemRow(r.Context(), pid, r)
	if err != nil {
		jsonErr(w, err.Error(), 404)
		return
	}
	if it.HouseID == nil {
		jsonErr(w, "this is your own rule; there is no studio version to restore", 400)
		return
	}
	if _, err := s.DB.ExecContext(r.Context(),
		`UPDATE project_stylesheet_items
		    SET col1 = (SELECT h.col1 FROM house_style h WHERE h.id = house_id),
		        col2 = (SELECT h.col2 FROM house_style h WHERE h.id = house_id),
		        col3 = (SELECT h.col3 FROM house_style h WHERE h.id = house_id),
		        body = (SELECT h.body FROM house_style h WHERE h.id = house_id),
		        status = 'accepted', updated_at = CURRENT_TIMESTAMP
		  WHERE id = ? AND project_id = ? AND house_id IS NOT NULL`, it.ID, pid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	s.pssTouch(r.Context(), pid)
	out, err := s.pssItemRow(r.Context(), pid, r)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, out)
}

// handleProjectStylesheetMD renders the *effective* sheet — accepted, edited
// and added rules; rejected ones omitted — for the copyeditor. Mirrors
// handleHouseStyleMD. Called from the /{client}/{project}/… catch-all, so the
// slugs arrive as arguments rather than path values.
func (s *Server) handleProjectStylesheetMD(w http.ResponseWriter, r *http.Request, clientSlug, projectSlug string) {
	var pid int64
	err := s.DB.QueryRowContext(r.Context(),
		`SELECT id FROM projects WHERE client_slug = ? AND project_slug = ?`, clientSlug, projectSlug).Scan(&pid)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	if !s.checkAuth(r, pid) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sheet, err := s.pssLoad(r.Context(), pid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var b strings.Builder
	title := sheet.ProjectName
	if title == "" {
		title = projectSlug
	}
	fmt.Fprintf(&b, "# Style sheet — %s\n\n", title)
	fmt.Fprintf(&b, "_[jdbb] studio editorial style sheet for this book · %s · generated %s_\n\n",
		pssKindLabel(sheet.BookKind), time.Now().UTC().Format("2006-01-02"))
	fmt.Fprintf(&b, "_%d studio rules: %d accepted, %d edited, %d rejected (omitted below) · %d added by the author_\n\n",
		sheet.Counts.House, sheet.Counts.Accepted, sheet.Counts.Edited, sheet.Counts.Rejected, sheet.Counts.Added)

	n := 0
	for _, sec := range sheet.Sections {
		var rules, prose []pssItem
		for _, it := range sec.Items {
			if it.Status == pssStatusRejected {
				continue
			}
			if it.Kind == "rule" {
				rules = append(rules, it)
			} else {
				prose = append(prose, it)
			}
		}
		if len(rules) == 0 && len(prose) == 0 {
			continue
		}
		n++
		fmt.Fprintf(&b, "## %d. %s\n\n", n, sec.Title)
		if len(rules) > 0 {
			b.WriteString("| Item | Rule | Example / note |\n|---|---|---|\n")
			for _, it := range rules {
				third := it.Col3
				if it.Note != "" {
					third = strings.TrimSpace(strings.TrimSpace(third) + " — " + it.Note)
				}
				fmt.Fprintf(&b, "| %s | %s | %s |\n", mdCell(it.Col1), mdCell(it.Col2), mdCell(third))
			}
			b.WriteString("\n")
		}
		for _, it := range prose {
			b.WriteString(strings.TrimSpace(it.Body) + "\n\n")
			if it.Note != "" {
				b.WriteString("> " + strings.TrimSpace(it.Note) + "\n\n")
			}
		}
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}

func pssKindLabel(kind string) string {
	switch kind {
	case "fiction":
		return "fiction"
	case "nonfiction":
		return "nonfiction"
	}
	return "rules common to fiction and nonfiction"
}
