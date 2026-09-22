package srv

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Custom styles: a style the author declares on the transmittal (Custom
// Styles) and marks in the manuscript ([[name]] or a Word style of that name).
// Design: docs/reviews/CUSTOM-STYLES-MARKERS-2026-09-22.md (option C).
//
// A declared paragraph style is *based on* one of the factory paragraph
// styles and may add one structural delta: an indent level (Verse, Block
// Quote, Code Block) or a space-before multiple (Section Break). A declared
// character style is based on italic, small caps, code or nothing. The factory
// owns typography; the author owns structure.
//
// Precedence when generating the Typst definition:
//  1. a hand-written `typst` snippet (differs from the preset's default) wins;
//  2. else a designed preset (tweet-block, metadata-*, ascii-block);
//  3. else parent + deltas (Normal with no deltas = a plain body paragraph).

// customStyleParents are the factory paragraph styles a custom style can be
// based on, in transmittal order. Names match the generated Word template.
var customStyleParents = []string{
	"Normal", "First Paragraph", "Heading 1", "Heading 2", "Heading 3",
	"Block Quote", "Epigraph", "Verse", "Code Block", "Section Break",
	"Copyright", "Signature", "Glossary Entry",
}

// customStyleCharParents are the bases a character custom style may use.
var customStyleCharParents = []string{"none", "italic", "small caps", "code"}

// customStyleIndentParents may carry an indent level (0 / 1 / 2 levels of 1.5 em).
var customStyleIndentParents = map[string]bool{"Verse": true, "Block Quote": true, "Code Block": true}

// customStyleSpaceParents may carry a space-before multiple (1 / 3 / 6 stanza gaps).
var customStyleSpaceParents = map[string]bool{"Section Break": true}

// designedPresets render from a studio-written preset rather than a parent.
var designedPresets = map[string]bool{
	"tweet-block": true, "metadata-paragraph": true, "metadata-inline": true, "ascii-block": true,
}

// typstReservedIdents are Typst keywords and literals that cannot name a
// function. Fotis (book 57, 2026-09-22) declared a style called "break" and the
// generated `#let break(content)` stopped the typesetter with an error that
// pointed him at his manuscript — the string existed only in our file.
var typstReservedIdents = map[string]bool{
	"let": true, "set": true, "show": true, "if": true, "else": true, "for": true, "in": true,
	"while": true, "break": true, "continue": true, "return": true, "import": true,
	"include": true, "as": true, "not": true, "and": true, "or": true, "none": true,
	"auto": true, "true": true, "false": true, "context": true,
}

func normalizeStyleKey(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}

// normalizeCustomStyleParent maps a based_on value to its canonical name, or
// the default (Normal / none) if it is empty or unknown.
func normalizeCustomStyleParent(styleType, basedOn string) string {
	key := normalizeStyleKey(basedOn)
	if styleType == "character" {
		switch key {
		case "italic", "ital", "emph", "italics":
			return "italic"
		case "small caps", "smallcaps", "sc", "small-caps":
			return "small caps"
		case "code", "mono", "monospace":
			return "code"
		}
		return "none"
	}
	for _, p := range customStyleParents {
		if normalizeStyleKey(p) == key {
			return p
		}
	}
	switch key {
	case "body", "body text", "paragraph", "normal paragraph":
		return "Normal"
	case "first para", "first-paragraph", "no indent":
		return "First Paragraph"
	case "blockquote", "block-quote", "quote":
		return "Block Quote"
	case "poem", "poetry":
		return "Verse"
	case "code", "codeblock", "code-block":
		return "Code Block"
	case "break", "section-break", "sectionbreak", "scene break":
		return "Section Break"
	case "glossary", "glossary-entry":
		return "Glossary Entry"
	}
	return "Normal"
}

func customStyleAllowsIndent(parent string) bool      { return customStyleIndentParents[parent] }
func customStyleAllowsSpaceBefore(parent string) bool { return customStyleSpaceParents[parent] }

// normalizeCustomStyle returns the spec form of one custom_styles[] entry.
// prev is the same-named entry from the spec being replaced (nil if none), so
// a preset / snippet Jenna wrote in the admin survives a transmittal re-pull.
func normalizeCustomStyle(m map[string]any, prev map[string]any) (map[string]any, bool) {
	str := func(src map[string]any, k string) string {
		if src == nil {
			return ""
		}
		v, _ := src[k].(string)
		return strings.TrimSpace(v)
	}
	num := func(src map[string]any, k string) (int, bool) {
		if src == nil {
			return 0, false
		}
		switch v := src[k].(type) {
		case float64:
			return int(v), true
		case int:
			return v, true
		case string:
			var n int
			if _, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &n); err == nil {
				return n, true
			}
		}
		return 0, false
	}

	name := str(m, "name")
	if name == "" {
		return nil, false
	}
	wordStyle := str(m, "word_style")
	if wordStyle == "" {
		wordStyle = name
	}
	styleType := str(m, "type")
	if styleType != "character" {
		styleType = "paragraph"
	}
	desc := str(m, "description")

	basedOn := str(m, "based_on")
	if basedOn == "" && prev != nil {
		basedOn = str(prev, "based_on")
	}
	parent := normalizeCustomStyleParent(styleType, basedOn)

	indent := 0
	if n, ok := num(m, "indent"); ok {
		indent = n
	} else if n, ok := num(prev, "indent"); ok {
		indent = n
	}
	if !customStyleAllowsIndent(parent) || indent < 0 {
		indent = 0
	}
	if indent > 2 {
		indent = 2
	}
	space := 1
	if n, ok := num(m, "space_before"); ok {
		space = n
	} else if n, ok := num(prev, "space_before"); ok {
		space = n
	}
	if !customStyleAllowsSpaceBefore(parent) || (space != 3 && space != 6) {
		space = 1
	}

	preset := str(m, "preset")
	typstCode := str(m, "typst")
	if preset == "" && typstCode == "" && prev != nil {
		// Transmittal rows carry neither; keep what the admin spec had.
		preset = str(prev, "preset")
		typstCode = str(prev, "typst")
	}
	if preset == "" {
		preset = defaultCustomStylePreset(name, styleType)
	}
	if typstCode == "" {
		typstCode = defaultCustomStyleTypst(name, styleType, preset)
	}

	return map[string]any{
		"name":         name,
		"word_style":   wordStyle,
		"type":         styleType,
		"description":  desc,
		"based_on":     parent,
		"indent":       indent,
		"space_before": space,
		"preset":       preset,
		"typst":        typstCode,
	}, true
}

// normalizeCustomStyles maps a list of rows (transmittal or admin spec) onto
// the spec form, carrying preset/snippet forward from prev by name.
func normalizeCustomStyles(rows []any, prev []any) []any {
	prevByName := map[string]map[string]any{}
	for _, p := range prev {
		if pm, ok := p.(map[string]any); ok {
			if n, _ := pm["name"].(string); strings.TrimSpace(n) != "" {
				prevByName[normalizeStyleKey(n)] = pm
			}
		}
	}
	out := []any{} // never nil: a JSON null here breaks the template generator
	for _, item := range rows {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		n, _ := m["name"].(string)
		if ns, ok := normalizeCustomStyle(m, prevByName[normalizeStyleKey(n)]); ok {
			out = append(out, ns)
		}
	}
	return out
}

// customStyleFields is the typed view of one spec entry.
type customStyleFields struct {
	Name, WordStyle, Type, Description, BasedOn, Preset, Typst string
	Indent, SpaceBefore                                        int
}

func customStyleFrom(m map[string]any) customStyleFields {
	ns, ok := normalizeCustomStyle(m, nil)
	if !ok {
		return customStyleFields{}
	}
	get := func(k string) string { v, _ := ns[k].(string); return v }
	return customStyleFields{
		Name: get("name"), WordStyle: get("word_style"), Type: get("type"), Description: get("description"),
		BasedOn: get("based_on"), Preset: get("preset"), Typst: get("typst"),
		Indent: ns["indent"].(int), SpaceBefore: ns["space_before"].(int),
	}
}

// customStyleHandWritten reports whether the snippet differs from the default
// for its preset — i.e. someone edited it in the admin.
func customStyleHandWritten(cs customStyleFields) bool {
	if cs.Typst == "" {
		return false
	}
	return strings.TrimSpace(cs.Typst) != strings.TrimSpace(defaultCustomStyleTypst(cs.Name, cs.Type, cs.Preset))
}

// customStyleSource says which rule produced the definition: "snippet",
// "preset" or "parent".
func customStyleSource(cs customStyleFields) string {
	switch {
	case customStyleHandWritten(cs):
		return "snippet"
	case designedPresets[cs.Preset]:
		return "preset"
	default:
		return "parent"
	}
}

// customStyleTypstDef returns the #let definition for one custom style.
func customStyleTypstDef(m map[string]any) string {
	cs := customStyleFrom(m)
	if cs.Name == "" {
		return ""
	}
	switch customStyleSource(cs) {
	case "snippet":
		return cs.Typst
	case "preset":
		return defaultCustomStyleTypst(cs.Name, cs.Type, cs.Preset)
	}
	ident := typstStyleIdent(cs.Name)
	if cs.Type == "character" {
		switch cs.BasedOn {
		case "italic":
			return fmt.Sprintf("#let %s(content) = emph(content)", ident)
		case "small caps":
			return fmt.Sprintf("#let %s(content) = sc(content)", ident)
		case "code":
			return fmt.Sprintf("#let %s(content) = text(font: config.code-font, size: 0.9em, content)", ident)
		}
		return fmt.Sprintf("#let %s(content) = content", ident)
	}
	switch cs.BasedOn {
	case "First Paragraph":
		return fmt.Sprintf("#let %s(content) = first-para(content)", ident)
	case "Heading 1", "Heading 2", "Heading 3":
		return fmt.Sprintf("#let %s(content) = heading(level: %s, content)", ident, cs.BasedOn[len(cs.BasedOn)-1:])
	case "Block Quote":
		return fmt.Sprintf("#let %s(content) = blockquote(indent: %d, content)", ident, cs.Indent)
	case "Epigraph":
		return fmt.Sprintf("#let %s(content) = epigraph(content)", ident)
	case "Verse":
		return fmt.Sprintf("#let %s(content) = poem(indent: %d, content)", ident, cs.Indent)
	case "Code Block":
		return fmt.Sprintf("#let %s(content) = code-block(indent: %d, content)", ident, cs.Indent)
	case "Section Break":
		return fmt.Sprintf("#let %s(content) = section-break-gap(gap: %d)", ident, cs.SpaceBefore)
	case "Copyright":
		return fmt.Sprintf("#let %s(content) = copyright-page(content)", ident)
	case "Signature":
		return fmt.Sprintf("#let %s(content) = signature(content)", ident)
	case "Glossary Entry":
		return fmt.Sprintf("#let %s(content) = glossary-entry(content)", ident)
	}
	// Normal: a plain body paragraph, exactly as before this change.
	return fmt.Sprintf("#let %s(content) = {\n  content\n}", ident)
}

// customStyleCoalesces reports whether consecutive paragraphs of this style
// are one block in the book (a poem's lines, a signature's lines), so the Lua
// filter merges them before wrapping.
func customStyleCoalesces(m map[string]any) bool {
	cs := customStyleFrom(m)
	if customStyleSource(cs) != "parent" || cs.Type == "character" {
		return false
	}
	return cs.BasedOn == "Verse" || cs.BasedOn == "Signature"
}

// customStyleResolvedForm is the one-line human description shown on the
// transmittal and in the admin: "verse2 → Verse, indent one level".
func customStyleResolvedForm(m map[string]any) string {
	cs := customStyleFrom(m)
	if cs.Name == "" {
		return ""
	}
	switch customStyleSource(cs) {
	case "snippet":
		return cs.Name + " → designed by the studio (custom typesetting)"
	case "preset":
		return cs.Name + " → designed by the studio (preset: " + cs.Preset + ")"
	}
	if cs.Type == "character" {
		if cs.BasedOn == "none" {
			return cs.Name + " → plain text (no change)"
		}
		return cs.Name + " → " + cs.BasedOn
	}
	out := cs.Name + " → " + cs.BasedOn
	switch cs.Indent {
	case 1:
		out += ", indent one level"
	case 2:
		out += ", indent two levels"
	}
	switch cs.SpaceBefore {
	case 3:
		out += ", 3× the stanza gap before"
	case 6:
		out += ", 6× the stanza gap before"
	}
	return out
}

// normalizeSpecCustomStyles normalises custom_styles[] inside a spec JSON
// saved from the admin (based_on / indent / space_before validated, defaults
// filled). Anything unparseable is returned unchanged.
func normalizeSpecCustomStyles(specJSON string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(specJSON), &data); err != nil {
		return specJSON
	}
	rows, ok := data["custom_styles"].([]any)
	if !ok {
		return specJSON
	}
	data["custom_styles"] = normalizeCustomStyles(rows, nil)
	out, err := json.Marshal(data)
	if err != nil {
		return specJSON
	}
	return string(out)
}
