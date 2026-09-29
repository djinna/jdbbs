package srv

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// House vocabulary (Help plan decision 3, 2026-09-29): client-facing copy
// never says "Word" or "Microsoft" — it says "your editor". The file is
// still a .docx. This scans the help pages and the client-facing UI and
// message sources; code comments are skipped. A genuinely needed mention
// goes in vocabAllow with the reason.
var vocabBanned = regexp.MustCompile(`\bWord\b|\bMicrosoft\b|\bMS Word\b`)

var vocabFiles = []string{
	"static/factory.html", "static/factory.js", "static/transmittal.js",
	"static/client.html", "static/portal.html",
	"static/project-stylesheet.html", "static/project-stylesheet.js",
	"static/housestyle.html",
	"builderr.go", "docx_budget.go",
}

// vocabAllow: "file:substring" → reason.
var vocabAllow = map[string]string{}

func vocabIsComment(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasPrefix(t, "//") || strings.HasPrefix(t, "/*") || strings.HasPrefix(t, "*") || strings.HasPrefix(t, "<!--") || strings.HasPrefix(t, "#")
}

func TestVocabularyNeverWord(t *testing.T) {
	check := func(path string, skipComments bool) {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			return
		}
		inFrontmatter := false
		for i, line := range strings.Split(string(b), "\n") {
			if strings.HasSuffix(path, ".md") && strings.TrimSpace(line) == "---" && (i == 0 || inFrontmatter) {
				inFrontmatter = !inFrontmatter
				continue
			}
			if inFrontmatter || (skipComments && vocabIsComment(line)) {
				continue
			}
			// Go line comments after code.
			if skipComments && strings.HasSuffix(path, ".go") {
				if j := strings.Index(line, " // "); j >= 0 {
					line = line[:j]
				}
			}
			m := vocabBanned.FindString(line)
			if m == "" {
				continue
			}
			allowed := false
			for k := range vocabAllow {
				f, sub, _ := strings.Cut(k, ":")
				if strings.HasSuffix(path, f) && strings.Contains(line, sub) {
					allowed = true
				}
			}
			if !allowed {
				t.Errorf("%s:%d says %q — client copy says \"your editor\" (or add a vocabAllow entry with a reason): %s", path, i+1, m, strings.TrimSpace(line))
			}
		}
	}
	helps, _ := filepath.Glob(filepath.Join("..", "docs", "help", "*.md"))
	for _, p := range helps {
		if filepath.Base(p) != "README.md" {
			check(p, false)
		}
	}
	for _, f := range vocabFiles {
		check(f, true)
	}
}

// Every finding type the detector knows (SECTION_META in
// detect-edge-cases.py) has a help page docs/help/finding-<type>.md, which
// the Factory page's "learn more" links point at (factory.js renderPreflight).
func TestHelpEveryFindingHasAPage(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "typesetting", "scripts", "detect-edge-cases.py"))
	if err != nil {
		t.Skip(err)
	}
	s := string(src)
	i := strings.Index(s, "SECTION_META")
	if i < 0 {
		t.Fatal("SECTION_META not found in detect-edge-cases.py")
	}
	block := s[i:]
	if j := strings.Index(block, "\n    }"); j > 0 {
		block = block[:j]
	}
	keys := regexp.MustCompile(`(?m)^\s*'([a-z_]+)'\s*:`).FindAllStringSubmatch(block, -1)
	if len(keys) < 10 {
		t.Fatalf("parsed only %d SECTION_META keys; parser out of date?", len(keys))
	}
	for _, k := range append(keys, []string{"", "book_map"}) {
		slug := "finding-" + strings.ReplaceAll(k[1], "_", "-")
		if _, err := os.Stat(filepath.Join("..", "docs", "help", slug+".md")); err != nil {
			t.Errorf("finding type %s has no help page docs/help/%s.md", k[1], slug)
		}
	}
	// bracketed_style_name comes from srv/preflight.go, not the detector.
	if _, err := os.Stat(filepath.Join("..", "docs", "help", "finding-bracketed-style-name.md")); err != nil {
		t.Error("finding-bracketed-style-name.md missing")
	}
}
