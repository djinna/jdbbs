package srv

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

// Inspect integration for the book map (P4 step 2): the map rides in the
// preflight report JSON as one `book_map` finding (not counted in the
// totals) plus one `book_map_warning` finding per warning (counted, medium)
// and one `book_map_note` per note (counted, low), and is rendered into the
// HTML report and the factory page.

const bookMapFindingType = "book_map"
const bookMapWarningType = "book_map_warning"
const bookMapNoteType = "book_map_note"

// bookMapFindings converts a map into preflight findings.
func bookMapFindings(m *BookMap) []map[string]any {
	if m == nil {
		return nil
	}
	// The map's own fields, then the finding envelope.
	item := map[string]any{}
	if raw, err := json.Marshal(m); err == nil {
		_ = json.Unmarshal(raw, &item)
	}
	item["type"] = bookMapFindingType
	item["severity"] = "low"
	item["suggestion"] = "Every section head is Heading 1; the build decides front / body / back matter from the heading text. To change a placement, rename or move the heading."
	out := []map[string]any{item}
	for _, e := range m.Errors {
		out = append(out, map[string]any{
			"type":       bookMapWarningType,
			"severity":   "high",
			"text":       e,
			"location":   "book map",
			"suggestion": "Fix the file and upload again before building a final.",
		})
	}
	for _, w := range m.Warnings {
		out = append(out, map[string]any{
			"type":       bookMapWarningType,
			"severity":   "medium",
			"text":       w,
			"location":   "book map",
			"suggestion": "Fix it in your editor and upload again, or leave it: the build applies the map as shown.",
		})
	}
	for _, n := range m.Notes {
		out = append(out, map[string]any{
			"type":       bookMapNoteType,
			"severity":   "low",
			"text":       n,
			"location":   "book map",
			"suggestion": "Nothing to do unless this is not what you meant.",
		})
	}
	return out
}

// appendBookMapFindings adds the map to a report JSON array.
func appendBookMapFindings(report []byte, m *BookMap) ([]byte, error) {
	var findings []map[string]any
	if len(strings.TrimSpace(string(report))) > 0 {
		if err := json.Unmarshal(report, &findings); err != nil {
			return nil, err
		}
	}
	findings = append(findings, bookMapFindings(m)...)
	return json.Marshal(findings)
}

// parseStoredBookMap pulls the map back out of a stored report JSON.
func parseStoredBookMap(raw string) *BookMap {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var findings []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &findings); err != nil {
		return nil
	}
	for _, f := range findings {
		var probe struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(f, &probe) != nil || probe.Type != bookMapFindingType {
			continue
		}
		var m BookMap
		if json.Unmarshal(f, &m) != nil {
			return nil
		}
		return &m
	}
	return nil
}

// bookMapHTML renders the map as a report section (same shell classes as the
// python-generated overview so it inherits the report's styling).
func bookMapHTML(m *BookMap) string {
	if m == nil {
		return ""
	}
	e := html.EscapeString
	var b strings.Builder
	b.WriteString(`<style>.book-map-table{width:100%;border-collapse:collapse;margin:12px 0;font-size:0.95em}.book-map-table th{text-align:left;font-weight:600;padding:4px 8px 6px 0;border-bottom:1px solid currentColor}.book-map-table td{padding:3px 8px 3px 0;vertical-align:top}.book-map-table td:nth-child(2),.book-map-table td:nth-child(3){white-space:nowrap}.book-map-warnings,.book-map-notes{margin:10px 0 0;padding-left:20px}.book-map-notes{opacity:.8}</style>`)
	b.WriteString(`<section class="overview book-map" id="book-map"><div class="overview-label">Book map</div><h2>How the build reads this manuscript</h2>`)
	b.WriteString(`<p class="overview-copy">` + e(m.Summary()) + `</p>`)
	b.WriteString(`<table class="book-map-table"><thead><tr><th>Section</th><th>Placement</th><th>Folios</th></tr></thead><tbody>`)
	if m.Title != "" || m.Subtitle != "" {
		t := m.Title
		if m.Subtitle != "" {
			t = strings.TrimSpace(t + ": " + m.Subtitle)
		}
		b.WriteString(`<tr><td>` + e(t) + `</td><td>Title / Subtitle style</td><td>dropped — title page is generated</td></tr>`)
	}
	for _, u := range m.UntitledFront {
		if u.Drop {
			b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>dropped — generated from the transmittal</td></tr>`, e(u.Preview), e(plural(u.Paras, "paragraph"))))
			continue
		}
		b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%s (untitled, %s)</td><td>roman</td></tr>`,
			e(u.Preview), e(titleCase(u.Name)), e(plural(u.Paras, "paragraph"))))
	}
	for _, s := range m.Sections {
		place, folio := "", ""
		switch s.Kind {
		case "front":
			place, folio = "Front matter", "roman"
		case "body":
			place, folio = "Body", "arabic"
		case "back":
			place, folio = "Back matter", "arabic"
		case "title":
			place, folio = "Title heading", "dropped — title page is generated"
		case "toc":
			place, folio = "Contents", "dropped — contents are generated"
		}
		b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td></tr>`, e(s.Title), place, folio))
	}
	b.WriteString(`</tbody></table>`)
	for _, er := range m.Errors {
		b.WriteString(`<p class="book-map-error"><strong>` + e(er) + `</strong></p>`)
	}
	if len(m.Warnings) > 0 {
		b.WriteString(`<ul class="book-map-warnings">`)
		for _, w := range m.Warnings {
			b.WriteString(`<li>` + e(w) + `</li>`)
		}
		b.WriteString(`</ul>`)
	}
	if len(m.Notes) > 0 {
		b.WriteString(`<ul class="book-map-notes">`)
		for _, n := range m.Notes {
			b.WriteString(`<li>` + e(n) + `</li>`)
		}
		b.WriteString(`</ul>`)
	}
	b.WriteString(`<p class="overview-copy">Rule: every section head is Heading 1. Front matter is a heading from the closed list (Foreword, Preface, Acknowledgments, Prologue, Note on the Text, List of Figures, …) that comes before the first chapter; back matter is one from its list (Notes, Appendix, Bibliography, Glossary, Index, About the Author, Colophon, …) after the last chapter. Introduction is body and gets page 1. To move a section, rename or reorder the heading.</p>`)
	b.WriteString(`</section>`)
	return b.String()
}

// injectBookMapHTML places the section after the report's overview section.
func injectBookMapHTML(report []byte, m *BookMap) []byte {
	sec := bookMapHTML(m)
	if sec == "" {
		return report
	}
	s := string(report)
	i := strings.Index(s, "</section>")
	if i < 0 {
		return report
	}
	i += len("</section>")
	return []byte(s[:i] + "\n\n    " + sec + s[i:])
}
