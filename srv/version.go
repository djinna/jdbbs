package srv

import (
	"encoding/json"
	"net/http"
)

// BuildVersion is the git describe of the running binary, set by the Makefile
// (-ldflags -X). Shown in small type on the factory step strip so Jenna can
// tell at a glance whether an attendee's tab needs a hard refresh (0.33,
// workshop Tue 22 Sep). "dev" when built without the Makefile.
var BuildVersion = "dev"

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]string{"version": BuildVersion})
}
