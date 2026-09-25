package srv

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

// A .docx is a zip, and the converters (pandoc, python-docx, our own XML
// scan) all inflate it in full. These budgets keep a crafted archive from
// turning a 50 MB upload into gigabytes of XML in memory or thousands of
// extracted media files, and keep entry names inside the extraction dir.
// They are checked by actually decompressing every entry, not by trusting
// the declared sizes in the central directory.
const (
	docxMaxEntries        = 10000
	docxMaxTotalInflated  = 1 << 30   // 1 GiB across the archive
	docxMaxEntryInflated  = 512 << 20 // any single part
	docxMaxRatioEntrySize = 8 << 20   // ratio is only judged on parts inflating past this
	docxMaxRatio          = 200       // XML compresses ~10–40:1 in real documents
)

var errNotDOCX = errors.New("not a Word .docx (expected a zip with word/document.xml)")

// checkDOCXBudget inflates the archive and returns a descriptive error if it
// is not a plausible Word document or exceeds the budgets above.
func checkDOCXBudget(data []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return errNotDOCX
	}
	if len(zr.File) > docxMaxEntries {
		return fmt.Errorf("docx has %d parts; the limit is %d", len(zr.File), docxMaxEntries)
	}
	var total uint64
	hasDocument := false
	for _, f := range zr.File {
		if unsafeZipName(f.Name) {
			return fmt.Errorf("docx part %q has an unsafe path", f.Name)
		}
		name := strings.ReplaceAll(f.Name, `\`, "/")
		if name == "word/document.xml" {
			hasDocument = true
		}
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("docx part %q is unreadable: %v", f.Name, err)
		}
		// Read one byte past the per-entry cap so an oversize part is detected
		// without inflating all of it.
		n, err := io.Copy(io.Discard, io.LimitReader(rc, int64(docxMaxEntryInflated)+1))
		rc.Close()
		if err != nil {
			return fmt.Errorf("docx part %q is corrupt: %v", f.Name, err)
		}
		if n > int64(docxMaxEntryInflated) {
			return fmt.Errorf("docx part %q inflates past %d MB", f.Name, docxMaxEntryInflated>>20)
		}
		total += uint64(n)
		if total > docxMaxTotalInflated {
			return fmt.Errorf("docx inflates past %d MB in total", docxMaxTotalInflated>>20)
		}
		if n > int64(docxMaxRatioEntrySize) && f.CompressedSize64 > 0 && uint64(n)/f.CompressedSize64 > docxMaxRatio {
			return fmt.Errorf("docx part %q has an implausible compression ratio", f.Name)
		}
	}
	if !hasDocument {
		return errNotDOCX
	}
	return nil
}

// unsafeZipName reports names that would escape an extraction directory:
// absolute paths or any ".." segment (either slash convention).
func unsafeZipName(name string) bool {
	name = strings.ReplaceAll(name, `\`, "/")
	if name == "" || strings.HasPrefix(name, "/") {
		return true
	}
	for _, seg := range strings.Split(name, "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}
