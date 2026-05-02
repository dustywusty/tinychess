package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"

	"tinychess/internal/game"
	"tinychess/internal/handlers"
	"tinychess/internal/logging"
	"tinychess/internal/storage"
)

//go:embed all:web/dist
var spaFS embed.FS

func main() {
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()
	logging.Debug = *debug

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		if _, err := storage.New(dsn); err != nil {
			log.Fatalf("failed to initialize database: %v", err)
		}
	}

	hub := game.NewHub()
	h := handlers.NewHandler(hub)

	dist, err := fs.Sub(spaFS, "web/dist")
	if err != nil {
		log.Fatalf("embed: %v", err)
	}

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("POST /api/games", h.HandleCreateGame)
	mux.HandleFunc("GET /api/sse/{gameId}", h.HandleSSE)
	mux.HandleFunc("POST /api/games/{gameId}/move", h.HandleMove)
	mux.HandleFunc("POST /api/games/{gameId}/react", h.HandleReact)
	mux.HandleFunc("POST /api/games/{gameId}/release", h.HandleRelease)
	mux.HandleFunc("GET /api/version", versionHandler)

	// Legacy redirect for <a href="/new">
	mux.HandleFunc("GET /new", h.HandleNewRedirect)

	// SPA + assets fallback (must be last)
	mux.HandleFunc("GET /", handlers.SpaHandler(dist))

	log.Printf("Tiny Chess listening on http://localhost:8080 …")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(`{"commit":"` + commit + `"}`))
}
