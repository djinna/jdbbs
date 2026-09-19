// Package indexer drafts a back-of-book index for a manuscript (punch list
// 5.13, docs/reviews/INDEX-ADDON-DESIGN-2026-09-19.md): it reads the pandoc-
// produced Typst, asks a generative model for entries with anchor sentences,
// merges them across chapters, and anchors #index(...) markers into the Typst
// so the series template can resolve the locators at compile time.
package indexer

import (
	"strings"
	"unicode"
)

// Rune is one character of manuscript text with the byte offset in the Typst
// source it came from. Markup (function calls, brackets, labels, comments)
// produces no runes, so a match against the extracted text maps straight back
// to an insertion point in the source.
type Rune struct {
	R   rune
	Off int // byte offset of the rune in the source
}

// Text is the manuscript text extracted from a Typst source.
type Text struct {
	Runes []Rune
}

// String returns the plain text.
func (t Text) String() string {
	var b strings.Builder
	for _, r := range t.Runes {
		b.WriteRune(r.R)
	}
	return b.String()
}

// ExtractText walks pandoc-flavoured Typst markup and returns the visible
// text with source offsets. It understands what the pandoc writer and our Lua
// filter emit: `#name`, `#name(args)`, `#name[content]` (nested), `= heading`
// markers, `<label>` lines, `\` escapes, `//` comments, and `*strong*` /
// `_emph_` delimiters. Anything else is text. Heading markers are dropped but
// heading text is kept so chapter titles can be anchored too.
func ExtractText(src string) Text {
	var out []Rune
	depth := 0 // open content brackets we skipped the `[` of
	i := 0
	n := len(src)
	lineStart := true
	for i < n {
		c := src[i]
		switch {
		case c == '\\' && i+1 < n:
			// Escaped char: literal.
			r, size := decodeRune(src[i+1:])
			out = append(out, Rune{R: r, Off: i})
			i += 1 + size
			lineStart = false
			continue
		case c == '/' && i+1 < n && src[i+1] == '/':
			for i < n && src[i] != '\n' {
				i++
			}
			continue
		case c == '#' && i+1 < n && (isIdentStart(src[i+1])):
			if kw := keywordAt(src, i+1); kw != "" {
				// `#import`, `#show`, `#set`, `#let`, `#include`: a statement, not
				// text. Skip to the end of the line, and on to the line that
				// closes any parenthesis opened (multi-line book.with(...)).
				i = skipStatement(src, i)
				lineStart = true
				continue
			}
			i = skipCall(src, i)
			// A following `[` opens content we keep.
			for i < n && src[i] == '[' {
				depth++
				i++
			}
			lineStart = false
			continue
		case c == '[':
			// Bare content block in markup (rare; e.g. our `[sic]` case). Treat
			// the brackets as markup, keep the inside.
			depth++
			i++
			continue
		case c == ']' && depth > 0:
			depth--
			i++
			continue
		case lineStart && c == '=':
			// Heading marker: `= `, `== `, …
			j := i
			for j < n && src[j] == '=' {
				j++
			}
			if j < n && src[j] == ' ' {
				i = j + 1
				continue
			}
		case lineStart && c == '<':
			// Label line `<khlongs-subaks>` produced by pandoc under headings.
			j := strings.IndexByte(src[i:], '>')
			if j > 0 && isLabel(src[i+1:i+j]) {
				i += j + 1
				continue
			}
		case c == '*' || c == '_':
			// pandoc emits #strong/#emph, but tolerate markup delimiters when
			// glued to a word.
			if (i+1 < n && isWordByte(src[i+1])) || (i > 0 && isWordByte(src[i-1])) {
				i++
				continue
			}
		}
		r, size := decodeRune(src[i:])
		out = append(out, Rune{R: r, Off: i})
		lineStart = r == '\n'
		i += size
	}
	return Text{Runes: out}
}

// keywordAt returns the statement keyword starting at i, or "".
func keywordAt(src string, i int) string {
	for _, kw := range []string{"import", "include", "show", "set", "let"} {
		if strings.HasPrefix(src[i:], kw) && (i+len(kw) >= len(src) || !isIdentByte(src[i+len(kw)])) {
			return kw
		}
	}
	return ""
}

// skipStatement skips from `#kw` to the end of the line, continuing over
// following lines while parentheses stay open.
func skipStatement(src string, i int) int {
	n := len(src)
	depth := 0
	for i < n {
		c := src[i]
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
		} else if c == '\n' && depth <= 0 {
			return i + 1
		}
		i++
	}
	return i
}

// skipCall skips `#ident(.ident)*` plus any balanced `(...)` argument lists
// that follow, returning the offset after them (a following `[` is left for
// the caller). Strings inside parens may contain brackets.
func skipCall(src string, i int) int {
	n := len(src)
	i++ // '#'
	for i < n && (isIdentByte(src[i]) || src[i] == '.') {
		if src[i] == '.' && (i+1 >= n || !isIdentStart(src[i+1])) {
			break
		}
		i++
	}
	for i < n && src[i] == '(' {
		depth := 0
		inStr := false
		for i < n {
			c := src[i]
			if inStr {
				if c == '\\' {
					i += 2
					continue
				}
				if c == '"' {
					inStr = false
				}
			} else if c == '"' {
				inStr = true
			} else if c == '(' {
				depth++
			} else if c == ')' {
				depth--
				if depth == 0 {
					i++
					break
				}
			}
			i++
		}
	}
	return i
}

func decodeRune(s string) (rune, int) {
	for _, r := range s {
		return r, len(string(r))
	}
	return 0, 1
}

func isIdentStart(c byte) bool { return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isIdentByte(c byte) bool  { return isIdentStart(c) || c == '-' || (c >= '0' && c <= '9') }
func isWordByte(c byte) bool {
	return c >= 0x80 || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
func isLabel(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !(isIdentByte(s[i]) || s[i] == ':' || s[i] == '.') {
			return false
		}
	}
	return true
}

// Fold normalises text for matching: lower-case, smart quotes and dashes
// folded, everything but letters/digits dropped, runs of whitespace collapsed
// to one space. Anchors from the model and text from the source both go
// through it, so inline markup and quote style cannot break a match.
func Fold(s string) string {
	var b strings.Builder
	space := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if space && b.Len() > 0 {
				b.WriteByte(' ')
			}
			space = false
			b.WriteRune(unicode.ToLower(foldLetter(r)))
		default:
			space = true
		}
	}
	return b.String()
}

// foldLetter strips the common Latin diacritics so “Éminence” matches
// “Eminence” whichever way the model copied it.
func foldLetter(r rune) rune {
	if r < 0x80 {
		return r
	}
	if f, ok := diacritics[r]; ok {
		return f
	}
	return r
}

var diacritics = map[rune]rune{
	'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a', 'ā': 'a',
	'À': 'A', 'Á': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'Å': 'A',
	'ç': 'c', 'Ç': 'C', 'ć': 'c', 'č': 'c',
	'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e', 'ē': 'e', 'È': 'E', 'É': 'E', 'Ê': 'E', 'Ë': 'E',
	'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i', 'ī': 'i', 'Ì': 'I', 'Í': 'I', 'Î': 'I', 'Ï': 'I',
	'ñ': 'n', 'Ñ': 'N', 'ń': 'n',
	'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ø': 'o', 'ō': 'o', 'Ò': 'O', 'Ó': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O', 'Ø': 'O',
	'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u', 'ū': 'u', 'Ù': 'U', 'Ú': 'U', 'Û': 'U', 'Ü': 'U',
	'ý': 'y', 'ÿ': 'y', 'Ý': 'Y', 'š': 's', 'Š': 'S', 'ž': 'z', 'Ž': 'Z', 'ß': 's',
}

// FoldedText is a folded string with, for each folded byte, the index of the
// source Rune it came from — so a match in the folded string maps back to
// source offsets.
type FoldedText struct {
	S    string
	From []int // len(S)+1 entries: source rune index for each byte, last = len(runes)
}

// FoldText folds an extracted Text keeping the offset map.
func FoldText(t Text) FoldedText {
	var b strings.Builder
	var from []int
	space := false
	for i, ru := range t.Runes {
		r := ru.R
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if space && b.Len() > 0 {
				b.WriteByte(' ')
				from = append(from, i)
			}
			space = false
			s := string(unicode.ToLower(foldLetter(r)))
			b.WriteString(s)
			for k := 0; k < len(s); k++ {
				from = append(from, i)
			}
		} else {
			space = true
		}
	}
	from = append(from, len(t.Runes))
	return FoldedText{S: b.String(), From: from}
}
