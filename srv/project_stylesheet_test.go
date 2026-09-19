package srv

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- helpers -------------------------------------------------------------

func pssDecodeSheet(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var sheet map[string]any
	decodeJSON(t, resp, &sheet)
	return sheet
}

func pssCount(t *testing.T, sheet map[string]any, name string) int {
	t.Helper()
	counts, ok := sheet["counts"].(map[string]any)
	if !ok {
		t.Fatalf("sheet has no counts: %v", sheet)
	}
	v, ok := counts[name].(float64)
	if !ok {
		t.Fatalf("counts has no %s: %v", name, counts)
	}
	return int(v)
}

// pssItems flattens the sections of a decoded sheet.
func pssItems(t *testing.T, sheet map[string]any) []map[string]any {
	t.Helper()
	var out []map[string]any
	secs, _ := sheet["sections"].([]any)
	for _, s := range secs {
		sec, _ := s.(map[string]any)
		items, _ := sec["items"].([]any)
		for _, i := range items {
			it, _ := i.(map[string]any)
			out = append(out, it)
		}
	}
	return out
}

func pssFirstRule(t *testing.T, sheet map[string]any) map[string]any {
	t.Helper()
	for _, it := range pssItems(t, sheet) {
		if it["kind"] == "rule" && it["house_id"] != nil {
			return it
		}
	}
	t.Fatalf("no house rule found in sheet")
	return nil
}

func pssItemByID(t *testing.T, sheet map[string]any, id float64) map[string]any {
	t.Helper()
	for _, it := range pssItems(t, sheet) {
		if it["id"] == id {
			return it
		}
	}
	return nil
}

// pssNewProject creates a project as admin and returns its id.
func pssNewProject(t *testing.T, ts *httptest.Server, client, project string) int64 {
	t.Helper()
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Style Sheet Book",
		"start_date":   "2026-09-01",
		"client_slug":  client,
		"project_slug": project,
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	return int64(created["ID"].(float64))
}

// --- tests ---------------------------------------------------------------

// TestProjectStylesheetSeedOnFirstOpen: the first GET creates the instance
// from the house sheet, copying the rules that apply to any book ('both'),
// every one already accepted — the encouraged path is to do nothing.
func TestProjectStylesheetSeedOnFirstOpen(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	pid := pssNewProject(t, ts, "sheetco", "book-001")

	var houseBoth int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM house_style WHERE book_kind = 'both'`).Scan(&houseBoth); err != nil {
		t.Fatalf("count house rows: %v", err)
	}
	if houseBoth == 0 {
		// house_style is seeded lazily on first read; force it and re-count.
		if _, err := s.DB.Exec(`SELECT 1`); err != nil {
			t.Fatal(err)
		}
	}

	resp := apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(pid)+"/stylesheet", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("first open: expected 200, got %d", resp.StatusCode)
	}
	sheet := pssDecodeSheet(t, resp)

	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM house_style WHERE book_kind = 'both'`).Scan(&houseBoth); err != nil {
		t.Fatalf("count house rows: %v", err)
	}
	if houseBoth == 0 {
		t.Fatal("house_style was not seeded")
	}
	if got := pssCount(t, sheet, "house"); got != houseBoth {
		t.Errorf("seeded %d rows, expected the %d 'both' house rows", got, houseBoth)
	}
	if got := pssCount(t, sheet, "accepted"); got != houseBoth {
		t.Errorf("expected every seeded row accepted, got %d of %d", got, houseBoth)
	}
	if sheet["book_kind"] != "both" || sheet["kind_chosen"] != false {
		t.Errorf("expected an unchosen 'both' sheet, got kind=%v chosen=%v", sheet["book_kind"], sheet["kind_chosen"])
	}

	// Second open is idempotent: no duplicate rows.
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(pid)+"/stylesheet", nil)
	again := pssDecodeSheet(t, resp)
	if pssCount(t, again, "house") != houseBoth {
		t.Errorf("second open duplicated rows: %d then %d", houseBoth, pssCount(t, again, "house"))
	}
}

// TestProjectStylesheetChooseKind: choosing fiction adds the fiction-only
// rules; switching to nonfiction drops the untouched fiction ones and keeps
// anything the client has decided on.
func TestProjectStylesheetChooseKind(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	pid := pssNewProject(t, ts, "kindco", "book-001")

	resp := apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(pid)+"/stylesheet", nil)
	base := pssCount(t, pssDecodeSheet(t, resp), "house")

	resp = apiRequestAdmin(t, ts, "PATCH", "/api/projects/"+itoa(pid)+"/stylesheet",
		map[string]string{"book_kind": "fiction"})
	if resp.StatusCode != 200 {
		t.Fatalf("choose fiction: expected 200, got %d", resp.StatusCode)
	}
	fiction := pssDecodeSheet(t, resp)
	if fiction["book_kind"] != "fiction" || fiction["kind_chosen"] != true {
		t.Errorf("expected a chosen fiction sheet, got %v/%v", fiction["book_kind"], fiction["kind_chosen"])
	}
	if pssCount(t, fiction, "house") <= base {
		t.Errorf("fiction should add rules: %d then %d", base, pssCount(t, fiction, "house"))
	}

	// Reject a fiction-only rule, then switch kinds: the decision survives.
	var keepID int64
	if err := s.DB.QueryRow(
		`SELECT i.id FROM project_stylesheet_items i JOIN house_style h ON h.id = i.house_id
		  WHERE i.project_id = ? AND h.book_kind = 'fiction' LIMIT 1`, pid).Scan(&keepID); err != nil {
		t.Fatalf("find a fiction row: %v", err)
	}
	resp = apiRequestAdmin(t, ts, "PATCH", "/api/projects/"+itoa(pid)+"/stylesheet/items/"+itoa(keepID),
		map[string]string{"status": "rejected"})
	if resp.StatusCode != 200 {
		t.Fatalf("reject fiction rule: got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "PATCH", "/api/projects/"+itoa(pid)+"/stylesheet",
		map[string]string{"book_kind": "nonfiction"})
	nonfic := pssDecodeSheet(t, resp)
	if nonfic["book_kind"] != "nonfiction" {
		t.Fatalf("expected nonfiction, got %v", nonfic["book_kind"])
	}
	if pssItemByID(t, nonfic, float64(keepID)) == nil {
		t.Errorf("a rejected fiction rule must not be pruned on a kind switch")
	}
	var pristineFiction int
	if err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM project_stylesheet_items i JOIN house_style h ON h.id = i.house_id
		  WHERE i.project_id = ? AND h.book_kind = 'fiction' AND i.status = 'accepted'`, pid).Scan(&pristineFiction); err != nil {
		t.Fatal(err)
	}
	if pristineFiction != 0 {
		t.Errorf("untouched fiction rules should be pruned when the book becomes nonfiction, %d left", pristineFiction)
	}
}

// TestProjectStylesheetEditRestoreRejectAcceptAll walks the decision model.
func TestProjectStylesheetEditRestoreRejectAcceptAll(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	pid := pssNewProject(t, ts, "editco", "book-001")
	base := "/api/projects/" + itoa(pid) + "/stylesheet"

	resp := apiRequestAdmin(t, ts, "GET", base, nil)
	sheet := pssDecodeSheet(t, resp)
	rule := pssFirstRule(t, sheet)
	id := int64(rule["id"].(float64))
	houseCol2 := rule["col2"].(string)

	// Edit the text: status becomes 'edited' and the studio original stays
	// reachable for Restore.
	resp = apiRequestAdmin(t, ts, "PATCH", base+"/items/"+itoa(id),
		map[string]string{"col2": "Serial comma: no, for this book."})
	if resp.StatusCode != 200 {
		t.Fatalf("edit rule: got %d", resp.StatusCode)
	}
	var edited map[string]any
	decodeJSON(t, resp, &edited)
	if edited["status"] != "edited" {
		t.Errorf("expected status edited, got %v", edited["status"])
	}
	sheet = pssDecodeSheet(t, apiRequestAdmin(t, ts, "GET", base, nil))
	if pssCount(t, sheet, "edited") != 1 {
		t.Errorf("expected 1 edited rule, got %d", pssCount(t, sheet, "edited"))
	}
	if h, ok := pssItemByID(t, sheet, float64(id))["house"].(map[string]any); !ok || h["col2"] != houseCol2 {
		t.Errorf("edited row must carry the house original, got %v", pssItemByID(t, sheet, float64(id))["house"])
	}

	// Restore: back to our words, accepted.
	resp = apiRequestAdmin(t, ts, "POST", base+"/items/"+itoa(id)+"/restore", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("restore: got %d", resp.StatusCode)
	}
	var restored map[string]any
	decodeJSON(t, resp, &restored)
	if restored["status"] != "accepted" || restored["col2"] != houseCol2 {
		t.Errorf("restore should return the house text and accept it, got %v / %v", restored["status"], restored["col2"])
	}

	// Reject with a note, then Accept all puts it back.
	resp = apiRequestAdmin(t, ts, "PATCH", base+"/items/"+itoa(id),
		map[string]string{"status": "rejected", "note": "House style differs for this series."})
	resp.Body.Close()
	sheet = pssDecodeSheet(t, apiRequestAdmin(t, ts, "GET", base, nil))
	if pssCount(t, sheet, "rejected") != 1 {
		t.Errorf("expected 1 rejected rule, got %d", pssCount(t, sheet, "rejected"))
	}

	resp = apiRequestAdmin(t, ts, "POST", base+"/accept-all", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("accept-all: got %d", resp.StatusCode)
	}
	after := pssDecodeSheet(t, resp)
	if pssCount(t, after, "rejected") != 0 {
		t.Errorf("accept-all should clear rejections, %d left", pssCount(t, after, "rejected"))
	}
	if pssItemByID(t, after, float64(id))["note"] != "House style differs for this series." {
		t.Errorf("accept-all must keep the client's note")
	}
}

// TestProjectStylesheetAddAndDelete: the client's own rules can be added and
// deleted; house rules can only be rejected.
func TestProjectStylesheetAddAndDelete(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	pid := pssNewProject(t, ts, "addco", "book-001")
	base := "/api/projects/" + itoa(pid) + "/stylesheet"

	sheet := pssDecodeSheet(t, apiRequestAdmin(t, ts, "GET", base, nil))
	houseRule := pssFirstRule(t, sheet)

	resp := apiRequestAdmin(t, ts, "POST", base+"/items", map[string]any{
		"col1": "Ship names", "col2": "Italic, no quotes", "col3": "the *Perihelion*",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("add rule: expected 201, got %d", resp.StatusCode)
	}
	var added map[string]any
	decodeJSON(t, resp, &added)
	if added["section"] != "Our additions" {
		t.Errorf("a rule with no section should land in Our additions, got %v", added["section"])
	}
	addedID := int64(added["id"].(float64))

	sheet = pssDecodeSheet(t, apiRequestAdmin(t, ts, "GET", base, nil))
	if pssCount(t, sheet, "added") != 1 {
		t.Errorf("expected 1 added rule, got %d", pssCount(t, sheet, "added"))
	}

	// Empty rule is rejected.
	resp = apiRequestAdmin(t, ts, "POST", base+"/items", map[string]any{"col3": "just an example"})
	if resp.StatusCode != 400 {
		t.Errorf("an empty rule should be refused, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// A house rule cannot be deleted.
	hid := int64(houseRule["id"].(float64))
	resp = apiRequestAdmin(t, ts, "DELETE", base+"/items/"+itoa(hid), nil)
	if resp.StatusCode != 400 {
		t.Errorf("deleting a house rule should be refused, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// The client's own rule can be.
	resp = apiRequestAdmin(t, ts, "DELETE", base+"/items/"+itoa(addedID), nil)
	if resp.StatusCode != 200 {
		t.Fatalf("delete own rule: got %d", resp.StatusCode)
	}
	resp.Body.Close()
	sheet = pssDecodeSheet(t, apiRequestAdmin(t, ts, "GET", base, nil))
	if pssCount(t, sheet, "added") != 0 {
		t.Errorf("added rule should be gone, %d left", pssCount(t, sheet, "added"))
	}
}

// TestProjectStylesheetMarkdownExport: the effective sheet is what the
// copyeditor gets — accepted + edited + added, rejected omitted.
func TestProjectStylesheetMarkdownExport(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	pid := pssNewProject(t, ts, "mdco", "book-001")
	base := "/api/projects/" + itoa(pid) + "/stylesheet"

	sheet := pssDecodeSheet(t, apiRequestAdmin(t, ts, "GET", base, nil))
	rule := pssFirstRule(t, sheet)
	id := int64(rule["id"].(float64))

	resp := apiRequestAdmin(t, ts, "PATCH", base+"/items/"+itoa(id),
		map[string]string{"col2": "REJECTED-RULE-MARKER"})
	resp.Body.Close()
	resp = apiRequestAdmin(t, ts, "PATCH", base+"/items/"+itoa(id),
		map[string]string{"status": "rejected"})
	resp.Body.Close()
	resp = apiRequestAdmin(t, ts, "POST", base+"/items", map[string]any{
		"col1": "ADDED-RULE-MARKER", "col2": "Always set in small caps",
	})
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "GET", "/mdco/book-001/stylesheet/index.md", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("markdown export: expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	md := string(body)
	if !strings.Contains(md, "# Style sheet — Style Sheet Book") {
		t.Errorf("export should be titled with the book: %.80q", md)
	}
	if strings.Contains(md, "REJECTED-RULE-MARKER") {
		t.Errorf("rejected rules must not appear in the effective sheet")
	}
	if !strings.Contains(md, "ADDED-RULE-MARKER") {
		t.Errorf("the client's own rules must appear in the effective sheet")
	}
	if !strings.Contains(md, "| Item | Rule | Example / note |") {
		t.Errorf("export should use the house table shape")
	}
}

// TestProjectStylesheetAuthGate: on a project whose client has a password,
// the API and the export answer 401 without auth; the page shell (no data)
// still renders, and the admin header passes.
func TestProjectStylesheetAuthGate(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	if _, err := s.DB.Exec(
		`INSERT INTO clients (slug, name, password_hash) VALUES ('gateco', 'Gate Co', 'x')`); err != nil {
		t.Fatalf("seed client: %v", err)
	}
	pid := pssNewProject(t, ts, "gateco", "book-001")
	base := "/api/projects/" + itoa(pid) + "/stylesheet"

	for _, path := range []string{base, base + "/items/1"} {
		resp := apiRequest(t, ts, "GET", path, nil)
		resp.Body.Close()
		if path == base && resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("anonymous GET %s: expected 401, got %d", path, resp.StatusCode)
		}
	}
	for _, m := range []struct{ method, path string }{
		{"PATCH", base},
		{"POST", base + "/accept-all"},
		{"POST", base + "/items"},
		{"PATCH", base + "/items/1"},
		{"DELETE", base + "/items/1"},
		{"POST", base + "/items/1/restore"},
	} {
		resp := apiRequest(t, ts, m.method, m.path, map[string]string{"status": "rejected"})
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("anonymous %s %s: expected 401, got %d", m.method, m.path, resp.StatusCode)
		}
	}

	resp := apiRequest(t, ts, "GET", "/gateco/book-001/stylesheet/index.md", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous markdown export: expected 401, got %d", resp.StatusCode)
	}

	resp = apiRequestAdmin(t, ts, "GET", base, nil)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("admin: expected 200, got %d", resp.StatusCode)
	}

	// The shell is public but carries no data (same contract as the factory).
	resp = apiRequest(t, ts, "GET", "/gateco/book-001/stylesheet/", nil)
	page, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("page shell: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(page), "data-client-nav") {
		t.Errorf("page shell should be the client-tier stylesheet page")
	}
}
