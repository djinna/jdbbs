package srv

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeHelpFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"send-manuscript.md": "---\ntitle: Send your manuscript\nsummary: Upload the .docx from your editor.\nroutes: [\"/{client}/{project}/factory/\"]\ncovers:\n  - srv/static/factory.js\norder: 10\n---\n\nUpload the file. See [Inspect](inspect.md#fixes).\n\n## Sizes\n\nAnything up to the limit.\n",
		"inspect.md":         "---\ntitle: Inspect\nsummary: What the factory found.\n---\n# Ignored H1\n\nInspect lists findings such as **orphaned footnotes**.\n",
		"secret-admin.md":    "---\ntitle: Maintaining help\nvisibility: admin\n---\nOnly the studio sees this. zebrafinch\n",
		"wip.md":             "---\ntitle: Work in progress\nstatus: draft\n---\nNot approved yet. quokka\n",
		"weird.md":           "---\ntitle: Bad tier\nvisibility: clients-only\n---\nFails closed. narwhal\n",
		"README.md":          "# not a page\n",
		"Bad Name.md":        "skipped\n",
	}
	for n, c := range files {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func helpGet(t *testing.T, url string, admin bool) (int, string, http.Header) {
	t.Helper()
	req, _ := http.NewRequest("GET", url, nil)
	if admin {
		req.Header.Set("X-ExeDev-UserID", "test-admin")
		req.Header.Set("X-ExeDev-Email", "owner@example.test")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b), resp.Header
}

func TestHelpTiersAndRendering(t *testing.T) {
	s, ts, done := testServer(t)
	defer done()
	s.help = newHelpStore(writeHelpFixture(t))

	code, body, _ := helpGet(t, ts.URL+"/help/", false)
	if code != 200 || !strings.Contains(body, "Send your manuscript") || !strings.Contains(body, "Inspect") {
		t.Fatalf("index: %d %q", code, body[:min(300, len(body))])
	}
	for _, hidden := range []string{"Maintaining help", "Work in progress", "Bad tier", "Not yet covered"} {
		if strings.Contains(body, hidden) {
			t.Errorf("public index shows %q", hidden)
		}
	}
	// Grouped from site_pages: the factory route is client-tier.
	if !strings.Contains(body, "Your books") {
		t.Errorf("index missing the client group heading")
	}

	code, body, _ = helpGet(t, ts.URL+"/help/send-manuscript", false)
	if code != 200 || !strings.Contains(body, `href="/help/inspect#fixes"`) || !strings.Contains(body, `id="sizes"`) {
		t.Fatalf("page: %d, md links/heading ids not rendered", code)
	}
	if !strings.Contains(body, "This page explains") || !strings.Contains(body, "data-public-nav") {
		t.Errorf("page missing routes footer or shared nav")
	}
	_, body, _ = helpGet(t, ts.URL+"/help/inspect", false)
	if strings.Contains(body, "Ignored H1") {
		t.Errorf("leading H1 should be dropped when frontmatter has a title")
	}

	for _, slug := range []string{"secret-admin", "wip", "weird", "README", "Bad%20Name", "nope"} {
		if code, _, _ := helpGet(t, ts.URL+"/help/"+slug, false); code != 404 {
			t.Errorf("public GET /help/%s = %d, want 404", slug, code)
		}
		if code, _, _ := helpGet(t, ts.URL+"/help/"+slug+".md", false); code != 404 {
			t.Errorf("public GET /help/%s.md = %d, want 404", slug, code)
		}
	}
	for _, slug := range []string{"secret-admin", "wip", "weird"} {
		code, body, h := helpGet(t, ts.URL+"/help/"+slug, true)
		if code != 200 || h.Get("X-Robots-Tag") != "noindex" {
			t.Errorf("admin GET /help/%s = %d robots=%q", slug, code, h.Get("X-Robots-Tag"))
		}
		if !strings.Contains(body, "badge") {
			t.Errorf("admin view of %s lacks a tier badge", slug)
		}
	}
	_, body, _ = helpGet(t, ts.URL+"/help/", true)
	if !strings.Contains(body, "Maintaining help") || !strings.Contains(body, "Not yet covered") {
		t.Errorf("admin index should list admin/draft pages and the coverage gap list")
	}
}

func TestHelpMarkdownAndLLMs(t *testing.T) {
	s, ts, done := testServer(t)
	defer done()
	s.help = newHelpStore(writeHelpFixture(t))

	code, body, h := helpGet(t, ts.URL+"/help/send-manuscript.md", false)
	if code != 200 || !strings.HasPrefix(h.Get("Content-Type"), "text/markdown") {
		t.Fatalf(".md: %d %s", code, h.Get("Content-Type"))
	}
	if !strings.HasPrefix(body, "# Send your manuscript\n\n> Upload the .docx") || strings.Contains(body, "covers:") {
		t.Errorf(".md body wrong: %q", body[:min(120, len(body))])
	}

	_, body, _ = helpGet(t, ts.URL+"/help/llms.txt", false)
	if !strings.HasPrefix(body, "# jdbb studio") || !strings.Contains(body, "/help/send-manuscript.md): Upload") {
		t.Errorf("llms.txt: %q", body)
	}
	if strings.Contains(body, "secret-admin") || strings.Contains(body, "wip") {
		t.Errorf("llms.txt leaks admin/draft pages")
	}
}

func TestHelpSearch(t *testing.T) {
	s, ts, done := testServer(t)
	defer done()
	s.help = newHelpStore(writeHelpFixture(t))

	search := func(q string, admin bool) []helpHit {
		_, body, _ := helpGet(t, ts.URL+"/api/help/search?q="+q, admin)
		var out struct{ Results []helpHit }
		if err := json.Unmarshal([]byte(body), &out); err != nil {
			t.Fatalf("search %q: %v %s", q, err, body)
		}
		return out.Results
	}
	// Porter stemming + prefix: "footnote" finds "footnotes"; "manu" finds manuscript.
	if hits := search("footnote", false); len(hits) != 1 || hits[0].Slug != "inspect" || !strings.Contains(hits[0].Snippet, "<mark>") {
		t.Errorf("footnote: %+v", hits)
	}
	if hits := search("manu", false); len(hits) == 0 || hits[0].Slug != "send-manuscript" {
		t.Errorf("manu: %+v", hits)
	}
	// Tier filter.
	for _, q := range []string{"zebrafinch", "quokka", "narwhal"} {
		if hits := search(q, false); len(hits) != 0 {
			t.Errorf("public search %q leaked %+v", q, hits)
		}
		if hits := search(q, true); len(hits) != 1 {
			t.Errorf("admin search %q: %+v", q, hits)
		}
	}
	// FTS syntax never reaches the engine.
	for _, q := range []string{`"`, "NEAR(", "a*+OR+b", "-", "%22%29%28"} {
		search(q, false)
	}
	code, body, _ := helpGet(t, ts.URL+"/help/search?q=%3Cscript%3E", false)
	if code != 200 || strings.Contains(body, "<script>") {
		t.Errorf("search page must escape the query")
	}
}

func TestHelpReloadsOnEdit(t *testing.T) {
	dir := writeHelpFixture(t)
	h := newHelpStore(dir)
	_, pages, _ := h.snapshot()
	if pages["inspect"] == nil {
		t.Fatal("inspect not loaded")
	}
	os.WriteFile(filepath.Join(dir, "new-page.md"), []byte("---\ntitle: New\n---\nhello axolotl\n"), 0o644)
	h.checked = h.checked.Add(-helpRecheck) // skip the recheck throttle
	if _, pages, _ = h.snapshot(); pages["new-page"] == nil {
		t.Fatal("new page not picked up")
	}
	if hits := h.search("axolotl", false, 5); len(hits) != 1 {
		t.Errorf("index not rebuilt: %+v", hits)
	}
}

// The real docs/help pages must parse, carry a title and summary, name only
// routes that exist in site_pages, and link only to pages that exist.
func TestHelpContentIsWellFormed(t *testing.T) {
	s, _, done := testServer(t)
	defer done()
	dir := filepath.Join("..", "docs", "help")
	order, pages, err := loadHelpPages(dir)
	if err != nil {
		t.Skipf("no docs/help: %v", err)
	}
	sp := s.loadSitePages()
	for _, p := range order {
		if p.Title == p.Slug || p.Summary == "" {
			t.Errorf("%s.md: needs a title and a summary in frontmatter", p.Slug)
		}
		for _, r := range p.Routes {
			if _, ok := sp[r]; !ok {
				t.Errorf("%s.md: route %q is not in site_pages", p.Slug, r)
			}
		}
		for _, m := range helpInternalLinkRe.FindAllStringSubmatch(p.HTML, -1) {
			if pages[m[1]] == nil {
				t.Errorf("%s.md links to missing help page %q", p.Slug, m[1])
			}
		}
	}
}

// Every client or public route in site_pages is explained by a live-or-draft
// help page (frontmatter routes:) or deliberately opted out in helpOptOut.
// Adding a route? Add a help page naming it, or an opt-out line with a reason.
func TestHelpRouteCoverage(t *testing.T) {
	s, _, done := testServer(t)
	defer done()
	order, _, err := loadHelpPages(filepath.Join("..", "docs", "help"))
	if err != nil {
		t.Fatalf("docs/help: %v", err)
	}
	covered := map[string]string{}
	for _, p := range order {
		for _, r := range p.Routes {
			covered[r] = p.Slug
		}
	}
	sp := s.loadSitePages()
	for route, info := range sp {
		if info.Visibility != "client" && info.Visibility != "public" {
			continue
		}
		if info.Listed == "retired" || info.Status == "retire" {
			continue
		}
		_, opted := helpOptOut[route]
		if slug, ok := covered[route]; ok && opted {
			t.Errorf("%s is explained by %s.md and also opted out; drop the opt-out", route, slug)
		}
		if _, ok := covered[route]; !ok && !opted {
			t.Errorf("site_pages route %s (%s) has no help page and no opt-out in helpOptOut", route, info.Title)
		}
	}
	// Some rows exist only in the live DB (added from /admin/#pages, not by a
	// migration), so a stale opt-out is a note, not a failure. The admin view
	// of /help/ checks the live registry.
	for route := range helpOptOut {
		if _, ok := sp[route]; !ok {
			t.Logf("helpOptOut names %s, which no migration adds to site_pages", route)
		}
	}
}

func TestHelpForPath(t *testing.T) {
	for _, c := range []struct {
		pat, path string
		want      bool
	}{
		{"/{client}/{project}/factory/", "/sample-press/spring-novel/factory/", true},
		{"/{client}/{project}/factory/", "/sample-press/spring-novel/factory", true},
		{"/{client}/{project}/factory/", "/sample-press/factory/", false},
		{"/{client}/", "/sample-press/", true},
		{"/{client}/", "/", false},
		{"/portal", "/portal", true},
		{"/factory", "/factory/api", false},
	} {
		if got := helpRouteMatch(c.pat, c.path); got != c.want {
			t.Errorf("helpRouteMatch(%q, %q) = %v", c.pat, c.path, got)
		}
	}

	s, ts, done := testServer(t)
	defer done()
	s.help = newHelpStore(writeHelpFixture(t))
	get := func(path string, admin bool) (page string, visible []string) {
		_, body, _ := helpGet(t, ts.URL+"/api/help/for?path="+path, admin)
		var out struct {
			Page    *struct{ Slug string }
			Visible []string
		}
		if err := json.Unmarshal([]byte(body), &out); err != nil {
			t.Fatal(err)
		}
		if out.Page != nil {
			page = out.Page.Slug
		}
		return page, out.Visible
	}
	if p, vis := get("/sample-press/spring-novel/factory/", false); p != "send-manuscript" || strings.Contains(strings.Join(vis, ","), "wip") {
		t.Errorf("public for factory: %q %v", p, vis)
	}
	if p, _ := get("/nowhere", false); p != "" {
		t.Errorf("unmatched path returned %q", p)
	}
	if _, vis := get("/", true); !strings.Contains(strings.Join(vis, ","), "wip") {
		t.Errorf("admin visible list should include drafts: %v", vis)
	}
}
