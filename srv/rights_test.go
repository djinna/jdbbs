package srv

import (
	"strings"
	"testing"
)

func TestRightsLine(t *testing.T) {
	if got := rightsLine("", "2026", "A. Writer"); got != "Copyright © 2026 A. Writer. All rights reserved." {
		t.Fatalf("default: %q", got)
	}
	if got := rightsLine("all_rights", "", ""); got != "All rights reserved." {
		t.Fatalf("empty: %q", got)
	}
	got := rightsLine("cc_by_nc_sa", "2026", "A. Writer")
	for _, want := range []string{"Copyright © 2026 A. Writer.", "Attribution-NonCommercial-ShareAlike 4.0 International License", "creativecommons.org/licenses/by-nc-sa/4.0/"} {
		if !strings.Contains(got, want) {
			t.Fatalf("cc_by_nc_sa missing %q in %q", want, got)
		}
	}
	if got := rightsLine("cc0", "2026", "A. Writer"); !strings.HasPrefix(got, "A. Writer has dedicated") || !strings.Contains(got, "publicdomain/zero/1.0/") {
		t.Fatalf("cc0: %q", got)
	}
	if got := rightsLine("bogus", "2026", "X"); got != "Copyright © 2026 X. All rights reserved." {
		t.Fatalf("unknown code should fall back: %q", got)
	}
	if got := rightsShort("cc_by", "2026", "X"); got != "© 2026 X. Licensed under Creative Commons Attribution 4.0 International (BY) — https://creativecommons.org/licenses/by/4.0/" {
		t.Fatalf("short: %q", got)
	}
}
