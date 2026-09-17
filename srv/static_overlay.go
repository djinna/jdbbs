package srv

import (
	"os"
	"path/filepath"
	"strings"
)

// ─── Static HTML overlay ───
//
// The portal pages under srv/static/ are compiled into the binary, which is
// what makes the deploy a single file. But it also means the admin doc editor
// (/admin/docs/) can't touch them without a rebuild. So: when the source tree
// is present next to the running service (the VM, or a local checkout), HTML
// pages are read from disk first and the embedded copy is only the fallback.
// At build time disk == embed, so nothing changes until someone saves an edit
// — and then it is live on the next request, exactly like a jdbbs-public page.
//
// Only *.html goes through the overlay. /static/theme.css, theme.js and the
// rest of the asset tree stay embedded.
//
// Override the tree with PRODCAL_REPO_DIR; set it to "-" to disable.

func repoDir() string {
	if d := os.Getenv("PRODCAL_REPO_DIR"); d != "" {
		if d == "-" {
			return ""
		}
		return d
	}
	const vm = "/home/exedev/prodcal"
	if st, err := os.Stat(filepath.Join(vm, "srv", "static")); err == nil && st.IsDir() {
		return vm
	}
	return ""
}

// readStatic returns the bytes for an embedded name like "static/factory.html",
// preferring the on-disk copy under repoDir()/srv/ when it exists.
func readStatic(name string) ([]byte, error) {
	if strings.HasSuffix(name, ".html") {
		if root := repoDir(); root != "" {
			if b, err := os.ReadFile(filepath.Join(root, "srv", filepath.FromSlash(name))); err == nil {
				return b, nil
			}
		}
	}
	return staticFS.ReadFile(name)
}
