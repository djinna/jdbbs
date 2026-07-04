// Command prodcal-local runs ProdCal as a persistent, single-user local service
// ("model A": its own data directory, separate from any prod/review database).
//
// It is a thin CLI over internal/localrun: a loopback-only reverse proxy that
// injects the admin header in front of the real server, so the admin UI works on
// a machine with no exe.dev proxy. It refuses any non-loopback address because
// the injected header makes the proxy an unauthenticated admin surface (safe only
// on a single-user machine). Shares its core with the prodcal-app desktop shell.
package main

import (
	"context"
	"flag"
	"log"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"srv.exe.dev/internal/localrun"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("prodcal-local: ")

	dataDir := flag.String("data", localrun.DefaultDataDir(), "data directory for the local ProdCal database and secret")
	addr := flag.String("addr", localrun.DefaultAddr(), "loopback address for the local admin proxy to listen on")
	flag.Parse()

	inst, err := localrun.Start(localrun.Options{DataDir: *dataDir, Addr: *addr})
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Printf("ProdCal local is running at %s", inst.URL)
	log.Printf("data dir: %s", inst.DataDir)
	log.Printf("database: %s", inst.DBPath)
	log.Printf("loopback-only admin proxy — do NOT expose this port; email is intentionally off")

	// Best-effort: open the local UI in the default browser.
	if err := exec.Command("open", inst.URL).Start(); err != nil {
		log.Printf("could not open browser (ignored): %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Print("shutting down…")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := inst.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
