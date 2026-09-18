package srv

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// runBookMapPandoc runs the real lua filter over a synthetic DOCX with the
// book map as metadata and returns the generated typst source.
func runBookMapPandoc(t *testing.T, docx string, m *BookMap, toc bool, extraMeta map[string]any) string {
	t.Helper()
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	dir := t.TempDir()
	payload := map[string]any{"book_map": map[string]any{"toc": toc, "sections": m.Sections, "untitled_front": m.UntitledFront}}
	for k, v := range extraMeta {
		payload[k] = v
	}
	meta, _ := json.Marshal(payload)
	metaPath := filepath.Join(dir, "meta.json")
	if err := os.WriteFile(metaPath, meta, 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "book.typ")
	cmd := exec.Command("pandoc", "--from=docx+styles", docx, "--lua-filter="+typstFilterPath(),
		"-t", "typst+smart", "--metadata-file="+metaPath, "-o", out)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pandoc: %v\n%s", err, b)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// hookLines extracts the P4 template calls and headings, in order.
func hookLines(typ string) []string {
	re := regexp.MustCompile(`(?m)^(#front-piece\([^)]*\)\[|\]|#contents-page\(\)|#front-section\(\)|#start-body\(\)|#start-back\(\)|#set-story-info\(.*|= .*)$`)
	var out []string
	for _, l := range strings.Split(typ, "\n") {
		if re.MatchString(l) {
			out = append(out, strings.TrimSpace(l))
		}
	}
	return out
}

func TestLuaFilterAppliesBookMap(t *testing.T) {
	docx := writeBookMapDOCX(t, []tp{
		{style: "Title", text: "My Book"},
		{text: "by Jane Author"},
		{text: "For my mother."},
		{text: "“Quote.” — Someone", brBef: true},
		{style: "Heading1", text: "Contents"}, {text: "[tk]"},
		{style: "Heading1", text: "Foreword"}, {text: "fw"},
		{style: "Heading1", text: "Chapter 1"}, {text: "body"},
		{style: "Heading1", text: "Notes"}, {text: "n"},
	})
	m, err := bookMapFromDOCX(docx, nil, "My Book", "Jane Author")
	if err != nil {
		t.Fatal(err)
	}
	typ := runBookMapPandoc(t, docx, m, true, nil)
	got := strings.Join(hookLines(typ), "\n")
	want := strings.Join([]string{
		`#front-piece(kind: "dedication")[`, `]`,
		`#front-piece(kind: "epigraph")[`, `]`,
		`#contents-page()`,
		`#front-section()`, `= Foreword`,
		`#start-body()`, `= Chapter 1`,
		`#start-back()`, `= Notes`,
	}, "\n")
	if got != want {
		t.Errorf("hooks\n got:\n%s\nwant:\n%s\n--- typst ---\n%s", got, want, typ)
	}
	if strings.Contains(typ, "[tk]") || strings.Contains(typ, "by Jane Author") {
		t.Errorf("typed Contents section / byline should be dropped:\n%s", typ)
	}
	if !strings.Contains(typ, "For my mother.") || !strings.Contains(typ, "Quote.") {
		t.Errorf("dedication/epigraph text missing:\n%s", typ)
	}
}

func TestLuaFilterAnthologyStoryInfoStaysWithHeading(t *testing.T) {
	docx := writeBookMapDOCX(t, []tp{
		{style: "Heading1", text: "Story One"}, {text: "a"},
		{style: "Heading1", text: "Story Two"}, {text: "b"},
	})
	m, err := bookMapFromDOCX(docx, nil, "Anthology", "")
	if err != nil {
		t.Fatal(err)
	}
	chapters := []map[string]string{{"title": "Story One", "author": "A"}, {"title": "Story Two", "author": "B"}}
	typ := runBookMapPandoc(t, docx, m, true, map[string]any{"chapters": chapters})
	got := strings.Join(hookLines(typ), "\n")
	want := strings.Join([]string{
		`#contents-page()`,
		`#start-body()`,
		`#set-story-info(title: "Story One", author: "A")`, `= Story One`,
		`#set-story-info(title: "Story Two", author: "B")`, `= Story Two`,
	}, "\n")
	if got != want {
		t.Errorf("hooks\n got:\n%s\nwant:\n%s", got, want)
	}
	if strings.Contains(typ, "front-piece") {
		t.Errorf("no untitled front matter expected:\n%s", typ)
	}
}

func TestLuaFilterNoHeadingsIsAllBody(t *testing.T) {
	docx := writeBookMapDOCX(t, []tp{{text: "just prose"}, {text: "more prose"}})
	m, err := bookMapFromDOCX(docx, nil, "", "")
	if err != nil {
		t.Fatal(err)
	}
	typ := runBookMapPandoc(t, docx, m, true, nil)
	got := strings.Join(hookLines(typ), "\n")
	if got != "#start-body()" {
		t.Errorf("hooks = %q\n%s", got, typ)
	}
}

// TestFrontMatterTypstCompiles renders the P4 hooks through the real template
// and checks the page structure: i–iv generated, roman folios, arabic 1 on a recto.
func TestFrontMatterTypstCompiles(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst not installed")
	}
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	dir := t.TempDir()
	src := `#import "` + seriesTemplatePath() + `": *
#let config = merge-config((body-font: "Libertinus Serif", heading-font: "Source Sans 3"))
#show: book.with(config: config, title: "My Book", author: "Jane Author",
  front-matter: (half-title: true, title-page: true, copyright-page: true, publisher: "Test Press", copyright-year: "2026"))
#front-piece(kind: "dedication")[For M.]
#contents-page()
#front-section()
= Foreword
#lorem(50)
#start-body()
= Chapter 1
#lorem(50)
= Chapter 2
#lorem(50)
`
	typPath := filepath.Join(dir, "t.typ")
	if err := os.WriteFile(typPath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	pdf := filepath.Join(dir, "t.pdf")
	cmd := exec.Command("typst", "compile", "--root", "/", "--font-path", fontsDirPath(), typPath, pdf)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("typst: %v\n%s", err, b)
	}
	page := func(n int) string {
		b, err := exec.Command("pdftotext", "-f", itoa(int64(n)), "-l", itoa(int64(n)), "-layout", pdf, "-").Output()
		if err != nil {
			t.Fatal(err)
		}
		return strings.Join(strings.Fields(string(b)), " ")
	}
	// i half-title, ii blank, iii title, iv copyright, v dedication, vi blank,
	// vii contents, viii blank, ix foreword, x blank, 1 chapter 1 (recto), 2 chapter 2.
	checks := []struct {
		n    int
		want []string
		not  []string
	}{
		{1, []string{"My Book"}, []string{"Jane", "i"}},
		{2, nil, []string{"My Book", "ii"}},
		{3, []string{"My Book", "Jane Author", "TEST PRESS"}, nil},
		{4, []string{"Copyright © 2026 Jane Author", "Published by Test Press"}, nil},
		{5, []string{"For M.", "v"}, nil},
		{6, nil, []string{"vi"}},
		{7, []string{"Contents", "Foreword", "ix", "Chapter 1", "vii"}, nil},
		{9, []string{"Foreword", "ix"}, []string{"JANE AUTHOR"}},
		{11, []string{"Chapter 1", "1"}, []string{"JANE AUTHOR", "xi"}},
		{12, []string{"Chapter 2", "2"}, nil},
	}
	for _, c := range checks {
		txt := page(c.n)
		for _, w := range c.want {
			if !strings.Contains(txt, w) {
				t.Errorf("page %d: want %q in %q", c.n, w, txt)
			}
		}
		for _, w := range c.not {
			if strings.Contains(" "+txt+" ", " "+w+" ") {
				t.Errorf("page %d: did not want %q in %q", c.n, w, txt)
			}
		}
	}
	if txt := page(6); strings.TrimSpace(txt) != "" {
		t.Errorf("page 6 should be blank, got %q", txt)
	}
}

// TestEPUBFilterAppliesBookMap runs the EPUB lua filter over a synthetic DOCX
// and checks sections, epub:types, drops, and epubcheck cleanliness.
func TestEPUBFilterAppliesBookMap(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	docx := writeBookMapDOCX(t, []tp{
		{style: "Title", text: "My Book"},
		{text: "by Jane Author"},
		{text: "For my mother."},
		{text: "“Quote.” — Someone", brBef: true},
		{style: "Heading1", text: "Contents"}, {text: "[tk]"},
		{style: "Heading1", text: "Foreword"}, {text: "fw"},
		{style: "Heading1", text: "Chapter 1"}, {text: "body"},
		{style: "Heading1", text: "Notes"}, {text: "n"},
	})
	m, err := bookMapFromDOCX(docx, nil, "My Book", "Jane Author")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	meta, _ := json.Marshal(map[string]any{"book_map": map[string]any{"toc": true, "sections": m.Sections, "untitled_front": m.UntitledFront}})
	metaPath := filepath.Join(dir, "meta.json")
	if err := os.WriteFile(metaPath, meta, 0644); err != nil {
		t.Fatal(err)
	}
	// Native pandoc output first: one section per H1, in order.
	native := filepath.Join(dir, "book.native")
	cmd := exec.Command("pandoc", "--from=docx+styles", docx, "--lua-filter="+epubFilterPath(),
		"-t", "native", "--metadata-file="+metaPath, "-o", native)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pandoc: %v\n%s", err, b)
	}
	nb, _ := os.ReadFile(native)
	ns := regexp.MustCompile(`\s+`).ReplaceAllString(string(nb), " ")
	var types []string
	for _, mm := range regexp.MustCompile(`\( "epub:type" , "([^"]*)" \)`).FindAllStringSubmatch(ns, -1) {
		types = append(types, mm[1])
	}
	want := []string{"dedication", "epigraph", "foreword", "chapter", "endnotes"}
	if strings.Join(types, ",") != strings.Join(want, ",") {
		t.Errorf("epub:types = %v, want %v\n%s", types, want, ns)
	}
	if !strings.Contains(ns, `[ Str "For" , Space , Str "my" , Space , Str "mother." ]`) || !strings.Contains(ns, `Str "\8220Quote.\8221"`) {
		t.Errorf("dedication/epigraph text missing:\n%s", ns)
	}
	for _, bad := range []string{`Str "[tk]"`, `Str "by" , Space , Str "Jane"`} {
		if strings.Contains(ns, bad) {
			t.Errorf("%q should have been dropped:\n%s", bad, ns)
		}
	}
	if _, err := exec.LookPath("epubcheck"); err != nil {
		return
	}
	epub := filepath.Join(dir, "book.epub")
	cmd = exec.Command("pandoc", "--from=docx+styles", docx, "--lua-filter="+epubFilterPath(),
		"-t", "epub3", "--toc", "--metadata=title:My Book", "--metadata=author:Jane Author",
		"--metadata-file="+metaPath, "-o", epub)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pandoc epub: %v\n%s", err, b)
	}
	if b, err := exec.Command("epubcheck", epub).CombinedOutput(); err != nil {
		t.Errorf("epubcheck: %v\n%s", err, b)
	}
}
