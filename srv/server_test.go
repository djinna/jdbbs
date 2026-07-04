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

// countRows runs a COUNT(*) query against the test DB.
func countRows(t *testing.T, s *Server, query string, args ...any) int {
	t.Helper()
	var n int
	if err := s.DB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return n
}

// createProjectT creates a project as admin and returns its ID.
func createProjectT(t *testing.T, ts *httptest.Server, name, clientSlug, projectSlug string) int64 {
	t.Helper()
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         name,
		"start_date":   "2025-01-01",
		"client_slug":  clientSlug,
		"project_slug": projectSlug,
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project %s/%s: expected 201, got %d", clientSlug, projectSlug, resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	return int64(created["ID"].(float64))
}

// TestDuplicateProjectReportsTasksCopied: a successful duplicate reports
// tasks_copied equal to the source's task count, and the new project really
// holds copies of the source tasks (not a fresh standard-workflow seed).
func TestDuplicateProjectReportsTasksCopied(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	srcID := createProjectT(t, ts, "Dup Source", "dupc", "src")

	// Add two custom tasks so the source count differs from the standard
	// seed count — if duplicate re-seeded instead of copying, counts and
	// titles below would not match.
	for _, title := range []string{"Custom Copy Task A", "Custom Copy Task B"} {
		resp := apiRequestAdmin(t, ts, "POST", "/api/projects/"+itoa(srcID)+"/tasks", map[string]any{
			"title":      title,
			"sort_order": 100,
		})
		if resp.StatusCode != 201 {
			t.Fatalf("add task %q: expected 201, got %d", title, resp.StatusCode)
		}
		resp.Body.Close()
	}

	resp := apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(srcID)+"/tasks", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list source tasks: expected 200, got %d", resp.StatusCode)
	}
	var srcTasks []map[string]any
	decodeJSON(t, resp, &srcTasks)
	n := len(srcTasks)
	if n < 2 {
		t.Fatalf("source project should have at least the 2 custom tasks, got %d", n)
	}

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+itoa(srcID)+"/duplicate", map[string]string{
		"name":         "Dup Copy",
		"start_date":   "2025-03-01",
		"client_slug":  "dupc",
		"project_slug": "copy",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("duplicate project: expected 200, got %d", resp.StatusCode)
	}
	var dup map[string]any
	decodeJSON(t, resp, &dup)
	copied, ok := dup["tasks_copied"].(float64)
	if !ok {
		t.Fatalf("expected numeric tasks_copied in response, got %+v", dup)
	}
	if int(copied) != n {
		t.Errorf("tasks_copied = %d, want %d (source task count)", int(copied), n)
	}
	newID := int64(dup["project"].(map[string]any)["ID"].(float64))

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(newID)+"/tasks", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list duplicated tasks: expected 200, got %d", resp.StatusCode)
	}
	var newTasks []map[string]any
	decodeJSON(t, resp, &newTasks)
	if len(newTasks) != n {
		t.Errorf("duplicated project has %d tasks, want %d", len(newTasks), n)
	}
	found := 0
	for _, task := range newTasks {
		if task["Title"] == "Custom Copy Task A" || task["Title"] == "Custom Copy Task B" {
			found++
		}
	}
	if found != 2 {
		t.Errorf("duplicated project should contain both custom tasks (proof of copy, not re-seed), found %d", found)
	}
}

// TestDuplicateProjectSlugCollisionRollsBack: duplicating onto an
// already-taken client/project slug pair returns 409 with the slug-conflict
// error and leaves the database untouched — no new project row, no orphan
// task rows.
func TestDuplicateProjectSlugCollisionRollsBack(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	srcID := createProjectT(t, ts, "Dup Source", "dupc", "src")
	createProjectT(t, ts, "Slug Owner", "dupc", "taken")

	projectsBefore := countRows(t, s, `SELECT COUNT(*) FROM projects`)
	tasksBefore := countRows(t, s, `SELECT COUNT(*) FROM tasks`)

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects/"+itoa(srcID)+"/duplicate", map[string]string{
		"name":         "Collision Copy",
		"start_date":   "2025-03-01",
		"client_slug":  "dupc",
		"project_slug": "taken",
	})
	if resp.StatusCode != 409 {
		t.Fatalf("duplicate onto taken slug: expected 409, got %d", resp.StatusCode)
	}
	var errBody map[string]string
	decodeJSON(t, resp, &errBody)
	if errBody["error"] != "a project with that client/project slug already exists" {
		t.Errorf("unexpected 409 error message: %q", errBody["error"])
	}

	if got := countRows(t, s, `SELECT COUNT(*) FROM projects`); got != projectsBefore {
		t.Errorf("project count changed after failed duplicate: %d -> %d", projectsBefore, got)
	}
	if got := countRows(t, s, `SELECT COUNT(*) FROM tasks`); got != tasksBefore {
		t.Errorf("task count changed after failed duplicate (orphan tasks leaked): %d -> %d", tasksBefore, got)
	}
}

// TestCreateProjectSlugCollision409: creating a project whose
// client/project slug pair already exists returns 409 (not 500) and does not
// insert a second row.
func TestCreateProjectSlugCollision409(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	createProjectT(t, ts, "First", "colc", "one")
	projectsBefore := countRows(t, s, `SELECT COUNT(*) FROM projects`)

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Second",
		"start_date":   "2025-02-01",
		"client_slug":  "colc",
		"project_slug": "one",
	})
	if resp.StatusCode != 409 {
		t.Fatalf("create with taken slug: expected 409, got %d", resp.StatusCode)
	}
	var errBody map[string]string
	decodeJSON(t, resp, &errBody)
	if errBody["error"] != "a project with that client/project slug already exists" {
		t.Errorf("unexpected 409 error message: %q", errBody["error"])
	}
	if got := countRows(t, s, `SELECT COUNT(*) FROM projects`); got != projectsBefore {
		t.Errorf("project count changed after 409 create: %d -> %d", projectsBefore, got)
	}
}

// TestUpdateProjectSlugCollision409: renaming a project onto another
// project's client/project slug pair returns 409 (not 500) and leaves the
// target project's slugs unchanged.
func TestUpdateProjectSlugCollision409(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	createProjectT(t, ts, "Owner", "colc", "one")
	otherID := createProjectT(t, ts, "Mover", "colc", "two")

	resp := apiRequestAdmin(t, ts, "PUT", "/api/projects/"+itoa(otherID), map[string]string{
		"name":         "Mover",
		"start_date":   "2025-01-01",
		"client_slug":  "colc",
		"project_slug": "one",
	})
	if resp.StatusCode != 409 {
		t.Fatalf("update onto taken slug: expected 409, got %d", resp.StatusCode)
	}
	var errBody map[string]string
	decodeJSON(t, resp, &errBody)
	if errBody["error"] != "a project with that client/project slug already exists" {
		t.Errorf("unexpected 409 error message: %q", errBody["error"])
	}

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(otherID), nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get project after failed update: expected 200, got %d", resp.StatusCode)
	}
	var getResp map[string]any
	decodeJSON(t, resp, &getResp)
	if slug := getResp["project"].(map[string]any)["ProjectSlug"]; slug != "two" {
		t.Errorf("failed update must not change slugs: got project_slug %v, want \"two\"", slug)
	}
}

// TestPublicConfigContactEmail: /api/public/config serves the CONTACT_EMAIL
// env var when set, and falls back to the documented public contact address
// when unset. The fallback IS the endpoint's payload contract (rendered as
// the public mailto link), not an incidental config default.
func TestPublicConfigContactEmail(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	t.Setenv("CONTACT_EMAIL", "ops@example.com")
	resp := apiRequest(t, ts, "GET", "/api/public/config", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("public config: expected 200, got %d", resp.StatusCode)
	}
	var cfg map[string]string
	decodeJSON(t, resp, &cfg)
	if cfg["contact_email"] != "ops@example.com" {
		t.Errorf("with CONTACT_EMAIL set: contact_email = %q, want \"ops@example.com\"", cfg["contact_email"])
	}

	t.Setenv("CONTACT_EMAIL", "")
	resp = apiRequest(t, ts, "GET", "/api/public/config", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("public config (no env): expected 200, got %d", resp.StatusCode)
	}
	decodeJSON(t, resp, &cfg)
	if cfg["contact_email"] != "j@djinna.com" {
		t.Errorf("without CONTACT_EMAIL: contact_email = %q, want \"j@djinna.com\"", cfg["contact_email"])
	}
}
