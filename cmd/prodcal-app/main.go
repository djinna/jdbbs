//go:build darwin

// Command prodcal-app is a minimal native macOS shell around the local ProdCal
// server: it starts the loopback admin instance (internal/localrun) on an
// ephemeral port and opens a WebKit window pointed at it. Personal single-user
// use — no distribution, no code signing. This is the Vellum-style desktop
// prototype: Word doc in → EPUB + print PDF out, with the same UI as the web app.
//
// It reuses the exact launcher core as cmd/prodcal-local, so the whole existing
// server + embedded SPA render unchanged inside the window.
package main

import (
	"context"
	"log"
	"runtime"
	"time"

	webview "github.com/webview/webview_go"

	"srv.exe.dev/internal/localrun"
)

func main() {
	// WebKit's UI must run on the main OS thread.
	runtime.LockOSThread()
	log.SetPrefix("prodcal-app: ")

	inst, err := localrun.Start(localrun.Options{Addr: "127.0.0.1:0"})
	if err != nil {
		log.Fatalf("start local server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = inst.Shutdown(ctx)
	}()

	log.Printf("serving %s in a native window (data: %s)", inst.URL, inst.DataDir)

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("ProdCal")
	w.SetSize(1280, 860, webview.HintNone)
	w.Navigate(inst.URL)
	w.Run()
}
