package srv

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"srv.exe.dev/db/dbgen"
)

func defaultSpecData() string {
	return `{
  "metadata": {
    "title": "", "subtitle": "", "author": "", "series": "",
    "publisher": "", "isbn_paper": "", "isbn_epub": "", "isbn_cloth": "",
    "copyright_year": "", "copyright_holder": "", "credit_lines": ""
  },
  "page": {
    "trim": "us-digest",
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
    "h1_size_pt": 16.67, "h1_weight": "bold",
    "h2_size_pt": 13.33, "h2_weight": 600,
    "h3_size_pt": 10,     "h3_weight": "medium"
  },
  "elements": {
    "section_break": "breve", "blockquote_style": "italic",
    "poem_size_pt": 7.5, "code_block_size_pt": 8,
    "footnote_size_pt": 7.5
  },
  "running_heads": {
    "enabled": true, "head_size_pt": 7.5,
    "verso": "author", "recto": "title"
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
  "typesetting": {
    "developmental_instructions": "", "copyeditor_instructions": "",
    "trim_guidance": "", "trim_size": "", "est_pages": "",
    "ppi": "", "spine_width": "", "complexity": "",
    "outside_designer": "", "reuse_previous": "", "design_notes": ""
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
		full, uErr := q.UpsertBookSpec(r.Context(), dbgen.UpsertBookSpecParams{
			ProjectID: pid,
			Data:      defaultSpecData(),
		})
		if uErr != nil {
			jsonErr(w, uErr.Error(), 500)
			return
		}
		spec = dbgen.GetBookSpecRow{
			ID: full.ID, ProjectID: full.ProjectID,
			Data: full.Data, CreatedAt: full.CreatedAt, UpdatedAt: full.UpdatedAt,
		}
		err = nil
	}
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}

	// Check if cover exists
	hasCover := false
	coverRow, cErr := q.GetBookSpecCover(r.Context(), pid)
	if cErr == nil && coverRow.CoverData != nil && len(coverRow.CoverData) > 0 {
		hasCover = true
	}

	var raw json.RawMessage
	json.Unmarshal([]byte(spec.Data), &raw)

	jsonOK(w, map[string]any{
		"id":         spec.ID,
		"project_id": spec.ProjectID,
		"data":       raw,
		"has_cover":  hasCover,
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
		mapField(book, "isbn_epub", meta, "isbn_epub")
		mapField(book, "isbn_cloth", meta, "isbn_cloth")
	}
	if pageIV, ok := tx["page_iv"].(map[string]any); ok {
		meta := ensureMap(specData, "metadata")
		mapField(pageIV, "copyright_year", meta, "copyright_year")
		mapField(pageIV, "held_by", meta, "copyright_holder")
		mapField(pageIV, "credit", meta, "credit_lines")
	}
	if editing, ok := tx["editing"].(map[string]any); ok {
		typesetting := ensureMap(specData, "typesetting")
		mapField(editing, "developmental_instructions", typesetting, "developmental_instructions")
		mapField(editing, "instructions", typesetting, "copyeditor_instructions")
	}
	if design, ok := tx["design"].(map[string]any); ok {
		typesetting := ensureMap(specData, "typesetting")
		mapField(design, "trim_guidance", typesetting, "trim_guidance")
		mapField(design, "trim", typesetting, "trim_size")
		mapField(design, "est_pages", typesetting, "est_pages")
		mapField(design, "ppi", typesetting, "ppi")
		mapField(design, "spine_width", typesetting, "spine_width")
		mapField(design, "complexity", typesetting, "complexity")
		mapField(design, "outside_designer", typesetting, "outside_designer")
		mapField(design, "reuse_previous", typesetting, "reuse_previous")
		mapField(design, "freeform_notes", typesetting, "design_notes")
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

	if styles, ok := tx["custom_styles"].([]any); ok {
		var mapped []any
		for _, item := range styles {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name, _ := m["name"].(string)
			styleType, _ := m["type"].(string)
			desc, _ := m["description"].(string)
			name = strings.TrimSpace(name)
			styleType = strings.TrimSpace(styleType)
			desc = strings.TrimSpace(desc)
			if name == "" {
				continue
			}
			if styleType == "" {
				styleType = "paragraph"
			}
			mapped = append(mapped, map[string]any{
				"name":        name,
				"word_style":  name,
				"type":        styleType,
				"description": desc,
			})
		}
		specData["custom_styles"] = mapped
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

// handleListFonts returns available font families by running `typst fonts`.
func (s *Server) handleListFonts(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}

	fonts := listTypstFonts()
	jsonOK(w, fonts)
}

// listTypstFonts runs `typst fonts --font-path <fontsDir>` and categorizes results.
func listTypstFonts() []map[string]string {
	cmd := exec.Command("typst", "fonts", "--font-path", fontsDirPath())
	out, err := cmd.Output()
	if err != nil {
		slog.Warn("typst fonts failed", "err", err)
		// Fallback
		return []map[string]string{
			{"family": "Libertinus Serif", "category": "serif"},
			{"family": "Source Sans 3", "category": "sans-serif"},
			{"family": "JetBrains Mono", "category": "monospace"},
		}
	}

	// Known categories for common fonts
	categories := map[string]string{
		"Libertinus Serif":    "serif",
		"New Computer Modern": "serif",
		"Nimbus Roman":        "serif",
		"P052":                "serif",
		"C059":                "serif",
		"URW Bookman":         "serif",
		"DejaVu Serif":        "serif",
		"Source Sans 3":       "sans-serif",
		"Nimbus Sans":         "sans-serif",
		"DejaVu Sans":         "sans-serif",
		"URW Gothic":          "sans-serif",
		"Droid Sans Fallback": "sans-serif",
		"JetBrains Mono":      "monospace",
		"JetBrains Mono NL":   "monospace",
		"Noto Mono":           "monospace",
		"Noto Sans Mono":      "monospace",
		"DejaVu Sans Mono":    "monospace",
		"Nimbus Mono PS":      "monospace",
	}

	// Fonts to skip (symbols, emoji, etc.)
	skip := map[string]bool{
		"D050000L":            true,
		"FontAwesome":         true,
		"Standard Symbols PS": true,
		"Symbola":             true,
		"Noto Color Emoji":    true,
		"Z003":                true,
		"New Computer Modern Math": true,
	}

	var fonts []map[string]string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name := strings.TrimSpace(line)
		if name == "" || skip[name] {
			continue
		}
		cat, ok := categories[name]
		if !ok {
			cat = "other"
		}
		fonts = append(fonts, map[string]string{"family": name, "category": cat})
	}

	if len(fonts) == 0 {
		fonts = []map[string]string{
			{"family": "Libertinus Serif", "category": "serif"},
		}
	}

	return fonts
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

// fmtEm converts a pt value to em relative to base, formatted to 3 decimal
// places with trailing zeros stripped.  E.g. 16.67/10 → "1.667", 8/10 → "0.8".
func fmtEm(pt, base float64) string {
	if base <= 0 {
		base = 10
	}
	v := pt / base
	s := fmt.Sprintf("%.3f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// fmtEmStr formats an em value (already in em) to 3 decimal places.
func fmtEmStr(em float64) string {
	s := fmt.Sprintf("%.3f", em)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// getSizePtOrEm checks for "<key>_pt" (new format) then "<key>_em" (old format).
// Returns the em value string and true if found. pt values are converted via baseSizePt.
func getSizePtOrEm(m map[string]any, key string, baseSizePt float64) (string, bool) {
	// Try new _pt format first
	if v, ok := m[key+"_pt"].(float64); ok && v > 0 {
		return fmtEm(v, baseSizePt), true
	}
	// Fall back to old _em format
	if v, ok := m[key+"_em"].(float64); ok && v > 0 {
		return fmtEmStr(v), true
	}
	return "", false
}

// specToTypstConfig converts a spec JSON map into Typst config override code.
func specToTypstConfig(data map[string]any) string {
	var lines []string
	lines = append(lines, "\n// Project-specific config overrides (from spec)")
	lines = append(lines, "#let config = merge-config((")

	// Determine base_size_pt for em conversions (used by headings, elements, running heads)
	baseSizePt := 10.0
	if typo, ok := data["typography"].(map[string]any); ok {
		if v, ok := typo["base_size_pt"].(float64); ok && v > 0 {
			baseSizePt = v
		}
	}

	if page, ok := data["page"].(map[string]any); ok {
		// Emit page-paper for named Typst sizes (informational)
		trim, _ := page["trim"].(string)
		if trim != "" && trim != "custom" {
			lines = append(lines, fmt.Sprintf("  // Paper: %s", trim))
		}
		// Always emit explicit dimensions (template uses width/height)
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
		} else if v, ok := typo["leading_em"].(float64); ok && v > 0 {
			// Old format: leading_em × base_size = pt
			leadPt := v * baseSizePt
			lines = append(lines, fmt.Sprintf("  leading: %gpt,", leadPt))
		}
		if v, ok := typo["paragraph_indent_em"].(float64); ok {
			lines = append(lines, fmt.Sprintf("  paragraph-indent: %gem,", v))
		}
	}



	// Headings
	if hdg, ok := data["headings"].(map[string]any); ok {
		lines = append(lines, "  // Headings")
		if em, ok := getSizePtOrEm(hdg, "h1_size", baseSizePt); ok {
			lines = append(lines, fmt.Sprintf("  h1-size: %sem,", em))
		}
		if v, ok := hdg["h1_weight"]; ok {
			switch w := v.(type) {
			case string:
				if w != "" {
					lines = append(lines, fmt.Sprintf(`  h1-weight: "%s",`, w))
				}
			case float64:
				if w > 0 {
					lines = append(lines, fmt.Sprintf("  h1-weight: %g,", w))
				}
			}
		}
		if em, ok := getSizePtOrEm(hdg, "h2_size", baseSizePt); ok {
			lines = append(lines, fmt.Sprintf("  h2-size: %sem,", em))
		}
		if v, ok := hdg["h2_weight"]; ok {
			switch w := v.(type) {
			case string:
				if w != "" {
					lines = append(lines, fmt.Sprintf(`  h2-weight: "%s",`, w))
				}
			case float64:
				if w > 0 {
					lines = append(lines, fmt.Sprintf("  h2-weight: %g,", w))
				}
			}
		}
		if em, ok := getSizePtOrEm(hdg, "h3_size", baseSizePt); ok {
			lines = append(lines, fmt.Sprintf("  h3-size: %sem,", em))
		}
		if v, ok := hdg["h3_weight"]; ok {
			switch w := v.(type) {
			case string:
				if w != "" {
					lines = append(lines, fmt.Sprintf(`  h3-weight: "%s",`, w))
				}
			case float64:
				if w > 0 {
					lines = append(lines, fmt.Sprintf("  h3-weight: %g,", w))
				}
			}
		}
	}

	// Elements
	if elem, ok := data["elements"].(map[string]any); ok {
		lines = append(lines, "  // Elements")
		if v, ok := elem["section_break"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf(`  section-break: "%s",`, v))
		}
		if v, ok := elem["blockquote_style"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf(`  blockquote-style: "%s",`, v))
		}
		if em, ok := getSizePtOrEm(elem, "poem_size", baseSizePt); ok {
			lines = append(lines, fmt.Sprintf("  poem-size: %sem,", em))
		}
		if em, ok := getSizePtOrEm(elem, "code_block_size", baseSizePt); ok {
			lines = append(lines, fmt.Sprintf("  code-block-size: %sem,", em))
		}
		if em, ok := getSizePtOrEm(elem, "footnote_size", baseSizePt); ok {
			lines = append(lines, fmt.Sprintf("  footnote-size: %sem,", em))
		}
	}

	// Running heads — check both new "running_heads" and old "running_headers" keys
	rh, _ := data["running_heads"].(map[string]any)
	if rh == nil {
		rh, _ = data["running_headers"].(map[string]any)
	}
	if rh != nil {
		lines = append(lines, "  // Running heads")
		if v, ok := rh["enabled"].(bool); ok {
			lines = append(lines, fmt.Sprintf("  running-heads-enabled: %t,", v))
		}
		// Check head_size_pt (new), then font_size_em (old)
		if v, ok := rh["head_size_pt"].(float64); ok && v > 0 {
			lines = append(lines, fmt.Sprintf("  running-heads-size: %sem,", fmtEm(v, baseSizePt)))
		} else if v, ok := rh["font_size_em"].(float64); ok && v > 0 {
			lines = append(lines, fmt.Sprintf("  running-heads-size: %sem,", fmtEmStr(v)))
		}
		if v, ok := rh["verso"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf(`  running-heads-verso: "%s",`, v))
		}
		if v, ok := rh["recto"].(string); ok && v != "" {
			lines = append(lines, fmt.Sprintf(`  running-heads-recto: "%s",`, v))
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

// handleGenerateWordTemplate generates a styled .docx template from the book spec.
func (s *Server) handleGenerateWordTemplate(w http.ResponseWriter, r *http.Request) {
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

	var specData map[string]any
	if err := json.Unmarshal([]byte(spec.Data), &specData); err != nil {
		jsonErr(w, "invalid spec data", 500)
		return
	}
	if styles, ok := specData["custom_styles"].([]any); ok {
		seen := map[string]bool{}
		for _, item := range styles {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name, _ := m["word_style"].(string)
			if strings.TrimSpace(name) == "" {
				name, _ = m["name"].(string)
			}
			key := strings.ToLower(strings.TrimSpace(name))
			if key == "" {
				continue
			}
			if seen[key] {
				jsonErr(w, "duplicate custom style names are not allowed in Word templates; use distinct names such as metadata-p and metadata-c", 400)
				return
			}
			seen[key] = true
		}
	}

	// Run python script with spec JSON on stdin
	script := filepath.Join(typesettingRoot(), "scripts", "generate-word-template.py")
	cmd := exec.Command("python3", script)
	cmd.Stdin = strings.NewReader(spec.Data)
	var outBuf bytes.Buffer
	var errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		slog.Error("word template generation failed", "err", err, "stderr", errBuf.String())
		jsonErr(w, "template generation failed: "+errBuf.String(), 500)
		return
	}

	// Derive filename from project name
	project, _ := q.GetProject(r.Context(), pid)
	filename := sanitizeFilename(project.Name)
	if filename == "" {
		filename = "template"
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-template.docx"`, filename))
	w.Header().Set("Content-Length", strconv.Itoa(outBuf.Len()))
	w.Write(outBuf.Bytes())
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
	// Typst paper names and legacy "W x H" formats → width/height in inches
	trimPresets := map[string][2]float64{
		// Typst paper names
		"us-digest":    {5.5, 8.5},                         // 139.7×215.9mm
		"us-trade":     {6.0, 9.0},                         // 152.4×228.6mm
		"uk-book-a":    {111.0 / 25.4, 178.0 / 25.4},       // 4.3701×7.0079
		"uk-book-b":    {129.0 / 25.4, 198.0 / 25.4},       // 5.0787×7.7953
		"a5":           {148.0 / 25.4, 210.0 / 25.4},       // 5.8268×8.2677
		"a4":           {210.0 / 25.4, 297.0 / 25.4},       // 8.2677×11.6929
		"a6":           {105.0 / 25.4, 148.0 / 25.4},       // 4.1339×5.8268
		"jis-b5":       {182.0 / 25.4, 257.0 / 25.4},       // 7.1654×10.1181
		"jis-b6":       {128.0 / 25.4, 182.0 / 25.4},       // 5.0394×7.1654
		"us-letter":    {8.5, 11.0},
		"us-legal":     {8.5, 14.0},
		"us-executive": {7.25, 10.5},                        // 184.15×266.7mm
		"us-statement": {5.5, 8.5},                          // same as digest
		// Legacy "W x H" format aliases
		"5.5 x 8.5": {5.5, 8.5},
		"6 x 9":     {6.0, 9.0},
		"5 x 8":     {5.0, 8.0},
		"8.5 x 11":  {8.5, 11.0},
	}
	if dims, ok := trimPresets[trim]; ok {
		page["width_in"] = dims[0]
		page["height_in"] = dims[1]
	}
}

// handleUploadCover accepts a cover image file upload and stores it in book_specs.
func (s *Server) handleUploadCover(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		jsonErr(w, "file too large", 400)
		return
	}

	file, header, err := r.FormFile("cover")
	if err != nil {
		jsonErr(w, "cover file required", 400)
		return
	}
	defer file.Close()

	ct := header.Header.Get("Content-Type")
	if ct != "image/jpeg" && ct != "image/png" {
		jsonErr(w, "cover must be JPEG or PNG", 400)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		jsonErr(w, "read error", 500)
		return
	}

	// Ensure spec row exists
	q := dbgen.New(s.DB)
	_, err = q.GetBookSpec(r.Context(), pid)
	if err == sql.ErrNoRows {
		_, err = q.UpsertBookSpec(r.Context(), dbgen.UpsertBookSpecParams{
			ProjectID: pid,
			Data:      defaultSpecData(),
		})
		if err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
	}

	err = q.UpdateBookSpecCover(r.Context(), dbgen.UpdateBookSpecCoverParams{
		CoverData: data,
		CoverType: ct,
		ProjectID: pid,
	})
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]any{"ok": true, "size": len(data), "type": ct})
}

// handleGetCover serves the cover image for a project's book spec.
func (s *Server) handleGetCover(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	q := dbgen.New(s.DB)
	row, err := q.GetBookSpecCover(r.Context(), pid)
	if err != nil || row.CoverData == nil || len(row.CoverData) == 0 {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", row.CoverType)
	w.Header().Set("Content-Length", strconv.Itoa(len(row.CoverData)))
	w.Header().Set("Cache-Control", "max-age=3600")
	w.Write(row.CoverData)
}
