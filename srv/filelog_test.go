package srv

import (
	"testing"
	"time"
)

func TestFileLogCRUD(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create project
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "FileLog Test", "client_slug": "fl", "project_slug": "test", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// Create file log entry
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/file-log", map[string]any{
		"direction":     "inbound",
		"filename":      "manuscript.docx",
		"file_type":     "Word",
		"sent_by":       "Author",
		"received_by":   "Editor",
		"notes":         "First draft",
		"transfer_date": "2025-01-15",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create file log: expected 201, got %d", resp.StatusCode)
	}
	var entry map[string]any
	decodeJSON(t, resp, &entry)
	entryID := itoa(int64(entry["id"].(float64)))

	if entry["filename"] != "manuscript.docx" {
		t.Errorf("expected filename 'manuscript.docx', got %v", entry["filename"])
	}
	if entry["direction"] != "inbound" {
		t.Errorf("expected direction 'inbound', got %v", entry["direction"])
	}

	// List file log
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/file-log", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list file log: expected 200, got %d", resp.StatusCode)
	}
	var entries []map[string]any
	decodeJSON(t, resp, &entries)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	// Delete entry
	resp = apiRequestAdmin(t, ts, "DELETE", "/api/projects/"+pid+"/file-log/"+entryID, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("delete file log: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify deletion
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/file-log", nil)
	decodeJSON(t, resp, &entries)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after delete, got %d", len(entries))
	}
}

func TestFileLogDefaults(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Defaults Test", "client_slug": "def", "project_slug": "test", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// Create entry without direction or date - should default
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/file-log", map[string]any{
		"filename": "test.pdf",
	})
	var entry map[string]any
	decodeJSON(t, resp, &entry)

	// Should default to "inbound"
	if entry["direction"] != "inbound" {
		t.Errorf("expected default direction 'inbound', got %v", entry["direction"])
	}

	// Should default to today
	today := time.Now().Format("2006-01-02")
	if entry["transfer_date"] != today {
		t.Errorf("expected default transfer_date '%s', got %v", today, entry["transfer_date"])
	}
}

func TestFileLogOutbound(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Outbound Test", "client_slug": "out", "project_slug": "t", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/file-log", map[string]any{
		"direction": "outbound",
		"filename":  "proofs.pdf",
		"sent_by":   "Editor",
		"received_by": "Author",
	})
	var entry map[string]any
	decodeJSON(t, resp, &entry)

	if entry["direction"] != "outbound" {
		t.Errorf("expected direction 'outbound', got %v", entry["direction"])
	}
}
