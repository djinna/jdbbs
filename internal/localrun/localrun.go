// Package localrun starts ProdCal as a loopback-only local instance: the real
// server runs on an ephemeral 127.0.0.1 port, fronted by a reverse proxy that
// injects the local-admin identity header (X-ExeDev-UserID) on every request.
//
// In production that header is supplied by the exe.dev proxy; locally there is
// none, so the admin UI would 302 to a login flow. Injecting it here makes the
// single-user machine its own trusted admin. Because that makes the proxy an
// unauthenticated admin surface, it is bound to loopback only and Start refuses
// any non-loopback address.
//
// This is the shared core behind both cmd/prodcal-local (headless service) and
// cmd/prodcal-app (desktop window).
package localrun

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"srv.exe.dev/srv"
)

// Options configure a local run.
type Options struct {
	// DataDir holds the local SQLite DB + secret. Empty → DefaultDataDir().
	DataDir string
	// Addr is the loopback front address (e.g. "127.0.0.1:8000"). Empty →
	// DefaultAddr(). Use "127.0.0.1:0" for an ephemeral port (the desktop app).
	Addr string
}

// Instance is a running local ProdCal. Call Shutdown to stop it.
type Instance struct {
	URL     string // e.g. http://127.0.0.1:8000
	DataDir string
	DBPath  string

	front    *http.Server
	internal *http.Server
}

// Start builds the server for the data dir and starts the loopback admin proxy.
// It is non-blocking: both HTTP servers run in background goroutines.
func Start(opts Options) (*Instance, error) {
	dataDir := opts.DataDir
	if dataDir == "" {
		dataDir = DefaultDataDir()
	}
	addr := opts.Addr
	if addr == "" {
		addr = DefaultAddr()
	}
	if !isLoopbackAddr(addr) {
		return nil, fmt.Errorf("addr %q is not loopback-only: this proxy injects an admin header and is "+
			"unauthenticated by design, so it must only listen on 127.0.0.1/localhost", addr)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir %q: %w", dataDir, err)
	}
	dbPath := filepath.Join(dataDir, "db.sqlite3")

	// Bind the front listener first so we know the final URL even with ":0".
	frontLn, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen (front): %w", err)
	}
	frontURL := "http://" + frontLn.Addr().String()

	// Keep generated links pointing at this local instance, not prod exe.xyz.
	if os.Getenv("PRODCAL_BASE_URL") == "" {
		_ = os.Setenv("PRODCAL_BASE_URL", frontURL)
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "localhost"
	}

	s, err := srv.New(dbPath, hostname)
	if err != nil {
		_ = frontLn.Close()
		return nil, fmt.Errorf("create server: %w", err)
	}

	// Internal server: the real ProdCal handler on an ephemeral loopback port.
	intLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		_ = frontLn.Close()
		return nil, fmt.Errorf("listen (internal): %w", err)
	}
	internal := &http.Server{Handler: s.Handler(), ReadHeaderTimeout: 30 * time.Second}
	go func() {
		if err := internal.Serve(intLn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("localrun: internal server: %v", err)
		}
	}()

	// Front proxy: inject the local-admin identity header and forward.
	target := &url.URL{Scheme: "http", Host: intLn.Addr().String()}
	proxy := httputil.NewSingleHostReverseProxy(target)
	baseDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		baseDirector(req)
		req.Header.Set("X-ExeDev-UserID", "local-admin")
		req.Header.Set("X-ExeDev-Email", "local@localhost")
	}
	front := &http.Server{Handler: proxy, ReadHeaderTimeout: 30 * time.Second}
	go func() {
		if err := front.Serve(frontLn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("localrun: front server: %v", err)
		}
	}()

	// Surface missing manuscript-pipeline tools now, at startup, instead of as
	// mysterious mid-conversion 500s. Never fatal: everything else still works.
	warnMissingPipelineDeps()

	return &Instance{URL: frontURL, DataDir: dataDir, DBPath: dbPath, front: front, internal: internal}, nil
}

// warnMissingPipelineDeps logs one clear WARNING per missing external tool the
// manuscript pipeline shells out to. Feature impact per tool:
//   - pandoc:      DOCX→EPUB and DOCX→print-PDF conversion
//   - typst:       print-PDF typesetting
//   - python-docx: preflight, corrections-apply, Word-template generation
func warnMissingPipelineDeps() {
	if _, err := exec.LookPath("pandoc"); err != nil {
		log.Printf("localrun: WARNING: pandoc not found on PATH — EPUB and print-PDF conversion will fail (brew install pandoc)")
	}
	if _, err := exec.LookPath("typst"); err != nil {
		log.Printf("localrun: WARNING: typst not found on PATH — print-PDF typesetting will fail (brew install typst)")
	}
	if _, err := exec.LookPath("python3"); err != nil {
		log.Printf("localrun: WARNING: python3 not found on PATH — preflight, corrections-apply, and Word-template generation will fail")
	} else if err := exec.Command("python3", "-c", "import docx").Run(); err != nil {
		log.Printf("localrun: WARNING: python-docx not importable by python3 — preflight, corrections-apply, and Word-template generation will fail (python3 -m pip install python-docx)")
	}
}

// Shutdown gracefully stops both servers.
func (i *Instance) Shutdown(ctx context.Context) error {
	frontErr := i.front.Shutdown(ctx)
	internalErr := i.internal.Shutdown(ctx)
	if frontErr != nil {
		return frontErr
	}
	return internalErr
}

// DefaultDataDir returns the default local data directory, honoring the
// PRODCAL_LOCAL_DATA env override.
func DefaultDataDir() string {
	if env := os.Getenv("PRODCAL_LOCAL_DATA"); env != "" {
		return env
	}
	return filepath.Join(homeDir(), "Library", "Application Support", "ProdCal")
}

// DefaultAddr returns the default loopback listen address, honoring the
// PRODCAL_LOCAL_ADDR env override.
func DefaultAddr() string {
	if env := os.Getenv("PRODCAL_LOCAL_ADDR"); env != "" {
		return env
	}
	return "127.0.0.1:8000"
}

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		return h
	}
	return os.Getenv("HOME")
}

// isLoopbackAddr reports whether addr (host:port) binds only to loopback. An
// empty host (":8000", "0.0.0.0:8000") is rejected because it would expose the
// header-injecting proxy on every interface.
func isLoopbackAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil || host == "" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, ip := range ips {
		if !ip.IsLoopback() {
			return false
		}
	}
	return true
}
