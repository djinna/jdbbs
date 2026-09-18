package srv

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminRunsPages(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "PUNCHLIST-2026-09-18.md"), []byte("# Day one\n\n- [x] 1.1 done\n\n  > **jenna** · ts  \n  > note\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Runs\n"), 0o644)
	t.Setenv("PRODCAL_RUNS_DIR", dir)
	s := &Server{}

	get := func(path, name string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("X-ExeDev-UserID", "admin")
		if name != "" {
			r.SetPathValue("name", name)
		}
		w := httptest.NewRecorder()
		if name == "" {
			s.handleAdminRunsIndex(w, r)
		} else {
			s.handleAdminRunFile(w, r)
		}
		return w
	}
	if w := get("/admin/runs/", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Day one") || strings.Contains(w.Body.String(), "README") {
		t.Fatalf("index: %d %q", w.Code, w.Body.String())
	}
	if w := get("/admin/runs/PUNCHLIST-2026-09-18", "PUNCHLIST-2026-09-18"); w.Code != 200 || !strings.Contains(w.Body.String(), "1.1 done") {
		t.Fatalf("file: %d", w.Code)
	}
	for _, bad := range []string{"../AGENTS", "README", ".hidden", "a/b"} {
		if w := get("/admin/runs/x", bad); w.Code != http.StatusNotFound {
			t.Errorf("%q: want 404, got %d", bad, w.Code)
		}
	}
	r := httptest.NewRequest("GET", "/admin/runs/", nil)
	w := httptest.NewRecorder()
	s.handleAdminRunsIndex(w, r)
	if w.Code != http.StatusFound {
		t.Errorf("anonymous: want redirect, got %d", w.Code)
	}
}
