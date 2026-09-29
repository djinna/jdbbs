package srv

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLandingHost(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "jdbb-net.html"), []byte("<h1>landing</h1>"), 0o644)
	t.Setenv("PRODCAL_PUBLIC_DOCS", dir)
	s := &Server{BaseURL: "https://studio.jdbb.net"}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(299) })
	h := s.landingHost(next)
	cases := []struct {
		host, target string
		code         int
		loc          string
	}{
		{"www.jdbb.net", "/x?y=1", 301, "https://jdbb.net/x?y=1"},
		{"jdbb.net", "/", 200, ""},
		{"jdbb.net", "/static/theme.css", 299, ""},
		{"jdbb.net", "/factory", 301, "https://studio.jdbb.net/factory"},
		{"jdbb.net", "/factory/api", 301, "https://studio.jdbb.net/factory/api"},
		{"jdbb.net", "/auth/link?t=a", 301, "https://studio.jdbb.net/auth/link?t=a"},
		{"jdbb.net", "/2024/03/some-blot-post", 302, "/"},
		{"jdbb.net", "/factoryish", 302, "/"},
		{"studio.jdbb.net", "/", 299, ""},
		{"localhost:8000", "/", 299, ""},
	}
	for _, c := range cases {
		req := httptest.NewRequest("GET", "http://"+c.host+c.target, nil)
		req.Host = c.host
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.code || rec.Header().Get("Location") != c.loc {
			t.Errorf("%s%s: got %d loc=%q", c.host, c.target, rec.Code, rec.Header().Get("Location"))
		}
	}
}
