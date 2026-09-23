package srv

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"srv.exe.dev/db/dbgen"
)

// tp is a test paragraph spec: style id, text, and flags.
type tp struct {
	style string // Word styleId: "Title", "Subtitle", "Heading1", "Normal" ("" = Normal)
	text  string
	brBef bool // pageBreakBefore in pPr
	brRun bool // a <w:br w:type="page"/> run before the text
	image bool // an inline drawing
}

// writeBookMapDOCX writes a minimal DOCX whose styles.xml maps Word's default
// styleIds to their names, the way Word and Google Docs both export.
func writeBookMapDOCX(t *testing.T, paras []tp) string {
	t.Helper()
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"><w:body>`)
	for _, p := range paras {
		b.WriteString("<w:p><w:pPr>")
		if p.style != "" {
			b.WriteString(`<w:pStyle w:val="` + p.style + `"/>`)
		}
		if p.brBef {
			b.WriteString("<w:pageBreakBefore/>")
		}
		b.WriteString("</w:pPr>")
		if p.brRun {
			b.WriteString(`<w:r><w:br w:type="page"/></w:r>`)
		}
		if p.image {
			b.WriteString("<w:r><w:drawing><wp:inline/></w:drawing></w:r>")
		}
		if p.text != "" {
			b.WriteString("<w:r><w:t>" + escapeXMLText(p.text) + "</w:t></w:r>")
		}
		b.WriteString("</w:p>")
	}
	b.WriteString("<w:sectPr/></w:body></w:document>")

	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Title"><w:name w:val="Title"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Subtitle"><w:name w:val="Subtitle"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="HalfTitle"><w:name w:val="Half Title"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Copyright"><w:name w:val="Copyright"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Dedication"><w:name w:val="Dedication"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Epigraph"><w:name w:val="Epigraph"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Verse"><w:name w:val="Verse"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="SectionBreak"><w:name w:val="Section Break"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Heading2"><w:name w:val="heading 2"/></w:style>` +
		`<w:style w:type="paragraph" w:styleId="Kapitel"><w:name w:val="Heading 1"/></w:style>` +
		`</w:styles>`

	path := filepath.Join(t.TempDir(), "bookmap.docx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	files := [][2]string{
		{"[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/></Types>`},
		{"_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`},
		{"word/_rels/document.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>`},
		{"word/document.xml", b.String()},
		{"word/styles.xml", styles},
	}
	for _, kv := range files {
		w, err := zw.Create(kv[0])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(kv[1])); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadDOCXParagraphs(t *testing.T) {
	path := writeBookMapDOCX(t, []tp{
		{style: "Title", text: "My Book"},
		{text: ""}, // empty: skipped
		{text: "For my mother."},
		{text: "", brRun: true}, // break-only paragraph: skipped, break carried
		{text: "Quote."},
		{style: "Kapitel", text: "Chapter 1", brBef: true}, // localized styleId, resolved by name
		{text: "img only", image: true},
		{style: "Heading2", text: "A section"},
	})
	got, err := readDOCXParagraphs(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []docxPara{
		{Style: "title", Text: "My Book"},
		{Style: "normal", Text: "For my mother."},
		{Style: "normal", Text: "Quote.", PageBreakBefore: true},
		{Style: "heading1", Text: "Chapter 1", PageBreakBefore: true},
		{Style: "normal", Text: "img only", HasImage: true},
		{Style: "heading2", Text: "A section"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("paragraphs\n got %+v\nwant %+v", got, want)
	}
}

func TestHeadingVocabulary(t *testing.T) {
	front := []string{"Foreword", "PREFACE", "A Note on the Text", "Note on the Translation", "List of Figures", "Acknowledgements", "Author’s Note", "Prologue"}
	for _, h := range front {
		if !isFrontMatterHeading(h) {
			t.Errorf("%q should be front matter", h)
		}
	}
	back := []string{"Notes", "Appendix A: Data Tables", "About the Author", "Bibliography", "Works Cited", "Colophon", "Afterword", "Index"}
	for _, h := range back {
		if !isBackMatterHeading(h) {
			t.Errorf("%q should be back matter", h)
		}
	}
	body := []string{"Introduction", "Chapter 1", "The Long Preface to a Short War", "Notes from Underground", "Epilogue", "1", "On Protocols", "Contents"}
	for _, h := range body {
		if isFrontMatterHeading(h) || isBackMatterHeading(h) {
			t.Errorf("%q should be body", h)
		}
	}
}

func TestBuildBookMap(t *testing.T) {
	p := func(style, text string) docxPara { return docxPara{Style: style, Text: text} }
	pb := func(style, text string) docxPara { return docxPara{Style: style, Text: text, PageBreakBefore: true} }

	cases := []struct {
		name     string
		paras    []docxPara
		names    []string // declared untitled pieces; nil = undeclared
		title    string   // book title from the transmittal
		author   string
		front    []string
		body     []string
		back     []string
		untitled []string
		warnHas  []string
		summary  string
	}{
		{
			name: "full book",
			paras: []docxPara{
				p("title", "My Book"), p("subtitle", "A Subtitle"),
				p("normal", "For my mother."),
				pb("normal", "“Quote.” — Someone"),
				p("heading1", "Foreword"), p("normal", "fw"),
				p("heading1", "Preface"), p("normal", "pf"),
				p("heading1", "Introduction"), p("normal", "intro"),
				p("heading1", "Chapter 1"), p("normal", "body"),
				p("heading1", "Chapter 2"), p("heading2", "Notes"), p("normal", "h2 named Notes is not an H1"),
				p("heading1", "Notes"), p("normal", "n"),
				p("heading1", "About the Author"), p("normal", "a"),
			},
			names:    []string{"dedication", "epigraph"},
			front:    []string{"Foreword", "Preface"},
			body:     []string{"Introduction", "Chapter 1", "Chapter 2"},
			back:     []string{"Notes", "About the Author"},
			untitled: []string{"dedication", "epigraph"},
			warnHas:  []string{"Title-styled paragraph “My Book” dropped", "Subtitle-styled"},
			summary:  "Front matter: Dedication, Epigraph, Foreword, Preface · Body starts at “Introduction” (3 sections) · Back matter: Notes, About the Author",
		},
		{
			name: "acknowledgments both ends; misplaced preface is body",
			paras: []docxPara{
				p("heading1", "Acknowledgments"), p("normal", "x"),
				p("heading1", "One"), p("normal", "x"),
				p("heading1", "Preface"), p("normal", "x"),
				p("heading1", "Two"), p("normal", "x"),
				p("heading1", "Acknowledgements"), p("normal", "x"),
			},
			front:    []string{"Acknowledgments"},
			body:     []string{"One", "Preface", "Two"},
			back:     []string{"Acknowledgements"},
			untitled: []string{},
			warnHas:  []string{"“Preface” looks like front matter but comes after the body starts"},
		},
		{
			name: "back-matter word mid-body stays body",
			paras: []docxPara{
				p("heading1", "Chapter 1"), p("normal", "x"),
				p("heading1", "Notes"), p("normal", "x"),
				p("heading1", "Chapter 2"), p("normal", "x"),
			},
			body:     []string{"Chapter 1", "Notes", "Chapter 2"},
			front:    []string{},
			back:     []string{},
			untitled: []string{},
			warnHas:  []string{"“Notes” looks like back matter but a chapter follows it"},
		},
		{
			name: "no page break between dedication and epigraph",
			paras: []docxPara{
				p("normal", "For my mother."), p("normal", "“Quote.”"),
				p("heading1", "Chapter 1"), p("normal", "x"),
			},
			names:    []string{"dedication", "epigraph"},
			front:    []string{},
			body:     []string{"Chapter 1"},
			back:     []string{},
			untitled: []string{"dedication"},
			warnHas:  []string{"One untitled block of 2 paragraphs", "Transmittal lists Epigraph but no untitled page was found"},
		},
		{
			name: "undeclared untitled pages get positional names",
			paras: []docxPara{
				p("normal", "a"), pb("normal", "b"), pb("normal", "c"),
				p("heading1", "Chapter 1"), p("normal", "x"),
			},
			front:    []string{},
			body:     []string{"Chapter 1"},
			back:     []string{},
			untitled: []string{"dedication", "epigraph", "untitled front-matter page 3"},
		},
		{
			name: "legacy template: title as first H1, typed Contents",
			paras: []docxPara{
				p("heading1", "The Twitter Years"), p("firstparagraph", "by Venkatesh Rao"), pb("firstparagraph", "[page iv]"),
				pb("heading1", "Contents"), p("normal", "[tk]"),
				pb("heading1", "Preface"), p("normal", "x"),
				pb("heading1", "1. Singles"), p("normal", "x"),
			},
			title:    "The Twitter Years",
			front:    []string{"Preface"},
			body:     []string{"1. Singles"},
			back:     []string{},
			untitled: []string{},
			warnHas:  []string{"Heading “The Twitter Years” matches the book title and was dropped with the 2 paragraphs under it (“by Venkatesh Rao”, “[page iv]”)", "“Contents” and the 1 paragraph under it dropped"},
			summary:  "Front matter: Preface · Body starts at “1. Singles” (1 section) · Back matter: none",
		},
		{
			// Seapunk, books 46–48: one H1 = the book title, the whole essay as
			// H2s under it. The old rule dropped the entire book.
			name: "title as the only H1 with the book under it is kept as chapter one",
			paras: []docxPara{
				p("heading1", "We Have Always Been Seapunks"),
				p("heading2", "How I got enamoured"), p("normal", "one two three four five six seven eight nine ten"),
				p("heading2", "A hole in the grey curtain"), p("normal", "one two three four five six seven eight nine ten"),
			},
			title:    "We Have Always Been Seapunks",
			front:    []string{},
			body:     []string{"We Have Always Been Seapunks"},
			back:     []string{},
			untitled: []string{},
			warnHas:  []string{"matches the book title but has 4 paragraphs of real text under it, so it is kept as chapter one"},
			summary:  "Front matter: none · Body starts at “We Have Always Been Seapunks” (1 section) · Back matter: none",
		},
		{
			name: "old template: title, subtitle, byline, Copyright style, Epigraph style, no page breaks",
			paras: []docxPara{
				p("title", "Messy Draft"), p("subtitle", "A Field Guide"),
				p("firstparagraph", "Mike Check"),
				p("copyright", "Copyright © 2026 Mike Check."),
				p("epigraph", "“Format nothing.” — an editor"),
				p("heading1", "Chapter 1"), p("normal", "x"),
			},
			title:    "Messy Draft",
			author:   "Mike Check",
			front:    []string{},
			body:     []string{"Chapter 1"},
			back:     []string{},
			untitled: []string{"epigraph"},
			warnHas:  []string{"Byline “Mike Check” dropped", "Copyright-styled 1 paragraph dropped"},
			summary:  "Front matter: Epigraph · Body starts at “Chapter 1” (1 section) · Back matter: none",
		},
		{
			name: "dedication then Epigraph style without page break splits on style",
			paras: []docxPara{
				p("normal", "For M."), p("epigraph", "“Quote.”"),
				p("heading1", "Chapter 1"), p("normal", "x"),
			},
			names:    []string{"dedication", "epigraph"},
			front:    []string{},
			body:     []string{"Chapter 1"},
			back:     []string{},
			untitled: []string{"dedication", "epigraph"},
		},
		{
			name:     "no headings at all",
			paras:    []docxPara{p("normal", "just prose"), p("normal", "more")},
			front:    []string{},
			body:     []string{},
			back:     []string{},
			untitled: []string{},
			warnHas:  []string{"No Heading 1 found"},
			summary:  "Front matter: none · Body: no section headings found · Back matter: none",
		},
		{
			name:     "all headings front-vocab collapse to body",
			paras:    []docxPara{p("heading1", "Foreword"), p("normal", "x"), p("heading1", "Preface"), p("normal", "y")},
			front:    []string{},
			body:     []string{"Foreword", "Preface"},
			back:     []string{},
			untitled: []string{},
			warnHas:  []string{"there is no body"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := buildBookMap(tc.paras, tc.names, tc.title, tc.author)
			if got := m.Front(); !reflect.DeepEqual(got, tc.front) {
				t.Errorf("front = %v, want %v", got, tc.front)
			}
			if got := m.Body(); !reflect.DeepEqual(got, tc.body) {
				t.Errorf("body = %v, want %v", got, tc.body)
			}
			if got := m.Back(); !reflect.DeepEqual(got, tc.back) {
				t.Errorf("back = %v, want %v", got, tc.back)
			}
			un := []string{}
			for _, u := range m.keptUntitled() {
				un = append(un, u.Name)
			}
			if !reflect.DeepEqual(un, tc.untitled) {
				t.Errorf("untitled = %v, want %v", un, tc.untitled)
			}
			all := strings.Join(m.Warnings, "\n") + "\n" + strings.Join(m.Notes, "\n")
			for _, w := range tc.warnHas {
				if !strings.Contains(all, w) {
					t.Errorf("warnings missing %q; got:\n%s", w, all)
				}
			}
			if tc.summary != "" && m.Summary() != tc.summary {
				t.Errorf("summary = %q, want %q", m.Summary(), tc.summary)
			}
		})
	}
}

func TestBookMapCrossCheckSpec(t *testing.T) {
	path := writeBookMapDOCX(t, []tp{
		{text: "For my mother."},
		{style: "Heading1", text: "Preface"}, {text: "x"},
		{style: "Heading1", text: "Chapter 1"}, {text: "x"},
		{style: "Heading1", text: "Works Cited"}, {text: "x"},
		{style: "Heading1", text: "Index"}, {text: "x"},
	})
	spec := map[string]any{
		"front_matter": map[string]any{"dedication": true, "epigraph": false, "foreword": true, "preface": false},
		"back_matter":  map[string]any{"bibliography": true, "index": false, "notes": true},
	}
	m, err := bookMapFromDOCX(path, spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(m.UntitledFront) != 1 || m.UntitledFront[0].Name != "dedication" {
		t.Errorf("untitled = %+v", m.UntitledFront)
	}
	all := strings.Join(m.Warnings, "\n") + "\n" + strings.Join(m.Notes, "\n")
	for _, w := range []string{
		"Transmittal lists Foreword but no “Foreword” heading was found.",
		"“Preface” found in the manuscript but not ticked",
		"Transmittal lists Notes but no “Notes” heading was found.",
		"“Index” found in the manuscript but not ticked",
	} {
		if !strings.Contains(all, w) {
			t.Errorf("missing warning %q\ngot:\n%s", w, all)
		}
	}
	if strings.Contains(all, "Bibliography but no") {
		t.Errorf("Works Cited should satisfy bibliography; got:\n%s", all)
	}
}

func TestRunManuscriptPreflightIncludesBookMap(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	s.preflightRunner = func(docxPath string, declaredStylesPath string) ([]byte, []byte, error) {
		html := []byte("<html><body><section class=\"overview\"><h2>Preflight summary</h2></section><section>rest</section></body></html>")
		report := []map[string]any{{"type": "manual_formatting", "severity": "high", "text": "bold text"}}
		jb, err := json.Marshal(report)
		return html, jb, err
	}

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Book Map", "start_date": "2026-04-12", "client_slug": "vgr", "project_slug": "book-map",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := int64(project["ID"].(float64))

	docx := writeBookMapDOCX(t, []tp{
		{style: "Title", text: "Book Map Test"},
		{text: "For M."},
		{style: "Heading1", text: "Foreword"}, {text: "x"},
		{style: "Heading1", text: "Chapter 1"}, {text: "x"},
		{style: "Heading1", text: "Notes"}, {text: "x"},
	})
	data, err := os.ReadFile(docx)
	if err != nil {
		t.Fatal(err)
	}
	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title: "Book Map Test", Author: "Tester", SourceFilename: "map.docx", SourceData: data,
		ProjectID: sql.NullInt64{Int64: pid, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+itoa(pid)+"/preflight", map[string]any{"book_id": book.ID})
	if resp.StatusCode != 200 {
		t.Fatalf("run preflight: %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	bm, _ := body["book_map"].(map[string]any)
	if bm == nil {
		t.Fatalf("no book_map in response: %v", body)
	}
	secs, _ := bm["sections"].([]any)
	if len(secs) != 3 {
		t.Fatalf("sections = %v", secs)
	}
	summary := body["summary"].(map[string]any)
	// 1 python finding + book-map notes (Title dropped, untitled page vs default
	// spec, Foreword / Notes found but not ticked); no warnings; the map itself
	// is not counted.
	if summary["total"] != float64(5) || summary["medium"] != float64(0) || summary["low"] != float64(4) {
		t.Fatalf("summary = %v", summary)
	}

	// GET returns the same map from the stored report, and the HTML has the section.
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(pid)+"/preflight?book_id="+itoa(book.ID), nil)
	decodeJSON(t, resp, &body)
	if bm, _ := body["book_map"].(map[string]any); bm == nil || bm["title"] != "Book Map Test" {
		t.Fatalf("stored book_map = %v", body["book_map"])
	}
	stored, err := q.GetLatestManuscriptPreflight(t.Context(), dbgen.GetLatestManuscriptPreflightParams{ProjectID: pid, BookID: book.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stored.ReportHtml, `id="book-map"`) || !strings.Contains(stored.ReportHtml, "Front matter: Dedication, Foreword") {
		t.Fatalf("html lacks book map section: %s", stored.ReportHtml)
	}
	if strings.Index(stored.ReportHtml, `id="book-map"`) < strings.Index(stored.ReportHtml, "Preflight summary") {
		t.Fatal("book map should follow the overview section")
	}
}
