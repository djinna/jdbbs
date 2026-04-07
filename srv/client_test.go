package srv

import (
	"net/http"
	"strings"
	"testing"
)

func seedClient(t *testing.T, s *Server, slug, name, password string) {
	t.Helper()
	_, err := s.DB.Exec(`INSERT INTO clients (slug, name, password_hash) VALUES (?, ?, ?)`, slug, name, hashToken(password))
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
	resp.Body.Close()

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
