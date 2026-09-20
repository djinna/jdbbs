package srv

import (
	"testing"
	"time"

	"srv.exe.dev/db/dbgen"
)

func TestPassIncludedFoldsAddOns(t *testing.T) {
	now := time.Date(2026, 9, 17, 14, 48, 3, 0, time.UTC)
	base := dbgen.Pass{BuildsIncluded: 3, FulfilledAt: now, ExpiresAt: now.AddDate(0, 6, 0)}
	if inc := passIncluded(base); inc.Builds != 3 || inc.Months != 6 || len(inc.AddOns) != 0 {
		t.Fatalf("base pass: %+v", inc)
	}
	bundle := dbgen.Pass{BuildsIncluded: 3, BuildsExtra: 3, FulfilledAt: now, ExpiresAt: now.AddDate(1, 0, 0)}
	inc := passIncluded(bundle)
	if inc.Builds != 6 || inc.Months != 12 {
		t.Fatalf("bundle: %+v", inc)
	}
	if len(inc.AddOns) != 2 || inc.AddOns[0] != "+3 builds" || inc.AddOns[1] != "+6 months storage" {
		t.Fatalf("add-ons: %v", inc.AddOns)
	}
	txt := passFulfillmentText(fulfillPassResult{Pass: bundle, Title: "T", ClientSlug: "c", Password: "p", PortalURL: "u"}, true)
	for _, want := range []string{"6 builds", "(12 months)", "Add-on: +3 builds", "Add-on: +6 months storage", "against your 6"} {
		if !contains(txt, want) {
			t.Errorf("email text missing %q", want)
		}
	}
	if contains(txt, "the Discord") {
		t.Error("stray 'the' before Discord")
	}
}
