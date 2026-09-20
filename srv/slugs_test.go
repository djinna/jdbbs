package srv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// adminJSON issues an admin API request and decodes the JSON response.
func adminJSON(t *testing.T, ts *httptest.Server, method, path string, body any, out any) int {
	t.Helper()
	resp := apiRequestAdmin(t, ts, method, path, body)
	defer resp.Body.Close()
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

func TestDeriveClientSlug(t *testing.T) {
	cases := map[string]string{
		"Mike Casey":           "mcasey",
		"  mike   casey ":      "mcasey",
		"Jean-Luc Picard":      "jpicard",
		"José Núñez":           "jnunez",
		"Ludwig van Beethoven": "lbeethoven",
		"Cher":                 "cher",
		"":                     "author",
		"Venkatesh Rao":        "vrao",
		"Protocol Institute":   "protocolinstitute",
		"Stripe Press":         "stripepress",
		"Summer of Protocols":  "sprotocols",
	}
	for in, want := range cases {
		if got := deriveClientSlug(in); got != want {
			t.Errorf("deriveClientSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUniqueClientSlugCollision(t *testing.T) {
	s, _, cleanup := testServer(t)
	defer cleanup()
	ctx := context.Background()
	got, err := uniqueClientSlug(ctx, s.DB, "Mike Casey")
	if err != nil || got != "mcasey" {
		t.Fatalf("first = %q, %v", got, err)
	}
	seedClient(t, s, "mcasey", "Mike Casey", "")
	got, _ = uniqueClientSlug(ctx, s.DB, "Mary Casey")
	if got != "mcasey2" {
		t.Fatalf("collision = %q, want mcasey2", got)
	}
	// An old slug still serving as an alias is not handed to a new client.
	if _, err := s.DB.Exec(`INSERT INTO slug_aliases (old_client_slug, new_client_slug) VALUES ('mcasey2', 'mcasey')`); err != nil {
		t.Fatal(err)
	}
	got, _ = uniqueClientSlug(ctx, s.DB, "Max Casey")
	if got != "mcasey3" {
		t.Fatalf("alias collision = %q, want mcasey3", got)
	}
}

func TestNextProjectSlugSequence(t *testing.T) {
	s, _, cleanup := testServer(t)
	defer cleanup()
	ctx := context.Background()
	seedClient(t, s, "mcasey", "Mike Casey", "")
	got, err := nextProjectSlug(ctx, s.DB, "mcasey")
	if err != nil || got != "book-001" {
		t.Fatalf("first = %q, %v", got, err)
	}
	for _, slug := range []string{"book-001", "book-002", "legacy-title"} {
		if _, err := s.DB.Exec(`INSERT INTO projects (name, client_slug, project_slug) VALUES ('x', 'mcasey', ?)`, slug); err != nil {
			t.Fatal(err)
		}
	}
	// Archiving book-002 must not free the number.
	if _, err := s.DB.Exec(`UPDATE projects SET archived_at = CURRENT_TIMESTAMP WHERE project_slug = 'book-002'`); err != nil {
		t.Fatal(err)
	}
	got, _ = nextProjectSlug(ctx, s.DB, "mcasey")
	if got != "book-003" {
		t.Fatalf("next = %q, want book-003", got)
	}
	// Other clients have their own sequence.
	got, _ = nextProjectSlug(ctx, s.DB, "other")
	if got != "book-001" {
		t.Fatalf("other client = %q, want book-001", got)
	}
}

func TestRenameClientAndProjectKeepsOldURLs(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	seedClient(t, s, "mike-casey", "Mike Casey", "")
	if _, err := s.DB.Exec(`INSERT INTO projects (id, name, client_slug, project_slug) VALUES (18, 'Building', 'mike-casey', 'building-in-the-wrong-ma')`); err != nil {
		t.Fatal(err)
	}
	// A sign-in link references clients.slug by FOREIGN KEY; the rename must
	// carry it along (regression: 2026-09-20, renaming snitkey → monstrous
	// failed with "FOREIGN KEY constraint failed").
	if _, err := s.DB.Exec(`INSERT INTO login_links (client_slug, token_hash, email, expires_at) VALUES ('mike-casey', 'h1', 'm@x.io', '2030-01-01')`); err != nil {
		t.Fatal(err)
	}

	// Suggestion endpoint.
	var sug map[string]string
	if code := adminJSON(t, ts, "GET", "/api/admin/slug-suggest?name=Mike+Casey&client=mike-casey", nil, &sug); code != 200 {
		t.Fatalf("suggest: %d", code)
	}
	if sug["client_slug"] != "mcasey" || sug["project_slug"] != "book-001" {
		t.Fatalf("suggest = %v", sug)
	}

	// Rename project slug, then client slug.
	var out map[string]any
	if code := adminJSON(t, ts, "POST", "/api/admin/projects/18/slug", map[string]any{"project_slug": "book-001"}, &out); code != 200 {
		t.Fatalf("rename project: %d %v", code, out)
	}
	if code := adminJSON(t, ts, "POST", "/api/admin/clients/mike-casey/rename", map[string]any{"slug": "mcasey"}, &out); code != 200 {
		t.Fatalf("rename client: %d %v", code, out)
	}
	var cs, ps string
	var ll string
	if err := s.DB.QueryRow(`SELECT client_slug FROM login_links WHERE token_hash = 'h1'`).Scan(&ll); err != nil || ll != "mcasey" {
		t.Fatalf("login_links not carried along: %q %v", ll, err)
	}
	if err := s.DB.QueryRow(`SELECT client_slug, project_slug FROM projects WHERE id = 18`).Scan(&cs, &ps); err != nil || cs != "mcasey" || ps != "book-001" {
		t.Fatalf("project row = %s/%s, %v", cs, ps, err)
	}

	noFollow := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	expect := func(path, want string) {
		t.Helper()
		resp, err := noFollow.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 301 || resp.Header.Get("Location") != want {
			t.Errorf("%s → %d %s, want 301 %s", path, resp.StatusCode, resp.Header.Get("Location"), want)
		}
	}
	// The link in the Sep 4 email (old client + old project, deeper path).
	expect("/mike-casey/building-in-the-wrong-ma/factory/", "/mcasey/book-001/factory/")
	// Old client portal.
	expect("/mike-casey/", "/mcasey/")
	// Old client slug + new project slug (client-level alias applies).
	expect("/mike-casey/book-001/", "/mcasey/book-001/")
	// Query string preserved.
	expect("/mike-casey/building-in-the-wrong-ma/?x=1", "/mcasey/book-001/?x=1")

	// The new address serves normally.
	resp, err := noFollow.Get(ts.URL + "/mcasey/book-001/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("new path status %d", resp.StatusCode)
	}

	// Renaming to a taken slug is refused.
	seedClient(t, s, "taken", "Taken", "")
	if code := adminJSON(t, ts, "POST", "/api/admin/clients/mcasey/rename", map[string]any{"slug": "taken"}, &out); code != 409 {
		t.Fatalf("rename to taken: %d", code)
	}
	if !strings.Contains(out["error"].(string), "exists") {
		t.Fatalf("error = %v", out["error"])
	}
}

func TestLooksLikeOrgAndGreeting(t *testing.T) {
	for name, org := range map[string]bool{
		"Protocol Institute": true, "Stripe Press": true, "Acme Publishing LLC": true,
		"Venkatesh Rao": false, "Mike Casey": false, "Cher": false, "": false,
	} {
		if got := looksLikeOrg(name); got != org {
			t.Errorf("looksLikeOrg(%q) = %v, want %v", name, got, org)
		}
	}
	if got := firstName("Protocol Institute"); got != "Protocol Institute" {
		t.Errorf("firstName(org) = %q", got)
	}
	if got := firstName("Venkatesh Rao"); got != "Venkatesh" {
		t.Errorf("firstName(person) = %q", got)
	}
}
