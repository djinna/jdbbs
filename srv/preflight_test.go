package srv

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"srv.exe.dev/db/dbgen"
)

func TestGetManuscriptPreflightReturnsExistsFalseWhenMissing(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Preflight Missing",
		"start_date":   "2026-04-12",
		"client_slug":  "vgr",
		"project_slug": "preflight-missing",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title:          "Missing Report",
		Author:         "Tester",
		Series:         "",
		SourceFilename: "missing.docx",
		SourceData:     []byte("fake-docx"),
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/preflight?book_id="+itoa(book.ID), nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get missing preflight: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["exists"] != false {
		t.Fatalf("expected exists=false, got %v", body["exists"])
	}
}

func TestRunManuscriptPreflightRequiresAdminHeader(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Preflight Auth",
		"start_date":   "2026-04-12",
		"client_slug":  "vgr",
		"project_slug": "preflight-auth",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := int64(project["ID"].(float64))

	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title:          "Auth Report",
		Author:         "Tester",
		Series:         "",
		SourceFilename: "auth.docx",
		SourceData:     []byte("fake-docx"),
		ProjectID:      sql.NullInt64{Int64: pid, Valid: true},
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	resp = apiRequest(t, ts, "POST", "/api/projects/"+itoa(pid)+"/preflight", map[string]any{"book_id": book.ID})
	if resp.StatusCode != 401 {
		t.Fatalf("run preflight without admin header: expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestGetManuscriptPreflightReportReturnsStoredHTML(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Preflight HTML",
		"start_date":   "2026-04-12",
		"client_slug":  "vgr",
		"project_slug": "preflight-html",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := int64(project["ID"].(float64))

	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title:          "HTML Report",
		Author:         "Tester",
		Series:         "",
		SourceFilename: "html.docx",
		SourceData:     []byte("fake-docx"),
		ProjectID:      sql.NullInt64{Int64: pid, Valid: true},
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	summary := `{"total":1,"high":1,"medium":0,"low":0,"by_type":{"manual_formatting":1}}`
	row, err := q.CreateManuscriptPreflight(t.Context(), dbgen.CreateManuscriptPreflightParams{
		ProjectID:      pid,
		BookID:         book.ID,
		Status:         "ready",
		SummaryJson:    summary,
		ReportJson:     `[{"type":"manual_formatting","severity":"high"}]`,
		ReportHtml:     "<html><body><h1>Stored Report</h1></body></html>",
		ErrorMsg:       "",
		SourceFilename: book.SourceFilename,
	})
	if err != nil {
		t.Fatalf("create preflight: %v", err)
	}
	if row.ProjectID != pid {
		t.Fatalf("expected project id %d, got %d", pid, row.ProjectID)
	}

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(pid)+"/preflight/report?book_id="+itoa(book.ID), nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get stored html report: expected 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("expected text/html content type, got %q", got)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read report body: %v", err)
	}
	_ = resp.Body.Close()
	if !strings.Contains(string(bodyBytes), "Stored Report") {
		t.Fatalf("expected stored report html body, got %q", string(bodyBytes))
	}

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(pid)+"/preflight/report?book_id="+itoa(book.ID)+"&preflight_id="+itoa(row.ID), nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get stored html report by preflight id: expected 200, got %d", resp.StatusCode)
	}
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read report body by preflight id: %v", err)
	}
	_ = resp.Body.Close()
	if !strings.Contains(string(bodyBytes), "Stored Report") {
		t.Fatalf("expected stored report html body by preflight id, got %q", string(bodyBytes))
	}
}

func TestGetManuscriptPreflightIncludesHistoryAndLatestReportURL(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Preflight History",
		"start_date":   "2026-04-12",
		"client_slug":  "vgr",
		"project_slug": "preflight-history",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := int64(project["ID"].(float64))

	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title:          "History Report",
		Author:         "Tester",
		Series:         "",
		SourceFilename: "history.docx",
		SourceData:     []byte("fake-docx"),
		ProjectID:      sql.NullInt64{Int64: pid, Valid: true},
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	first, err := q.CreateManuscriptPreflight(t.Context(), dbgen.CreateManuscriptPreflightParams{
		ProjectID:      pid,
		BookID:         book.ID,
		Status:         "ready",
		SummaryJson:    `{"total":1,"high":0,"medium":1,"low":0,"by_type":{"manual_formatting":1}}`,
		ReportJson:     `[{"type":"manual_formatting","severity":"medium"}]`,
		ReportHtml:     "<html><body><h1>First Report</h1></body></html>",
		ErrorMsg:       "",
		SourceFilename: book.SourceFilename,
	})
	if err != nil {
		t.Fatalf("create first preflight: %v", err)
	}
	second, err := q.CreateManuscriptPreflight(t.Context(), dbgen.CreateManuscriptPreflightParams{
		ProjectID:      pid,
		BookID:         book.ID,
		Status:         "ready",
		SummaryJson:    `{"total":2,"high":1,"medium":1,"low":0,"by_type":{"manual_formatting":2}}`,
		ReportJson:     `[{"type":"manual_formatting","severity":"high"}]`,
		ReportHtml:     "<html><body><h1>Second Report</h1></body></html>",
		ErrorMsg:       "",
		SourceFilename: book.SourceFilename,
	})
	if err != nil {
		t.Fatalf("create second preflight: %v", err)
	}

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(pid)+"/preflight?book_id="+itoa(book.ID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get preflight: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["report_url"] != "/api/projects/"+itoa(pid)+"/preflight/report?book_id="+itoa(book.ID)+"&preflight_id="+itoa(second.ID) {
		t.Fatalf("unexpected latest report_url: %v", body["report_url"])
	}
	history, ok := body["history"].([]any)
	if !ok {
		t.Fatalf("expected history array, got %#v", body["history"])
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
	latest := history[0].(map[string]any)
	if latest["id"] != float64(second.ID) || latest["latest"] != true {
		t.Fatalf("expected newest history entry first, got %#v", latest)
	}
	earlier := history[1].(map[string]any)
	if earlier["id"] != float64(first.ID) {
		t.Fatalf("expected earlier report second, got %#v", earlier)
	}
}

func TestRunManuscriptPreflightStoresSummaryOnSuccess(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	s.preflightRunner = func(docxPath string, declaredStylesPath string) ([]byte, []byte, error) {
		html := []byte("<html><body><h1>Edge Case Review Report</h1></body></html>")
		report := []map[string]any{
			{"type": "manual_formatting", "severity": "high", "text": "bold text"},
			{"type": "manual_list", "severity": "low", "text": "1. one"},
			{"type": "manual_list", "severity": "medium", "text": "2. two"},
		}
		jb, err := json.Marshal(report)
		return html, jb, err
	}

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Preflight Success",
		"start_date":   "2026-04-12",
		"client_slug":  "vgr",
		"project_slug": "preflight-success",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := int64(project["ID"].(float64))

	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title:          "Success Report",
		Author:         "Tester",
		Series:         "",
		SourceFilename: "success.docx",
		SourceData:     []byte("fake-docx"),
		ProjectID:      sql.NullInt64{Int64: pid, Valid: true},
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+itoa(pid)+"/preflight", map[string]any{"book_id": book.ID})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("run preflight: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["exists"] != true {
		t.Fatalf("expected exists=true, got %v", body["exists"])
	}
	if body["status"] != "ready" {
		t.Fatalf("expected status ready, got %v", body["status"])
	}
	summary := body["summary"].(map[string]any)
	if summary["total"] != float64(3) {
		t.Fatalf("expected total 3, got %v", summary["total"])
	}
	if summary["high"] != float64(1) || summary["medium"] != float64(1) || summary["low"] != float64(1) {
		t.Fatalf("unexpected severity counts: %#v", summary)
	}
	byType := summary["by_type"].(map[string]any)
	if byType["manual_list"] != float64(2) {
		t.Fatalf("expected manual_list count 2, got %v", byType["manual_list"])
	}

	stored, err := q.GetLatestManuscriptPreflight(t.Context(), dbgen.GetLatestManuscriptPreflightParams{ProjectID: pid, BookID: book.ID})
	if err != nil {
		t.Fatalf("get stored preflight: %v", err)
	}
	if stored.Status != "ready" {
		t.Fatalf("expected stored status ready, got %s", stored.Status)
	}
	if !strings.Contains(stored.ReportHtml, "Edge Case Review Report") {
		t.Fatalf("expected stored html report, got %q", stored.ReportHtml)
	}
}

func TestRunManuscriptPreflightWarnsOnObservedStyleMissingFromSpec(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	s.preflightRunner = func(docxPath string, declaredStylesPath string) ([]byte, []byte, error) {
		html := []byte("<html><body><h1>Edge Case Review Report</h1></body></html>")
		report := []map[string]any{
			{"type": "observed_style", "style_name": "tweet-p-ascii", "style_kind": "paragraph", "severity": "low", "text": "tweet-p-ascii"},
		}
		jb, err := json.Marshal(report)
		return html, jb, err
	}

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Preflight Missing Style",
		"start_date":   "2026-04-12",
		"client_slug":  "vgr",
		"project_slug": "preflight-missing-style",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := int64(project["ID"].(float64))

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+itoa(pid)+"/book-spec", map[string]any{
		"data": map[string]any{
			"metadata":      map[string]any{"title": "Test", "author": "Tester"},
			"typography":    map[string]any{},
			"headings":      map[string]any{},
			"elements":      map[string]any{},
			"front_matter":  map[string]any{},
			"back_matter":   map[string]any{},
			"page":          map[string]any{},
			"running_heads": map[string]any{},
			"epub":          map[string]any{},
			"custom_styles": []map[string]any{
				{"name": "tweet-p", "word_style": "tweet-p", "type": "paragraph", "description": "Tweet block"},
			},
		},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update book spec: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title:          "Style Report",
		Author:         "Tester",
		Series:         "",
		SourceFilename: "style.docx",
		SourceData:     []byte("fake-docx"),
		ProjectID:      sql.NullInt64{Int64: pid, Valid: true},
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+itoa(pid)+"/preflight", map[string]any{"book_id": book.ID})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("run preflight: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	summary := body["summary"].(map[string]any)
	byType := summary["by_type"].(map[string]any)
	if byType["undeclared_custom_style"] != float64(1) {
		t.Fatalf("expected undeclared_custom_style count 1, got %v", byType["undeclared_custom_style"])
	}

	stored, err := q.GetLatestManuscriptPreflight(t.Context(), dbgen.GetLatestManuscriptPreflightParams{ProjectID: pid, BookID: book.ID})
	if err != nil {
		t.Fatalf("get stored preflight: %v", err)
	}
	if !strings.Contains(stored.ReportJson, "undeclared_custom_style") {
		t.Fatalf("expected stored report json to include undeclared custom style warning, got %q", stored.ReportJson)
	}
}

func TestRunManuscriptPreflightRecordsDeclaredObservedCustomStyleUsage(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	s.preflightRunner = func(docxPath string, declaredStylesPath string) ([]byte, []byte, error) {
		html := []byte("<html><body><h1>Edge Case Review Report</h1></body></html>")
		report := []map[string]any{
			{"type": "observed_style", "style_name": "tweet-p-ascii", "style_kind": "paragraph", "severity": "low", "text": "ascii block"},
		}
		jb, err := json.Marshal(report)
		return html, jb, err
	}

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Preflight Declared Style",
		"start_date":   "2026-04-12",
		"client_slug":  "vgr",
		"project_slug": "preflight-declared-style",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := int64(project["ID"].(float64))

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+itoa(pid)+"/book-spec", map[string]any{
		"data": map[string]any{
			"metadata":      map[string]any{"title": "Test", "author": "Tester"},
			"typography":    map[string]any{},
			"headings":      map[string]any{},
			"elements":      map[string]any{},
			"front_matter":  map[string]any{},
			"back_matter":   map[string]any{},
			"page":          map[string]any{},
			"running_heads": map[string]any{},
			"epub":          map[string]any{},
			"custom_styles": []map[string]any{
				{"name": "tweet-p-ascii", "word_style": "tweet-p-ascii", "type": "paragraph", "description": "ASCII tweet block"},
			},
		},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update book spec: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title:          "Declared Style Report",
		Author:         "Tester",
		Series:         "",
		SourceFilename: "declared-style.docx",
		SourceData:     []byte("fake-docx"),
		ProjectID:      sql.NullInt64{Int64: pid, Valid: true},
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+itoa(pid)+"/preflight", map[string]any{"book_id": book.ID})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("run preflight: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	summary := body["summary"].(map[string]any)
	byType := summary["by_type"].(map[string]any)
	if byType["declared_custom_style_used"] != float64(1) {
		t.Fatalf("expected declared_custom_style_used count 1, got %v", byType["declared_custom_style_used"])
	}
	if _, exists := byType["undeclared_custom_style"]; exists {
		t.Fatalf("did not expect undeclared_custom_style warning when style is declared: %#v", byType)
	}
}
