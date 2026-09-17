package srv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNavConvergence keeps the two sides of the app from drifting apart.
// Every admin page carries <nav data-admin-nav>, every customer page
// <nav data-client-nav>; theme.js fills both from one list, so a page that
// opts out silently loses the shared strip (which is how the four customer
// pages shipped with an empty top nav for weeks). Add a new page to the
// right list here when you add it to the router.
func TestNavConvergence(t *testing.T) {
	admin := []string{"admin.html", "store-admin.html", "factory-admin.html", "registrations.html", "docs-editor.html", "content-review.html"}
	client := []string{"client.html", "index.html", "factory.html", "transmittal.html"}
	check := func(files []string, attr string) {
		for _, f := range files {
			b, err := os.ReadFile(filepath.Join("static", f))
			if err != nil {
				t.Fatalf("%s: %v", f, err)
			}
			s := string(b)
			if !strings.Contains(s, `class="jdbb-masthead"`) {
				t.Errorf("%s: missing .jdbb-masthead header", f)
			}
			if !strings.Contains(s, "<nav "+attr) {
				t.Errorf("%s: masthead nav missing %s (shared nav strip won't render)", f, attr)
			}
			if !strings.Contains(s, "/static/theme.js") {
				t.Errorf("%s: does not load /static/theme.js (fills the nav)", f)
			}
		}
	}
	check(admin, "data-admin-nav")
	check(client, "data-client-nav")

	// Server-rendered admin pages live in Go source; same contract.
	b, err := os.ReadFile("email_preview.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "<nav data-admin-nav") {
		t.Errorf("email_preview.go: gallery masthead missing data-admin-nav")
	}

	// Every admin route the server registers should be reachable from the
	// shared nav list in theme.js (or be a sub-route of one that is).
	js, err := os.ReadFile(filepath.Join("static", "theme.js"))
	if err != nil {
		t.Fatal(err)
	}
	srv, err := os.ReadFile("server.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(srv), "\n") {
		i := strings.Index(line, `"GET /admin/`)
		if i < 0 {
			continue
		}
		route := line[i+len(`"GET `):]
		route = route[:strings.Index(route, `"`)]
		route = strings.TrimSuffix(route, "{$}")
		if strings.Contains(route, "{") { // /admin/email-preview/{kind}
			route = route[:strings.Index(route, "{")]
		}
		if !strings.Contains(string(js), "'"+route+"'") {
			t.Errorf("server.go registers %s but theme.js ADMIN_NAV has no entry for it", route)
		}
	}
}
