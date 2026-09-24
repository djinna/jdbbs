package srv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFactoryEventsFeedAndBoard(t *testing.T) {
	s, _, cleanup := testServer(t)
	defer cleanup()
	pid := newTestProject(t, s, "feed-test")

	s.factoryEvent(pid, "", "manuscript.uploaded", "client:feedtest", "x.docx (1 KB)")
	s.factoryEvent(0, "feedtest", "login.failed", "", "wrong password")

	req := httptest.NewRequest("GET", "/api/admin/factory/events?limit=10", nil)
	req.Header.Set("X-ExeDev-UserID", "admin")
	req.Header.Set("X-ExeDev-Email", "owner@example.test")
	rr := httptest.NewRecorder()
	s.handleAdminFactoryEvents(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("events: %d %s", rr.Code, rr.Body.String())
	}
	var evs []factoryEventRow
	if err := json.Unmarshal(rr.Body.Bytes(), &evs); err != nil {
		t.Fatal(err)
	}
	if len(evs) != 2 {
		t.Fatalf("want 2 events, got %d", len(evs))
	}
	// Newest first; the anonymous login row has actor "anon" and no project.
	if evs[0].Kind != "login.failed" || evs[0].Actor != "anon" || evs[0].ProjectID != 0 || evs[0].ClientSlug != "feedtest" {
		t.Errorf("unexpected first row: %+v", evs[0])
	}
	if evs[1].Kind != "manuscript.uploaded" || evs[1].ProjectID != pid || evs[1].ProjectName == "" {
		t.Errorf("unexpected second row: %+v", evs[1])
	}

	// after=ID returns only newer rows.
	req = httptest.NewRequest("GET", "/api/admin/factory/events?after="+itoa(evs[1].ID), nil)
	req.Header.Set("X-ExeDev-UserID", "admin")
	req.Header.Set("X-ExeDev-Email", "owner@example.test")
	rr = httptest.NewRecorder()
	s.handleAdminFactoryEvents(rr, req)
	var newer []factoryEventRow
	_ = json.Unmarshal(rr.Body.Bytes(), &newer)
	if len(newer) != 1 || newer[0].ID != evs[0].ID {
		t.Errorf("after: want 1 newer row, got %+v", newer)
	}

	// Board: no passes yet → empty list, not an error.
	req = httptest.NewRequest("GET", "/api/admin/factory/board", nil)
	req.Header.Set("X-ExeDev-UserID", "admin")
	req.Header.Set("X-ExeDev-Email", "owner@example.test")
	rr = httptest.NewRecorder()
	s.handleAdminFactoryBoard(rr, req)
	if rr.Code != http.StatusOK || rr.Body.String() != "[]\n" && rr.Body.String() != "[]" {
		t.Fatalf("board: %d %q", rr.Code, rr.Body.String())
	}

	// Not admin → 401/403.
	rr = httptest.NewRecorder()
	s.handleAdminFactoryEvents(rr, httptest.NewRequest("GET", "/api/admin/factory/events", nil))
	if rr.Code == http.StatusOK {
		t.Errorf("events should require admin, got %d", rr.Code)
	}
}
