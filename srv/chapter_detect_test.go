package srv

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"srv.exe.dev/db/dbgen"
)

// --- AST builders -----------------------------------------------------------
// These mirror the pandoc JSON AST shape (verified against pandoc 3.9), so the
// detection logic can be tested without invoking pandoc or shipping DOCX
// fixtures.

func astDoc(blocks ...any) []byte {
	d := map[string]any{
		"pandoc-api-version": []int{1, 23, 1, 1},
		"meta":               map[string]any{},
		"blocks":             blocks,
	}
	b, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}
	return b
}

// words turns a plain string into a Str/Space inline sequence the way pandoc
// would, so inlinesText's whitespace handling is exercised.
func words(s string) []any {
	var out []any
	for i, w := range strings.Fields(s) {
		if i > 0 {
			out = append(out, map[string]any{"t": "Space"})
		}
		out = append(out, map[string]any{"t": "Str", "c": w})
	}
	return out
}

func header(level int, inlines ...any) map[string]any {
	return map[string]any{"t": "Header", "c": []any{level, []any{"", []any{}, []any{}}, inlines}}
}

func para(inlines ...any) map[string]any {
	return map[string]any{"t": "Para", "c": inlines}
}

func emph(inlines ...any) map[string]any {
	return map[string]any{"t": "Emph", "c": inlines}
}

func softBreak() map[string]any   { return map[string]any{"t": "SoftBreak"} }
func space() map[string]any       { return map[string]any{"t": "Space"} }
func str(s string) map[string]any { return map[string]any{"t": "Str", "c": s} }

// customStyleDiv wraps blocks in a Div carrying a Word custom-style, as
// `--from=docx+styles` emits for a paragraph styled "Author"/"Byline"/etc.
func customStyleDiv(style string, blocks ...any) map[string]any {
	attr := []any{"", []any{}, [][]string{{"custom-style", style}}}
	return map[string]any{"t": "Div", "c": []any{attr, blocks}}
}

// --- detectChaptersFromAST --------------------------------------------------

func TestDetectChaptersFromAST(t *testing.T) {
	tests := []struct {
		name string
		ast  []byte
		want []detectedChapter
	}{
		{
			name: "by-prefix byline, By stripped",
			ast: astDoc(
				header(1, words("Soda Sweet as Blood")...),
				para(words("By Spencer Nitkey")...),
				para(words("The body was found at dawn.")...),
			),
			want: []detectedChapter{{Title: "Soda Sweet as Blood", Author: "Spencer Nitkey"}},
		},
		{
			name: "lowercase by-prefix byline",
			ast: astDoc(
				header(1, words("A Chapter")...),
				para(words("by Claire Pichelin")...),
			),
			want: []detectedChapter{{Title: "A Chapter", Author: "Claire Pichelin"}},
		},
		{
			name: "short all-italic paragraph is a byline",
			ast: astDoc(
				header(1, words("Chapter Two")...),
				para(emph(words("Zach Hyman")...)),
			),
			want: []detectedChapter{{Title: "Chapter Two", Author: "Zach Hyman"}},
		},
		{
			name: "Word Author custom-style wins (highest confidence)",
			ast: astDoc(
				header(1, words("Chapter Three")...),
				customStyleDiv("Author", para(words("Daniel Fernández")...)),
			),
			want: []detectedChapter{{Title: "Chapter Three", Author: "Daniel Fernández"}},
		},
		{
			name: "Word Byline style with a By prefix gets stripped",
			ast: astDoc(
				header(1, words("Chapter Four")...),
				customStyleDiv("Byline", para(words("By Jane Roe")...)),
			),
			want: []detectedChapter{{Title: "Chapter Four", Author: "Jane Roe"}},
		},
		{
			name: "intro chapter with no byline → title-only suggestion",
			ast: astDoc(
				header(1, words("00 CHUA Introduction")...),
				para(words("This collection began as a question.")...),
			),
			want: []detectedChapter{{Title: "00 CHUA Introduction", Author: ""}},
		},
		{
			name: "front-matter h1 without byline still emitted (title-only)",
			ast: astDoc(
				header(1, words("Acknowledgments")...),
				para(words("Thanks to everyone who made this possible.")...),
				header(1, words("Real Chapter")...),
				para(words("By Spencer Nitkey")...),
			),
			want: []detectedChapter{
				{Title: "Acknowledgments", Author: ""},
				{Title: "Real Chapter", Author: "Spencer Nitkey"},
			},
		},
		{
			name: "no level-1 headings (Twitter Years) → empty",
			ast: astDoc(
				header(2, words("A Subhead")...),
				para(words("By Tuesday everything had changed.")...),
			),
			want: []detectedChapter{},
		},
		{
			name: "prose opener 'By the time…' does not false-match",
			ast: astDoc(
				header(1, words("Chapter Five")...),
				para(words("By the time he arrived at the station the train had gone.")...),
			),
			want: []detectedChapter{{Title: "Chapter Five", Author: ""}},
		},
		{
			name: "italic verse epigraph does not beat a real By byline",
			ast: astDoc(
				header(1, words("Soda Sweet as Blood")...),
				// 4-line italic epigraph (one Emph wrapping SoftBreaks) — long + multi-line.
				para(emph(joinInlines(
					words("The sea remembered everything it had ever swallowed"),
					[]any{softBreak()},
					words("and forgave nothing at all, not even the moon"),
				)...)),
				para(words("By Spencer Nitkey")...),
			),
			want: []detectedChapter{{Title: "Soda Sweet as Blood", Author: "Spencer Nitkey"}},
		},
		{
			name: "long italic line is not a byline (length cap)",
			ast: astDoc(
				header(1, words("Chapter Six")...),
				para(emph(words("This is a full italic sentence that runs well beyond sixty characters in length.")...)),
			),
			want: []detectedChapter{{Title: "Chapter Six", Author: ""}},
		},
		{
			name: "italic with em-dash attribution is not a byline",
			ast: astDoc(
				header(1, words("Chapter Seven")...),
				para(emph(joinInlines(words("Hope"), []any{map[string]any{"t": "Str", "c": "—"}}, words("E. Dickinson"))...)),
			),
			want: []detectedChapter{{Title: "Chapter Seven", Author: ""}},
		},
		{
			name: "byline outside the adjacency window is ignored",
			ast: astDoc(
				header(1, words("Chapter Eight")...),
				para(words("First paragraph of body text here.")...),
				para(words("Second paragraph of body text here.")...),
				para(words("Third paragraph of body text here.")...),
				para(words("By Spencer Nitkey")...), // 4th content block — out of window
			),
			want: []detectedChapter{{Title: "Chapter Eight", Author: ""}},
		},
		{
			name: "formatted heading title is flattened",
			ast: astDoc(
				header(1, str("The"), space(), emph(words("Real")...), space(), str("Story")),
				para(words("By Jane Doe")...),
			),
			want: []detectedChapter{{Title: "The Real Story", Author: "Jane Doe"}},
		},
		{
			name: "blank paragraph between heading and byline does not consume window",
			ast: astDoc(
				header(1, words("Chapter Nine")...),
				para(), // empty
				para(),
				para(),
				para(words("By Spencer Nitkey")...),
			),
			want: []detectedChapter{{Title: "Chapter Nine", Author: "Spencer Nitkey"}},
		},

		// --- Real DOCX shape: `--from=docx+styles` wraps every paragraph in a
		// Div tagged with its Word paragraph style. Byline paras are therefore
		// nested inside style Divs, never top-level Paras. These lock the
		// regression the live Ghosts/synthetic test surfaced.
		{
			name: "docx-wrapped By byline (First Paragraph style div)",
			ast: astDoc(
				header(1, words("Soda Sweet as Blood")...),
				customStyleDiv("First Paragraph", para(words("By Spencer Nitkey")...)),
				customStyleDiv("Body Text", para(words("Body paragraph one.")...)),
			),
			want: []detectedChapter{{Title: "Soda Sweet as Blood", Author: "Spencer Nitkey"}},
		},
		{
			name: "docx-wrapped italic byline (First Paragraph style div)",
			ast: astDoc(
				header(1, words("In Every Lifetime")...),
				customStyleDiv("First Paragraph", para(emph(words("Claire Pichelin")...))),
				customStyleDiv("Body Text", para(words("Body text.")...)),
			),
			want: []detectedChapter{{Title: "In Every Lifetime", Author: "Claire Pichelin"}},
		},
		{
			name: "docx-wrapped Word Author style div",
			ast: astDoc(
				header(1, words("The House")...),
				customStyleDiv("Author", para(words("Daniel Fernández")...)),
				customStyleDiv("Body Text", para(words("Body text.")...)),
			),
			want: []detectedChapter{{Title: "The House", Author: "Daniel Fernández"}},
		},
		{
			name: "docx-wrapped: Author style outranks an earlier italic line",
			ast: astDoc(
				header(1, words("Chapter X")...),
				customStyleDiv("First Paragraph", para(emph(words("An Italic Epigraph")...))),
				customStyleDiv("Author", para(words("Real Author")...)),
			),
			want: []detectedChapter{{Title: "Chapter X", Author: "Real Author"}},
		},
		{
			name: "docx-wrapped body prose (Body Text div) → title-only",
			ast: astDoc(
				header(1, words("Latency")...),
				customStyleDiv("Body Text", para(words("By the time she reached her floor, six had cleared.")...)),
			),
			want: []detectedChapter{{Title: "Latency", Author: ""}},
		},
		{
			name: "docx-wrapped byline beyond the window is ignored",
			ast: astDoc(
				header(1, words("Chapter Y")...),
				customStyleDiv("Body Text", para(words("First body paragraph.")...)),
				customStyleDiv("Body Text", para(words("Second body paragraph.")...)),
				customStyleDiv("Body Text", para(words("Third body paragraph.")...)),
				customStyleDiv("First Paragraph", para(words("By Too Late")...)),
			),
			want: []detectedChapter{{Title: "Chapter Y", Author: ""}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := detectChaptersFromAST(tc.ast)
			if err != nil {
				t.Fatalf("detectChaptersFromAST: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("count: got %d %+v, want %d %+v", len(got), got, len(tc.want), tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("chapter %d: got %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// joinInlines concatenates inline slices into one []any.
func joinInlines(groups ...[]any) []any {
	var out []any
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

func TestDetectChaptersFromAST_BadJSON(t *testing.T) {
	if _, err := detectChaptersFromAST([]byte("not json")); err == nil {
		t.Fatal("expected error for malformed AST JSON")
	}
}

// --- storeChapterSuggestions (DB round-trip) --------------------------------

func newTestProject(t *testing.T, s *Server, slug string) int64 {
	t.Helper()
	p, err := dbgen.New(s.DB).CreateProject(context.Background(), dbgen.CreateProjectParams{
		Name:        "Anthology " + slug,
		StartDate:   "2026-01-01",
		ClientSlug:  "client-" + slug,
		ProjectSlug: slug,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return p.ID
}

// specMap reads the stored spec JSON for a project as a generic map.
func specMap(t *testing.T, s *Server, projectID int64) map[string]any {
	t.Helper()
	row, err := dbgen.New(s.DB).GetBookSpec(context.Background(), projectID)
	if err != nil {
		t.Fatalf("get spec: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(row.Data), &m); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}
	return m
}

func TestStoreChapterSuggestions_CreatesSpecAndStores(t *testing.T) {
	s, _, cleanup := testServer(t)
	defer cleanup()
	ctx := context.Background()
	pid := newTestProject(t, s, "create")

	chs := []detectedChapter{
		{Title: "Soda Sweet as Blood", Author: "Spencer Nitkey"},
		{Title: "00 CHUA Introduction", Author: ""},
	}
	if err := s.storeChapterSuggestions(ctx, pid, chs); err != nil {
		t.Fatalf("store: %v", err)
	}

	m := specMap(t, s, pid)
	sug, ok := m["chapters_suggested"].([]any)
	if !ok || len(sug) != 2 {
		t.Fatalf("chapters_suggested: got %#v, want 2 entries", m["chapters_suggested"])
	}
	first := sug[0].(map[string]any)
	if first["title"] != "Soda Sweet as Blood" || first["author"] != "Spencer Nitkey" {
		t.Errorf("suggestion[0] = %#v", first)
	}
	// json_set must have stored a real JSON array, not a quoted string.
	if _, isStr := m["chapters_suggested"].(string); isStr {
		t.Error("chapters_suggested stored as string, expected array (json() wrapping failed)")
	}
}

func TestStoreChapterSuggestions_LeavesChaptersUntouched(t *testing.T) {
	s, _, cleanup := testServer(t)
	defer cleanup()
	ctx := context.Background()
	pid := newTestProject(t, s, "untouched")

	// Pre-populate a spec that already has a user-curated chapters[] list.
	seed := `{"metadata":{"title":"X"},"chapters":[{"title":"Hand-Entered","author":"Real Author","file":""}]}`
	if _, err := dbgen.New(s.DB).UpsertBookSpec(ctx, dbgen.UpsertBookSpecParams{ProjectID: pid, Data: seed}); err != nil {
		t.Fatalf("seed spec: %v", err)
	}

	if err := s.storeChapterSuggestions(ctx, pid, []detectedChapter{{Title: "Detected", Author: "Auto"}}); err != nil {
		t.Fatalf("store: %v", err)
	}

	m := specMap(t, s, pid)
	chapters, ok := m["chapters"].([]any)
	if !ok || len(chapters) != 1 {
		t.Fatalf("chapters[] mutated: got %#v", m["chapters"])
	}
	if c0 := chapters[0].(map[string]any); c0["title"] != "Hand-Entered" || c0["author"] != "Real Author" {
		t.Errorf("chapters[0] changed: %#v", c0)
	}
	if sug, _ := m["chapters_suggested"].([]any); len(sug) != 1 {
		t.Errorf("chapters_suggested: got %#v, want 1", m["chapters_suggested"])
	}
	// Unrelated keys survive the json_set merge.
	if meta, _ := m["metadata"].(map[string]any); meta["title"] != "X" {
		t.Errorf("metadata clobbered: %#v", m["metadata"])
	}
}

func TestStoreChapterSuggestions_EmptyNoSpec_DoesNotCreate(t *testing.T) {
	s, _, cleanup := testServer(t)
	defer cleanup()
	ctx := context.Background()
	pid := newTestProject(t, s, "empty-nospec")

	if err := s.storeChapterSuggestions(ctx, pid, nil); err != nil {
		t.Fatalf("store: %v", err)
	}
	// A single-author upload (zero detections) shouldn't materialize a spec row.
	var n int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM book_specs WHERE project_id = ?`, pid).Scan(&n); err != nil {
		t.Fatalf("count specs: %v", err)
	}
	if n != 0 {
		t.Errorf("expected no spec row created for empty suggestions, found %d", n)
	}
}

func TestStoreChapterSuggestions_EmptyClearsStale(t *testing.T) {
	s, _, cleanup := testServer(t)
	defer cleanup()
	ctx := context.Background()
	pid := newTestProject(t, s, "clear")

	if err := s.storeChapterSuggestions(ctx, pid, []detectedChapter{{Title: "Old", Author: "Stale"}}); err != nil {
		t.Fatalf("store initial: %v", err)
	}
	// A re-scan that finds nothing must clear the stale suggestions.
	if err := s.storeChapterSuggestions(ctx, pid, nil); err != nil {
		t.Fatalf("store empty: %v", err)
	}
	m := specMap(t, s, pid)
	if sug, ok := m["chapters_suggested"].([]any); !ok || len(sug) != 0 {
		t.Errorf("chapters_suggested not cleared: got %#v", m["chapters_suggested"])
	}
}
