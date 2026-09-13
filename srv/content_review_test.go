package srv

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestContentReviewImportDecideExport(t *testing.T) {
	_, ts, cleanup := testServer(t)
	t.Cleanup(cleanup)

	// Import a batch.
	var imp struct{ Added int }
	code := adminJSON(t, ts, "POST", "/api/admin/content-review", map[string]any{
		"batch": "2026-09-13",
		"items": []map[string]any{
			{"ref": "A-F1", "page": "/workshop", "file": "jdbbs-public/workshop.html", "location": "dek", "current": "old", "proposed": "new", "reason": "buried ask", "severity": "MAJOR", "pass": 1},
			{"ref": "A-F2", "page": "/workshop", "current": "x", "proposed": "y"},
		},
	}, &imp)
	if code != 200 || imp.Added != 2 {
		t.Fatalf("import: code=%d added=%d", code, imp.Added)
	}
	// Re-import is idempotent.
	code = adminJSON(t, ts, "POST", "/api/admin/content-review", map[string]any{
		"batch": "2026-09-13", "items": []map[string]any{{"ref": "A-F1", "proposed": "clobber"}},
	}, &imp)
	if code != 200 || imp.Added != 0 {
		t.Fatalf("re-import: code=%d added=%d", code, imp.Added)
	}

	var list struct{ Items []contentReviewItem }
	if code = adminJSON(t, ts, "GET", "/api/admin/content-review", nil, &list); code != 200 || len(list.Items) != 2 {
		t.Fatalf("list: code=%d n=%d", code, len(list.Items))
	}
	if list.Items[0].Decision != "pending" || list.Items[0].Proposed != "new" || list.Items[1].Severity != "MINOR" {
		t.Fatalf("defaults wrong: %+v", list.Items)
	}

	// Decide: accept one, edit the other; bad decision rejected.
	id1, id2 := list.Items[0].ID, list.Items[1].ID
	var it contentReviewItem
	if code = adminJSON(t, ts, "PUT", "/api/admin/content-review/"+itoa(id1), map[string]any{"decision": "accepted"}, &it); code != 200 || it.Decision != "accepted" || it.DecidedAt == "" {
		t.Fatalf("accept: code=%d item=%+v", code, it)
	}
	if code = adminJSON(t, ts, "PUT", "/api/admin/content-review/"+itoa(id2), map[string]any{"decision": "edited"}, nil); code != 400 {
		t.Fatalf("edited without text should 400, got %d", code)
	}
	if code = adminJSON(t, ts, "PUT", "/api/admin/content-review/"+itoa(id2), map[string]any{"decision": "edited", "edited_text": "z", "note": "shorter"}, &it); code != 200 || it.EditedText != "z" {
		t.Fatalf("edit: code=%d item=%+v", code, it)
	}
	if code = adminJSON(t, ts, "PUT", "/api/admin/content-review/"+itoa(id2), map[string]any{"decision": "bogus"}, nil); code != 400 {
		t.Fatalf("bogus decision should 400, got %d", code)
	}

	// Export lists both, with the edited text for the edited one.
	resp := apiRequestAdmin(t, ts, "GET", "/api/admin/content-review/export", nil)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	md := string(body)
	if resp.StatusCode != 200 || !strings.Contains(md, "## A-F1") || !strings.Contains(md, "## A-F2") || !strings.Contains(md, "\nz\n") || strings.Contains(md, "\ny\n") {
		t.Fatalf("export: code=%d body=%s", resp.StatusCode, md)
	}

	// Non-admin gets nothing.
	r, _ := http.Get(ts.URL + "/api/admin/content-review")
	if r.StatusCode == 200 {
		t.Fatalf("unauthenticated list should not be 200")
	}
	r.Body.Close()
	noFollow := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	r, _ = noFollow.Get(ts.URL + "/admin/content-review/")
	if r.StatusCode != http.StatusFound {
		t.Fatalf("unauthenticated page should redirect to login, got %d", r.StatusCode)
	}
	r.Body.Close()
}
