package srv

import (
	"bytes"
	"context"
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
}

func (s *Server) handleAdminDocsPage(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdmin(w, r) {
		return
	}
	s.serveStaticHTML(w, "static/docs-editor.html")
}

// docSourcePath validates that source is a registered jdbbs-public file and
// returns its absolute path inside publicDocsDir.
func (s *Server) docSourcePath(ctx context.Context, source string) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" || strings.HasPrefix(source, "/") || strings.Contains(source, "..") || strings.Contains(source, `\`) {
		return "", errors.New("invalid source")
	}
	var n int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM site_pages WHERE owner = 'jdbbs-public' AND source = ?`, source).Scan(&n); err != nil {
		return "", err
	}
	if n == 0 {
		return "", errors.New("source is not a registered jdbbs-public page")
	}
	root := publicDocsDir()
	full := filepath.Join(root, filepath.FromSlash(source))
	rel, err := filepath.Rel(root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", errors.New("invalid source")
	}
	return full, nil
}

// gitStatusMap runs one `git status --porcelain` in publicDocsDir and returns
// path → status code. A missing git or non-repo directory yields an empty map.
func gitStatusMap() map[string]string {
	out := map[string]string{}
	cmd := exec.Command("git", "-C", publicDocsDir(), "status", "--porcelain", "--untracked-files=all")
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
		SELECT id, route, title, source, listed, status, page_type, visibility
		FROM site_pages WHERE owner = 'jdbbs-public' AND source <> '' ORDER BY route`)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	git := gitStatusMap()
	files := []docFile{}
	for rows.Next() {
		var f docFile
		if err := rows.Scan(&f.ID, &f.Route, &f.Title, &f.Source, &f.Listed, &f.Status, &f.PageType, &f.Visibility); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		if st, err := os.Stat(filepath.Join(publicDocsDir(), filepath.FromSlash(f.Source))); err == nil && !st.IsDir() {
			f.Exists = true
			f.Size = st.Size()
			f.ModTime = st.ModTime().UTC().Format(time.RFC3339Nano)
		}
		f.GitStatus = git[f.Source]
		files = append(files, f)
	}
	settings, err := s.listSettings(r.Context())
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{"dir": publicDocsDir(), "files": files, "settings": settings})
}

func (s *Server) handleAdminDocsRead(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	source := r.URL.Query().Get("source")
	full, err := s.docSourcePath(r.Context(), source)
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
		"git_status": gitStatusMap()[source],
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
	full, err := s.docSourcePath(r.Context(), in.Source)
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
		"git_status": gitStatusMap()[in.Source],
	})
}

type docCommitInput struct {
	Source  string `json:"source"`
	Message string `json:"message"`
}

// handleAdminDocsCommit stages one registered file and commits it in
// publicDocsDir's repo. No push: that stays a deliberate shell step.
func (s *Server) handleAdminDocsCommit(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	var in docCommitInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid body", http.StatusBadRequest)
		return
	}
	if _, err := s.docSourcePath(r.Context(), in.Source); err != nil {
		jsonErr(w, err.Error(), http.StatusBadRequest)
		return
	}
	msg := clip(strings.TrimSpace(in.Message), 300)
	if msg == "" {
		msg = "Edit " + in.Source + " (admin doc editor)"
	}
	dir := publicDocsDir()
	run := func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		var out bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &out
		err := cmd.Run()
		return strings.TrimSpace(out.String()), err
	}
	if st := gitStatusMap()[in.Source]; st == "" {
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
