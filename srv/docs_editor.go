package srv

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ─── Admin doc editor: /admin/docs/ ───
//
// The pages under publicDocsDir (talk deck, workshop guide, field notes …) are
// served straight from disk, so a saved edit is live on the next request. This
// editor lets the admin open any registry row with owner = jdbbs-public in a
// textarea, save it, and commit it to that directory's git repo. Only sources
// named in site_pages can be opened: the registry is the allow-list.
//
// Saves are guarded by the file's mtime as seen at load time (409 if the file
// moved on underneath, e.g. an edit from the shell) and written via temp file +
// rename so a half-written page is never served.

type docFile struct {
	ID         int64  `json:"id"`
	Route      string `json:"route"`
	Title      string `json:"title"`
	Source     string `json:"source"`
	Listed     string `json:"listed"`
	Status     string `json:"status"`
	Exists     bool   `json:"exists"`
	Size       int64  `json:"size"`
	ModTime    string `json:"mtime"`
	GitStatus  string `json:"git_status"` // "" clean, "M" modified, "??" untracked …
	PageType   string `json:"page_type"`
	Visibility string `json:"visibility"`
	Owner      string `json:"owner"`
}

func (s *Server) handleAdminDocsPage(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdmin(w, r) {
		return
	}
	s.serveStaticHTML(w, "static/docs-editor.html")
}

// docRoot maps a registry owner to the directory its sources are relative to.
// jdbbs-public pages live in publicDocsDir; prodcal pages are the *.html under
// srv/static/ in the source tree, served through the disk overlay (see
// static_overlay.go) so a save is live without a rebuild.
func docRoot(owner string) string {
	switch owner {
	case "jdbbs-public":
		return publicDocsDir()
	case "prodcal":
		return repoDir()
	}
	return ""
}

// docEditable reports whether a registry row may be opened in the editor.
func docEditable(owner, source string) bool {
	switch owner {
	case "jdbbs-public":
		return source != ""
	case "prodcal":
		return strings.HasPrefix(source, "srv/static/") && strings.HasSuffix(source, ".html") && repoDir() != ""
	}
	return false
}

// docSourcePath validates that source is a registered, editable file and
// returns its owner's root and its absolute path inside that root.
func (s *Server) docSourcePath(ctx context.Context, source string) (root, full string, err error) {
	source = strings.TrimSpace(source)
	if source == "" || strings.HasPrefix(source, "/") || strings.Contains(source, "..") || strings.Contains(source, `\`) {
		return "", "", errors.New("invalid source")
	}
	var owner string
	if err := s.DB.QueryRowContext(ctx, `SELECT owner FROM site_pages WHERE source = ? AND owner IN ('jdbbs-public','prodcal') LIMIT 1`, source).Scan(&owner); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", errors.New("source is not a registered page")
		}
		return "", "", err
	}
	if !docEditable(owner, source) {
		return "", "", errors.New("this page can't be edited here")
	}
	root = docRoot(owner)
	full = filepath.Join(root, filepath.FromSlash(source))
	rel, err := filepath.Rel(root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", "", errors.New("invalid source")
	}
	return root, full, nil
}

// gitStatusMap runs one `git status --porcelain` in dir and returns
// path → status code. A missing git or non-repo directory yields an empty map.
func gitStatusMap(dir string) map[string]string {
	out := map[string]string{}
	if dir == "" {
		return out
	}
	cmd := exec.Command("git", "-C", dir, "status", "--porcelain", "--untracked-files=all")
	b, err := cmd.Output()
	if err != nil {
		return out
	}
	for _, ln := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
		if len(ln) < 4 {
			continue
		}
		out[strings.TrimSpace(ln[3:])] = strings.TrimSpace(ln[:2])
	}
	return out
}

func (s *Server) handleAdminDocsList(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, route, title, source, listed, status, page_type, visibility, owner
		FROM site_pages WHERE owner IN ('jdbbs-public','prodcal') AND source <> '' ORDER BY owner DESC, route`)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	git := map[string]map[string]string{}
	files := []docFile{}
	for rows.Next() {
		var f docFile
		if err := rows.Scan(&f.ID, &f.Route, &f.Title, &f.Source, &f.Listed, &f.Status, &f.PageType, &f.Visibility, &f.Owner); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		if !docEditable(f.Owner, f.Source) {
			continue
		}
		root := docRoot(f.Owner)
		if _, ok := git[root]; !ok {
			git[root] = gitStatusMap(root)
		}
		if st, err := os.Stat(filepath.Join(root, filepath.FromSlash(f.Source))); err == nil && !st.IsDir() {
			f.Exists = true
			f.Size = st.Size()
			f.ModTime = st.ModTime().UTC().Format(time.RFC3339Nano)
		}
		f.GitStatus = git[root][f.Source]
		files = append(files, f)
	}
	settings, err := s.listSettings(r.Context())
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{"dir": publicDocsDir(), "repo_dir": repoDir(), "files": files, "settings": settings})
}

func (s *Server) handleAdminDocsRead(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	source := r.URL.Query().Get("source")
	root, full, err := s.docSourcePath(r.Context(), source)
	if err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := os.ReadFile(full)
	if err != nil {
		jsonErr(w, "file not found on disk", http.StatusNotFound)
		return
	}
	st, _ := os.Stat(full)
	jsonOK(w, map[string]any{
		"source":     source,
		"content":    string(data),
		"size":       len(data),
		"mtime":      st.ModTime().UTC().Format(time.RFC3339Nano),
		"git_status": gitStatusMap(root)[source],
	})
}

type docWriteInput struct {
	Source    string `json:"source"`
	Content   string `json:"content"`
	BaseMtime string `json:"base_mtime"` // mtime the editor loaded; "" skips the check
}

func (s *Server) handleAdminDocsWrite(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	var in docWriteInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&in); err != nil {
		jsonErr(w, "invalid body (max 4 MB)", http.StatusBadRequest)
		return
	}
	root, full, err := s.docSourcePath(r.Context(), in.Source)
	if err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(in.Content) == "" {
		jsonErr(w, "refusing to save an empty file", http.StatusBadRequest)
		return
	}
	if st, err := os.Stat(full); err == nil && in.BaseMtime != "" {
		if cur := st.ModTime().UTC().Format(time.RFC3339Nano); cur != in.BaseMtime {
			jsonErr(w, "file changed on disk since you opened it — reload before saving", http.StatusConflict)
			return
		}
	}
	// Normalise line endings; textareas send CRLF.
	content := strings.ReplaceAll(in.Content, "\r\n", "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	tmp, err := os.CreateTemp(filepath.Dir(full), "."+filepath.Base(full)+".tmp-*")
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		jsonErr(w, err.Error(), 500)
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		jsonErr(w, err.Error(), 500)
		return
	}
	os.Chmod(tmpName, 0o644)
	if err := os.Rename(tmpName, full); err != nil {
		os.Remove(tmpName)
		jsonErr(w, err.Error(), 500)
		return
	}
	st, _ := os.Stat(full)
	jsonOK(w, map[string]any{
		"ok":         true,
		"size":       len(content),
		"mtime":      st.ModTime().UTC().Format(time.RFC3339Nano),
		"git_status": gitStatusMap(root)[in.Source],
	})
}

type docCommitInput struct {
	Source  string `json:"source"`
	Message string `json:"message"`
}

// handleAdminDocsCommit stages one registered file and commits it in its
// owner's repo (jdbbs-public or the prodcal source tree). No push: that stays a deliberate shell step.
func (s *Server) handleAdminDocsCommit(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	var in docCommitInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid body", http.StatusBadRequest)
		return
	}
	dir, _, err := s.docSourcePath(r.Context(), in.Source)
	if err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	msg := clip(strings.TrimSpace(in.Message), 300)
	if msg == "" {
		msg = "Edit " + in.Source + " (admin doc editor)"
	}
	run := func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		var out bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &out
		err := cmd.Run()
		return strings.TrimSpace(out.String()), err
	}
	if st := gitStatusMap(dir)[in.Source]; st == "" {
		jsonErr(w, "nothing to commit — file matches the last commit", http.StatusBadRequest)
		return
	}
	if out, err := run("add", "--", in.Source); err != nil {
		jsonErr(w, "git add: "+out, 500)
		return
	}
	if out, err := run("commit", "-m", msg, "--", in.Source); err != nil {
		jsonErr(w, "git commit: "+out, 500)
		return
	}
	hash, _ := run("rev-parse", "--short", "HEAD")
	jsonOK(w, map[string]any{"ok": true, "commit": hash, "message": msg})
}

// handleAdminSaveSetting stores one studio setting (PUT /api/admin/settings/{key}).
func (s *Server) handleAdminSaveSetting(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	key := r.PathValue("key")
	if !isSettingKey(key) {
		jsonErr(w, "unknown setting", http.StatusNotFound)
		return
	}
	var in struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid body", http.StatusBadRequest)
		return
	}
	if strings.Contains(in.Value, "<") {
		jsonErr(w, "plain text only — no HTML tags", http.StatusBadRequest)
		return
	}
	if err := s.saveSetting(r.Context(), key, in.Value); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	list, err := s.listSettings(r.Context())
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{"ok": true, "settings": list, "preview_html": emailSignoff(), "preview_footer": emailFooterHTML()})
}
