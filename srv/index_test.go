package srv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"srv.exe.dev/db/dbgen"
	"srv.exe.dev/srv/indexer"
)

// ─── helpers ───

// indexTestDOCX is a two-chapter manuscript with sentences the index
// anchors point at.
func indexTestDOCX(t *testing.T) []byte {
	t.Helper()
	path := writeBookMapDOCX(t, []tp{
		{style: "Title", text: "Rice and Water"},
		{style: "Heading1", text: "Khlongs"},
		{text: "The khlongs carried water to the paddies and rice was the region's dietary staple for a thousand years."},
		{text: "A spirit house stood at every lock, and the lock keepers left rice for the spirits each morning."},
		{style: "Heading1", text: "Subaks"},
		{text: "The subak is a cooperative of farmers who share one water temple and one irrigation schedule."},
		{text: "Rice again: the subaks time their planting so that pests find no continuous field to feed on."},
	})
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// uploadDocx uploads real .docx bytes to a project as the given actor.
func uploadDocx(t *testing.T, ts *httptest.Server, projectID string, cookie *http.Cookie, admin bool, docx []byte) string {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("title", "Rice and Water")
	_ = mw.WriteField("author", "A. Tester")
	_ = mw.WriteField("project_id", projectID)
	fw, _ := mw.CreateFormFile("file", "rice.docx")
	_, _ = fw.Write(docx)
	_ = mw.Close()
	req, _ := http.NewRequest("POST", ts.URL+"/api/books/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if cookie != nil {
		req.AddCookie(cookie)
	}
	if admin {
		req.Header.Set("X-ExeDev-UserID", "test-admin")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("upload: %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	return itoa(int64(created["id"].(float64)))
}

// clientJSON sends a JSON request with the client's cookie.
func clientJSON(t *testing.T, ts *httptest.Server, cookie *http.Cookie, method, path string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, ts.URL+path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func indexStatusOf(t *testing.T, s *Server, bookID string) string {
	t.Helper()
	var st string
	if err := s.DB.QueryRow(`SELECT index_status FROM books WHERE id = ?`, bookID).Scan(&st); err != nil {
		t.Fatal(err)
	}
	return st
}

func waitIndexStatus(t *testing.T, s *Server, bookID string, want ...string) string {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		st := indexStatusOf(t, s, bookID)
		for _, w := range want {
			if st == w {
				return st
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("index_status did not reach %v (is %q)", want, indexStatusOf(t, s, bookID))
	return ""
}

func factoryEventDetail(t *testing.T, s *Server, kind string) string {
	t.Helper()
	var d string
	_ = s.DB.QueryRow(`SELECT detail FROM factory_events WHERE kind = ? ORDER BY id DESC LIMIT 1`, kind).Scan(&d)
	return d
}

// fakeIndexAnswers answers the chapter calls with two entries each (one
// anchor of which is a paraphrase that will not match) and the merge call
// with a synonym merge.
func fakeIndexAnswers(system, user string) (string, error) {
	if strings.Contains(system, "reading your draft headings from every chapter together") {
		return `{"merge":[{"from":"khlong","to":"khlongs"}],"see":[],"see_also":[{"heading":"khlongs","also":"subaks"}],"drop":[]}`, nil
	}
	if strings.Contains(user, "Chapter 1:") {
		return `{"entries":[{"heading":"khlong","subheading":"","see":"","see_also":[],"anchors":["The khlongs carried water to the paddies"]},` +
			`{"heading":"rice","subheading":"as staple","see":"","see_also":[],"anchors":["rice was the region's dietary staple","the paraphrase that is not in the book"]},` +
			`{"heading":"spirit houses","subheading":"","see":"","see_also":[],"anchors":["A spirit house stood at every lock"]}]}`, nil
	}
	return `{"entries":[{"heading":"subaks","subheading":"","see":"","see_also":[],"anchors":["The subak is a cooperative of farmers"]},` +
		`{"heading":"rice","subheading":"pest control","see":"","see_also":[],"anchors":["pests find no continuous field"]}]}`, nil
}

// ─── entitlement gating ───

// TestIndexEntitlementGating: without the add-on the draft/put/convert-with-
// index calls are 402 with addon "index"; the admin grant flips the pass and
// the pass JSON says so.
func TestIndexEntitlementGating(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Index Buyer", "Indexed Book")
	cookie := clientCookie(t, ts, clientSlug, password)
	bookID := uploadDocx(t, ts, itoa(pass.ProjectID), cookie, false, indexTestDOCX(t))

	// GET is allowed (the page needs to know), but says not entitled.
	resp := clientJSON(t, ts, cookie, "GET", "/api/books/"+bookID+"/index", nil)
	var got indexResponse
	decodeJSON(t, resp, &got)
	if resp.StatusCode != 200 || got.Status != "off" || got.Entitled {
		t.Fatalf("GET index before purchase: %d %+v", resp.StatusCode, got)
	}

	for _, c := range []struct {
		method, path string
		body         any
	}{
		{"POST", "/api/books/" + bookID + "/index/draft", nil},
		{"PUT", "/api/books/" + bookID + "/index", map[string]any{"entries": []any{}}},
		{"POST", "/api/books/" + bookID + "/convert", map[string]any{"index": true, "kind": "proof"}},
	} {
		resp := clientJSON(t, ts, cookie, c.method, c.path, c.body)
		var body map[string]any
		decodeJSON(t, resp, &body)
		if resp.StatusCode != http.StatusPaymentRequired || body["addon"] != "index" {
			t.Errorf("%s %s without the add-on: %d %v, want 402 addon=index", c.method, c.path, resp.StatusCode, body)
		}
	}
	if st := indexStatusOf(t, s, bookID); st != "off" {
		t.Errorf("402s must not change index_status (is %q)", st)
	}

	// Pass JSON: not included.
	resp = clientJSON(t, ts, cookie, "GET", fmt.Sprintf("/api/projects/%d/pass", pass.ProjectID), nil)
	var pj map[string]any
	decodeJSON(t, resp, &pj)
	if pj["index_included"] != false {
		t.Errorf("pass JSON index_included = %v, want false", pj["index_included"])
	}

	// A stranger's cookie (another client) is 403 even after the grant.
	_, otherPw, otherSlug, _ := grantedPass(t, s, ts, "Other Client", "Other Book")
	otherCookie := clientCookie(t, ts, otherSlug, otherPw)

	// Admin grant.
	resp = apiRequestAdmin(t, ts, "POST", fmt.Sprintf("/api/admin/passes/%d/index", pass.ID), nil)
	if resp.StatusCode != 200 {
		t.Fatalf("admin grant: %d", resp.StatusCode)
	}
	resp.Body.Close()
	if n := countLedger(t, s, pass.ID, "index_grant"); n != 1 {
		t.Errorf("ledger rows for the grant = %d, want 1", n)
	}
	resp = clientJSON(t, ts, cookie, "GET", fmt.Sprintf("/api/projects/%d/pass", pass.ProjectID), nil)
	decodeJSON(t, resp, &pj)
	if pj["index_included"] != true {
		t.Errorf("pass JSON index_included after grant = %v, want true", pj["index_included"])
	}
	resp = clientJSON(t, ts, otherCookie, "POST", "/api/books/"+bookID+"/index/draft", nil)
	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("other client's draft: %d, want 401/403", resp.StatusCode)
	}
	resp.Body.Close()

	// Entitled now, but no draft yet: convert with index is a 400, not a build.
	resp = clientJSON(t, ts, cookie, "POST", "/api/books/"+bookID+"/convert", map[string]any{"index": true, "kind": "proof"})
	if resp.StatusCode != 400 {
		t.Errorf("convert with index and no draft: %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// The store item exists at $100 and grants the entitlement.
	it := storeItemByKey("index")
	if it == nil || it.Amount != 10000 || !it.Index {
		t.Fatalf("store catalog index item: %+v", it)
	}
}

// ─── draft state machine ───

func TestIndexDraftStateMachine(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	s, ts, cleanup := testServer(t)
	defer cleanup()
	s.IndexClient = &indexer.Client{Model: "fake", Fake: fakeIndexAnswers}

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Index Buyer", "Indexed Book")
	cookie := clientCookie(t, ts, clientSlug, password)
	if _, err := s.DB.Exec(`UPDATE passes SET index_included = 1 WHERE id = ?`, pass.ID); err != nil {
		t.Fatal(err)
	}
	bookID := uploadDocx(t, ts, itoa(pass.ProjectID), cookie, false, indexTestDOCX(t))

	// Draft.
	resp := clientJSON(t, ts, cookie, "POST", "/api/books/"+bookID+"/index/draft", nil)
	var started map[string]any
	decodeJSON(t, resp, &started)
	if resp.StatusCode != 200 || started["status"] != "drafting" {
		t.Fatalf("draft: %d %v", resp.StatusCode, started)
	}
	if st := waitIndexStatus(t, s, bookID, "draft", "error"); st != "draft" {
		t.Fatalf("draft ended in %q: %s", st, factoryEventDetail(t, s, "index.draft.failed"))
	}
	resp = clientJSON(t, ts, cookie, "GET", "/api/books/"+bookID+"/index", nil)
	var got indexResponse
	decodeJSON(t, resp, &got)
	if got.Status != "draft" || !got.Entitled || got.Index == nil {
		t.Fatalf("GET after draft: %+v", got)
	}
	heads := map[string]int{}
	sees := map[string]string{}
	for _, e := range got.Index.Entries {
		heads[e.Heading]++
		sees[e.Heading] = e.See
	}
	// "khlong" merged into "khlongs" and, sharing no word, left as a see.
	if heads["khlongs"] != 1 || heads["rice"] != 2 || heads["subaks"] != 1 || sees["khlong"] != "khlongs" {
		t.Errorf("entries after merge: %v %v", heads, sees)
	}
	if len(got.Index.Unmatched) != 1 || got.Index.Unmatched[0].Heading != "rice" {
		t.Errorf("unmatched anchors flagged: %+v", got.Index.Unmatched)
	}
	if d := factoryEventDetail(t, s, "index.drafted"); !strings.Contains(d, "1 unmatched") {
		t.Errorf("index.drafted event: %q", d)
	}

	// Review: drop the spirit houses entry, fix the bad anchor, add a see.
	var edited []indexer.Entry
	for _, e := range got.Index.Entries {
		if e.Heading == "spirit houses" {
			continue
		}
		if e.Heading == "rice" && e.Subheading == "as staple" {
			e.Anchors = []indexer.Anchor{{Chapter: 1, Text: "rice was the region's dietary staple"}}
		}
		edited = append(edited, e)
	}
	edited = append(edited, indexer.Entry{Heading: "canals", See: "khlongs"})
	resp = clientJSON(t, ts, cookie, "PUT", "/api/books/"+bookID+"/index", map[string]any{"entries": edited})
	got = indexResponse{}
	decodeJSON(t, resp, &got)
	if resp.StatusCode != 200 || got.Status != "reviewed" {
		t.Fatalf("PUT: %d %+v", resp.StatusCode, got)
	}
	if len(got.Index.Unmatched) != 0 {
		t.Errorf("unmatched after the fix: %+v", got.Index.Unmatched)
	}
	seen := map[string]bool{}
	for _, e := range got.Index.Entries {
		seen[e.Heading+"|"+e.See] = true
	}
	if seen["spirit houses|"] || !seen["canals|khlongs"] {
		t.Errorf("reviewed entries: %v", seen)
	}
	if st := indexStatusOf(t, s, bookID); st != "reviewed" {
		t.Errorf("index_status = %q, want reviewed", st)
	}

	// Validation.
	resp = clientJSON(t, ts, cookie, "PUT", "/api/books/"+bookID+"/index", map[string]any{"entries": []map[string]any{{"heading": "", "anchors": []any{}}}})
	if resp.StatusCode != 400 {
		t.Errorf("empty heading: %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
	resp = clientJSON(t, ts, cookie, "PUT", "/api/books/"+bookID+"/index", map[string]any{"nope": 1})
	if resp.StatusCode != 400 {
		t.Errorf("no entries: %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// While drafting: a second draft and an edit are 409.
	if _, err := s.DB.Exec(`UPDATE books SET index_status = 'drafting' WHERE id = ?`, bookID); err != nil {
		t.Fatal(err)
	}
	resp = clientJSON(t, ts, cookie, "POST", "/api/books/"+bookID+"/index/draft", nil)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("draft while drafting: %d, want 409", resp.StatusCode)
	}
	resp.Body.Close()
	resp = clientJSON(t, ts, cookie, "PUT", "/api/books/"+bookID+"/index", map[string]any{"entries": edited})
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("edit while drafting: %d, want 409", resp.StatusCode)
	}
	resp.Body.Close()
	// While a sibling build runs: 409 too.
	if _, err := s.DB.Exec(`UPDATE books SET index_status = 'reviewed', status = 'converting' WHERE id = ?`, bookID); err != nil {
		t.Fatal(err)
	}
	resp = clientJSON(t, ts, cookie, "POST", "/api/books/"+bookID+"/index/draft", nil)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("draft during a build: %d, want 409", resp.StatusCode)
	}
	resp.Body.Close()

	// A failed draft (model error) lands in error with the reason readable.
	if _, err := s.DB.Exec(`UPDATE books SET status = 'uploaded' WHERE id = ?`, bookID); err != nil {
		t.Fatal(err)
	}
	s.IndexClient = &indexer.Client{Model: "fake", Fake: func(string, string) (string, error) { return "", fmt.Errorf("gateway down") }}
	resp = clientJSON(t, ts, cookie, "POST", "/api/books/"+bookID+"/index/draft", nil)
	resp.Body.Close()
	waitIndexStatus(t, s, bookID, "error")
	resp = clientJSON(t, ts, cookie, "GET", "/api/books/"+bookID+"/index", nil)
	got = indexResponse{}
	decodeJSON(t, resp, &got)
	if got.Status != "error" || !strings.Contains(got.Error, "gateway down") {
		t.Errorf("failed draft: %+v", got)
	}
	// The previous reviewed document survives a failed re-draft.
	if got.Index == nil || len(got.Index.Entries) != len(edited) {
		t.Errorf("reviewed entries lost on a failed draft: %+v", got.Index)
	}
}

// ─── convert with index ───

// TestConvertWithIndexPlacesMarkers runs the real pipeline (pandoc + typst)
// with a stored draft and index:true, and checks the markers were placed
// and the index page set.
func TestConvertWithIndexPlacesMarkers(t *testing.T) {
	for _, bin := range []string{"pandoc", "typst"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not installed", bin)
		}
	}
	s, ts, cleanup := testServer(t)
	defer cleanup()
	s.epubRunner = func(int64, dbgen.Book) error { return nil }

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Index Buyer", "Indexed Book")
	cookie := clientCookie(t, ts, clientSlug, password)
	if _, err := s.DB.Exec(`UPDATE passes SET index_included = 1 WHERE id = ?`, pass.ID); err != nil {
		t.Fatal(err)
	}
	bookID := uploadDocx(t, ts, itoa(pass.ProjectID), cookie, false, indexTestDOCX(t))

	idx := indexer.Index{Book: "Rice and Water", Entries: []indexer.Entry{
		{Heading: "khlongs", Anchors: []indexer.Anchor{{Chapter: 1, Text: "The khlongs carried water to the paddies"}}},
		{Heading: "rice", Subheading: "as staple", Anchors: []indexer.Anchor{{Chapter: 1, Text: "rice was the region's dietary staple"}}},
		{Heading: "rice", Subheading: "pest control", Anchors: []indexer.Anchor{{Chapter: 2, Text: "pests find no continuous field"}, {Chapter: 2, Text: "not in the manuscript at all"}}},
		{Heading: "canals", See: "khlongs", Anchors: []indexer.Anchor{}},
	}}
	b, _ := json.Marshal(idx)
	if _, err := s.DB.Exec(`UPDATE books SET index_json = ?, index_status = 'draft' WHERE id = ?`, string(b), bookID); err != nil {
		t.Fatal(err)
	}

	resp := clientJSON(t, ts, cookie, "POST", "/api/books/"+bookID+"/convert", map[string]any{"index": true, "kind": "proof", "format": "pdf"})
	if resp.StatusCode != 200 {
		var body map[string]any
		decodeJSON(t, resp, &body)
		t.Fatalf("convert: %d %v", resp.StatusCode, body)
	}
	resp.Body.Close()
	if d := factoryEventDetail(t, s, "build.started"); !strings.Contains(d, "with index") {
		t.Errorf("build.started event: %q", d)
	}

	deadline := time.Now().Add(3 * time.Minute)
	var status, errMsg string
	for time.Now().Before(deadline) {
		_ = s.DB.QueryRow(`SELECT status, error_msg FROM books WHERE id = ?`, bookID).Scan(&status, &errMsg)
		if status == "ready" || status == "error" {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if status != "ready" {
		t.Fatalf("build status %q: %s", status, errMsg)
	}
	d := factoryEventDetail(t, s, "index.placed")
	if !strings.Contains(d, "3 anchors placed of 4") || !strings.Contains(d, "1 unmatched") || !strings.Contains(d, "1 cross-references") {
		t.Errorf("index.placed event: %q", d)
	}

	// The PDF has an index page with the entries and Typst-resolved folios.
	var pdf []byte
	if err := s.DB.QueryRow(`SELECT output_data FROM book_outputs WHERE book_id = ? AND output_format = 'pdf' ORDER BY id DESC LIMIT 1`, bookID).Scan(&pdf); err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("pdftotext"); err == nil {
		f := t.TempDir() + "/out.pdf"
		_ = os.WriteFile(f, pdf, 0644)
		out, err := exec.Command("pdftotext", "-layout", f, "-").Output()
		if err != nil {
			t.Fatal(err)
		}
		text := string(out)
		for _, want := range []string{"Index", "khlongs, ", "rice: as staple, ", "pest control, ", "canals. See khlongs"} {
			if !strings.Contains(text, want) {
				t.Errorf("PDF text lacks %q", want)
			}
		}
	}

	// Without index:true the same book builds without an index page.
	if _, err := s.DB.Exec(`UPDATE books SET status = 'ready' WHERE id = ?`, bookID); err != nil {
		t.Fatal(err)
	}
	resp = clientJSON(t, ts, cookie, "POST", "/api/books/"+bookID+"/convert", map[string]any{"kind": "proof", "format": "pdf"})
	if resp.StatusCode != 200 {
		t.Fatalf("plain convert: %d", resp.StatusCode)
	}
	resp.Body.Close()
	var n int
	deadline = time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		_ = s.DB.QueryRow(`SELECT status FROM books WHERE id = ?`, bookID).Scan(&status)
		if status == "ready" || status == "error" {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM factory_events WHERE kind = 'index.placed'`).Scan(&n)
	if n != 1 {
		t.Errorf("a build without index:true placed markers (index.placed events = %d)", n)
	}
}
