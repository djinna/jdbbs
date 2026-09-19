package srv

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Runs: read-only archive of punch lists and run logs.
//
// docs/runs/*.md is written by scripts/punchlist-export.py (checklist + every
// note thread) and committed; /admin/runs/ renders those files from disk so
// Jenna can read any past run from any device without opening the repo.
// Nothing here writes. Markdown is rendered with pandoc (gfm) when present,
// otherwise shown as preformatted text so the page still works locally.

// runsDir is where the exported punch lists live. Override with PRODCAL_RUNS_DIR.
func runsDir() string {
	if d := os.Getenv("PRODCAL_RUNS_DIR"); d != "" {
		return d
	}
	return "/home/exedev/prodcal/docs/runs"
}

type runFile struct {
	Name    string // file name, e.g. PUNCHLIST-2026-09-18.md
	Title   string // first H1, else the name
	ModTime time.Time
}

func listRunFiles(dir string) ([]runFile, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []runFile
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".md") || n == "README.md" {
			continue
		}
		info, _ := e.Info()
		rf := runFile{Name: n, Title: strings.TrimSuffix(n, ".md")}
		if info != nil {
			rf.ModTime = info.ModTime()
		}
		if b, err := os.ReadFile(filepath.Join(dir, n)); err == nil {
			for _, ln := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(ln, "# ") {
					rf.Title = strings.TrimSpace(ln[2:])
					break
				}
			}
		}
		out = append(out, rf)
	}
	// newest first: names carry the date, so reverse-lexical is chronological
	sort.Slice(out, func(i, j int) bool { return out[i].Name > out[j].Name })
	return out, nil
}

// renderRunMarkdown returns HTML for a markdown file. pandoc gfm handles the
// task-list boxes, nested lists and the note blockquotes the export writes.
func renderRunMarkdown(ctx context.Context, md []byte) string {
	if p, err := exec.LookPath("pandoc"); err == nil {
		cmd := exec.CommandContext(ctx, p, "-f", "gfm", "-t", "html")
		cmd.Stdin = bytes.NewReader(md)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			return out.String()
		}
	}
	return "<pre>" + html.EscapeString(string(md)) + "</pre>"
}

var runsPageCSS = `*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--text);font:14px/1.55 var(--body)}.wrap{max-width:var(--shell-max);margin:auto;padding:0 var(--shell-gutter) 64px}.hero{padding:34px 0 20px;border-bottom:1px solid var(--border-strong)}.kicker{font:700 11px/1.2 var(--mono);letter-spacing:.12em;text-transform:uppercase;color:var(--accent)}h1{font:600 clamp(24px,4vw,34px)/1.15 var(--display);margin:6px 0 8px}.sub{color:var(--text-secondary);max-width:70ch}.mono{font-family:var(--mono)}.small{font-size:12px;color:var(--text-secondary)}
.runs{list-style:none;padding:0;margin:20px 0}.runs li{padding:10px 0;border-bottom:1px solid var(--border)}.runs a{color:var(--text);text-decoration:underline;text-underline-offset:3px}.runs .n{display:block;font:12px var(--mono);color:var(--text-secondary);margin-top:2px}
.doc{max-width:78ch;margin:24px 0}.doc h1{font-size:26px;margin-top:0}.doc h2{font:600 18px/1.25 var(--display);margin:34px 0 10px;padding-top:14px;border-top:1px solid var(--border-strong)}.doc h3{font:600 15px/1.3 var(--display);margin:22px 0 6px}.doc ul{padding-left:1.4em}.doc li{margin:6px 0}.doc li p{margin:0}.doc input[type=checkbox]{margin-right:6px;accent-color:var(--accent)}.doc blockquote{margin:8px 0 8px 0;padding:6px 12px;border-left:2px solid var(--accent);background:color-mix(in srgb,var(--accent) 6%,transparent);font-size:13px}.doc blockquote p{margin:4px 0}.doc code{font:12px var(--mono);background:color-mix(in srgb,var(--text) 6%,transparent);padding:1px 4px}.doc pre{font:12px/1.5 var(--mono);white-space:pre-wrap;overflow-wrap:anywhere}.doc table{border-collapse:collapse;font-size:13px}.doc td,.doc th{border-bottom:1px solid var(--border);padding:4px 8px;text-align:left;vertical-align:top}.doc a{color:var(--text);text-decoration:underline;text-underline-offset:3px}.doc em{color:var(--text-secondary)}
.back{font:12px var(--mono);margin-top:18px}.back a{color:var(--text-secondary)}`

func runsShell(title, kicker, h1, sub, body string) string {
	return `<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>` + html.EscapeString(title) + ` — jdbb studio</title><link rel="icon" href="/static/favicon.svg?v=2" type="image/svg+xml"><link rel="stylesheet" href="/static/theme.css"><style>` + runsPageCSS + `</style></head><body><div class="wrap">
<header class="jdbb-masthead"><a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span><span class="studio">studio</span></a><nav data-admin-nav><a href="/admin/">Admin</a><div id="theme-bar"></div></nav></header>
<section class="hero"><div class="kicker">` + html.EscapeString(kicker) + `</div><h1>` + html.EscapeString(h1) + `</h1><div class="sub">` + sub + `</div></section>
` + body + `
<footer class="jdbb-footer"><a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span></a><nav aria-label="Footer"><a href="/admin/">Admin</a><a href="/">Home</a><a href="/field-notes">Field notes</a><a href="/workshop">Workshop</a></nav><span class="copy">&copy; 2026 Jenna Dixon</span></footer>
</div><script src="/static/theme.js"></script><script>JdbbTheme.mount(document.getElementById('theme-bar'));</script></body></html>`
}

// GET /admin/runs/ — index of every exported run.
func (s *Server) handleAdminRunsIndex(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdmin(w, r) {
		return
	}
	files, err := listRunFiles(runsDir())
	var b strings.Builder
	if err != nil || len(files) == 0 {
		b.WriteString(`<p class="small">No runs exported yet. Run <code class="mono">python3 scripts/punchlist-export.py</code> and commit <code class="mono">docs/runs/</code>.</p>`)
	} else {
		b.WriteString(`<ul class="runs">`)
		for _, f := range files {
			fmt.Fprintf(&b, `<li><a href="/admin/runs/%s">%s</a><span class="n">%s · %s</span></li>`,
				html.EscapeString(strings.TrimSuffix(f.Name, ".md")), html.EscapeString(f.Title),
				html.EscapeString(f.Name), f.ModTime.Format("Mon 2 Jan 2006 15:04"))
		}
		b.WriteString(`</ul>`)
	}
	sub := `Punch lists and run logs, read-only, one file per day or event. The live list is the <a href="https://jdbbs.exe.xyz:8766/">punch-list page</a>; this is its archive (<span class="mono">docs/runs/</span>).`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	fmt.Fprint(w, runsShell("Runs", "Admin", "Runs", sub, b.String()))
}

// GET /admin/runs/{name} — one exported run, rendered.
func (s *Server) handleAdminRunFile(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdmin(w, r) {
		return
	}
	name := strings.TrimSuffix(r.PathValue("name"), ".md")
	if name == "" || name == "README" || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") || strings.HasPrefix(name, ".") {
		http.NotFound(w, r)
		return
	}
	md, err := os.ReadFile(filepath.Join(runsDir(), name+".md"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	body := `<article class="doc">` + renderRunMarkdown(r.Context(), md) + `</article><p class="back"><a href="/admin/runs/">← all runs</a></p>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	fmt.Fprint(w, runsShell(name, "Runs", name, `<span class="mono small">docs/runs/`+html.EscapeString(name)+`.md · read-only</span>`, body))
}

// handleAdminRunImage serves docs/runs/img/<name> — screenshots pasted on the
// punch list, copied into the archive by scripts/punchlist-export.py.
func (s *Server) handleAdminRunImage(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdmin(w, r) {
		return
	}
	name := r.PathValue("name")
	if r.PathValue("kind") != "img" || name == "" || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") || strings.HasPrefix(name, ".") {
		http.NotFound(w, r)
		return
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
	default:
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(runsDir(), "img", name))
}
