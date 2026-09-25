package srv

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestToolCommandKillsProcessGroup: when a converter overruns its deadline
// the whole process group dies, including a grandchild that would otherwise
// outlive the parent and keep the build's temp dir (and CPU) busy.
func TestToolCommandKillsProcessGroup(t *testing.T) {
	cmd, cancel := toolCommand(context.Background(), 300*time.Millisecond, "sh", "-c", "sleep 30 & sleep 30")
	defer cancel()
	start := time.Now()
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected the command to be killed")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatalf("kill took %v", time.Since(start))
	}
	if !strings.Contains(toolErr(cmd, err).Error(), "timed out") {
		t.Errorf("toolErr should describe the timeout, got %v", toolErr(cmd, err))
	}
	// The background `sleep 30` was in the same group; give the kernel a
	// moment, then the group must be empty.
	pgid := cmd.Process.Pid
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(-pgid, 0); errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("process group %d still has members after kill", pgid)
}

func zipDOCX(t *testing.T, mutate func(zw *zip.Writer)) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range map[string]string{
		"[Content_Types].xml": "<Types/>",
		"word/document.xml":   "<w:document><w:body><w:p><w:r><w:t>Hello</w:t></w:r></w:p></w:body></w:document>",
	} {
		w, _ := zw.Create(name)
		w.Write([]byte(body))
	}
	if mutate != nil {
		mutate(zw)
	}
	zw.Close()
	return buf.Bytes()
}

// TestCheckDOCXBudget: uploads must be real Word archives that inflate to a
// sane size, with no path-escaping part names.
func TestCheckDOCXBudget(t *testing.T) {
	if err := checkDOCXBudget(zipDOCX(t, nil)); err != nil {
		t.Fatalf("minimal docx rejected: %v", err)
	}
	if err := checkDOCXBudget([]byte("fake-docx")); !errors.Is(err, errNotDOCX) {
		t.Errorf("non-zip: got %v", err)
	}
	var plain bytes.Buffer
	zw := zip.NewWriter(&plain)
	w, _ := zw.Create("readme.txt")
	w.Write([]byte("hi"))
	zw.Close()
	if err := checkDOCXBudget(plain.Bytes()); !errors.Is(err, errNotDOCX) {
		t.Errorf("zip without word/document.xml: got %v", err)
	}
	for _, bad := range []string{"../evil.xml", "word/../../evil", "/abs/path", `word\..\..\evil`} {
		data := zipDOCX(t, func(zw *zip.Writer) { w, _ := zw.Create(bad); w.Write([]byte("x")) })
		if err := checkDOCXBudget(data); err == nil || !strings.Contains(err.Error(), "unsafe path") {
			t.Errorf("%q: got %v", bad, err)
		}
	}
	// A 16 MB run of zeros deflates to a few KB: a >1000:1 part past the
	// ratio threshold — the classic bomb shape.
	bomb := zipDOCX(t, func(zw *zip.Writer) {
		w, _ := zw.Create("word/media/image1.png")
		w.Write(make([]byte, 16<<20))
	})
	if len(bomb) > 1<<20 {
		t.Fatalf("test bomb did not compress: %d bytes", len(bomb))
	}
	if err := checkDOCXBudget(bomb); err == nil || !strings.Contains(err.Error(), "compression ratio") {
		t.Errorf("bomb: got %v", err)
	}
	// Too many parts.
	many := zipDOCX(t, func(zw *zip.Writer) {
		for i := 0; i <= docxMaxEntries; i++ {
			zw.Create(fmt.Sprintf("word/media/i%d.png", i))
		}
	})
	if err := checkDOCXBudget(many); err == nil || !strings.Contains(err.Error(), "parts") {
		t.Errorf("too many entries: got %v", err)
	}
	// Realistic size at realistic ratio stays fine: 12 MB of varied bytes.
	big := make([]byte, 12<<20)
	for i := range big {
		big[i] = byte(i*7 + i>>8)
	}
	ok := zipDOCX(t, func(zw *zip.Writer) { w, _ := zw.Create("word/media/image1.bin"); w.Write(big) })
	if err := checkDOCXBudget(ok); err != nil {
		t.Errorf("realistic image rejected: %v", err)
	}
}

// minimalDOCX is the smallest archive checkDOCXBudget accepts; tests that
// only need the upload to land (and never convert it) use it in place of
// arbitrary bytes.
func minimalDOCX(t *testing.T) []byte { return zipDOCX(t, nil) }

// TestUploadRejectsNonDOCX: the budget check runs at upload, so a bogus
// file is refused with a clear 400 rather than failing later in a build.
func TestUploadRejectsNonDOCX(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	pass, _, _, _ := grantedPass(t, s, ts, "Bogus Author", "Bogus Book")
	pid := itoa(pass.ProjectID)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("title", "Bogus")
	_ = mw.WriteField("author", "Nobody")
	_ = mw.WriteField("project_id", pid)
	fw, _ := mw.CreateFormFile("file", "bogus.docx")
	_, _ = fw.Write([]byte("not-a-real-docx"))
	_ = mw.Close()
	req, _ := http.NewRequest("POST", ts.URL+"/api/books/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-ExeDev-UserID", "test-admin")
	req.Header.Set("X-ExeDev-Email", "owner@example.test")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 400 || !strings.Contains(string(body), "not a Word .docx") {
		t.Fatalf("got %d %s", resp.StatusCode, body)
	}
}

// TestRealBuildThroughToolCommand: a genuine .docx goes upload → convert →
// done through pandoc and typst under toolCommand's process-group deadline,
// proving the wrapper does not disturb a healthy build. Skipped without the
// pipeline (VM parity job).
func TestRealBuildThroughToolCommand(t *testing.T) {
	for _, bin := range []string{"pandoc", "typst"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not installed", bin)
		}
	}
	s, ts, cleanup := testServer(t)
	defer cleanup()
	pass, _, _, _ := grantedPass(t, s, ts, "Real Author", "Real Book")
	pid := itoa(pass.ProjectID)
	bid := uploadDocx(t, ts, pid, nil, true, indexTestDOCX(t))
	resp := apiRequestAdmin(t, ts, "POST", "/api/books/"+bid+"/convert", map[string]string{"format": "pdf"})
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("convert: %d %s", resp.StatusCode, b)
	}
	resp.Body.Close()
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		var st map[string]any
		r := apiRequestAdmin(t, ts, "GET", "/api/books/"+bid, nil)
		decodeJSON(t, r, &st)
		switch st["status"] {
		case "ready":
			return
		case "error":
			t.Fatalf("build failed: %v", st)
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("build did not finish in time")
}
