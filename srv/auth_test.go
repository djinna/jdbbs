package srv

import (
	"net/http"
	"testing"
)

func TestAuthRequired(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create a project with auth
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Auth Test", "client_slug": "auth", "project_slug": "test", "start_date": "2025-01-01",
	})
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := itoa(int64(created["ID"].(float64)))

	// Set password on project
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/auth", map[string]string{
		"password": "secret123",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("set auth: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Request without auth should fail
	resp = apiRequest(t, ts, "GET", "/api/projects/"+pid+"/tasks", nil)
	if resp.StatusCode != 401 {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Request with admin header should succeed
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/tasks", nil)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 with admin header, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestVerifyAuth(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create project and set password
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Verify Test", "client_slug": "v", "project_slug": "t", "start_date": "2025-01-01",
	})
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := itoa(int64(created["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/auth", map[string]string{"password": "mypass"})
	resp.Body.Close()

	// Verify with wrong password
	resp = apiRequest(t, ts, "POST", "/api/projects/"+pid+"/verify", map[string]string{"password": "wrong"})
	if resp.StatusCode != 401 {
		t.Errorf("expected 401 for wrong password, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify with correct password - should set cookie
	resp = apiRequest(t, ts, "POST", "/api/projects/"+pid+"/verify", map[string]string{"password": "mypass"})
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 for correct password, got %d", resp.StatusCode)
	}
	// Check cookie was set
	cookies := resp.Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "prodcal_auth_"+pid {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected auth cookie to be set")
	}
	resp.Body.Close()
}

func TestNoAuthRequired(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create project WITHOUT setting auth
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Open Project", "client_slug": "open", "project_slug": "proj", "start_date": "2025-01-01",
	})
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := itoa(int64(created["ID"].(float64)))

	// Request without auth should succeed (no auth configured)
	resp = apiRequest(t, ts, "GET", "/api/projects/"+pid+"/tasks", nil)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 for project without auth, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAuthWithCookie(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create project with auth
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Cookie Test", "client_slug": "c", "project_slug": "t", "start_date": "2025-01-01",
	})
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := itoa(int64(created["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/auth", map[string]string{"password": "cookiepass"})
	resp.Body.Close()

	// Make request with cookie
	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL+"/api/projects/"+pid+"/tasks", nil)
	req.AddCookie(&http.Cookie{Name: "prodcal_auth_" + pid, Value: "cookiepass"})
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request with cookie: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 with valid cookie, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestRemoveProjectAuthRestoresOpenAccess(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Toggle Auth", "client_slug": "tog", "project_slug": "auth", "start_date": "2025-01-01",
	})
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := itoa(int64(created["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/auth", map[string]string{"password": "***"})
	if resp.StatusCode != 200 {
		t.Fatalf("set auth: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequest(t, ts, "GET", "/api/projects/"+pid+"/tasks", nil)
	if resp.StatusCode != 401 {
		t.Fatalf("expected 401 before auth removal, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "DELETE", "/api/projects/"+pid+"/auth", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("remove auth: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequest(t, ts, "GET", "/api/projects/"+pid+"/tasks", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 after auth removal, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestListProjectsRequiresAdmin(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Without admin header should fail
	resp := apiRequest(t, ts, "GET", "/api/projects", nil)
	if resp.StatusCode != 401 {
		t.Errorf("expected 401 without admin, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// With admin header should succeed
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects", nil)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 with admin, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestDownloadBookRequiresAuth(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	// Create a project with auth
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Book Auth Test", "client_slug": "ba", "project_slug": "test", "start_date": "2025-01-01",
	})
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := int64(created["ID"].(float64))

	// Set auth on the project
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+itoa(pid)+"/auth", map[string]string{"password": "secret"})
	resp.Body.Close()

	// Insert a book linked to the project, with its compiled PDF in
	// book_outputs (the single artifact store since migration 018).
	_, err := s.DB.Exec(`INSERT INTO books (title, author, status, project_id) VALUES (?, ?, ?, ?)`,
		"Test Book", "Author", "ready", pid)
	if err != nil {
		t.Fatalf("insert book: %v", err)
	}
	_, err = s.DB.Exec(`INSERT INTO book_outputs (book_id, output_format, output_data) VALUES (1, 'pdf', ?)`,
		[]byte("fakepdf"))
	if err != nil {
		t.Fatalf("insert book output: %v", err)
	}

	// Download without auth should fail
	resp = apiRequest(t, ts, "GET", "/api/books/1/download/pdf", nil)
	if resp.StatusCode != 401 {
		t.Errorf("expected 401 without auth for book download, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Download with admin should succeed
	resp = apiRequestAdmin(t, ts, "GET", "/api/books/1/download/pdf", nil)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 with admin for book download, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestCoverRequiresAuth(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create a project with auth
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Cover Auth Test", "client_slug": "ca", "project_slug": "test", "start_date": "2025-01-01",
	})
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := itoa(int64(created["ID"].(float64)))

	// Set auth on the project
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/auth", map[string]string{"password": "secret"})
	resp.Body.Close()

	// Cover without auth should fail
	resp = apiRequest(t, ts, "GET", "/api/projects/"+pid+"/book-spec/cover", nil)
	if resp.StatusCode != 401 {
		t.Errorf("expected 401 without auth for cover, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
