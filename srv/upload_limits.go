package srv

import (
	"errors"
	"net/http"
)

// parseUploadForm bounds the entire request before multipart parsing can spool
// files to disk. The argument to ParseMultipartForm alone is only a RAM budget.
// Callers must also check the file size and defer MultipartForm.RemoveAll.
func parseUploadForm(w http.ResponseWriter, r *http.Request, fileLimit int64) bool {
	const formOverhead = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, fileLimit+formOverhead)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		status := http.StatusBadRequest
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		jsonErr(w, "file too large or bad form", status)
		return false
	}
	return true
}
