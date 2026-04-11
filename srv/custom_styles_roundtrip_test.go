package srv

import "testing"

func TestAdminSpecCustomStylesPreserveTypeAndDescriptionOnSave(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Custom Style Roundtrip",
		"start_date":   "2026-04-09",
		"client_slug":  "vgr",
		"project_slug": "custom-style-roundtrip",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/book-spec", map[string]any{
		"data": map[string]any{
			"metadata": map[string]any{"title": "Test", "author": "Tester", "isbn_epub": "9780000009999"},
			"typography": map[string]any{},
			"headings": map[string]any{},
			"elements": map[string]any{},
			"front_matter": map[string]any{},
			"back_matter": map[string]any{},
			"page": map[string]any{},
			"running_heads": map[string]any{},
			"epub": map[string]any{},
			"custom_styles": []map[string]any{
				{"name": "tweet", "word_style": "tweet", "type": "paragraph", "description": "Tweet block"},
				{"name": "metadata-c", "word_style": "metadata-c", "type": "character", "description": "Inline metadata"},
				{"name": "metadata-p", "word_style": "metadata-p", "type": "paragraph", "description": "Metadata paragraph"},
			},
		},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update book spec: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/book-spec", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get book spec: expected 200, got %d", resp.StatusCode)
	}
	var result map[string]any
	decodeJSON(t, resp, &result)
	data := result["data"].(map[string]any)
	meta := data["metadata"].(map[string]any)
	if meta["isbn_epub"] != "9780000009999" {
		t.Fatalf("expected isbn_epub preserved after spec save, got %v", meta["isbn_epub"])
	}
	styles := data["custom_styles"].([]any)
	if len(styles) != 3 {
		t.Fatalf("expected 3 custom styles after spec save, got %d", len(styles))
	}
	assertStyle := func(idx int, name, styleType, desc string) {
		style := styles[idx].(map[string]any)
		if style["name"] != name {
			t.Fatalf("style %d expected name %s, got %v", idx, name, style["name"])
		}
		if style["word_style"] != name {
			t.Fatalf("style %d expected word_style %s, got %v", idx, name, style["word_style"])
		}
		if style["type"] != styleType {
			t.Fatalf("style %d expected type %s, got %v", idx, styleType, style["type"])
		}
		if style["description"] != desc {
			t.Fatalf("style %d expected description %q, got %v", idx, desc, style["description"])
		}
	}
	assertStyle(0, "tweet", "paragraph", "Tweet block")
	assertStyle(1, "metadata-c", "character", "Inline metadata")
	assertStyle(2, "metadata-p", "paragraph", "Metadata paragraph")
}
