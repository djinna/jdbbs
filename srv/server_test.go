package srv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"srv.exe.dev/db"
)

// testServer creates a test server with an in-memory SQLite database.
func testServer(t *testing.T) (*Server, *httptest.Server, func()) {
	t.Helper()

	memDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}

	if err := db.RunMigrations(memDB); err != nil {
		memDB.Close()
		t.Fatalf("run migrations: %v", err)
	}

	s := &Server{
		DB:       memDB,
		Hostname: "test.local",
		Email:    nil, // No email for most tests
	}

	ts := httptest.NewServer(s.Handler())

	cleanup := func() {
		ts.Close()
		memDB.Close()
	}

	return s, ts, cleanup
}

// apiRequest makes a request to the test server and returns the response.
func apiRequest(t *testing.T, ts *httptest.Server, method, path string, body any) *http.Response {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, ts.URL+path, reqBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

// apiRequestAdmin makes a request with the admin header.
func apiRequestAdmin(t *testing.T, ts *httptest.Server, method, path string, body any) *http.Response {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, ts.URL+path, reqBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ExeDev-UserID", "test-admin")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

// decodeJSON decodes the response body into v.
func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

// --- Smoke Tests ---

func TestHealthz(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequest(t, ts, "GET", "/healthz", nil)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	decodeJSON(t, resp, &body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body)
	}
}

func TestProjectCRUD(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create project (as admin)
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Test Project",
		"start_date":   "2025-01-01",
		"client_slug":  "test",
		"project_slug": "proj1",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	// Response uses Go struct field names (ID, Name, etc.)
	projectID := int64(created["ID"].(float64))

	// Get project
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(projectID), nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get project: expected 200, got %d", resp.StatusCode)
	}
	var getResp map[string]any
	decodeJSON(t, resp, &getResp)
	project := getResp["project"].(map[string]any)
	if project["Name"] != "Test Project" {
		t.Errorf("expected name 'Test Project', got %v", project["Name"])
	}

	// Update project
	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+itoa(projectID), map[string]string{
		"name":         "Updated Project",
		"start_date":   "2025-02-01",
		"client_slug":  "test",
		"project_slug": "proj1",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update project: expected 200, got %d", resp.StatusCode)
	}

	// Verify update
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(projectID), nil)
	decodeJSON(t, resp, &getResp)
	project = getResp["project"].(map[string]any)
	if project["Name"] != "Updated Project" {
		t.Errorf("expected name 'Updated Project', got %v", project["Name"])
	}

	// Delete endpoint is disabled in favor of archive
	resp = apiRequestAdmin(t, ts, "DELETE", "/api/projects/"+itoa(projectID), nil)
	if resp.StatusCode != 405 {
		t.Fatalf("delete project: expected 405, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Archive project instead
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+itoa(projectID)+"/archive", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("archive project: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Archived projects should not resolve by public/project path APIs anymore
	resp = apiRequestAdmin(t, ts, "GET", "/api/project-by-path/test/proj1", nil)
	if resp.StatusCode != 404 {
		t.Errorf("expected 404 after archive by path, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestArchiveProjectRequiresAdminHeaderEvenForOpenProject(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Archive Header Guard", "client_slug": "ahg", "project_slug": "one", "start_date": "2025-01-01",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := itoa(int64(created["ID"].(float64)))

	resp = apiRequest(t, ts, "POST", "/api/projects/"+pid+"/archive", nil)
	if resp.StatusCode != 401 {
		t.Fatalf("archive project without admin header: expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("project should still exist after unauthorized archive, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}
