package srv

import (
	"strings"
	"testing"
)

func TestAnnouncementMinimalMarkdown(t *testing.T) {
	body := "See the [session guide](https://jdbbs.exe.xyz/2026-pi-symposium/workshop) and https://example.com/x. **Plan on all four**, *please*.\n2 * 3 = 6"
	plain := announcementPlain(body)
	for _, want := range []string{"session guide (https://jdbbs.exe.xyz/2026-pi-symposium/workshop)", "Plan on all four, please.", "2 * 3 = 6"} {
		if !strings.Contains(plain, want) {
			t.Errorf("plain missing %q in %q", want, plain)
		}
	}
	if strings.ContainsAny(plain, "[]") {
		t.Errorf("plain kept brackets: %q", plain)
	}
	h := announcementBodyHTML(body)
	for _, want := range []string{
		`<a href="https://jdbbs.exe.xyz/2026-pi-symposium/workshop" style="color:` + emailAccent + `">session guide</a>`,
		`<a href="https://example.com/x" style="color:` + emailAccent + `">https://example.com/x</a>.`,
		`<strong>Plan on all four</strong>`, `<em>please</em>`, `2 * 3 = 6`, `<br>`,
	} {
		if !strings.Contains(h, want) {
			t.Errorf("html missing %q in %q", want, h)
		}
	}
	if strings.Count(h, "<a ") != 2 {
		t.Errorf("expected 2 links, got %d: %q", strings.Count(h, "<a "), h)
	}
	// Merge fields and angle brackets stay safe.
	if x := announcementBodyHTML("<b>hi</b> {{code}}"); !strings.Contains(x, "&lt;b&gt;hi&lt;/b&gt; {{code}}") {
		t.Errorf("escape broke: %q", x)
	}
}
