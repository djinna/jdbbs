package srv

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// TestBookMapCLI prints the map for BOOKMAP_DOCX (manual smoke; skipped otherwise).
func TestBookMapCLI(t *testing.T) {
	path := os.Getenv("BOOKMAP_DOCX")
	if path == "" {
		t.Skip("set BOOKMAP_DOCX")
	}
	if os.Getenv("BOOKMAP_PARAS") != "" {
		ps, err := readDOCXParagraphs(path)
		if err != nil {
			t.Fatal(err)
		}
		for i, p := range ps {
			if i > 40 {
				break
			}
			fmt.Printf("%3d %-10s br=%v %s\n", i, p.Style, p.PageBreakBefore, preview(p.Text))
		}
		return
	}
	m, err := bookMapFromDOCX(path, nil, os.Getenv("BOOKMAP_TITLE"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(m.Summary())
	b, _ := json.MarshalIndent(m, "", " ")
	fmt.Println(string(b))
}
