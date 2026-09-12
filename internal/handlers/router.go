package handlers

import (
	"net/http"
	"os"

	"tinychess/internal/game"
)

// NewRouter serves the API. The deployment ingress must preserve /api.
func NewRouter(hub *game.Hub, version string) *http.ServeMux {
	h := NewHandler(hub)
	api := http.NewServeMux()
	api.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		WriteJSON(w, http.StatusOK, map[string]any{
			"version":      version,
			"coachEnabled": os.Getenv("OPENAI_API_KEY") != "",
		})
	})
	api.HandleFunc("/sse/", h.HandleSSE)
	api.HandleFunc("/move/", h.HandleMove)
	api.HandleFunc("/react/", h.HandleReact)
	api.HandleFunc("/release/", h.HandleRelease)
	api.HandleFunc("/coach", h.HandleCoach)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte("ok\n"))
	})
	return mux
}
