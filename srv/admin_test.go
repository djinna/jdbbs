package srv

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminCreateClientWithOptionalPassword(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/admin/clients", map[string]string{
		"name":     "Venkatesh Rao",
		"slug":     "vgr",
		"password": "pw123",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create client: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	if created["slug"] != "vgr" {
		t.Fatalf("expected slug vgr, got %v", created["slug"])
	}
	if created["name"] != "Venkatesh Rao" {
		t.Fatalf("expected client name, got %v", created["name"])
	}
	if created["has_auth"] != true {
		t.Fatalf("expected has_auth true, got %v", created["has_auth"])
	}

	resp = apiRequestAdmin(t, ts, "GET", "/api/admin/clients", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list clients: expected 200, got %d", resp.StatusCode)
	}
	var clients []map[string]any
	decodeJSON(t, resp, &clients)
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}
	if clients[0]["project_count"] != float64(0) {
		t.Fatalf("expected zero projects for new client, got %v", clients[0]["project_count"])
	}
}

func TestAdminCreateClientWithoutPassword(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/admin/clients", map[string]string{
		"name": "Open Client",
		"slug": "open-client",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create open client: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	if created["has_auth"] != false {
		t.Fatalf("expected has_auth false, got %v", created["has_auth"])
	}
}

func TestAdminCreateClientDuplicateSlug(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/admin/clients", map[string]string{
		"name": "Client One",
		"slug": "dup-client",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("initial create client: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "POST", "/api/admin/clients", map[string]string{
		"name": "Client Two",
		"slug": "dup-client",
	})
	if resp.StatusCode != 409 {
		t.Fatalf("duplicate create client: expected 409, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAdminProjectListSeparatesArchivedProjects(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Active Project", "client_slug": "act", "project_slug": "one", "start_date": "2025-01-01",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create active project: expected 201, got %d", resp.StatusCode)
	}
	var active map[string]any
	decodeJSON(t, resp, &active)

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Archive Me", "client_slug": "arc", "project_slug": "one", "start_date": "2025-01-01",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create archived project: expected 201, got %d", resp.StatusCode)
	}
	var archived map[string]any
	decodeJSON(t, resp, &archived)
	archivedID := itoa(int64(archived["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+archivedID+"/archive", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("archive project: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "GET", "/api/admin/projects", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list active projects: expected 200, got %d", resp.StatusCode)
	}
	var activeList []map[string]any
	decodeJSON(t, resp, &activeList)
	if len(activeList) != 1 || activeList[0]["name"] != "Active Project" {
		t.Fatalf("expected only active project in main admin list, got %v", activeList)
	}

	resp = apiRequestAdmin(t, ts, "GET", "/api/admin/projects?archived=1", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list archived projects: expected 200, got %d", resp.StatusCode)
	}
	var archivedList []map[string]any
	decodeJSON(t, resp, &archivedList)
	if len(archivedList) != 1 || archivedList[0]["name"] != "Archive Me" {
		t.Fatalf("expected only archived project in archived admin list, got %v", archivedList)
	}

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+archivedID+"/restore", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("restore project: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "GET", "/api/admin/projects", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list active projects after restore: expected 200, got %d", resp.StatusCode)
	}
	decodeJSON(t, resp, &activeList)
	if len(activeList) != 2 {
		t.Fatalf("expected restored project back in active list, got %v", activeList)
	}
}

func TestAdminDashboardIncludesProjectAndClientActions(t *testing.T) {
	resp := apiRequestAdmin(t, testServerForHTML(t), "GET", "/admin/", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("admin dashboard: expected 200, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)
	if !contains(text, "new-project-btn") {
		t.Fatalf("expected admin dashboard to include new-project-btn")
	}
	if !contains(text, "new-client-btn") {
		t.Fatalf("expected admin dashboard to include new-client-btn")
	}
	if !contains(text, "Archived") {
		t.Fatalf("expected admin dashboard to include Archived navigation")
	}
	if !contains(text, "archiveProject(") {
		t.Fatalf("expected admin dashboard to include archiveProject action")
	}
	if !contains(text, "Project password") {
		t.Fatalf("expected admin dashboard to include clearer auth label")
	}
	if !contains(text, "Open access") {
		t.Fatalf("expected admin dashboard to include open-access auth label")
	}
	if !contains(text, "Auth labels") {
		t.Fatalf("expected admin dashboard to include auth help note")
	}
}

func TestAdminNewProjectModalIncludesClientSuggestions(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	resp := apiRequestAdmin(t, ts, "POST", "/api/admin/clients", map[string]string{
		"name": "Venkatesh Rao",
		"slug": "vgr",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create client for suggestions: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "GET", "/admin/", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("admin dashboard: expected 200, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)
	if !contains(text, "np-client-list") {
		t.Fatalf("expected datalist for client suggestions")
	}
	if !contains(text, "Type an existing client slug or enter a new one.") {
		t.Fatalf("expected helper copy for client slug field")
	}
}

func testServerForHTML(t *testing.T) *httptest.Server {
	_, ts, cleanup := testServer(t)
	t.Cleanup(cleanup)
	return ts
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
