package srv

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"srv.exe.dev/db/dbgen"
)

type testFile struct{ name, body string }

// buildTestEPUB synthesizes a minimal EPUB-shaped zip with a mimetype entry,
// a nav file, and the given chapter XHTML files in the order provided.
func buildTestEPUB(t *testing.T, chapters []testFile) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	w, err := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	if err != nil {
		t.Fatalf("mimetype: %v", err)
	}
	w.Write([]byte("application/epub+zip"))

	// A "nav" file that has no <h1> so it should not be matched.
	w, _ = zw.Create("EPUB/nav.xhtml")
	w.Write([]byte(`<html><body><nav><h2>Contents</h2></nav></body></html>`))

	for _, c := range chapters {
		w, err := zw.Create(c.name)
		if err != nil {
			t.Fatalf("create %s: %v", c.name, err)
		}
		w.Write([]byte(c.body))
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return buf.Bytes()
}

func readEPUBFile(t *testing.T, data []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			rc.Close()
			return string(b)
		}
	}
	t.Fatalf("file %s not found", name)
	return ""
}

func TestInjectChapterAuthors_AddsBylineAfterH1(t *testing.T) {
	epub := buildTestEPUB(t, []testFile{
		{"EPUB/ch01.xhtml", `<html><body><h1 class="ch">Soda Sweet as Blood</h1><p>...</p></body></html>`},
		{"EPUB/ch02.xhtml", `<html><body><h1>In Every Lifetime</h1><p>...</p></body></html>`},
	})

	out, err := injectChapterAuthors(epub, []epubChapter{
		{Title: "Soda Sweet as Blood", Author: "Spencer Nitkey"},
		{Title: "In Every Lifetime", Author: "Lara Dal Molin"},
	})
	if err != nil {
		t.Fatalf("inject: %v", err)
	}

	ch1 := readEPUBFile(t, out, "EPUB/ch01.xhtml")
	if !strings.Contains(ch1, `<p class="chapter-author">Spencer Nitkey</p>`) {
		t.Errorf("ch01 missing byline; got:\n%s", ch1)
	}
	// Byline must come *after* the h1.
	h1End := strings.Index(ch1, "</h1>")
	byline := strings.Index(ch1, `<p class="chapter-author">`)
	if h1End < 0 || byline < 0 || byline <= h1End {
		t.Errorf("ch01 byline not after </h1>; h1End=%d byline=%d", h1End, byline)
	}

	ch2 := readEPUBFile(t, out, "EPUB/ch02.xhtml")
	if !strings.Contains(ch2, `<p class="chapter-author">Lara Dal Molin</p>`) {
		t.Errorf("ch02 missing byline; got:\n%s", ch2)
	}
}

func TestInjectChapterAuthors_NoChaptersIsNoop(t *testing.T) {
	epub := buildTestEPUB(t, []testFile{
		{"EPUB/ch01.xhtml", `<html><body><h1>Only</h1></body></html>`},
	})
	out, err := injectChapterAuthors(epub, nil)
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	got := readEPUBFile(t, out, "EPUB/ch01.xhtml")
	if strings.Contains(got, "chapter-author") {
		t.Errorf("unexpected byline injected: %s", got)
	}
}

func TestInjectChapterAuthors_SkipsNav(t *testing.T) {
	// Nav with no <h1> should be left alone. The chapter file should still
	// receive index-0's byline (i.e. nav must not consume the slot).
	epub := buildTestEPUB(t, []testFile{
		{"EPUB/ch01.xhtml", `<html><body><h1>First</h1></body></html>`},
	})
	out, err := injectChapterAuthors(epub, []epubChapter{{Author: "A. Writer"}})
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	ch := readEPUBFile(t, out, "EPUB/ch01.xhtml")
	if !strings.Contains(ch, "A. Writer") {
		t.Errorf("byline not applied to first chapter: %s", ch)
	}
	nav := readEPUBFile(t, out, "EPUB/nav.xhtml")
	if strings.Contains(nav, "chapter-author") {
		t.Errorf("nav was modified: %s", nav)
	}
}

func TestInjectChapterAuthors_EscapesXML(t *testing.T) {
	epub := buildTestEPUB(t, []testFile{
		{"EPUB/ch01.xhtml", `<html><body><h1>X</h1></body></html>`},
	})
	out, err := injectChapterAuthors(epub, []epubChapter{
		{Author: `Smith & <Co>`},
	})
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	ch := readEPUBFile(t, out, "EPUB/ch01.xhtml")
	if !strings.Contains(ch, `Smith &amp; &lt;Co&gt;`) {
		t.Errorf("xml not escaped: %s", ch)
	}
}

func TestInjectChapterAuthors_PreservesMimetypeFirst(t *testing.T) {
	epub := buildTestEPUB(t, []testFile{
		{"EPUB/ch01.xhtml", `<html><body><h1>X</h1></body></html>`},
	})
	out, err := injectChapterAuthors(epub, []epubChapter{{Author: "A"}})
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	zr, _ := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if len(zr.File) == 0 || zr.File[0].Name != "mimetype" {
		t.Fatalf("mimetype must be first entry; got first=%q", func() string {
			if len(zr.File) == 0 {
				return ""
			}
			return zr.File[0].Name
		}())
	}
	if zr.File[0].Method != zip.Store {
		t.Errorf("mimetype must be stored (uncompressed); got method=%d", zr.File[0].Method)
	}
}

func TestParseEPUBSpec_Chapters_TopLevel(t *testing.T) {
	// TRK-DEV-012: canonical location is spec.chapters[] (top-level).
	json := `{"chapters":[
		{"title":"A","author":"Alice","file":"01.md"},
		{"title":"B","author":"Bob"},
		{"title":"","author":""}
	]}`
	spec := parseEPUBSpec(json, dbgen.Book{Title: "T", Author: "A"})
	if len(spec.Chapters) != 2 {
		t.Fatalf("expected 2 chapters (empty stub dropped); got %d: %+v", len(spec.Chapters), spec.Chapters)
	}
	if spec.Chapters[0].Author != "Alice" || spec.Chapters[0].File != "01.md" {
		t.Errorf("chapter[0]: %+v", spec.Chapters[0])
	}
	if spec.Chapters[1].Author != "Bob" {
		t.Errorf("chapter[1]: %+v", spec.Chapters[1])
	}
}

func TestParseEPUBSpec_Chapters_LegacyEpubFallback(t *testing.T) {
	// TRK-DEV-012: pre-existing DEV-009 specs stored chapters at
	// spec.epub.chapters[]. Backend reads them when spec.chapters[] is absent.
	json := `{"epub":{"chapters":[
		{"title":"Old","author":"Legacy"}
	]}}`
	spec := parseEPUBSpec(json, dbgen.Book{Title: "T", Author: "A"})
	if len(spec.Chapters) != 1 || spec.Chapters[0].Author != "Legacy" {
		t.Fatalf("expected legacy chapter to be read; got %+v", spec.Chapters)
	}
}

func TestParseEPUBSpec_Chapters_TopLevelWinsOverLegacy(t *testing.T) {
	// If both exist (mid-migration), prefer canonical location.
	json := `{
		"chapters":[{"title":"New","author":"NewAuthor"}],
		"epub":{"chapters":[{"title":"Old","author":"OldAuthor"}]}
	}`
	spec := parseEPUBSpec(json, dbgen.Book{Title: "T", Author: "A"})
	if len(spec.Chapters) != 1 || spec.Chapters[0].Author != "NewAuthor" {
		t.Fatalf("expected top-level to win; got %+v", spec.Chapters)
	}
}
