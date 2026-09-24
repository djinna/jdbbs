package srv

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The production build compiles with --root set to the job directory and
// imports the staged template by a root-relative path. Customer-authored
// custom-style snippets are executable Typst, so a build must not be able to
// read files outside its own job directory (2026-09-24 security review).
func TestTypstRootConfinedToJobDir(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst not installed")
	}
	dir := t.TempDir()
	if _, err := writeSpecialisedTemplate(dir, ""); err != nil {
		t.Fatal(err)
	}
	// A canary outside the job directory, in a path typst can resolve.
	outside := t.TempDir()
	canary := filepath.Join(outside, "canary.txt")
	if err := os.WriteFile(canary, []byte("SECRET-CANARY"), 0644); err != nil {
		t.Fatal(err)
	}
	compile := func(body string) (string, error) {
		src := "#import \"/templates/series-template.typ\": *\n" +
			"#show: book.with(title: \"T\", author: \"A\")\n" + body + "\n"
		typPath := filepath.Join(dir, "main.typ")
		if err := os.WriteFile(typPath, []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command("typst", "compile", "--root", dir, "--font-path", fontsDirPath(),
			typPath, filepath.Join(dir, "out.pdf")).CombinedOutput()
		return string(out), err
	}
	if out, err := compile("= Chapter\n#lorem(20)"); err != nil {
		t.Fatalf("baseline build under confined root failed: %v\n%s", err, out)
	}
	out, err := compile("#read(\"" + canary + "\")")
	if err == nil {
		t.Fatalf("typst read of %s outside the job root should fail", canary)
	}
	if !strings.Contains(out, "outside of project root") && !strings.Contains(out, "file not found") {
		t.Fatalf("unexpected typst error:\n%s", out)
	}
}
