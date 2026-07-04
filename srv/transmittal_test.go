package srv

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestTransmittalDefaultsIncludeEPUBISBNAndChecklistStatus(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Transmittal Defaults",
		"start_date":   "2026-04-08",
		"client_slug":  "vgr",
		"project_slug": "defaults",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/transmittal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get transmittal: expected 200, got %d", resp.StatusCode)
	}
	var tx map[string]any
	decodeJSON(t, resp, &tx)

	data := tx["data"].(map[string]any)
	book := data["book"].(map[string]any)
	if _, ok := book["isbn_epub"]; !ok {
		t.Fatalf("expected default book data to include isbn_epub, got keys %+v", book)
	}

	checklist := data["checklist"].([]any)
	firstChecklist := checklist[0].(map[string]any)
	if _, ok := firstChecklist["status"]; !ok {
		t.Fatalf("expected checklist item to include status, got %+v", firstChecklist)
	}
	if firstChecklist["status"] != "" {
		t.Fatalf("expected default checklist status to be empty, got %v", firstChecklist["status"])
	}

	backmatter := data["backmatter"].([]any)
	firstBackmatter := backmatter[0].(map[string]any)
	if _, ok := firstBackmatter["status"]; !ok {
		t.Fatalf("expected backmatter item to include status, got %+v", firstBackmatter)
	}
	if firstBackmatter["status"] != "" {
		t.Fatalf("expected default backmatter status to be empty, got %v", firstBackmatter["status"])
	}

	styles, ok := data["custom_styles"].([]any)
	if !ok {
		t.Fatalf("expected default transmittal data to include custom_styles array, got %T", data["custom_styles"])
	}
	if len(styles) != 0 {
		t.Fatalf("expected default custom_styles to be empty, got %d items", len(styles))
	}
}

func TestDuplicateTransmittalClearsChecklistStatusesAndDates(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	createProject := func(name, slug string) int64 {
		resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
			"name":         name,
			"start_date":   "2026-04-08",
			"client_slug":  "vgr",
			"project_slug": slug,
		})
		if resp.StatusCode != 201 {
			t.Fatalf("create project %s: expected 201, got %d", slug, resp.StatusCode)
		}
		var project map[string]any
		decodeJSON(t, resp, &project)
		return int64(project["ID"].(float64))
	}

	sourcePID := createProject("Source", "source")
	targetPID := createProject("Target", "target")
	sourcePIDStr := itoa(sourcePID)
	targetPIDStr := itoa(targetPID)

	resp := apiRequestAdmin(t, ts, "GET", "/api/projects/"+sourcePIDStr+"/transmittal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get source transmittal: expected 200, got %d", resp.StatusCode)
	}
	var tx map[string]any
	decodeJSON(t, resp, &tx)
	data := tx["data"].(map[string]any)

	book := data["book"].(map[string]any)
	book["title"] = "Filled Source"
	book["isbn_epub"] = "9780000000001"

	checklist := data["checklist"].([]any)
	checklist[0].(map[string]any)["status"] = "included"
	checklist[0].(map[string]any)["here_now"] = true
	checklist[1].(map[string]any)["status"] = "later"
	checklist[1].(map[string]any)["to_come_when"] = "2026-05-01"

	backmatter := data["backmatter"].([]any)
	backmatter[0].(map[string]any)["status"] = "not_in_book"
	backmatter[0].(map[string]any)["to_come_when"] = "2026-05-02"

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+sourcePIDStr+"/transmittal", map[string]any{
		"status": "draft",
		"data":   data,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update source transmittal: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "POST", "/api/transmittals/"+sourcePIDStr+"/duplicate", map[string]any{
		"target_project_id": targetPID,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("duplicate transmittal: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+targetPIDStr+"/transmittal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get target transmittal: expected 200, got %d", resp.StatusCode)
	}
	var targetTx map[string]any
	decodeJSON(t, resp, &targetTx)
	duplicated := targetTx["data"].(map[string]any)

	dupBook := duplicated["book"].(map[string]any)
	if dupBook["title"] != "" {
		t.Fatalf("expected duplicated transmittal title to be cleared, got %v", dupBook["title"])
	}

	dupChecklist := duplicated["checklist"].([]any)
	for i := 0; i < 2; i++ {
		item := dupChecklist[i].(map[string]any)
		if item["status"] != "" {
			t.Fatalf("expected duplicated checklist item %d status to be cleared, got %v", i, item["status"])
		}
		if item["to_come_when"] != "" {
			t.Fatalf("expected duplicated checklist item %d to_come_when to be cleared, got %v", i, item["to_come_when"])
		}
		if item["here_now"] != false {
			t.Fatalf("expected duplicated checklist item %d here_now to be false, got %v", i, item["here_now"])
		}
	}

	dupBackmatter := duplicated["backmatter"].([]any)
	bm := dupBackmatter[0].(map[string]any)
	if bm["status"] != "" {
		t.Fatalf("expected duplicated backmatter status to be cleared, got %v", bm["status"])
	}
	if bm["to_come_when"] != "" {
		t.Fatalf("expected duplicated backmatter to_come_when to be cleared, got %v", bm["to_come_when"])
	}
	if bm["here_now"] != false {
		t.Fatalf("expected duplicated backmatter here_now to be false, got %v", bm["here_now"])
	}
}

// TestTransmittalVersionCapEnforced seeds more versions than the cap, then
// triggers one more snapshot via PUT and asserts pruning keeps exactly the
// newest maxTransmittalVersions rows.
func TestTransmittalVersionCapEnforced(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Version Cap",
		"start_date":   "2026-04-08",
		"client_slug":  "vgr",
		"project_slug": "version-cap",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// First PUT creates the transmittal row (no version yet: nothing to snapshot).
	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/transmittal", map[string]any{
		"status": "draft",
		"data":   map[string]any{"v": "initial"},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("initial put: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	var txID int64
	if err := s.DB.QueryRow(`SELECT id FROM transmittals WHERE project_id = ?`, pid).Scan(&txID); err != nil {
		t.Fatalf("lookup transmittal id: %v", err)
	}

	// Seed 60 versions, all older than the 5-minute snapshot throttle, with
	// distinct saved_at so prune ordering is deterministic (seed 59 newest).
	for i := range 60 {
		_, err := s.DB.Exec(
			`INSERT INTO transmittal_versions (transmittal_id, data, status, saved_at)
			 VALUES (?, ?, 'draft', datetime('now', ?))`,
			txID, fmt.Sprintf(`{"marker":"seed-%d"}`, i), fmt.Sprintf("-%d minutes", 70-i),
		)
		if err != nil {
			t.Fatalf("seed version %d: %v", i, err)
		}
	}

	// Second PUT snapshots the pre-update state (61st version) and must prune
	// down to the cap in the same transaction.
	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/transmittal", map[string]any{
		"status": "draft",
		"data":   map[string]any{"v": "updated"},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("second put: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	var count int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM transmittal_versions WHERE transmittal_id = ?`, txID).Scan(&count); err != nil {
		t.Fatalf("count versions: %v", err)
	}
	if count != maxTransmittalVersions {
		t.Fatalf("expected %d versions after prune, got %d", maxTransmittalVersions, count)
	}

	countData := func(data string) int {
		var n int
		if err := s.DB.QueryRow(
			`SELECT COUNT(*) FROM transmittal_versions WHERE transmittal_id = ? AND data = ?`,
			txID, data,
		).Scan(&n); err != nil {
			t.Fatalf("count %q: %v", data, err)
		}
		return n
	}
	// Newest retained: the fresh snapshot of the pre-update state plus the 49
	// newest seeds (59 down to 11); seeds 10 and older are pruned.
	if n := countData(`{"v":"initial"}`); n != 1 {
		t.Fatalf("expected fresh snapshot to be retained, found %d rows", n)
	}
	if n := countData(`{"marker":"seed-11"}`); n != 1 {
		t.Fatalf("expected seed-11 (within newest 50) retained, found %d rows", n)
	}
	if n := countData(`{"marker":"seed-10"}`); n != 0 {
		t.Fatalf("expected seed-10 (beyond cap) pruned, found %d rows", n)
	}
}

// TestTransmittalVersionInsertFailureFailsUpdate forces the version INSERT to
// fail (RAISE(ABORT) trigger) and asserts the PUT returns 500 and rolls back,
// leaving the transmittal untouched instead of silently succeeding.
func TestTransmittalVersionInsertFailureFailsUpdate(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Version Failure",
		"start_date":   "2026-04-08",
		"client_slug":  "vgr",
		"project_slug": "version-failure",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create project: expected 201, got %d", resp.StatusCode)
	}
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/transmittal", map[string]any{
		"status": "draft",
		"data":   map[string]any{"v": "original"},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("initial put: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	if _, err := s.DB.Exec(
		`CREATE TRIGGER fail_version_insert BEFORE INSERT ON transmittal_versions
		 BEGIN SELECT RAISE(ABORT, 'version insert forced to fail'); END`,
	); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/transmittal", map[string]any{
		"status": "draft",
		"data":   map[string]any{"v": "should-not-land"},
	})
	if resp.StatusCode != 500 {
		t.Fatalf("expected 500 when version insert fails, got %d", resp.StatusCode)
	}
	var errBody map[string]string
	decodeJSON(t, resp, &errBody)
	if errBody["error"] == "" {
		t.Fatalf("expected error message in 500 body, got %+v", errBody)
	}

	// The whole write must have rolled back: transmittal keeps original data.
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/transmittal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get transmittal: expected 200, got %d", resp.StatusCode)
	}
	var tx struct {
		Data json.RawMessage `json:"data"`
	}
	decodeJSON(t, resp, &tx)
	var data map[string]any
	if err := json.Unmarshal(tx.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data["v"] != "original" {
		t.Fatalf("expected transmittal data unchanged after failed version insert, got %v", data["v"])
	}

	// No version row may exist either (the snapshot itself failed).
	var versions int
	if err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM transmittal_versions v
		 JOIN transmittals t ON t.id = v.transmittal_id
		 WHERE t.project_id = ?`, pid,
	).Scan(&versions); err != nil {
		t.Fatalf("count versions: %v", err)
	}
	if versions != 0 {
		t.Fatalf("expected 0 versions after rollback, got %d", versions)
	}
}
