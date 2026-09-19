package srv

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Build-failure explanations for customers.
//
// A failed build used to say "We couldn't build this file. Run Inspect for
// clues" — true and useless (Jenna, 2026-09-19). The pipeline's stderr is
// specific; we translate the common shapes into plain language, point at the
// text near the failure so the author can find it in Word, and keep the first
// technical line under a "detail" fold. The full trace stays in the logs.
//
// The stored error_msg has up to three parts, newline-separated:
//
//	<plain explanation and what to do>
//	Near: “<text quoted from the offending line>”
//	Technical detail: <first meaningful stderr line>

var (
	unknownTypstVariableRE = regexp.MustCompile(`(?mi)unknown variable:\s*([A-Za-z0-9_-]+)`)
	typstLabelRE           = regexp.MustCompile("(?m)label `<([^>]*)>` does not exist")
	typstFontRE            = regexp.MustCompile(`(?mi)unknown font family:\s*(.+)$`)
	typstFileRE            = regexp.MustCompile(`(?mi)file not found \(searched at ([^)]+)\)`)
	typstLocRE             = regexp.MustCompile(`(?m)┌─\s+(\S+\.typ):(\d+):(\d+)`)
	typstErrorLineRE       = regexp.MustCompile(`(?m)^error:\s*(.+)$`)
)

// buildFailure is the analysed form of a pipeline error.
type buildFailure struct {
	Message string // plain language + action
	Near    string // quoted document text near the failure, if we can find it
	Detail  string // first technical line
}

// String renders the parts as stored in books.error_msg.
func (f buildFailure) String() string {
	parts := []string{f.Message}
	if f.Near != "" {
		parts = append(parts, "Near: “"+f.Near+"”")
	}
	if f.Detail != "" {
		parts = append(parts, "Technical detail: "+f.Detail)
	}
	return strings.Join(parts, "\n")
}

// diagnoseBuildFailure explains raw pipeline stderr. typPath, when non-empty,
// is the generated book.typ so the offending line can be quoted.
func diagnoseBuildFailure(raw, typPath string) buildFailure {
	f := buildFailure{Detail: firstErrorLine(raw)}
	lower := strings.ToLower(raw)

	// Where in the generated typst did it fail? Only the book's own file is
	// useful to quote; a trace into the template means the cause is upstream.
	if m := typstLocRE.FindStringSubmatch(raw); len(m) == 4 && strings.HasSuffix(m[1], "book.typ") && typPath != "" {
		if n, err := strconv.Atoi(m[2]); err == nil {
			f.Near = typstLineText(typPath, n)
		}
	}

	switch {
	case unknownTypstVariableRE.MatchString(raw):
		style := unknownTypstVariableRE.FindStringSubmatch(raw)[1]
		f.Message = fmt.Sprintf("Your file uses a Word style (%s) that isn't in your template. Inspect lists styles not in your transmittal — remove or remap it, or ask us to add it.", style)
	case typstLabelRE.MatchString(raw):
		f.Message = "An @ sign in your text was read as a cross-reference by the typesetter. This is our bug, not yours — email us and we'll fix the build; as a workaround, put the address in a plain Body paragraph."
	case strings.Contains(lower, "pagebreaks are not allowed inside of containers"):
		f.Message = "A chapter or section start landed inside a styled block (a Signature, Epigraph or quote paragraph that runs right up to the next heading). This is our bug — email us and we'll fix the build; as a workaround, add an empty Body paragraph before the heading."
	case typstFontRE.MatchString(raw):
		font := strings.TrimSpace(typstFontRE.FindStringSubmatch(raw)[1])
		f.Message = fmt.Sprintf("Your template asks for a font (%s) that isn't installed on our press. Pick another typeface on the transmittal, or email us the font licence.", font)
	case typstFileRE.MatchString(raw):
		f.Message = "An image in your file couldn't be read. Re-insert it in Word (Insert → Pictures, not a linked or pasted preview) and rebuild."
	case strings.Contains(lower, "pandoc"):
		f.Message = "We couldn't read this Word file. Re-save it as .docx from Word (File → Save As, Word Document) and try again. If it came from Pages or a converter, open and re-save it in Word or LibreOffice first."
	case strings.Contains(lower, "epub:"):
		f.Message = "The print PDF built but the EPUB didn't. Your PDF is still ready; email us and we'll sort the EPUB."
	case strings.Contains(lower, "typst"):
		f.Message = "The typesetter stopped on this file. Run Inspect on it and look at the High findings first; if the text quoted below is the problem, edit that paragraph and rebuild. Otherwise email us the message — failed builds aren't counted."
	default:
		f.Message = "We couldn't build this file. Failed builds aren't counted. Run Inspect for clues, and email us the message below if it keeps happening."
	}
	return f
}

// customerBuildError keeps the old single-string entry point (tests, callers
// without a typst file).
func customerBuildError(raw string) string {
	return diagnoseBuildFailure(raw, "").Message
}

// firstErrorLine returns the first "error: …" line from typst/pandoc stderr,
// or the first non-empty line.
func firstErrorLine(raw string) string {
	if m := typstErrorLineRE.FindStringSubmatch(raw); len(m) == 2 {
		return clip(strings.TrimSpace(m[1]), 300)
	}
	for _, l := range strings.Split(raw, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			return clip(t, 300)
		}
	}
	return ""
}

var typstMarkupRE = regexp.MustCompile(`#[a-zA-Z-]+(\([^)]*\))?\[?|[\[\]]|<[a-z0-9.-]+>|\\`)

// typstLineText quotes the human text around line n of a generated .typ, with
// typst markup stripped so the author can search for it in Word. Empty when
// the line carries no prose (a bare "]" or a template call).
func typstLineText(path string, n int) string {
	data, err := os.ReadFile(path)
	if err != nil || n < 1 {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	if n > len(lines) {
		return ""
	}
	// The line itself, plus the previous one when the line is short (typst
	// reports the token, the sentence often started a line earlier).
	text := strings.TrimSpace(typstMarkupRE.ReplaceAllString(lines[n-1], ""))
	if len(text) < 40 && n >= 2 {
		prev := strings.TrimSpace(typstMarkupRE.ReplaceAllString(lines[n-2], ""))
		if prev != "" {
			text = strings.TrimSpace(prev + " " + text)
		}
	}
	text = strings.Join(strings.Fields(text), " ")
	if len(text) < 4 {
		return ""
	}
	return clip(text, 160)
}
