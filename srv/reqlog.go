package srv

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// requestLog is the factory's flight recorder at the HTTP layer. It logs
// every request that changes something (non-GET) and every request that
// failed (status >= 400), with who asked: the exe.dev admin email if the
// proxy injected one, else the client slug(s) whose auth cookie is present.
// Successful GETs and static assets stay silent so the journal stays
// readable during a live session. See docs/reviews/SESSION-HANDOFF-2026-09-13.md
// (Addendum 2026-09-16b) for the monitoring plan this belongs to.
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") || r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		if r.Method == http.MethodGet && rw.status < 400 {
			return
		}
		attrs := []any{
			"method", r.Method, "path", r.URL.Path, "status", rw.status,
			"ms", time.Since(start).Milliseconds(),
		}
		if who := requestActor(r); who != "" {
			attrs = append(attrs, "who", who)
		}
		if rw.status >= 500 {
			slog.Error("http", attrs...)
		} else if rw.status >= 400 {
			slog.Warn("http", attrs...)
		} else {
			slog.Info("http", attrs...)
		}
	})
}

// requestActor names the caller for the log: admin email via the exe.dev
// proxy, otherwise "client:<slug>" for each client auth cookie present.
// It never reveals the cookie values themselves.
func requestActor(r *http.Request) string {
	if r.Header.Get("X-ExeDev-UserID") != "" {
		if e := r.Header.Get("X-ExeDev-Email"); e != "" {
			return "admin:" + e
		}
		return "admin"
	}
	var slugs []string
	for _, c := range r.Cookies() {
		if strings.HasPrefix(c.Name, "prodcal_client_") && c.Value != "" {
			slugs = append(slugs, strings.TrimPrefix(c.Name, "prodcal_client_"))
		}
	}
	if len(slugs) == 0 {
		return ""
	}
	return "client:" + strings.Join(slugs, ",")
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Flush keeps streaming handlers (if any) working through the wrapper.
func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
