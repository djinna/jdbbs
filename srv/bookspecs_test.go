package srv

import (
	"strings"
	"testing"
)

func TestPullTransmittalMapsCustomStyles(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Custom Style Mapping",
		"start_date":   "2026-04-09",
		"client_slug":  "vgr",
		"project_slug": "custom-style-mapping",
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
	data["custom_styles"] = []map[string]any{
		{
			"name":        "vgr-tweet",
			"type":        "paragraph",
			"description": "Tweet / short social post",
		},
		{
			"name":        "vgr-handle",
			"type":        "character",
			"description": "Social media handle",
		},
	}

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/transmittal", map[string]any{
		"status": "draft",
		"data":   data,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update transmittal: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/book-spec/pull-transmittal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("pull transmittal into spec: expected 200, got %d", resp.StatusCode)
	}
	var result map[string]any
	decodeJSON(t, resp, &result)
	dataOut := result["data"].(map[string]any)
	styles, ok := dataOut["custom_styles"].([]any)
	if !ok {
		t.Fatalf("expected custom_styles array in pulled spec, got %T", dataOut["custom_styles"])
	}
	if len(styles) != 2 {
		t.Fatalf("expected 2 custom styles, got %d", len(styles))
	}

	first := styles[0].(map[string]any)
	if first["name"] != "vgr-tweet" {
		t.Fatalf("expected first custom style name vgr-tweet, got %v", first["name"])
	}
	if first["word_style"] != "vgr-tweet" {
		t.Fatalf("expected first custom style word_style vgr-tweet, got %v", first["word_style"])
	}
	if first["type"] != "paragraph" {
		t.Fatalf("expected first custom style type paragraph, got %v", first["type"])
	}
	if first["description"] != "Tweet / short social post" {
		t.Fatalf("expected first custom style description preserved, got %v", first["description"])
	}

	second := styles[1].(map[string]any)
	if second["name"] != "vgr-handle" {
		t.Fatalf("expected second custom style name vgr-handle, got %v", second["name"])
	}
	if second["word_style"] != "vgr-handle" {
		t.Fatalf("expected second custom style word_style vgr-handle, got %v", second["word_style"])
	}
	if second["type"] != "character" {
		t.Fatalf("expected second custom style type character, got %v", second["type"])
	}
}

func TestPullTransmittalMapsSharedTypesettingFields(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name":         "Shared Typesetting Mapping",
		"start_date":   "2026-04-12",
		"client_slug":  "vgr",
		"project_slug": "shared-typesetting-mapping",
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

	editing := data["editing"].(map[string]any)
	editing["developmental_instructions"] = "Focus on structure and pacing."
	editing["instructions"] = "Preserve house italics conventions."

	design := data["design"].(map[string]any)
	design["trim_guidance"] = "Gift-book feel; flexible if cost pushes smaller trim."
	design["trim"] = "6 x 9"
	design["est_pages"] = "240"
	design["ppi"] = "300"
	design["spine_width"] = "0.65 in"
	design["complexity"] = "complex_jdbb"
	design["outside_designer"] = "Jane Doe"
	design["reuse_previous"] = "Series Vol. 1"
	design["freeform_notes"] = "Needs room for image-heavy openers."

	resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/transmittal", map[string]any{
		"status": "draft",
		"data":   data,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update transmittal: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/book-spec/pull-transmittal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("pull transmittal into spec: expected 200, got %d", resp.StatusCode)
	}
	var result map[string]any
	decodeJSON(t, resp, &result)
	dataOut := result["data"].(map[string]any)
	typesetting, ok := dataOut["typesetting"].(map[string]any)
	if !ok {
		t.Fatalf("expected typesetting map in pulled spec, got %T", dataOut["typesetting"])
	}

	if typesetting["developmental_instructions"] != "Focus on structure and pacing." {
		t.Fatalf("expected developmental instructions mapped, got %v", typesetting["developmental_instructions"])
	}
	if typesetting["copyeditor_instructions"] != "Preserve house italics conventions." {
		t.Fatalf("expected copyeditor instructions mapped, got %v", typesetting["copyeditor_instructions"])
	}
	if typesetting["trim_guidance"] != "Gift-book feel; flexible if cost pushes smaller trim." {
		t.Fatalf("expected trim guidance mapped, got %v", typesetting["trim_guidance"])
	}
	if typesetting["trim_size"] != "6 x 9" {
		t.Fatalf("expected trim size mapped, got %v", typesetting["trim_size"])
	}
	if typesetting["design_notes"] != "Needs room for image-heavy openers." {
		t.Fatalf("expected design notes mapped, got %v", typesetting["design_notes"])
	}

	page := dataOut["page"].(map[string]any)
	if page["trim"] != "6 x 9" {
		t.Fatalf("expected page.trim still mapped from design.trim, got %v", page["trim"])
	}
}

func TestTypstStyleIdent(t *testing.T) {
	cases := map[string]string{
		"Field Note": "field-note", "tweet-p": "tweet-p", "  Term ": "term",
		"Sidebar__Box!": "sidebar-box", "2nd Level": "s-2nd-level", "": "my-style", "Metadata (c)": "metadata-c",
	}
	for in, want := range cases {
		if got := typstStyleIdent(in); got != want {
			t.Errorf("typstStyleIdent(%q) = %q, want %q", in, got, want)
		}
	}
	// The generated #let must use the same identifier.
	if code := defaultCustomStyleTypst("Field Note", "paragraph", ""); !strings.Contains(code, "#let field-note(") {
		t.Errorf("defaultCustomStyleTypst did not use the ident:\n%s", code)
	}
}

func TestDeclaredStylesForPandoc(t *testing.T) {
	spec := `{"custom_styles":[{"name":"Field Note","type":"paragraph"},{"name":"Term","word_style":"Term Char","type":"character"},{"name":"  "}]}`
	got := declaredStylesForPandoc(spec)
	if len(got) != 2 {
		t.Fatalf("want 2 styles, got %d: %#v", len(got), got)
	}
	if got[0] != (pandocStyle{WordStyle: "Field Note", Ident: "field-note", Type: "paragraph"}) {
		t.Errorf("first: %#v", got[0])
	}
	if got[1] != (pandocStyle{WordStyle: "Term Char", Ident: "term", Type: "character"}) {
		t.Errorf("second: %#v", got[1])
	}
	if declaredStylesForPandoc(`{}`) != nil || declaredStylesForPandoc(`nope`) != nil {
		t.Error("no styles should yield nil")
	}
}

func TestParseTrimFreeForm(t *testing.T) {
	cases := map[string][2]float64{
		"7 x 10":         {7, 10},
		"7x10":           {7, 10},
		"6.14 × 9.21 in": {6.14, 9.21},
		"6 x 9":          {6, 9}, // registry alias still wins
	}
	for in, want := range cases {
		page := map[string]any{}
		parseTrim(in, page)
		if page["width_in"] != want[0] || page["height_in"] != want[1] {
			t.Errorf("parseTrim(%q) = %v x %v, want %v", in, page["width_in"], page["height_in"], want)
		}
	}
	page := map[string]any{}
	parseTrim("coffee table", page)
	if _, ok := page["width_in"]; ok {
		t.Errorf("parseTrim of prose should not set a width")
	}
}

func TestStudioTrimLeavesPageAlone(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	for _, v := range []string{"studio", "dont_care"} {
		resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
			"name": "Studio trim " + v, "start_date": "2026-04-12",
			"client_slug": "vgr", "project_slug": "studio-trim-" + v,
		})
		var project map[string]any
		decodeJSON(t, resp, &project)
		pid := itoa(int64(project["ID"].(float64)))

		resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/transmittal", nil)
		var tx map[string]any
		decodeJSON(t, resp, &tx)
		data := tx["data"].(map[string]any)
		data["design"].(map[string]any)["trim"] = v
		resp = apiRequestAdmin(t, ts, "PUT", "/api/projects/"+pid+"/transmittal", map[string]any{"status": "draft", "data": data})
		resp.Body.Close()

		resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/book-spec/pull-transmittal", nil)
		var result map[string]any
		decodeJSON(t, resp, &result)
		page := result["data"].(map[string]any)["page"].(map[string]any)
		if page["trim"] != "protocolized" {
			t.Errorf("trim %q should leave the series trim alone, got page.trim=%v", v, page["trim"])
		}
	}
}
