package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"tinychess/internal/game"
	"tinychess/internal/handlers"
	"tinychess/internal/logging"
	"tinychess/internal/storage"
	"tinychess/internal/templates"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()
	logging.Debug = *debug

	templates.SetVersion(commit)

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		if _, err := storage.New(dsn); err != nil {
			log.Fatalf("failed to initialize database: %v", err)
		}
	}

	hub := game.NewHub()
	h := handlers.NewHandler(hub)

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("POST /api/games", h.HandleCreateGame)
	mux.HandleFunc("GET /api/sse/{gameId}", h.HandleSSE)
	mux.HandleFunc("POST /api/games/{gameId}/move", h.HandleMove)
	mux.HandleFunc("POST /api/games/{gameId}/react", h.HandleReact)
	mux.HandleFunc("POST /api/games/{gameId}/release", h.HandleRelease)

	// Pages (legacy inline-React templates until SPA cutover)
	mux.HandleFunc("GET /new", h.HandleNewRedirect)
	mux.HandleFunc("GET /", h.HandlePage)

	log.Printf("Tiny Chess listening on http://localhost:8080 …")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
