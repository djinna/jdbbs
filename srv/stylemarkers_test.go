package srv

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSummarizeStyleMarkers(t *testing.T) {
	report := []styleMarkerReport{
		{Marker: "computer text", Resolved: "Code Block", Via: "alias", Start: 118, End: 118, Paragraphs: 1},
		{Marker: "poem", Resolved: "Verse", Via: "alias", Start: 8, End: 10, Paragraphs: 3},
		{Marker: "caption", Start: 179, End: 179, Paragraphs: 1},
	}
	res, unres, summary := summarizeStyleMarkers(report)
	if res != 2 || unres != 1 {
		t.Fatalf("resolved=%d unresolved=%d", res, unres)
	}
	for _, want := range []string{"[[computer text]]→Code Block ¶118", "[[poem]]→Verse ¶8–10", "unresolved: [[caption]] ¶179"} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary %q lacks %q", summary, want)
		}
	}
}

// TestApplyStyleMarkersScript runs the real pre-pass on a docx built by
// python-docx: an opener on one paragraph and a closer three later restyle
// four paragraphs and strip the marker text. Skips if python-docx is absent.
func TestApplyStyleMarkersScript(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	if err := exec.Command("python3", "-c", "import docx").Run(); err != nil {
		t.Skip("python-docx not installed")
	}
	dir := t.TempDir()
	docx := filepath.Join(dir, "input.docx")
	mk := `
from docx import Document
d = Document()
for l in ["Intro.", "[[code block]]one", "two", "three", "four [[/code block]]", "[[caption]] stays", "[[break]]", "[[verse]]a line"]:
    d.add_paragraph(l)
d.save(%q)
`
	if out, err := exec.Command("python3", "-c", strings.ReplaceAll(mk, "%q", `"`+docx+`"`)).CombinedOutput(); err != nil {
		t.Fatalf("build docx: %v\n%s", err, out)
	}
	report, err := applyStyleMarkers(docx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(report) != 4 {
		t.Fatalf("want 4 markers, got %+v", report)
	}
	if r := report[2]; r.Resolved != "Section Break" || r.Start != 7 {
		t.Errorf("break marker: %+v", r)
	}
	if r := report[0]; r.Resolved != "Code Block" || r.Start != 2 || r.End != 5 || r.Paragraphs != 4 || r.Via != "factory" {
		t.Errorf("code block marker: %+v", r)
	}
	if r := report[1]; r.Resolved != "" || r.Start != 6 {
		t.Errorf("caption marker: %+v", r)
	}
	check := `
from docx import Document
d = Document(%q)
ps = d.paragraphs
assert [p.style.name for p in ps[1:5]] == ["Code Block"]*4, [p.style.name for p in ps]
assert ps[1].text == "one" and ps[4].text == "four", [p.text for p in ps]
assert ps[5].text.startswith("[[caption]]")
# a bare [[break]] must not leave an empty paragraph (pandoc would drop it)
assert ps[6].style.name == "Section Break" and ps[6].text.strip() != "", (ps[6].style.name, ps[6].text)
assert ps[7].style.name == "Verse" and ps[7].text == "a line"
`
	if out, err := exec.Command("python3", "-c", strings.ReplaceAll(check, "%q", `"`+docx+`"`)).CombinedOutput(); err != nil {
		t.Fatalf("verify docx: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "style-markers.json")); err != nil {
		t.Errorf("report file missing: %v", err)
	}
}
