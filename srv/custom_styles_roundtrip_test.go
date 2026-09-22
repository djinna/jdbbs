package srv

import (
	"context"
	"strings"
	"testing"

	"srv.exe.dev/db/dbgen"
)

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
			"metadata":      map[string]any{"title": "Test", "author": "Tester", "isbn_epub": "9780000009999"},
			"typography":    map[string]any{},
			"headings":      map[string]any{},
			"elements":      map[string]any{},
			"front_matter":  map[string]any{},
			"back_matter":   map[string]any{},
			"page":          map[string]any{},
			"running_heads": map[string]any{},
			"epub":          map[string]any{},
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

// A style declared on a final transmittal must reach the build even when an
// incidental spec write (chapter suggestions at upload, cover upload) happens
// afterwards. Before the fix those writes bumped book_specs.updated_at, the
// sync gate read "spec newer than transmittal" and skipped the re-pull, and
// [[verse2]] printed literally (mcheck book 58, 2026-09-22).
func TestIncidentalSpecWritesDoNotBlockTransmittalRePull(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]any{
		"name": "Re-pull", "start_date": "2026-09-01", "client_slug": "vgr", "project_slug": "repull",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := int64(project["ID"].(float64))
	pidStr := itoa(pid)

	put := func(styles []map[string]any) {
		resp := apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pidStr+"/transmittal", map[string]any{
			"status": "final",
			"data":   map[string]any{"custom_styles": styles},
		})
		if resp.StatusCode != 200 {
			t.Fatalf("put transmittal: %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
	// First final: spec pulled with one style.
	put([]map[string]any{{"name": "aside", "type": "paragraph", "based_on": "Block Quote"}})
	if _, err := s.syncSpecFromTransmittal(context.Background(), pid, true); err != nil {
		t.Fatalf("first sync: %v", err)
	}
	// Author declares a second style; the transmittal row is now the newer one.
	// (Timestamps are whole seconds: push the transmittal clearly ahead.)
	put([]map[string]any{
		{"name": "aside", "type": "paragraph", "based_on": "Block Quote"},
		{"name": "verse2", "type": "paragraph", "based_on": "Verse", "indent": 1},
	})
	if _, err := s.DB.Exec(`UPDATE book_specs SET updated_at = datetime('now','-10 seconds') WHERE project_id = ?`, pid); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB.Exec(`UPDATE transmittals SET updated_at = datetime('now','-5 seconds') WHERE project_id = ?`, pid); err != nil {
		t.Fatal(err)
	}
	// Then uploads a manuscript: chapter detection writes suggestions; and a cover.
	if err := s.storeChapterSuggestions(context.Background(), pid, []detectedChapter{}); err != nil {
		t.Fatalf("store suggestions: %v", err)
	}
	if err := dbgen.New(s.DB).UpdateBookSpecCover(context.Background(), dbgen.UpdateBookSpecCoverParams{
		CoverData: []byte("x"), CoverType: "image/png", ProjectID: pid,
	}); err != nil {
		t.Fatal(err)
	}
	data, err := s.syncSpecFromTransmittal(context.Background(), pid, true)
	if err != nil {
		t.Fatalf("sync after incidental writes: %v", err)
	}
	if !strings.Contains(data, `"verse2"`) {
		t.Fatalf("verse2 not pulled into spec after incidental writes; spec custom_styles: %s", data)
	}
}
