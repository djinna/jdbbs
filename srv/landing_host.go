package srv

import (
	"net/http"
	"strings"
)

// landingHost serves the jdbb.net landing page (jdbbs-public/jdbb-net.html)
// when a request arrives for the apex or www host (punch list 8.3). It is
// inert until jdbb.net's DNS points at this VM and `domain add` is done.
//
//   - www.jdbb.net/*          → 301 https://jdbb.net/* (one canonical host)
//   - jdbb.net/               → the landing page
//   - jdbb.net/static/*, /favicon.ico, /robots.txt → served as usual
//   - jdbb.net/<studio page>  → 301 to the same path on BaseURL
//     (someone typing jdbb.net/factory lands on the factory)
//   - anything else (old Blot post URLs) → 302 to jdbb.net/
//
// On the studio host the draft is previewed at /jdbb-net.
func (s *Server) landingHost(next http.Handler) http.Handler {
	studio := strings.TrimRight(s.BaseURL, "/")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(stripPort(r.Host))
		if host == "www.jdbb.net" {
			target := "https://jdbb.net" + r.URL.EscapedPath()
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}
		if host != "jdbb.net" {
			next.ServeHTTP(w, r)
			return
		}
		p := r.URL.Path
		switch {
		case p == "/" && (r.Method == http.MethodGet || r.Method == http.MethodHead):
			w.Header().Set("Link", `<https://jdbb.net/>; rel="canonical"`)
			s.servePublicDoc(w, "jdbb-net.html")
		case strings.HasPrefix(p, "/static/") || p == "/favicon.ico" || p == "/robots.txt" || p == "/healthz":
			next.ServeHTTP(w, r)
		case studioPath(p) && studio != "":
			target := studio + r.URL.EscapedPath()
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		default:
			http.Redirect(w, r, "/", http.StatusFound)
		}
	})
}

// studioPath reports whether p is a public studio page worth forwarding from
// the landing host rather than folding into the landing page.
func studioPath(p string) bool {
	for _, pre := range []string{"/factory", "/press", "/workshop", "/field-notes", "/field-guide", "/litmags", "/word-free", "/store", "/stylesheet", "/auth/", "/help", "/2026-pi-symposium", "/exedeck", "/admin", "/api/"} {
		if p == pre || strings.HasPrefix(p, pre+"/") || (strings.HasSuffix(pre, "/") && strings.HasPrefix(p, pre)) {
			return true
		}
	}
	return false
}
