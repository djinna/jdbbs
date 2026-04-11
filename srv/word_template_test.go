package srv

import (
	"bytes"
	"encoding/json"
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
			"metadata": map[string]any{"title": "Test", "author": "Tester"},
			"typography": map[string]any{},
			"headings": map[string]any{},
			"elements": map[string]any{},
			"front_matter": map[string]any{},
			"back_matter": map[string]any{},
			"page": map[string]any{},
			"running_heads": map[string]any{},
			"epub": map[string]any{},
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
			"metadata": map[string]any{"title": "Test", "author": "Tester"},
			"typography": map[string]any{},
			"headings": map[string]any{},
			"elements": map[string]any{},
			"front_matter": map[string]any{},
			"back_matter": map[string]any{},
			"page": map[string]any{},
			"running_heads": map[string]any{},
			"epub": map[string]any{},
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
