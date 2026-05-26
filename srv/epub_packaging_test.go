package srv

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"time"
)

// TestStripMimetypeExtraField builds a zip with a mimetype entry whose LFH
// carries an Extended-Timestamp extra field (Go's archive/zip writer attaches
// one whenever FileHeader.Modified is non-zero — the same 9-byte extra
// pandoc's zip writer emits, which epubcheck flags as PKG-005).
//
// After stripMimetypeExtraField, we assert:
//   - mimetype is still the first entry
//   - mimetype's LFH extra-field-length is zero
//   - mimetype is stored uncompressed
//   - mimetype body is unchanged
//   - other entries are still readable with their bodies intact
func TestStripMimetypeExtraField(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// mimetype first, with a non-zero Modified to provoke the 9-byte extra.
	mh := &zip.FileHeader{
		Name:     "mimetype",
		Method:   zip.Store,
		Modified: time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC),
	}
	w, err := zw.CreateHeader(mh)
	if err != nil {
		t.Fatalf("create mimetype: %v", err)
	}
	w.Write([]byte("application/epub+zip"))

	// A couple of other entries, also with Modified set.
	for _, name := range []string{"META-INF/container.xml", "EPUB/nav.xhtml"} {
		fh := &zip.FileHeader{
			Name:     name,
			Method:   zip.Deflate,
			Modified: time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC),
		}
		w, err := zw.CreateHeader(fh)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		w.Write([]byte("<x/>"))
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	in := buf.Bytes()

	// Sanity: input has a non-zero extra field on the mimetype LFH.
	if got := lfhExtraLen(t, in, 0); got == 0 {
		t.Fatalf("test setup invalid: expected mimetype LFH to have a non-zero extra field, got 0")
	}

	out, err := stripMimetypeExtraField(in)
	if err != nil {
		t.Fatalf("stripMimetypeExtraField: %v", err)
	}

	// LFH extra-field-length on the first entry must be zero.
	if got := lfhExtraLen(t, out, 0); got != 0 {
		t.Fatalf("mimetype LFH extra-field-length = %d, want 0", got)
	}

	zr, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if len(zr.File) == 0 || zr.File[0].Name != "mimetype" {
		t.Fatalf("mimetype not first entry; got files=%v", names(zr.File))
	}
	if zr.File[0].Method != zip.Store {
		t.Fatalf("mimetype method = %d, want zip.Store", zr.File[0].Method)
	}
	if mb := readEntry(t, zr.File[0]); string(mb) != "application/epub+zip" {
		t.Fatalf("mimetype body = %q, want %q", mb, "application/epub+zip")
	}

	// Other entries still present and intact.
	wantOther := map[string]string{
		"META-INF/container.xml": "<x/>",
		"EPUB/nav.xhtml":         "<x/>",
	}
	for _, f := range zr.File[1:] {
		want, ok := wantOther[f.Name]
		if !ok {
			t.Fatalf("unexpected entry %s", f.Name)
		}
		if got := string(readEntry(t, f)); got != want {
			t.Fatalf("%s body = %q, want %q", f.Name, got, want)
		}
		delete(wantOther, f.Name)
	}
	if len(wantOther) != 0 {
		t.Fatalf("missing entries: %v", wantOther)
	}
}

// lfhExtraLen returns the extra-field-length of the local file header located
// at byteOffset within the zip. Assumes a valid LFH signature.
func lfhExtraLen(t *testing.T, data []byte, byteOffset int) int {
	t.Helper()
	if len(data) < byteOffset+30 {
		t.Fatalf("data too short for LFH at offset %d", byteOffset)
	}
	if sig := binary.LittleEndian.Uint32(data[byteOffset : byteOffset+4]); sig != 0x04034b50 {
		t.Fatalf("offset %d: not an LFH signature (got 0x%08x)", byteOffset, sig)
	}
	return int(binary.LittleEndian.Uint16(data[byteOffset+28 : byteOffset+30]))
}

func readEntry(t *testing.T, f *zip.File) []byte {
	t.Helper()
	rc, err := f.Open()
	if err != nil {
		t.Fatalf("open %s: %v", f.Name, err)
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read %s: %v", f.Name, err)
	}
	return b
}

func names(files []*zip.File) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Name
	}
	return out
}
