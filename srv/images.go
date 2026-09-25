package srv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Print images. The EPUB keeps the author's originals (colour and all); the
// print PDF is a black-and-white interior unless the spec says otherwise, so
// colour images are converted here rather than left to the printer's RIP,
// which flattens them. The conversion is what a careful operator does first
// on most photographs: luminance grey, auto-level, a gentle S-curve. It is
// not the judgement call (channel mixing, a chart whose two lines became the
// same grey) — that is the tuned-conversion add-on.

// colourFractionThreshold: share of pixels with clear chroma (max(r,g,b) −
// min(r,g,b) above 10 %) for an image to count as colour. A mean-saturation
// test misses a white chart with two thin coloured lines; a pixel-fraction
// test catches it (≈4 % of pixels) while grey JPEGs with chroma noise stay at 0.
const colourFractionThreshold = 0.005

var rasterExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".tif": true, ".tiff": true, ".bmp": true, ".webp": true}

// colourArgs is the ImageMagick recipe shared with detect-edge-cases.py
// (_image_is_colour): keep the two in step.
func colourArgs(src string) []string {
	return []string{src, "-resize", "400x400>", "-colorspace", "sRGB", "-fx", "max(r,g,b)-min(r,g,b)", "-threshold", "10%", "-format", "%[fx:mean]", "info:"}
}

// imageColourFraction returns the share (0..1) of clearly coloured pixels.
func imageColourFraction(path string) (float64, error) {
	cmd, cancel := toolCommand(context.Background(), toolTimeoutImage, "convert", colourArgs(path+"[0]")...)
	defer cancel()
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("identify %s: %w", filepath.Base(path), err)
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

// greyForPrint rewrites a colour image in place as tuned greyscale.
func greyForPrint(path string) error {
	cmd, cancel := toolCommand(context.Background(), toolTimeoutImage, "convert", path, "-colorspace", "Gray", "-auto-level", "-sigmoidal-contrast", "3,50%", path)
	defer cancel()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("grey %s: %s", filepath.Base(path), strings.TrimSpace(string(out)))
	}
	return nil
}

// greyscaleMediaDir converts every colour raster image under dir for a
// black-and-white print interior. Returns (images seen, images converted).
// Images that are already grey are left untouched so an author's own tonal
// work survives. Errors on a single image are logged and skipped — a build
// should not fail because one PNG confused ImageMagick.
func greyscaleMediaDir(dir string, logf func(string, ...any)) (seen, converted int) {
	_ = filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() || !rasterExts[strings.ToLower(filepath.Ext(p))] {
			return nil
		}
		seen++
		frac, err := imageColourFraction(p)
		if err != nil {
			logf("images: %v", err)
			return nil
		}
		if frac < colourFractionThreshold {
			return nil
		}
		if err := greyForPrint(p); err != nil {
			logf("images: %v", err)
			return nil
		}
		converted++
		return nil
	})
	return
}

// specPrintColour reports whether the book is a colour interior — i.e. keep
// colour images in the print PDF. Off by default.
func specPrintColour(spec map[string]any) bool {
	im, _ := spec["images"].(map[string]any)
	v, _ := im["print_colour"].(bool)
	return v
}
