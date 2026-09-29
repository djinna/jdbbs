package srv

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// Help: the studio's user help (punch list 8.5, plan in
// docs/NEXT_SESSION_PROMPT_2026-09-29.md Phase 4).
//
// Content is docs/help/*.md, one page per file, with a small frontmatter
// block (see docs/help/README.md). The server reads the directory from disk
// and re-reads it when anything changes, so an edited page is live without a
// rebuild — the same model as /admin/runs/ and the jdbbs-public docs.
//
// Tiers: content is public by default ("garage door open"). Pages marked
// `visibility: admin` and every `status: draft` page are shown only to the
// admin; everyone else gets a 404, in the page view, the index, search,
// llms.txt and the .md view alike.
//
// Routes: /help/ (index, grouped), /help/{slug}, /help/{slug}.md (raw, for
// agents), /help/llms.txt, /help/search?q=, /api/help/search?q= (JSON).

// helpDir is where the help pages live. PRODCAL_HELP_DIR overrides; otherwise
// docs/help under the working directory (the systemd unit runs in the repo),
// falling back to the VM checkout.
func helpDir() string {
	if d := os.Getenv("PRODCAL_HELP_DIR"); d != "" {
		return d
	}
	if st, err := os.Stat("docs/help"); err == nil && st.IsDir() {
		if abs, err := filepath.Abs("docs/help"); err == nil {
			return abs
		}
	}
	return "/home/exedev/prodcal/docs/help"
}

// helpPage is one docs/help/<slug>.md file.
type helpPage struct {
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Summary      string   `json:"summary,omitempty"`
	Audience     string   `json:"audience,omitempty"`      // authors | api | studio — who it is written for
	Group        string   `json:"group,omitempty"`         // index heading; derived from routes when empty
	Visibility   string   `json:"visibility"`              // public | admin
	Status       string   `json:"status"`                  // live | draft
	Routes       []string `json:"routes,omitempty"`        // the pages it explains (site_pages routes)
	Covers       []string `json:"covers,omitempty"`        // source globs it describes (self-update, 8.9)
	LastVerified string   `json:"last_verified,omitempty"` // commit the text was last checked against
	Owner        string   `json:"owner,omitempty"`
	Order        int      `json:"order"`

	Body    string    `json:"-"` // markdown without frontmatter
	HTML    string    `json:"-"`
	Text    string    `json:"-"` // plain text, for search
	ModTime time.Time `json:"-"`
}

// adminOnly reports whether only the admin may see the page.
func (p *helpPage) adminOnly() bool {
	return p.Visibility == "admin" || p.Status != "live"
}

// helpStore caches the parsed pages and a full-text index. It reloads when
// the directory's signature (names, sizes, mtimes) changes, checked at most
// every helpRecheck.
type helpStore struct {
	dir string

	mu      sync.Mutex
	sig     string
	checked time.Time
	pages   map[string]*helpPage
	order   []*helpPage
	fts     *sql.DB
	loadErr error
}

const helpRecheck = 2 * time.Second

func (s *Server) helpStore() *helpStore {
	s.helpOnce.Do(func() {
		if s.help == nil {
			s.help = newHelpStore(helpDir())
		}
	})
	return s.help
}

func newHelpStore(dir string) *helpStore { return &helpStore{dir: dir} }

var helpSlugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// snapshot returns the current pages, reloading from disk if they changed.
func (h *helpStore) snapshot() ([]*helpPage, map[string]*helpPage, *sql.DB) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if time.Since(h.checked) < helpRecheck && h.pages != nil {
		return h.order, h.pages, h.fts
	}
	h.checked = time.Now()
	sig := dirSignature(h.dir)
	if sig == h.sig && h.pages != nil {
		return h.order, h.pages, h.fts
	}
	order, pages, err := loadHelpPages(h.dir)
	if err != nil {
		h.loadErr = err
		slog.Warn("help: load", "dir", h.dir, "err", err)
	}
	fts, ferr := buildHelpIndex(order)
	if ferr != nil {
		slog.Warn("help: index", "err", ferr)
	}
	if h.fts != nil {
		h.fts.Close()
	}
	h.sig, h.order, h.pages, h.fts = sig, order, pages, fts
	return h.order, h.pages, h.fts
}

func dirSignature(dir string) string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "err:" + err.Error()
	}
	var b strings.Builder
	for _, e := range ents {
		if info, err := e.Info(); err == nil {
			fmt.Fprintf(&b, "%s/%d/%d;", e.Name(), info.Size(), info.ModTime().UnixNano())
		}
	}
	return b.String()
}

func loadHelpPages(dir string) ([]*helpPage, map[string]*helpPage, error) {
	pages := map[string]*helpPage{}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, pages, err
	}
	var order []*helpPage
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".md") || n == "README.md" {
			continue
		}
		slug := strings.TrimSuffix(n, ".md")
		if !helpSlugRe.MatchString(slug) {
			slog.Warn("help: skipping file with a non-slug name", "file", n)
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			continue
		}
		p := parseHelpPage(slug, raw)
		if info, err := e.Info(); err == nil {
			p.ModTime = info.ModTime()
		}
		pages[slug] = p
		order = append(order, p)
	}
	sort.SliceStable(order, func(i, j int) bool {
		if order[i].Order != order[j].Order {
			return order[i].Order < order[j].Order
		}
		return order[i].Title < order[j].Title
	})
	return order, pages, nil
}

// parseHelpPage splits the frontmatter from the body. The frontmatter is a
// small YAML subset: `key: value`, `key: [a, b]`, or `key:` followed by
// `  - item` lines. Unknown keys are ignored.
func parseHelpPage(slug string, raw []byte) *helpPage {
	p := &helpPage{Slug: slug, Visibility: "public", Status: "live", Order: 100}
	body := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if strings.HasPrefix(body, "---\n") {
		if end := strings.Index(body[4:], "\n---"); end >= 0 {
			fm := body[4 : 4+end]
			rest := body[4+end+4:]
			body = strings.TrimPrefix(rest, "\n")
			parseHelpFrontmatter(p, fm)
		}
	}
	// Title falls back to the first H1, which is then dropped from the body
	// (the page shell prints the title itself).
	lines := strings.Split(body, "\n")
	for i, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		if strings.HasPrefix(ln, "# ") {
			if p.Title == "" {
				p.Title = strings.TrimSpace(ln[2:])
			}
			body = strings.Join(lines[i+1:], "\n")
		}
		break
	}
	if p.Title == "" {
		p.Title = slug
	}
	p.Body = strings.TrimLeft(body, "\n")
	p.HTML = renderHelpMarkdown(p.Body)
	p.Text = helpPlainText(p.HTML)
	return p
}

func parseHelpFrontmatter(p *helpPage, fm string) {
	var listKey string
	set := func(k string, vals []string) {
		switch k {
		case "routes":
			p.Routes = append(p.Routes, vals...)
		case "covers":
			p.Covers = append(p.Covers, vals...)
		}
	}
	for _, ln := range strings.Split(fm, "\n") {
		t := strings.TrimSpace(ln)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if strings.HasPrefix(t, "- ") && listKey != "" {
			set(listKey, []string{unquoteYAML(strings.TrimSpace(t[2:]))})
			continue
		}
		k, v, ok := strings.Cut(t, ":")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(strings.ToLower(k)), strings.TrimSpace(v)
		listKey = ""
		if v == "" {
			listKey = k
			continue
		}
		if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
			var vals []string
			for _, x := range strings.Split(v[1:len(v)-1], ",") {
				if x = unquoteYAML(strings.TrimSpace(x)); x != "" {
					vals = append(vals, x)
				}
			}
			set(k, vals)
			continue
		}
		v = unquoteYAML(v)
		switch k {
		case "title":
			p.Title = v
		case "summary":
			p.Summary = v
		case "audience":
			p.Audience = strings.ToLower(v)
		case "group":
			p.Group = v
		case "visibility":
			p.Visibility = strings.ToLower(v)
		case "status":
			p.Status = strings.ToLower(v)
		case "last_verified":
			p.LastVerified = v
		case "owner":
			p.Owner = v
		case "order":
			if n, err := strconv.Atoi(v); err == nil {
				p.Order = n
			}
		case "routes", "covers":
			set(k, []string{v})
		}
	}
	// Anything other than the two known values fails closed.
	if p.Visibility != "public" {
		p.Visibility = "admin"
	}
	if p.Status != "live" {
		p.Status = "draft"
	}
}

func unquoteYAML(s string) string {
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}

var helpMD = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.Footnote),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

// Links between help pages are written as plain file links (`inspect.md`,
// `inspect.md#fixes`) so they also work on GitHub; the renderer points them at
// /help/.
var helpMDLinkRe = regexp.MustCompile(`href="(?:\./)?([a-z0-9][a-z0-9-]*)\.md(#[^"]*)?"`)

// helpInternalLinkRe finds rendered links to other help pages.
var helpInternalLinkRe = regexp.MustCompile(`href="/help/([a-z0-9][a-z0-9-]*)(?:#[^"]*)?"`)

func renderHelpMarkdown(md string) string {
	var buf bytes.Buffer
	if err := helpMD.Convert([]byte(md), &buf); err != nil {
		return "<pre>" + html.EscapeString(md) + "</pre>"
	}
	return helpMDLinkRe.ReplaceAllString(buf.String(), `href="/help/$1$2"`)
}

var (
	helpTagRe   = regexp.MustCompile(`<[^>]*>`)
	helpSpaceRe = regexp.MustCompile(`\s+`)
)

func helpPlainText(h string) string {
	t := helpTagRe.ReplaceAllString(h, " ")
	return strings.TrimSpace(helpSpaceRe.ReplaceAllString(html.UnescapeString(t), " "))
}

// buildHelpIndex makes an in-memory FTS5 index over every page (admin-only
// pages included; queries filter by tier). It is rebuilt whenever the pages
// reload, so it never needs a migration.
func buildHelpIndex(pages []*helpPage) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // one connection = one in-memory database
	if _, err := db.Exec(`CREATE VIRTUAL TABLE help_fts USING fts5(slug UNINDEXED, admin_only UNINDEXED, title, summary, body, tokenize='porter unicode61')`); err != nil {
		db.Close()
		return nil, err
	}
	for _, p := range pages {
		ao := 0
		if p.adminOnly() {
			ao = 1
		}
		if _, err := db.Exec(`INSERT INTO help_fts (slug, admin_only, title, summary, body) VALUES (?,?,?,?,?)`, p.Slug, ao, p.Title, p.Summary, p.Text); err != nil {
			db.Close()
			return nil, err
		}
	}
	return db, nil
}

type helpHit struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	Snippet string `json:"snippet"` // HTML: escaped text with <mark> around matches
	URL     string `json:"url"`
}

var helpWordRe = regexp.MustCompile(`[\p{L}\p{N}]+`)

// helpFTSQuery turns free text into a safe FTS5 query: each word becomes a
// quoted prefix term, all ANDed. No user syntax reaches FTS5.
func helpFTSQuery(q string) string {
	words := helpWordRe.FindAllString(strings.ToLower(q), 12)
	var terms []string
	for _, w := range words {
		terms = append(terms, `"`+w+`"*`)
	}
	return strings.Join(terms, " ")
}

func (h *helpStore) search(q string, admin bool, limit int) []helpHit {
	_, pages, db := h.snapshot()
	fq := helpFTSQuery(q)
	if db == nil || fq == "" {
		return nil
	}
	// Snippet markers are control characters so the text can be escaped
	// first and the markers swapped for <mark> after.
	rows, err := db.Query(`SELECT slug, admin_only, snippet(help_fts, 4, char(1), char(2), '…', 14)
		FROM help_fts WHERE help_fts MATCH ? ORDER BY bm25(help_fts, 10.0, 4.0, 1.0) LIMIT ?`, fq, limit*3)
	if err != nil {
		slog.Warn("help: search", "q", q, "err", err)
		return nil
	}
	defer rows.Close()
	var hits []helpHit
	for rows.Next() && len(hits) < limit {
		var slug, snip string
		var ao int
		if rows.Scan(&slug, &ao, &snip) != nil || (ao == 1 && !admin) {
			continue
		}
		p := pages[slug]
		if p == nil {
			continue
		}
		snip = html.EscapeString(snip)
		snip = strings.NewReplacer("\x01", "<mark>", "\x02", "</mark>").Replace(snip)
		hits = append(hits, helpHit{Slug: slug, Title: p.Title, Summary: p.Summary, Snippet: snip, URL: "/help/" + slug})
	}
	return hits
}

// ---- grouping: derived from the site_pages registry ----

// helpGroups orders the index headings. A page's group is its frontmatter
// `group`, else it comes from the tier of the first route it explains, as
// recorded in site_pages — so help and the route registry can't drift.
var helpGroupOrder = []string{"Start here", "Your books", "The studio", "For developers", "Studio admin", "More"}

type sitePageInfo struct {
	Route, Title, Visibility, Listed, Status, PageType string
}

func (s *Server) loadSitePages() map[string]sitePageInfo {
	out := map[string]sitePageInfo{}
	if s.DB == nil {
		return out
	}
	rows, err := s.DB.Query(`SELECT route, title, visibility, listed, status, page_type FROM site_pages`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var p sitePageInfo
		if rows.Scan(&p.Route, &p.Title, &p.Visibility, &p.Listed, &p.Status, &p.PageType) == nil {
			out[p.Route] = p
		}
	}
	return out
}

func helpGroupFor(p *helpPage, sp map[string]sitePageInfo) string {
	if p.Group != "" {
		return p.Group
	}
	if p.Audience == "api" {
		return "For developers"
	}
	if p.Visibility == "admin" || p.Audience == "studio" {
		return "Studio admin"
	}
	for _, r := range p.Routes {
		if info, ok := sp[r]; ok {
			switch info.Visibility {
			case "client":
				return "Your books"
			case "admin":
				return "Studio admin"
			default:
				return "The studio"
			}
		}
	}
	return "More"
}

// ---- handlers ----

func (s *Server) visibleHelp(r *http.Request) (admin bool, pages []*helpPage) {
	admin = s.isAdmin(r)
	order, _, _ := s.helpStore().snapshot()
	for _, p := range order {
		if admin || !p.adminOnly() {
			pages = append(pages, p)
		}
	}
	return admin, pages
}

// GET /help/ — every visible page, grouped.
func (s *Server) handleHelpIndex(w http.ResponseWriter, r *http.Request) {
	admin, pages := s.visibleHelp(r)
	sp := s.loadSitePages()
	groups := map[string][]*helpPage{}
	for _, p := range pages {
		g := helpGroupFor(p, sp)
		groups[g] = append(groups[g], p)
	}
	names := append([]string{}, helpGroupOrder...)
	for g := range groups {
		if !containsStr(names, g) {
			names = append(names, g)
		}
	}
	var b strings.Builder
	b.WriteString(helpSearchForm(""))
	if len(pages) == 0 {
		b.WriteString(`<p class="help-empty">Help is being written. In the meantime, write to the studio: <a href="mailto:j@djinna.com">j@djinna.com</a>.</p>`)
	}
	for _, g := range names {
		list := groups[g]
		if len(list) == 0 {
			continue
		}
		fmt.Fprintf(&b, `<section class="help-group"><h2>%s</h2><ul class="help-list">`, html.EscapeString(g))
		for _, p := range list {
			fmt.Fprintf(&b, `<li><a href="/help/%s">%s</a>%s`, p.Slug, html.EscapeString(p.Title), helpBadges(p, admin))
			if p.Summary != "" {
				fmt.Fprintf(&b, `<span class="d">%s</span>`, html.EscapeString(p.Summary))
			}
			b.WriteString(`</li>`)
		}
		b.WriteString(`</ul></section>`)
	}
	if admin {
		b.WriteString(s.helpCoverageHTML(sp))
	}
	b.WriteString(`<p class="help-agents">For agents and scripts: <a href="/help/llms.txt">llms.txt</a> · every page also as markdown at <span class="mono">/help/&lt;page&gt;.md</span>.</p>`)
	s.writeHelpShell(w, "Help", "jdbb studio", "Help", "How the studio works: send your manuscript, read what the factory found, and download your book.", b.String(), "")
}

// helpCoverageHTML lists client routes and public tools no help page explains yet
// (admin view only; 8.6 turns this into a test).
func (s *Server) helpCoverageHTML(sp map[string]sitePageInfo) string {
	order, _, _ := s.helpStore().snapshot()
	covered := map[string]bool{}
	for _, p := range order {
		for _, r := range p.Routes {
			covered[r] = true
		}
	}
	var missing []string
	for route, info := range sp {
		if _, opted := helpOptOut[route]; opted {
			continue
		}
		if covered[route] || info.Listed == "retired" || info.Status == "retire" || info.Visibility == "admin" || info.Visibility == "cohort" {
			continue
		}
		if info.Visibility == "client" || info.PageType == "tool" {
			missing = append(missing, route)
		}
	}
	sort.Strings(missing)
	if len(missing) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<section class="help-group help-admin"><h2>Not yet covered <span class="badge">admin</span></h2><p class="d">Client routes and public tools (from site_pages) that no help page names in <span class="mono">routes:</span>.</p><ul class="help-list">`)
	for _, r := range missing {
		fmt.Fprintf(&b, `<li><span class="mono">%s</span><span class="d">%s</span></li>`, html.EscapeString(r), html.EscapeString(sp[r].Title))
	}
	b.WriteString(`</ul></section>`)
	return b.String()
}

func helpBadges(p *helpPage, admin bool) string {
	if !admin {
		return ""
	}
	var b strings.Builder
	if p.Status != "live" {
		b.WriteString(` <span class="badge draft">draft</span>`)
	}
	if p.Visibility == "admin" {
		b.WriteString(` <span class="badge">admin</span>`)
	}
	return b.String()
}

func helpSearchForm(q string) string {
	return `<form class="help-search" action="/help/search" method="get" role="search"><label for="help-q" class="sr">Search help</label><input id="help-q" name="q" type="search" placeholder="Search help" value="` + html.EscapeString(q) + `" autocomplete="off"><button type="submit">Search</button></form>`
}

// GET /help/{slug} and /help/{slug}.md
func (s *Server) handleHelpPage(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("slug")
	if name == "llms.txt" {
		s.handleHelpLLMs(w, r)
		return
	}
	if name == "search" {
		s.handleHelpSearch(w, r)
		return
	}
	raw := strings.HasSuffix(name, ".md")
	slug := strings.TrimSuffix(name, ".md")
	_, pages, _ := s.helpStore().snapshot()
	p := pages[slug]
	if !helpSlugRe.MatchString(slug) || p == nil || (p.adminOnly() && !s.isAdmin(r)) {
		http.NotFound(w, r)
		return
	}
	if p.adminOnly() {
		w.Header().Set("X-Robots-Tag", "noindex")
	}
	if raw {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		fmt.Fprintf(w, "# %s\n\n", p.Title)
		if p.Summary != "" {
			fmt.Fprintf(w, "> %s\n\n", p.Summary)
		}
		fmt.Fprint(w, p.Body)
		if !strings.HasSuffix(p.Body, "\n") {
			fmt.Fprint(w, "\n")
		}
		fmt.Fprintf(w, "\n---\nSource: %s/help/%s\n", s.helpBase(), p.Slug)
		return
	}
	admin := s.isAdmin(r)
	var b strings.Builder
	b.WriteString(`<article class="help-doc jdbb-prose">`)
	b.WriteString(p.HTML)
	b.WriteString(`</article>`)
	b.WriteString(s.helpPageFooter(p, admin))
	sub := html.EscapeString(p.Summary) + helpBadges(p, admin)
	s.writeHelpShell(w, p.Title, "Help", p.Title, sub, b.String(), `<link rel="alternate" type="text/markdown" href="/help/`+p.Slug+`.md">`)
}

// helpPageFooter: the pages this entry explains (titles from site_pages),
// the markdown twin, and for the admin the source facts.
func (s *Server) helpPageFooter(p *helpPage, admin bool) string {
	var b strings.Builder
	b.WriteString(`<aside class="help-meta">`)
	sp := s.loadSitePages()
	var routes []string
	for _, r := range p.Routes {
		info, ok := sp[r]
		if !ok || info.Visibility == "admin" && !admin {
			continue
		}
		label := html.EscapeString(r)
		if !strings.Contains(r, "{") {
			label = `<a href="` + html.EscapeString(r) + `">` + label + `</a>`
		}
		routes = append(routes, `<li><span class="mono">`+label+`</span> <span class="d">`+html.EscapeString(info.Title)+`</span></li>`)
	}
	if len(routes) > 0 {
		b.WriteString(`<h3>This page explains</h3><ul class="help-list">` + strings.Join(routes, "") + `</ul>`)
	}
	b.WriteString(`<p class="help-links"><a href="/help/">← All help</a> · <a href="/help/` + p.Slug + `.md">View as markdown</a></p>`)
	if admin {
		fmt.Fprintf(&b, `<p class="help-src mono">docs/help/%s.md · %s · %s · last verified %s · updated %s</p>`,
			p.Slug, p.Visibility, p.Status, html.EscapeString(orDash(p.LastVerified)), p.ModTime.Format("2 Jan 2006 15:04"))
	}
	b.WriteString(`</aside>`)
	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// GET /help/search?q=
func (s *Server) handleHelpSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) > 200 {
		q = q[:200]
	}
	admin := s.isAdmin(r)
	var b strings.Builder
	b.WriteString(helpSearchForm(q))
	if q != "" {
		hits := s.helpStore().search(q, admin, 20)
		if len(hits) == 0 {
			fmt.Fprintf(&b, `<p class="help-empty">Nothing in help matches <strong>%s</strong>. Try a shorter word, or <a href="/help/">browse all help</a>.</p>`, html.EscapeString(q))
		} else {
			b.WriteString(`<ul class="help-list help-hits">`)
			for _, h := range hits {
				fmt.Fprintf(&b, `<li><a href="%s">%s</a><span class="d">%s</span></li>`, h.URL, html.EscapeString(h.Title), h.Snippet)
			}
			b.WriteString(`</ul>`)
		}
	}
	b.WriteString(`<p class="help-links"><a href="/help/">← All help</a></p>`)
	s.writeHelpShell(w, "Search help", "Help", "Search help", "", b.String(), `<meta name="robots" content="noindex">`)
}

// GET /api/help/search?q= — JSON, for the "?" panel and for agents.
func (s *Server) handleHelpSearchAPI(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) > 200 {
		q = q[:200]
	}
	hits := s.helpStore().search(q, s.isAdmin(r), 10)
	if hits == nil {
		hits = []helpHit{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"query": q, "results": hits})
}

// GET /help/llms.txt — the llms.txt convention: a markdown index of the help,
// each entry linking its .md twin.
func (s *Server) handleHelpLLMs(w http.ResponseWriter, r *http.Request) {
	admin, pages := s.visibleHelp(r)
	sp := s.loadSitePages()
	base := s.helpBase()
	groups := map[string][]*helpPage{}
	for _, p := range pages {
		g := helpGroupFor(p, sp)
		groups[g] = append(groups[g], p)
	}
	var b strings.Builder
	b.WriteString("# jdbb studio — help\n\n")
	b.WriteString("> The studio's book factory turns a manuscript (.docx from your editor) into an EPUB and a print-ready PDF. These pages explain the client side: sending a manuscript, reading what Inspect found, building and downloading, and the style sheet.\n\n")
	fmt.Fprintf(&b, "Every page is also available as markdown by adding `.md` to its URL. Factory API reference: %s/factory/api\n", base)
	names := append([]string{}, helpGroupOrder...)
	for g := range groups {
		if !containsStr(names, g) {
			names = append(names, g)
		}
	}
	for _, g := range names {
		if len(groups[g]) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n## %s\n\n", g)
		for _, p := range groups[g] {
			line := fmt.Sprintf("- [%s](%s/help/%s.md)", p.Title, base, p.Slug)
			if p.Summary != "" {
				line += ": " + p.Summary
			}
			if admin && p.adminOnly() {
				line += " (admin only)"
			}
			b.WriteString(line + "\n")
		}
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, b.String())
}

func (s *Server) helpBase() string {
	if s.BaseURL != "" {
		return strings.TrimRight(s.BaseURL, "/")
	}
	return "https://studio.jdbb.net"
}

func containsStr(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

var helpPageCSS = `body{margin:0;background:var(--bg);color:var(--text);font:15px/1.6 var(--body)}
.hero{padding:34px 0 18px;border-bottom:1px solid var(--border-strong)}.kicker{font:700 11px/1.2 var(--mono);letter-spacing:.12em;text-transform:uppercase;color:var(--accent)}.kicker a{color:inherit;text-decoration:none}
.hero h1{font:600 clamp(26px,4.5vw,38px)/1.2 var(--mono);letter-spacing:-.015em;margin:6px 0 10px}.sub{color:var(--text-secondary);max-width:70ch}
.mono{font-family:var(--mono)}.sr{position:absolute;left:-9999px}
.badge{font:700 10px/1 var(--mono);letter-spacing:.08em;text-transform:uppercase;border:1px solid var(--border-strong);padding:2px 5px;color:var(--text-secondary);vertical-align:2px}.badge.draft{color:var(--accent);border-color:var(--accent)}
.help-search{display:flex;gap:8px;margin:22px 0 8px;max-width:560px}.help-search input{flex:1;font:14px var(--body);padding:8px 10px;border:1px solid var(--border-strong);background:var(--bg);color:var(--text)}.help-search button{font:600 12px var(--mono);letter-spacing:.06em;text-transform:uppercase;padding:8px 14px;border:1px solid var(--text);background:var(--text);color:var(--bg);cursor:pointer}
.help-group h2{font:600 13px/1.2 var(--mono);letter-spacing:.1em;text-transform:uppercase;color:var(--text-secondary);margin:30px 0 6px;padding-top:12px;border-top:1px solid var(--border-strong)}
.help-list{list-style:none;padding:0;margin:0;max-width:78ch}.help-list li{padding:9px 0;border-bottom:1px solid var(--border)}.help-list a{color:var(--text);font-weight:600;text-decoration:underline;text-underline-offset:3px}.help-list .d{display:block;color:var(--text-secondary);font-size:13.5px;margin-top:2px}
.help-hits mark{background:color-mix(in srgb,var(--accent) 22%,transparent);color:inherit;padding:0 1px}
.help-empty{color:var(--text-secondary);margin:18px 0}
.help-doc{margin:26px 0}.help-doc h2{font:600 20px/1.25 var(--mono);margin:32px 0 8px}.help-doc h3{font:600 16px/1.3 var(--mono);margin:22px 0 6px}.help-doc a{color:var(--text);text-decoration:underline;text-underline-offset:3px;text-decoration-color:var(--accent)}
.help-doc code{font:13px var(--mono);background:color-mix(in srgb,var(--text) 6%,transparent);padding:1px 4px}.help-doc pre{font:13px/1.5 var(--mono);background:color-mix(in srgb,var(--text) 5%,transparent);padding:10px 12px;overflow-x:auto}.help-doc pre code{background:none;padding:0}
.help-doc blockquote{margin:14px 0;padding:8px 14px;border-left:3px solid var(--accent);background:color-mix(in srgb,var(--accent) 6%,transparent)}.help-doc blockquote p{margin:4px 0}
.help-doc table{border-collapse:collapse;font-size:14px;margin:12px 0}.help-doc td,.help-doc th{border-bottom:1px solid var(--border);padding:5px 10px 5px 0;text-align:left;vertical-align:top}.help-doc th{font:600 12px var(--mono);letter-spacing:.05em;text-transform:uppercase;color:var(--text-secondary)}
.help-doc li{margin:4px 0}.help-doc img{max-width:100%;border:1px solid var(--border)}
.help-meta{margin:34px 0 10px;padding-top:12px;border-top:1px solid var(--border-strong);max-width:78ch}.help-meta h3{font:600 12px/1.2 var(--mono);letter-spacing:.1em;text-transform:uppercase;color:var(--text-secondary);margin:0 0 4px}.help-meta .help-list li{padding:6px 0}.help-meta .help-list .d{display:inline;margin-left:6px}
.help-links,.help-agents,.help-src{font:12px var(--mono);color:var(--text-secondary);margin:14px 0}.help-links a,.help-agents a{color:var(--text-secondary)}
.jdbb-shell{padding-bottom:48px}`

// writeHelpShell prints a help page in the standard 1240 shell with the
// public nav (help is public-tier; admin-only pages use the same chrome).
func (s *Server) writeHelpShell(w http.ResponseWriter, title, kicker, h1, subHTML, body, head string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, `<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>`+html.EscapeString(title)+` · jdbb studio</title><link rel="icon" href="/static/favicon.svg?v=2" type="image/svg+xml"><link rel="stylesheet" href="/static/theme.css">`+head+`<style>`+helpPageCSS+`</style></head><body><div class="jdbb-shell">
<header class="jdbb-masthead"><a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span><span class="studio">studio</span></a><nav data-public-nav><div id="theme-bar"></div></nav></header>
<section class="hero"><div class="kicker"><a href="/help/">`+html.EscapeString(kicker)+`</a></div><h1>`+html.EscapeString(h1)+`</h1>`)
	if subHTML != "" {
		fmt.Fprint(w, `<div class="sub">`+subHTML+`</div>`)
	}
	fmt.Fprint(w, `</section>
<main>`+body+`</main>
<footer class="jdbb-footer"><a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span></a><nav aria-label="Footer"><a href="/">Home</a><a href="/help/">Help</a><a href="/factory">Factory</a><a href="/portal">Client portal</a></nav><span class="copy">&copy; 2026 Jenna Dixon</span></footer>
</div><script src="/static/theme.js"></script></body></html>`)
}

// ---- 8.6: "?" per page, "learn more" links, route coverage ----

// helpOptOut lists site_pages routes that deliberately have no help page,
// with the reason. TestHelpRouteCoverage fails when a client or public route
// is neither explained by a help page (frontmatter routes:) nor listed here —
// the same contract as nav_convergence_test.go for the navs.
var helpOptOut = map[string]string{
	"/":                               "homepage (marketing)",
	"/jdbb-net":                       "jdbb.net landing page (marketing)",
	"/press":                          "about the press (marketing)",
	"/workshop":                       "workshop registration page; the page is its own explanation",
	"/field-notes":                    "essay",
	"/word-free":                      "essay",
	"/litmags":                        "reference ledger, not a tool",
	"/field-guide":                    "artifact for collaborators, not clients",
	"/work-notes-standard":            "artifact",
	"/architecture-plan":              "artifact",
	"/exedeck":                        "talk deck",
	"/2026-pi-symposium/map":          "workshop material",
	"/2026-pi-symposium/og-protocols": "talk deck",
	"/2026-pi-symposium/talk":         "talk deck",
	"/2026-pi-symposium/why-book":     "workshop material",
	"/2026-pi-symposium/workshop":     "workshop handout",
	"/stylesheet-pi/":                 "symposium editorial review tool (cohort only in practice)",
	"/stylesheet-pi/authors":          "symposium editorial review tool",
	"/{client}/{project}/":            "project calendar: hidden from DIY clients, no help pages in v1 (open question 8.15)",
	"/help/":                          "help itself",
	"/help/llms.txt":                  "help itself",
}

// helpRouteMatch reports whether a request path matches a site_pages route
// pattern, where a {name} segment matches any one segment. Trailing slashes
// are ignored on both sides.
func helpRouteMatch(pattern, path string) bool {
	ps := strings.Split(strings.Trim(pattern, "/"), "/")
	xs := strings.Split(strings.Trim(path, "/"), "/")
	if len(ps) != len(xs) {
		return false
	}
	for i, p := range ps {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			if xs[i] == "" {
				return false
			}
			continue
		}
		if p != xs[i] {
			return false
		}
	}
	return true
}

// helpForPath picks the help page for a request path among the pages the
// viewer may see: the most specific matching route (fewest {placeholders})
// wins, then the page's order.
func helpForPath(pages []*helpPage, path string) *helpPage {
	var best *helpPage
	bestWild := 1 << 30
	for _, p := range pages {
		for _, r := range p.Routes {
			if !helpRouteMatch(r, path) {
				continue
			}
			wild := strings.Count(r, "{")
			if wild < bestWild || wild == bestWild && best != nil && p.Order < best.Order {
				best, bestWild = p, wild
			}
		}
	}
	return best
}

// GET /api/help/for?path=/sample-press/spring-novel/factory/
// → {"page": {slug,title,url} | null, "visible": [slug…]}. theme.js uses it
// to add the masthead "?" and to reveal [data-help] "learn more" links —
// only for pages this viewer can open, so drafts never surface to clients.
func (s *Server) handleHelpFor(w http.ResponseWriter, r *http.Request) {
	_, pages := s.visibleHelp(r)
	type ref struct {
		Slug  string `json:"slug"`
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	out := struct {
		Page    *ref     `json:"page"`
		Visible []string `json:"visible"`
	}{Visible: []string{}}
	if p := helpForPath(pages, r.URL.Query().Get("path")); p != nil {
		out.Page = &ref{p.Slug, p.Title, "/help/" + p.Slug}
	}
	for _, p := range pages {
		out.Visible = append(out.Visible, p.Slug)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, max-age=30")
	json.NewEncoder(w).Encode(out)
}
