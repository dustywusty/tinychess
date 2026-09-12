// Package frontend serves static files for local development and browser tests.
// Production uses the App Platform static site component.
package frontend

import (
	"io/fs"
	"net/http"
	"os"
	"strings"
)

func Handler(dir string) http.Handler {
	files := os.DirFS(dir)
	server := http.FileServer(http.FS(files))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		path := strings.Trim(r.URL.Path, "/")
		if path == "" {
			path = "."
		}
		if _, err := fs.Stat(files, path); err == nil {
			server.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, dir+"/game.html")
	})
}
