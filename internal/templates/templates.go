package templates

import (
	"html/template"
	"net/http"
	"os"
	"strings"
)

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
	_, _ = w.Write(content)
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
	_, _ = w.Write([]byte(html))
}

// LoadTemplate loads and parses an HTML template
func LoadTemplate(name, content string) (*template.Template, error) {
	return template.New(name).Parse(content)
}
