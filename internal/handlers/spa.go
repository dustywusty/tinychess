package handlers

import (
	"io/fs"
	"net/http"
	"strings"
)

// SpaHandler serves the SPA bundle from the given filesystem. For any path
// that resolves to a regular file (e.g. /assets/index-XXXX.js), the file is
// served directly. Anything else falls back to /index.html so the SPA's
// client-side router can take over (/, /:gameId, /review/:id, etc).
func SpaHandler(dist fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if info, err := fs.Stat(dist, name); err == nil && !info.IsDir() {
				http.ServeFileFS(w, r, dist, name)
				return
			}
		}
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFileFS(w, r, dist, "index.html")
	}
}
