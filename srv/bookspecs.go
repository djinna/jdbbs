package srv

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"srv.exe.dev/db/dbgen"
)

func defaultSpecData() string {
	return `{
  "metadata": {
    "title": "", "subtitle": "", "author": "", "series": "",
    "publisher": "", "isbn_paper": "", "isbn_cloth": "",
    "copyright_year": "", "copyright_holder": "", "credit_lines": ""
  },
  "page": {
    "trim": "5.5 x 8.5",
    "width_in": 5.5, "height_in": 8.5,
    "margin_top": "0.75in", "margin_bottom": "0.75in",
    "margin_inside": "0.7in", "margin_outside": "0.6in"
  },
  "typography": {
    "body_font": "Libertinus Serif", "heading_font": "Source Sans 3",
    "code_font": "JetBrains Mono",
    "base_size_pt": 10, "leading_pt": 2,
    "paragraph_indent_em": 0.75,
    "justify": true, "hyphenate": true
  },
  "headings": {
    "h1_size_em": 1.667, "h1_weight": "bold",
    "h2_size_em": 1.333, "h2_weight": 600,
    "h3_size_em": 1.0,   "h3_weight": "medium"
  },
  "elements": {
    "section_break": "breve",
    "code_block_size_em": 0.8, "poem_size_em": 0.75,
    "blockquote_style": "italic",
    "footnote_size_em": 0.75, "drop_caps": false
  },
  "running_headers": {
    "enabled": true, "verso": "author", "recto": "title",
    "font_size_em": 0.75
  },
  "front_matter": {
    "half_title": true, "series_title": false, "title_page": true,
    "copyright_page": true, "dedication": false, "epigraph": false,
    "toc": true, "foreword": false, "preface": false,
    "acknowledgments": false, "introduction": false
  },
  "back_matter": {
    "notes": false, "appendix": false, "glossary": false,
    "bibliography": false, "index": false
  },
  "custom_styles": [],
  "epub": {
    "toc_depth": 2, "landmarks": true,
    "chapter_break": "page", "section_break": "breve",
    "body_font_size": "inherit", "embed_fonts": false,
    "custom_css": "", "language": "en",
    "subject": "", "description": ""
  }
}`
}

// handleGetBookSpec returns spec for a project (auto-creates if missing).
func (s *Server) handleGetBookSpec(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	q := dbgen.New(s.DB)
	spec, err := q.GetBookSpec(r.Context(), pid)
	if err == sql.ErrNoRows {
		// Auto-create with defaults
		spec, err = q.UpsertBookSpec(r.Context(), dbgen.UpsertBookSpecParams{
			ProjectID: pid,
			Data:      defaultSpecData(),
		})
	}
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}

	var raw json.RawMessage
	json.Unmarshal([]byte(spec.Data), &raw)

	jsonOK(w, map[string]any{
		"id":         spec.ID,
		"project_id": spec.ProjectID,
		"data":       raw,
		"created_at": spec.CreatedAt,
		"updated_at": spec.UpdatedAt,
	})
}

// handleUpdateBookSpec saves spec JSON.
func (s *Server) handleUpdateBookSpec(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	var body struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}

	dataStr := string(body.Data)
	if dataStr == "" || dataStr == "null" {
		dataStr = defaultSpecData()
	}

	q := dbgen.New(s.DB)
	spec, err := q.UpsertBookSpec(r.Context(), dbgen.UpsertBookSpecParams{
		ProjectID: pid,
		Data:      dataStr,
	})
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]any{"ok": true, "updated_at": spec.UpdatedAt})
}

// handlePullTransmittalToSpec imports transmittal fields into spec.
func (s *Server) handlePullTransmittalToSpec(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	// Get transmittal data
	var txDataStr string
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT data FROM transmittals WHERE project_id = ?`, pid,
	).Scan(&txDataStr)
	if err != nil {
		jsonErr(w, "no transmittal found for this project", 404)
		return
	}

	// Parse transmittal
	var tx map[string]any
	if err := json.Unmarshal([]byte(txDataStr), &tx); err != nil {
		jsonErr(w, "bad transmittal data", 500)
		return
	}

	// Get current spec (or defaults)
	q := dbgen.New(s.DB)
	spec, err := q.GetBookSpec(r.Context(), pid)
	var specData map[string]any
	if err == sql.ErrNoRows {
		json.Unmarshal([]byte(defaultSpecData()), &specData)
	} else if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	} else {
		json.Unmarshal([]byte(spec.Data), &specData)
	}

	// Map transmittal → spec
	if book, ok := tx["book"].(map[string]any); ok {
		meta := ensureMap(specData, "metadata")
		mapField(book, "title", meta, "title")
		mapField(book, "subtitle", meta, "subtitle")
		mapField(book, "author", meta, "author")
		mapField(book, "series", meta, "series")
		mapField(book, "publisher", meta, "publisher")
		mapField(book, "isbn_paper", meta, "isbn_paper")
		mapField(book, "isbn_cloth", meta, "isbn_cloth")
	}
	if pageIV, ok := tx["page_iv"].(map[string]any); ok {
		meta := ensureMap(specData, "metadata")
		mapField(pageIV, "copyright_year", meta, "copyright_year")
		mapField(pageIV, "held_by", meta, "copyright_holder")
		mapField(pageIV, "credit", meta, "credit_lines")
	}
	if design, ok := tx["design"].(map[string]any); ok {
		if trim, ok := design["trim"].(string); ok && trim != "" {
			page := ensureMap(specData, "page")
			page["trim"] = trim
			parseTrim(trim, page)
		}
	}

	// Map front matter from checklist
	if cl, ok := tx["checklist"].([]any); ok {
		fm := ensureMap(specData, "front_matter")
		for _, item := range cl {
			if m, ok := item.(map[string]any); ok {
				comp, _ := m["component"].(string)
				here, _ := m["here_now"].(bool)
				switch comp {
				case "Half title pg":
					fm["half_title"] = here
				case "Series title/Frontis.":
					fm["series_title"] = here
				case "Title pg":
					fm["title_page"] = here
				case "Copyright pg":
					fm["copyright_page"] = here
				case "Dedication":
					fm["dedication"] = here
				case "Epigraph":
					fm["epigraph"] = here
				case "Contents":
					fm["toc"] = here
				case "Foreword":
					fm["foreword"] = here
				case "Preface":
					fm["preface"] = here
				case "Acknowledgments":
					fm["acknowledgments"] = here
				case "Introduction":
					fm["introduction"] = here
				}
			}
		}
	}

	// Map back matter
	if bm, ok := tx["backmatter"].([]any); ok {
		bmSpec := ensureMap(specData, "back_matter")
		for _, item := range bm {
			if m, ok := item.(map[string]any); ok {
				comp, _ := m["component"].(string)
				here, _ := m["here_now"].(bool)
				switch comp {
				case "Notes":
					bmSpec["notes"] = here
				case "Appendix(es)":
					bmSpec["appendix"] = here
				case "Glossary":
					bmSpec["glossary"] = here
				case "Bibliography":
					bmSpec["bibliography"] = here
				case "Index":
					bmSpec["index"] = here
				}
			}
		}
	}

	// Save
	newData, _ := json.Marshal(specData)
	_, err = q.UpsertBookSpec(r.Context(), dbgen.UpsertBookSpecParams{
		ProjectID: pid,
		Data:      string(newData),
	})
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}

	var raw json.RawMessage
	json.Unmarshal(newData, &raw)
	jsonOK(w, map[string]any{"ok": true, "data": raw})
}

// handleListFonts returns available font families from the fonts directory.
func (s *Server) handleListFonts(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	// Hardcoded for now — could scan fonts/ dir later
	fonts := []map[string]string{
		{"family": "Source Sans 3", "category": "sans-serif"},
		{"family": "JetBrains Mono", "category": "monospace"},
		// System fonts that Typst can use
		{"family": "Libertinus Serif", "category": "serif"},
	}
	jsonOK(w, fonts)
}

// handleGenerateConfig returns the Typst config that would be generated from the spec.
func (s *Server) handleGenerateConfig(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	q := dbgen.New(s.DB)
	spec, err := q.GetBookSpec(r.Context(), pid)
	if err != nil {
		jsonErr(w, "no spec found", 404)
		return
	}

	var data map[string]any
	json.Unmarshal([]byte(spec.Data), &data)

	config := specToTypstConfig(data)
	jsonOK(w, map[string]any{"config": config})
}

// specToTypstConfig converts a spec JSON map into Typst config override code.
func specToTypstConfig(data map[string]any) string {
	var lines []string
	lines = append(lines, "\n// Project-specific config overrides (from spec)")
	lines = append(lines, "#let config = merge-config((")

	if page, ok := data["page"].(map[string]any); ok {
		if w, ok := page["width_in"].(float64); ok && w > 0 {
			lines = append(lines, fmt.Sprintf("  page-width: %gin,", w))
		}
		if h, ok := page["height_in"].(float64); ok && h > 0 {
			lines = append(lines, fmt.Sprintf("  page-height: %gin,", h))
		}
		if v, ok := page["margin_top"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf("  margin-top: %s,", v))
		}
		if v, ok := page["margin_bottom"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf("  margin-bottom: %s,", v))
		}
		if v, ok := page["margin_inside"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf("  margin-inside: %s,", v))
		}
		if v, ok := page["margin_outside"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf("  margin-outside: %s,", v))
		}
	}

	if typo, ok := data["typography"].(map[string]any); ok {
		if v, ok := typo["body_font"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf(`  body-font: "%s",`, v))
		}
		if v, ok := typo["heading_font"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf(`  heading-font: "%s",`, v))
		}
		if v, ok := typo["code_font"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf(`  code-font: "%s",`, v))
		}
		if v, ok := typo["base_size_pt"].(float64); ok && v > 0 {
			lines = append(lines, fmt.Sprintf("  base-size: %gpt,", v))
		}
		if v, ok := typo["leading_pt"].(float64); ok {
			lines = append(lines, fmt.Sprintf("  leading: %gpt,", v))
		}
		if v, ok := typo["paragraph_indent_em"].(float64); ok {
			lines = append(lines, fmt.Sprintf("  paragraph-indent: %gem,", v))
		}
	}

	lines = append(lines, "))")

	// Custom styles
	if styles, ok := data["custom_styles"].([]any); ok {
		var styleCodes []string
		for _, s := range styles {
			if m, ok := s.(map[string]any); ok {
				if code, ok := m["typst"].(string); ok && code != "" {
					styleCodes = append(styleCodes, code)
				}
			}
		}
		if len(styleCodes) > 0 {
			lines = append(lines, "\n// Custom styles")
			lines = append(lines, styleCodes...)
		}
	}

	return strings.Join(lines, "\n")
}

// helpers

func ensureMap(parent map[string]any, key string) map[string]any {
	if m, ok := parent[key].(map[string]any); ok {
		return m
	}
	m := map[string]any{}
	parent[key] = m
	return m
}

func mapField(src map[string]any, srcKey string, dst map[string]any, dstKey string) {
	if v, ok := src[srcKey]; ok {
		dst[dstKey] = v
	}
}

func parseTrim(trim string, page map[string]any) {
	// Parse "5.5 x 8.5" → width/height
	trimPresets := map[string][2]float64{
		"5.5 x 8.5":  {5.5, 8.5},
		"6 x 9":      {6.0, 9.0},
		"8.5 x 11":   {8.5, 11.0},
		"5 x 8":      {5.0, 8.0},
	}
	if dims, ok := trimPresets[trim]; ok {
		page["width_in"] = dims[0]
		page["height_in"] = dims[1]
	}
}
