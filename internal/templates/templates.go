package templates

import (
	"html/template"
	"net/http"
	"os"
	"strings"
)

var commit = "dev"
var buildDate = ""

func SetVersion(c, d string) {
	commit = c
	buildDate = d
}

// WriteHomeHTML serves the home page template
func WriteHomeHTML(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	content, err := os.ReadFile("internal/templates/home.html")
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	html := strings.ReplaceAll(string(content), "{{COMMIT}}", commit)
	html = strings.ReplaceAll(html, "{{BUILD_DATE}}", buildDate)
	_, _ = w.Write([]byte(html))
}

// WriteGameHTML serves the game page template with game ID substitution
func WriteGameHTML(w http.ResponseWriter, gameID string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	content, err := os.ReadFile("internal/templates/game.html")
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	html := strings.ReplaceAll(string(content), "{{GAME_ID}}", gameID)
	html = strings.ReplaceAll(html, "{{COMMIT}}", commit)
	html = strings.ReplaceAll(html, "{{BUILD_DATE}}", buildDate)
	_, _ = w.Write([]byte(html))
}

// LoadTemplate loads and parses an HTML template
func LoadTemplate(name, content string) (*template.Template, error) {
	return template.New(name).Parse(content)
}
