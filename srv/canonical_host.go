package srv

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// canonicalHost moves public page traffic from the legacy exe.dev host
// (jdbbs.exe.xyz) to the app's public host (PRODCAL_BASE_URL, e.g.
// studio.jdbb.net) and tells search engines which URL is canonical.
//
//   - GET/HEAD on the legacy host → 301 to the same path + query on BaseURL,
//     so old magic links (/auth/link?t=…) still redeem, on the new host.
//   - Never redirected: /api/ (scripts, webhooks, POST bodies), /admin
//     (admin stays reachable on the exe.dev host as a fallback), /static/,
//     /healthz, and any non-GET method.
//   - Every other GET carries `Link: <BaseURL+path>; rel="canonical"`.
//
// A no-op when BaseURL is unset/unparseable or already is the legacy host
// (local runs, tests, before the switch).
func (s *Server) canonicalHost(next http.Handler) http.Handler {
	base, err := url.Parse(strings.TrimRight(s.BaseURL, "/"))
	if err != nil || base.Host == "" {
		return next
	}
	legacy := ""
	if s.Hostname != "" {
		legacy = strings.ToLower(s.Hostname) + ".exe.xyz"
	}
	if strings.EqualFold(base.Host, legacy) {
		legacy = ""
	}
	origin := base.Scheme + "://" + base.Host
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (r.Method != http.MethodGet && r.Method != http.MethodHead) || !canonicalEligible(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		if legacy != "" && strings.EqualFold(stripPort(r.Host), legacy) {
			target := origin + r.URL.EscapedPath()
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}
		w.Header().Set("Link", "<"+origin+r.URL.EscapedPath()+`>; rel="canonical"`)
		next.ServeHTTP(w, r)
	})
}

func canonicalEligible(path string) bool {
	for _, p := range []string{"/api/", "/admin", "/static/", "/healthz", "/__exe.dev/"} {
		if strings.HasPrefix(path, p) {
			return false
		}
	}
	return true
}

func stripPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}
