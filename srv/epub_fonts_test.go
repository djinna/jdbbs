package srv

import (
	"archive/zip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeFontTestDOCX(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "font-test.docx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create docx: %v", err)
	}
	zw := zip.NewWriter(f)
	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`,
		"_rels/.rels":         `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`,
		"word/document.xml":   `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>` + escapeXMLText(text) + `</w:t></w:r></w:p><w:sectPr/></w:body></w:document>`,
	}
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "word/document.xml"} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(files[name])); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close docx zip: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close docx: %v", err)
	}
	return path
}

func buildFontTestEPUB(t *testing.T, text string) ([]string, int64) {
	t.Helper()
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	docxPath := writeFontTestDOCX(t, text)
	fontPaths, err := epubFontPathsForDOCX(docxPath, fontsDirPath())
	if err != nil {
		t.Fatalf("select epub fonts: %v", err)
	}
	fontArgs, err := epubEmbedFontArgs(fontPaths)
	if err != nil {
		t.Fatalf("build font args: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "font-test.epub")
	args := []string{"--from=docx", "--to=epub3", "--metadata=title:Font test", "-o", outPath}
	args = append(args, fontArgs...)
	args = append(args, docxPath)
	if out, err := exec.Command("pandoc", args...).CombinedOutput(); err != nil {
		t.Fatalf("pandoc font test: %v\n%s", err, out)
	}
	zr, err := zip.OpenReader(outPath)
	if err != nil {
		t.Fatalf("open epub: %v", err)
	}
	defer zr.Close()
	var embedded []string
	for _, f := range zr.File {
		if strings.Contains(strings.ToLower(f.Name), "noto") {
			embedded = append(embedded, filepath.Base(f.Name))
		}
	}
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat epub: %v", err)
	}
	return embedded, info.Size()
}

// TestConditionalEPUBFontEmbedding proves the output-zip behavior: a Latin
// manuscript carries no Noto font files and stays small; CJK text carries the
// TC regular and bold fallbacks.
func TestConditionalEPUBFontEmbedding(t *testing.T) {
	t.Run("Latin only", func(t *testing.T) {
		embedded, size := buildFontTestEPUB(t, "A short chapter in plain Latin text.")
		if len(embedded) != 0 {
			t.Fatalf("Latin EPUB embedded fonts: %v; want none", embedded)
		}
		if size >= 500*1024 {
			t.Errorf("Latin EPUB size = %d bytes; want under 500 KB", size)
		}
	})
	t.Run("CJK", func(t *testing.T) {
		embedded, _ := buildFontTestEPUB(t, "第一章：工廠裡的一本書。")
		joined := strings.Join(embedded, " ")
		for _, want := range []string{"NotoSerifTC-Regular.otf", "NotoSerifTC-Bold.otf"} {
			if !strings.Contains(joined, want) {
				t.Errorf("CJK EPUB embedded fonts = %v; missing %s", embedded, want)
			}
		}
	})
}

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
