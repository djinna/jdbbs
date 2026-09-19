package srv

import "strings"

// Rights on the copyright page (5.22). The transmittal stores a code under
// page_iv.rights; the print PDF, the Word template and the EPUB's dc:rights
// all derive their wording from here so the three never drift.
//
//	""/all_rights  Copyright © 2026 Name. All rights reserved.
//	cc_by …        Copyright © 2026 Name. This work is licensed under a Creative
//	               Commons Attribution 4.0 International License. To view a copy
//	               of this license, visit creativecommons.org/licenses/by/4.0/
//	cc0            Name has dedicated this work to the public domain (CC0 1.0) …
type rightsOption struct {
	Code  string
	Label string // transmittal select
	Name  string // licence name as CC prints it
	Path  string // creativecommons.org path, "" for all-rights
}

var rightsOptions = []rightsOption{
	{"all_rights", "All rights reserved", "", ""},
	{"cc_by", "CC BY — attribution", "Attribution 4.0 International", "licenses/by/4.0/"},
	{"cc_by_sa", "CC BY-SA — attribution, share-alike", "Attribution-ShareAlike 4.0 International", "licenses/by-sa/4.0/"},
	{"cc_by_nc", "CC BY-NC — attribution, non-commercial", "Attribution-NonCommercial 4.0 International", "licenses/by-nc/4.0/"},
	{"cc_by_nc_sa", "CC BY-NC-SA — attribution, non-commercial, share-alike", "Attribution-NonCommercial-ShareAlike 4.0 International", "licenses/by-nc-sa/4.0/"},
	{"cc_by_nd", "CC BY-ND — attribution, no derivatives", "Attribution-NoDerivatives 4.0 International", "licenses/by-nd/4.0/"},
	{"cc_by_nc_nd", "CC BY-NC-ND — attribution, non-commercial, no derivatives", "Attribution-NonCommercial-NoDerivatives 4.0 International", "licenses/by-nc-nd/4.0/"},
	{"cc0", "CC0 — public domain dedication", "CC0 1.0 Universal", "publicdomain/zero/1.0/"},
}

func rightsLookup(code string) rightsOption {
	code = strings.TrimSpace(code)
	for _, o := range rightsOptions {
		if o.Code == code {
			return o
		}
	}
	return rightsOptions[0]
}

// rightsLine is the copyright line as printed on p. iv: © statement plus the
// licence sentence. year/holder may be empty (a transmittal still being
// filled in); the line degrades the way the Typst template always has.
func rightsLine(code, year, holder string) string {
	o := rightsLookup(code)
	who := strings.TrimSpace(strings.Join([]string{strings.TrimSpace(year), strings.TrimSpace(holder)}, " "))
	cline := "Copyright © " + who + "."
	if who == "" {
		cline = ""
	}
	switch {
	case o.Code == "cc0":
		h := strings.TrimSpace(holder)
		if h == "" {
			h = "The author"
		}
		return h + " has dedicated this work to the public domain under the Creative Commons CC0 1.0 Universal dedication. To view a copy, visit https://creativecommons.org/" + o.Path
	case o.Path != "":
		lic := "This work is licensed under a Creative Commons " + o.Name + " License. To view a copy of this license, visit https://creativecommons.org/" + o.Path
		if cline == "" {
			return lic
		}
		return cline + " " + lic
	default:
		if cline == "" {
			return "All rights reserved."
		}
		return cline + " All rights reserved."
	}
}

// rightsShort is the one-line form for machine metadata (EPUB dc:rights).
func rightsShort(code, year, holder string) string {
	o := rightsLookup(code)
	who := strings.TrimSpace(strings.Join([]string{strings.TrimSpace(year), strings.TrimSpace(holder)}, " "))
	switch {
	case o.Code == "cc0":
		return "CC0 1.0 Universal (public domain dedication) — https://creativecommons.org/" + o.Path
	case o.Path != "":
		s := "Licensed under Creative Commons " + o.Name + " (" + ccAbbrev(o.Code) + ") — https://creativecommons.org/" + o.Path
		if who != "" {
			return "© " + who + ". " + s
		}
		return s
	default:
		if who != "" {
			return "© " + who + ". All rights reserved."
		}
		return "All rights reserved."
	}
}

func ccAbbrev(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimPrefix(code, "cc_"), "_", "-"))
}
