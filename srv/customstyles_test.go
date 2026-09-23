package srv

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestTypstStyleIdentAvoidsKeywords(t *testing.T) {
	// Fotis, book 57: a style named "break" produced `#let break(content)`.
	// cblass, book 60: "center" produced `#let center(content)`, which
	// shadowed the alignment and broke align(center) in the template.
	// Fotis's BOLD/ITAL would have shadowed the template's own bold/ital.
	for in, want := range map[string]string{
		"break": "cs-break", "Break": "cs-break", "quote": "cs-quote", "let": "cs-let",
		"none": "cs-none", "center": "cs-center", "right-justified": "right-justified",
		"BOLD": "cs-bold", "Ital": "cs-ital", "poem": "cs-poem", "Chapter": "cs-chapter",
		"verse2": "verse2", "Field Note": "field-note", "commentary": "commentary",
	} {
		if got := typstStyleIdent(in); got != want {
			t.Errorf("typstStyleIdent(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeCustomStyleParentAwareDeltas(t *testing.T) {
	row := func(m map[string]any) map[string]any {
		ns, ok := normalizeCustomStyle(m, nil)
		if !ok {
			t.Fatalf("row dropped: %v", m)
		}
		return ns
	}
	v := row(map[string]any{"name": "verse2", "based_on": "Verse", "indent": float64(1)})
	if v["based_on"] != "Verse" || v["indent"] != 1 || v["space_before"] != 1 {
		t.Fatalf("verse2: %v", v)
	}
	// indent is only for Verse / Block Quote / Code Block
	n := row(map[string]any{"name": "note", "based_on": "Normal", "indent": float64(2)})
	if n["indent"] != 0 {
		t.Fatalf("Normal must not carry an indent: %v", n)
	}
	// space_before only for Section Break, and only 3 or 6
	b := row(map[string]any{"name": "break2", "based_on": "Section Break", "space_before": "3"})
	if b["space_before"] != 3 {
		t.Fatalf("break2: %v", b)
	}
	b4 := row(map[string]any{"name": "break4", "based_on": "Section Break", "space_before": float64(4)})
	if b4["space_before"] != 1 {
		t.Fatalf("space_before 4 must fall back to 1: %v", b4)
	}
	// unknown / empty parent → Normal; aliases resolve
	if row(map[string]any{"name": "x"})["based_on"] != "Normal" {
		t.Fatal("default parent must be Normal")
	}
	if row(map[string]any{"name": "x", "based_on": "poem"})["based_on"] != "Verse" {
		t.Fatal("poem alias → Verse")
	}
	// character styles
	c := row(map[string]any{"name": "code-c", "type": "character", "based_on": "Code"})
	if c["based_on"] != "code" || c["indent"] != 0 {
		t.Fatalf("character: %v", c)
	}
	if row(map[string]any{"name": "c", "type": "character"})["based_on"] != "none" {
		t.Fatal("character default parent must be none")
	}
}

func TestCustomStyleTypstDefPrecedence(t *testing.T) {
	def := func(m map[string]any) string { return customStyleTypstDef(m) }
	// parent + deltas
	if got := def(map[string]any{"name": "verse2", "based_on": "Verse", "indent": float64(1)}); got != "#let verse2(content) = poem(indent: 1, content)" {
		t.Errorf("verse2: %q", got)
	}
	if got := def(map[string]any{"name": "verse3", "based_on": "Verse", "indent": float64(2)}); got != "#let verse3(content) = poem(indent: 2, content)" {
		t.Errorf("verse3: %q", got)
	}
	if got := def(map[string]any{"name": "break2", "based_on": "Section Break", "space_before": float64(3)}); got != "#let break2(content) = section-break-gap(gap: 3)" {
		t.Errorf("break2: %q", got)
	}
	if got := def(map[string]any{"name": "aside", "based_on": "Block Quote", "indent": float64(1)}); got != "#let aside(content) = blockquote(indent: 1, content)" {
		t.Errorf("aside: %q", got)
	}
	if got := def(map[string]any{"name": "sub", "based_on": "Heading 2"}); got != "#let cs-sub(content) = heading(level: 2, content)" {
		t.Errorf("sub: %q", got)
	}
	// Normal, no deltas: unchanged plain body paragraph
	if got := def(map[string]any{"name": "narration"}); got != "#let narration(content) = {\n  content\n}" {
		t.Errorf("narration: %q", got)
	}
	// legacy spec rows saved before based_on existed carry the generic default
	// snippet — that is not hand-written, so the parent rule applies
	legacy := map[string]any{"name": "verse2", "type": "paragraph", "preset": "generic-paragraph",
		"typst": "#let verse2(content) = {\n  content\n}", "based_on": "Verse", "indent": float64(1)}
	if got := def(legacy); !strings.Contains(got, "poem(indent: 1") {
		t.Errorf("legacy default snippet must not win over parent: %q", got)
	}
	// designed preset wins over parent
	tweet := map[string]any{"name": "tweet", "based_on": "Verse", "indent": float64(1)}
	if got := def(tweet); !strings.Contains(got, "config.tweet-font") {
		t.Errorf("tweet preset expected: %q", got)
	}
	// hand-written snippet wins over everything
	hand := map[string]any{"name": "verse2", "based_on": "Verse", "indent": float64(1), "typst": "#let verse2(c) = box(c)"}
	if got := def(hand); got != "#let verse2(c) = box(c)" {
		t.Errorf("snippet must win: %q", got)
	}
	// keyword-named style is safe
	if got := def(map[string]any{"name": "break", "based_on": "Section Break"}); !strings.HasPrefix(got, "#let cs-break(") {
		t.Errorf("break → cs-break: %q", got)
	}
	// character parents
	if got := def(map[string]any{"name": "ITAL", "type": "character", "based_on": "italic"}); got != "#let cs-ital(content) = emph(content)" {
		t.Errorf("ITAL: %q", got)
	}
}

func TestCustomStyleResolvedForm(t *testing.T) {
	cases := map[string]map[string]any{
		"verse2 → Verse, indent one level":                     {"name": "verse2", "based_on": "Verse", "indent": float64(1)},
		"verse3 → Verse, indent two levels":                    {"name": "verse3", "based_on": "Verse", "indent": float64(2)},
		"break2 → Section Break, 3× the stanza gap before":     {"name": "break2", "based_on": "Section Break", "space_before": float64(3)},
		"narration → Normal":                                   {"name": "narration"},
		"tweet → designed by the studio (preset: tweet-block)": {"name": "tweet"},
		"ITAL → italic":                                        {"name": "ITAL", "type": "character", "based_on": "italic"},
		"x → designed by the studio (custom typesetting)":      {"name": "x", "typst": "#let x(c) = c"},
	}
	for want, m := range cases {
		if got := customStyleResolvedForm(m); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestNormalizeCustomStylesCarriesSnippetForward(t *testing.T) {
	prev := []any{map[string]any{"name": "tweet", "type": "paragraph", "preset": "tweet-block", "typst": "#let tweet(c) = box(c)", "based_on": "Normal"}}
	rows := []any{map[string]any{"name": "tweet", "type": "paragraph", "description": "a tweet"}, map[string]any{"name": ""}}
	out := normalizeCustomStyles(rows, prev)
	if len(out) != 1 {
		t.Fatalf("want 1 row, got %d", len(out))
	}
	m := out[0].(map[string]any)
	if m["typst"] != "#let tweet(c) = box(c)" || m["preset"] != "tweet-block" || m["description"] != "a tweet" {
		t.Fatalf("snippet/preset not carried forward: %v", m)
	}
	if !customStyleCoalesces(map[string]any{"name": "verse2", "based_on": "Verse"}) || customStyleCoalesces(map[string]any{"name": "n"}) {
		t.Fatal("coalesce: Verse-based yes, Normal no")
	}
}

// TestTypstReservedCoversTemplateLets: every `#let name(` our templates define
// must be reserved, or a declared style with that name silently redefines it.
func TestTypstReservedCoversTemplateLets(t *testing.T) {
	files, _ := filepath.Glob("../typesetting/templates/*.typ")
	if len(files) == 0 {
		t.Skip("templates not found")
	}
	re := regexp.MustCompile(`(?m)^#let ([a-z][a-z0-9-]*)`)
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range re.FindAllStringSubmatch(string(b), -1) {
			if !typstReservedIdents[m[1]] {
				t.Errorf("%s defines #let %s but it is not in typstReservedIdents", filepath.Base(f), m[1])
			}
		}
	}
}

func TestRewriteSnippetIdent(t *testing.T) {
	// A snippet stored before the reserved list grew (cblass's `center`).
	got := customStyleTypstDef(map[string]any{"name": "center", "type": "paragraph", "preset": "generic-paragraph",
		"typst": "#let center(content) = {\n  text(style: \"italic\", content)\n}"})
	if !strings.HasPrefix(got, "#let cs-center(content)") {
		t.Errorf("stored snippet head not rewritten: %q", got)
	}
	if rewriteSnippetIdent("#let verse2(content) = poem(content)", "verse2") != "#let verse2(content) = poem(content)" {
		t.Error("matching ident must be left alone")
	}
}
