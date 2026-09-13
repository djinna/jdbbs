package srv

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPullTransmittalMapsEPUBISBN(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "EPUB ISBN Mapping",
		"start_date":   "2026-04-09",
		"client_slug":  "vgr",
		"project_slug": "epub-isbn-mapping",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/transmittal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get transmittal: expected 200, got %d", resp.StatusCode)
	}
	var tx map[string]any
	decodeJSON(t, resp, &tx)
	data := tx["data"].(map[string]any)
	book := data["book"].(map[string]any)
	book["isbn_epub"] = "9780000009999"

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/transmittal", map[string]any{
		"status": "draft",
		"data":   data,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update transmittal: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/book-spec/pull-transmittal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("pull transmittal into spec: expected 200, got %d", resp.StatusCode)
	}
	var result map[string]any
	decodeJSON(t, resp, &result)
	dataOut := result["data"].(map[string]any)
	meta := dataOut["metadata"].(map[string]any)
	if meta["isbn_epub"] != "9780000009999" {
		t.Fatalf("expected isbn_epub to map into spec metadata, got %v", meta["isbn_epub"])
	}
}

func TestWordTemplateGenerationRejectsDuplicateCustomStyleNames(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Duplicate Style Names",
		"start_date":   "2026-04-09",
		"client_slug":  "vgr",
		"project_slug": "duplicate-style-names",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/book-spec", map[string]any{
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
				{"name": "metadata", "word_style": "metadata", "type": "character", "description": "meta char"},
				{"name": "metadata", "word_style": "metadata", "type": "paragraph", "description": "meta para"},
			},
		},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update book spec: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/book-spec/word-template", nil)
	if resp.StatusCode != 400 {
		t.Fatalf("word template duplicate styles: expected 400, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	msg, _ := body["error"].(string)
	if msg == "" {
		t.Fatalf("expected duplicate style validation error message")
	}
}

func TestWordTemplateGenerationAllowsUniqueCustomStyleNames(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Unique Style Names",
		"start_date":   "2026-04-09",
		"client_slug":  "vgr",
		"project_slug": "unique-style-names",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/book-spec", map[string]any{
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
				{"name": "metadata-c", "word_style": "metadata-c", "type": "character", "description": "meta char"},
				{"name": "metadata-p", "word_style": "metadata-p", "type": "paragraph", "description": "meta para"},
			},
		},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update book spec: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/book-spec/word-template", nil)
	if resp.StatusCode != 200 {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		t.Fatalf("word template unique styles: expected 200, got %d (%v)", resp.StatusCode, body)
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read word template response: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected non-empty word template response body")
	}
}

// TestClientWordTemplateSelfServe: the client can download the Word template
// themselves once the transmittal is final; before that the route says so
// instead of handing out a template built from nothing.
func TestClientWordTemplateSelfServe(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Template Tester", "Template Book")
	pid := itoa(pass.ProjectID)
	cookie := clientCookie(t, ts, clientSlug, password)
	get := func() *http.Response {
		req, _ := http.NewRequest("GET", ts.URL+"/api/projects/"+pid+"/word-template", nil)
		req.AddCookie(cookie)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("get template: %v", err)
		}
		return resp
	}

	// No cookie at all: unauthorized.
	resp := apiRequest(t, ts, "GET", "/api/projects/"+pid+"/word-template", nil)
	if resp.StatusCode != 401 {
		t.Fatalf("anonymous: expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Draft transmittal: refused with a reason.
	resp = get()
	if resp.StatusCode != 409 {
		t.Fatalf("draft transmittal: expected 409, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Client marks the transmittal final (same call transmittal.js makes).
	var tx map[string]any
	json.Unmarshal([]byte(defaultTransmittalData()), &tx)
	tx["book"].(map[string]any)["title"] = "Template Book"
	markFinal := func() {
		req, _ := http.NewRequest("PUT", ts.URL+"/api/projects/"+pid+"/transmittal",
			bytes.NewReader(mustJSONBytes(t, map[string]any{"status": "final", "data": tx})))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("mark final: %v / %d", err, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// The default transmittal has an empty custom_styles list; that used to
	// reach the generator as JSON null and crash it.
	markFinal()
	resp = get()
	if resp.StatusCode != 200 {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		t.Fatalf("final transmittal, no custom styles: expected 200, got %d (%v)", resp.StatusCode, body)
	}
	resp.Body.Close()

	tx["custom_styles"] = []map[string]any{{"name": "verse", "type": "paragraph"}}
	markFinal()
	resp = get()
	if resp.StatusCode != 200 {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		t.Fatalf("final transmittal: expected 200, got %d (%v)", resp.StatusCode, body)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "wordprocessingml") {
		t.Fatalf("content-type: %q", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "-template.docx") {
		t.Fatalf("content-disposition: %q", cd)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.Len() < 4 || string(buf.Bytes()[:2]) != "PK" {
		t.Fatalf("expected a .docx (zip) body, got %d bytes", buf.Len())
	}

	// The pull happened: the spec now carries the transmittal's custom style.
	var specData string
	if err := s.DB.QueryRow(`SELECT data FROM book_specs WHERE project_id = ?`, pass.ProjectID).Scan(&specData); err != nil {
		t.Fatalf("spec after pull: %v", err)
	}
	if !strings.Contains(specData, `"verse"`) {
		t.Fatalf("expected transmittal custom style pulled into spec, got %s", specData)
	}
}

func mustJSONBytes(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}
