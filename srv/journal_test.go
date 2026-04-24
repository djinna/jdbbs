package srv

import (
	"testing"
)

func TestJournalCRUD(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create project
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Journal Test", "client_slug": "jrn", "project_slug": "test", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// Create journal entry
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/journal", map[string]any{
		"entry_type": "call",
		"content":    "Discussed timeline with author",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create journal: expected 201, got %d", resp.StatusCode)
	}
	var entry map[string]any
	decodeJSON(t, resp, &entry)
	entryID := itoa(int64(entry["id"].(float64)))

	if entry["entry_type"] != "call" {
		t.Errorf("expected entry_type 'call', got %v", entry["entry_type"])
	}
	if entry["content"] != "Discussed timeline with author" {
		t.Errorf("unexpected content: %v", entry["content"])
	}

	// List journal
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/journal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list journal: expected 200, got %d", resp.StatusCode)
	}
	var entries []map[string]any
	decodeJSON(t, resp, &entries)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	// Delete entry
	resp = apiRequestAdmin(t, ts, "DELETE", "/api/projects/"+pid+"/journal/"+entryID, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("delete journal: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify deletion
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/journal", nil)
	decodeJSON(t, resp, &entries)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after delete, got %d", len(entries))
	}
}

func TestJournalContentRequired(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Content Test", "client_slug": "cr", "project_slug": "t", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// Try to create entry without content
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/journal", map[string]any{
		"entry_type": "note",
	})
	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for empty content, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Empty string should also fail
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/journal", map[string]any{
		"entry_type": "note",
		"content":    "",
	})
	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for empty string content, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestJournalDefaultType(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Type Test", "client_slug": "ty", "project_slug": "t", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// Create entry without type - should default to "note"
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/journal", map[string]any{
		"content": "Just a quick note",
	})
	var entry map[string]any
	decodeJSON(t, resp, &entry)

	if entry["entry_type"] != "note" {
		t.Errorf("expected default entry_type 'note', got %v", entry["entry_type"])
	}
}

func TestJournalEntryTypes(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Types Test", "client_slug": "tp", "project_slug": "t", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	types := []string{"call", "decision", "approval", "note"}
	for _, typ := range types {
		resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/journal", map[string]any{
			"entry_type": typ,
			"content":    "Entry of type " + typ,
		})
		if resp.StatusCode != 201 {
			t.Errorf("failed to create entry of type %s: %d", typ, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// Verify all created
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/journal", nil)
	var entries []map[string]any
	decodeJSON(t, resp, &entries)
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}
}
