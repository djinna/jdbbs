package srv

import "testing"

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
