package srv

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHouseStylePublic(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	get := func(path string) (*http.Response, string) {
		res, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(res.Body)
		res.Body.Close()
		return res, string(b)
	}

	// Public, no auth. Seeds from the embedded JSON on first hit.
	res, body := get("/api/stylesheet/house")
	if res.StatusCode != 200 {
		t.Fatalf("house json: %d", res.StatusCode)
	}
	var all struct {
		Sections []hsSection `json:"sections"`
	}
	if err := json.Unmarshal([]byte(body), &all); err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, s := range all.Sections {
		total += len(s.Items)
	}
	if total < 50 {
		t.Fatalf("expected a seeded sheet, got %d items", total)
	}
	for _, leak := range []string{"Zoothesia", "Protocolized", "Protocol Institute", "Mayaford", "Langdon"} {
		if strings.Contains(body, leak) {
			t.Fatalf("public sheet leaks %q", leak)
		}
	}

	// kind filter drops the other kind but keeps 'both'.
	_, body = get("/api/stylesheet/house?kind=nonfiction")
	var nf struct {
		Sections []hsSection `json:"sections"`
	}
	_ = json.Unmarshal([]byte(body), &nf)
	nfTotal := 0
	for _, s := range nf.Sections {
		for _, it := range s.Items {
			nfTotal++
			if it.BookKind == "fiction" {
				t.Fatalf("nonfiction view contains a fiction item: %+v", it)
			}
		}
	}
	if nfTotal == 0 || nfTotal >= total {
		t.Fatalf("nonfiction filter: %d of %d", nfTotal, total)
	}

	res, body = get("/stylesheet/index.md?kind=fiction")
	if res.StatusCode != 200 || !strings.HasPrefix(body, "# House Editorial Stylesheet — Fiction") {
		t.Fatalf("md export: %d %q", res.StatusCode, body[:min(60, len(body))])
	}

	// Split: the internal tool moved; its old sub-paths redirect, the root does not.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	r2, _ := client.Get(ts.URL + "/stylesheet/authors")
	r2.Body.Close()
	if r2.StatusCode != 301 || r2.Header.Get("Location") != "/stylesheet-pi/authors" {
		t.Fatalf("legacy tool path: %d -> %q", r2.StatusCode, r2.Header.Get("Location"))
	}
	r3, _ := client.Get(ts.URL + "/stylesheet/")
	r3.Body.Close()
	if r3.StatusCode != 200 {
		t.Fatalf("public root: %d", r3.StatusCode)
	}
}
