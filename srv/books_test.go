package srv

import "testing"

func TestLiteralTypstMentionsLeavesEscapedAndQuoted(t *testing.T) {
	cases := map[string]string{
		`production\@protocol-institute.org`:                                           `production\@protocol-institute.org`,
		`#link("https://substack.com/@drewaustin")[https://substack.com/\@drewaustin]`: `#link("https://substack.com/@drewaustin")[https://substack.com/\@drewaustin]`,
		`#link("@handle")`:     `#link("@handle")`,
		`Follow @jenna on X`:   `Follow #sym.at#h(0em)jenna on X`,
		`@jenna at line start`: `#sym.at#h(0em)jenna at line start`,
	}
	for in, want := range cases {
		if got := literalTypstMentions(in); got != want {
			t.Errorf("literalTypstMentions(%q) = %q, want %q", in, got, want)
		}
	}
}
