package srv

import (
	"net/http"
	"strings"
	"testing"
)

func seedClient(t *testing.T, s *Server, slug, name, password string) {
	t.Helper()
	hash, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	_, err = s.DB.Exec(`INSERT INTO clients (slug, name, password_hash) VALUES (?, ?, ?)`, slug, name, hash)
	if err != nil {
		t.Fatalf("seed client: %v", err)
	}
}

func TestClientCreateProjectRequiresClientAuth(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	seedClient(t, s, "vgr", "VGR", "pw123")

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Auth Seed",
		"start_date":   "2025-01-01",
		"client_slug":  "vgr",
		"project_slug": "seed",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("seed project: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequest(t, ts, "POST", "/api/clients/vgr/projects", map[string]string{
		"name":         "New Client Project",
		"project_slug": "new-client-project",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without client auth, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestClientCreateProjectWithClientAuth(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	seedClient(t, s, "vgr", "VGR", "pw123")

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Auth Seed",
		"start_date":   "2025-01-01",
		"client_slug":  "vgr",
		"project_slug": "seed",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("seed project: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	client := &http.Client{}
	req, err := http.NewRequest("POST", ts.URL+"/api/clients/vgr/verify", strings.NewReader(`{"password":"pw123"}`))
	if err != nil {
		t.Fatalf("create verify request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("verify client auth: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("verify client auth: expected 200, got %d", resp.StatusCode)
	}

	var authCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "prodcal_client_vgr" {
			authCookie = c
			break
		}
	}
	resp.Body.Close()
	if authCookie == nil {
		t.Fatal("expected client auth cookie")
	}

	req, err = http.NewRequest("POST", ts.URL+"/api/clients/vgr/projects", strings.NewReader(`{"name":"Second Project","project_slug":"second-project","start_date":"2025-02-01"}`))
	if err != nil {
		t.Fatalf("create project request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(authCookie)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("create client project: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("create client project: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	projectID := itoa(int64(created["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+projectID+"/tasks", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list seeded client tasks: expected 200, got %d", resp.StatusCode)
	}
	var tasks []map[string]any
	decodeJSON(t, resp, &tasks)
	if len(tasks) != 31 {
		t.Fatalf("expected 31 seeded tasks for client project, got %d", len(tasks))
	}
	if tasks[0]["Title"] != "Ms transmittal" {
		t.Fatalf("expected first seeded task to be 'Ms transmittal', got %v", tasks[0]["Title"])
	}
	if tasks[30]["Title"] != "Log in mechs" {
		t.Fatalf("expected last seeded task to be 'Log in mechs', got %v", tasks[30]["Title"])
	}

	req, err = http.NewRequest("GET", ts.URL+"/api/clients/vgr/projects", nil)
	if err != nil {
		t.Fatalf("list client projects request: %v", err)
	}
	req.AddCookie(authCookie)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("list client projects: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("list client projects: expected 200, got %d", resp.StatusCode)
	}
	var projects []map[string]any
	decodeJSON(t, resp, &projects)
	if len(projects) != 2 {
		t.Fatalf("expected 2 client projects, got %d", len(projects))
	}
	found := false
	for _, p := range projects {
		if p["project_slug"] == "second-project" || p["ProjectSlug"] == "second-project" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected second-project in client project list, got %+v", projects)
	}
}

func TestAdminCreateProjectSeedsStandardWorkflow(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Admin Seeded",
		"start_date":   "2026-04-09",
		"client_slug":  "vgr",
		"project_slug": "admin-seeded",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create admin project: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	projectID := itoa(int64(created["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+projectID+"/tasks", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list seeded admin tasks: expected 200, got %d", resp.StatusCode)
	}
	var tasks []map[string]any
	decodeJSON(t, resp, &tasks)
	if len(tasks) != 31 {
		t.Fatalf("expected 31 seeded tasks for admin project, got %d", len(tasks))
	}
	if tasks[0]["Title"] != "Ms transmittal" {
		t.Fatalf("expected first seeded task to be 'Ms transmittal', got %v", tasks[0]["Title"])
	}
	if tasks[23]["Title"] != "Send mechs to printer" {
		t.Fatalf("expected seeded printer task, got %v", tasks[23]["Title"])
	}
}

// TestPortalCreateProjectPasswordlessClientAdminOnly: a passwordless client
// (empty password_hash) has no auth gate, so portal project creation is
// forbidden (403) for anonymous requests and creates nothing; the admin
// header (X-ExeDev-UserID) is still allowed through and gets 201.
func TestPortalCreateProjectPasswordlessClientAdminOnly(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	if _, err := s.DB.Exec(`INSERT INTO clients (slug, name, password_hash) VALUES ('nopw', 'No Password', '')`); err != nil {
		t.Fatalf("seed passwordless client: %v", err)
	}

	resp := apiRequest(t, ts, "POST", "/api/clients/nopw/projects", map[string]string{
		"name":         "Anon Project",
		"project_slug": "anon-project",
		"start_date":   "2025-05-01",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("anonymous create for passwordless client: expected 403, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	if n := countRows(t, s, `SELECT COUNT(*) FROM projects WHERE client_slug = 'nopw'`); n != 0 {
		t.Fatalf("403 must not create a project, found %d rows for client", n)
	}

	resp = apiRequestAdmin(t, ts, "POST", "/api/clients/nopw/projects", map[string]string{
		"name":         "Admin Project",
		"project_slug": "admin-project",
		"start_date":   "2025-05-01",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("admin create for passwordless client: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	if created["ProjectSlug"] != "admin-project" {
		t.Errorf("expected created project slug \"admin-project\", got %v", created["ProjectSlug"])
	}

	resp = apiRequest(t, ts, "GET", "/api/project-by-path/nopw/admin-project", nil)
	if resp.StatusCode != 200 {
		t.Errorf("created project should resolve by path, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
