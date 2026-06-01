package srv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEpubEmbedFontArgs_RejectsLicensed is the core TRK-DESIGN-002 guard:
// any font path under a licensed/ directory must be refused before it can be
// handed to pandoc's --epub-embed-font. Licensed fonts are print-only;
// embedding one in an EPUB zip is redistribution the desktop license doesn't
// cover.
func TestEpubEmbedFontArgs_RejectsLicensed(t *testing.T) {
	licensed := []string{
		"/home/exedev/prodcal/typesetting/fonts/licensed/plantin-mt-pro/PlantinMTPro-Regular.otf",
		"typesetting/fonts/licensed/proxima-nova/ProximaNova-Regular.otf",
		filepath.Join("a", "licensed", "b.otf"),
	}
	for _, p := range licensed {
		args, err := epubEmbedFontArgs([]string{p})
		if err == nil {
			t.Errorf("epubEmbedFontArgs(%q) = %v, nil; want error", p, args)
			continue
		}
		if !strings.Contains(err.Error(), "licensed") {
			t.Errorf("epubEmbedFontArgs(%q) error = %q; want it to mention \"licensed\"", p, err)
		}
	}
}

// TestEpubEmbedFontArgs_RejectsLicensedAmongValid ensures the guard fires even
// when a licensed path is mixed in with legitimate OFL paths — the function
// must fail closed, not embed the safe ones and skip the licensed one.
func TestEpubEmbedFontArgs_RejectsLicensedAmongValid(t *testing.T) {
	paths := []string{
		filepath.Join(t.TempDir(), "NotoSerifTC-Regular.otf"), // doesn't exist → would be skipped
		"/srv/typesetting/fonts/licensed/plantin/Plantin.otf", // must trip the guard
	}
	if _, err := epubEmbedFontArgs(paths); err == nil {
		t.Fatal("expected error when a licensed path is present, got nil")
	}
}

// TestEpubEmbedFontArgs_AllowsOFL covers the happy path: non-licensed paths
// that exist on disk become --epub-embed-font args; non-licensed paths that
// don't exist are silently skipped (a fresh checkout may lack the bundled
// fonts). No error in either case.
func TestEpubEmbedFontArgs_AllowsOFL(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "NotoSerifTC-Regular.otf")
	if err := os.WriteFile(existing, []byte("OTTO"), 0644); err != nil {
		t.Fatalf("seed font file: %v", err)
	}
	missing := filepath.Join(dir, "NotoSerifThai-Bold.ttf") // not created

	args, err := epubEmbedFontArgs([]string{existing, missing})
	if err != nil {
		t.Fatalf("epubEmbedFontArgs returned error for OFL paths: %v", err)
	}
	want := "--epub-embed-font=" + existing
	if len(args) != 1 || args[0] != want {
		t.Fatalf("args = %v; want exactly [%q] (missing path should be skipped)", args, want)
	}
}
