package srv

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCanonicalHost(t *testing.T) {
	s := &Server{BaseURL: "https://studio.jdbb.net", Hostname: "jdbbs"}
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	h := s.canonicalHost(ok)
	cases := []struct {
		method, host, target string
		code                 int
		loc, link            string
	}{
		{"GET", "jdbbs.exe.xyz", "/factory", 301, "https://studio.jdbb.net/factory", ""},
		{"GET", "jdbbs.exe.xyz", "/auth/link?t=abc", 301, "https://studio.jdbb.net/auth/link?t=abc", ""},
		{"HEAD", "jdbbs.exe.xyz:443", "/mike-casey/", 301, "https://studio.jdbb.net/mike-casey/", ""},
		{"GET", "jdbbs.exe.xyz", "/api/version", 200, "", ""},
		{"GET", "jdbbs.exe.xyz", "/admin/", 200, "", ""},
		{"GET", "jdbbs.exe.xyz", "/static/theme.css", 200, "", ""},
		{"POST", "jdbbs.exe.xyz", "/auth/link", 200, "", ""},
		{"GET", "studio.jdbb.net", "/factory", 200, "", `<https://studio.jdbb.net/factory>; rel="canonical"`},
		{"GET", "localhost:8000", "/press", 200, "", `<https://studio.jdbb.net/press>; rel="canonical"`},
		{"GET", "studio.jdbb.net", "/admin/", 200, "", ""},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, "http://"+c.host+c.target, nil)
		req.Host = c.host
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.code || rec.Header().Get("Location") != c.loc || rec.Header().Get("Link") != c.link {
			t.Errorf("%s %s%s: got %d loc=%q link=%q", c.method, c.host, c.target, rec.Code, rec.Header().Get("Location"), rec.Header().Get("Link"))
		}
	}
	// No-op before the switch: BaseURL is the legacy host.
	s2 := &Server{BaseURL: "https://jdbbs.exe.xyz", Hostname: "jdbbs"}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://jdbbs.exe.xyz/factory", nil)
	s2.canonicalHost(ok).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("legacy base: got %d", rec.Code)
	}
}
