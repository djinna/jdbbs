package srv

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGreyscaleMediaDirConvertsOnlyColour(t *testing.T) {
	if _, err := exec.LookPath("convert"); err != nil {
		t.Skip("ImageMagick not installed")
	}
	dir := t.TempDir()
	mk := func(name string, args ...string) {
		a := append(args, filepath.Join(dir, name))
		if out, err := exec.Command("convert", a...).CombinedOutput(); err != nil {
			t.Fatalf("convert %s: %v\n%s", name, err, out)
		}
	}
	mk("colour.jpg", "-size", "120x80", "gradient:red-blue")
	mk("grey.jpg", "-size", "120x80", "gradient:black-white")
	mk("lineart.png", "-size", "120x80", "xc:none", "-fill", "black", "-draw", "line 0,0 119,79")
	// a chart: white ground, two thin coloured lines (≈3 % of pixels) — the case a mean-saturation test misses
	mk("chart.png", "-size", "600x400", "xc:white", "-stroke", "#e63946", "-strokewidth", "6", "-draw", "line 20,380 580,40", "-stroke", "#2a9d8f", "-draw", "line 20,300 580,120")
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644)

	seen, conv := greyscaleMediaDir(dir, t.Logf)
	if seen != 4 || conv != 2 {
		t.Fatalf("seen=%d converted=%d, want 4 seen, 2 converted (colour.jpg, chart.png)", seen, conv)
	}
	for _, n := range []string{"colour.jpg", "chart.png"} {
		frac, err := imageColourFraction(filepath.Join(dir, n))
		if err != nil || frac > 0 {
			t.Errorf("%s still has colour after conversion: frac=%.4f err=%v", n, frac, err)
		}
	}
	// alpha survives on line art
	out, _ := exec.Command("identify", "-format", "%[channels]", filepath.Join(dir, "lineart.png")).Output()
	if string(out) == "" || string(out) == "gray" || string(out) == "srgb" {
		t.Errorf("lineart.png lost its alpha channel: %q", out)
	}
}

func TestSpecPrintColour(t *testing.T) {
	if specPrintColour(map[string]any{}) {
		t.Fatal("default must be a black-and-white interior")
	}
	if !specPrintColour(map[string]any{"images": map[string]any{"print_colour": true}}) {
		t.Fatal("print_colour=true not honoured")
	}
}
