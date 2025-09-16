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

	var store *storage.Store
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := storage.New(dsn)
		if err != nil {
			log.Fatalf("failed to initialize database: %v", err)
		}
		store = storage.NewStore(db)
	}

	// Initialize game hub
	hub := game.NewHub(store)

	// Initialize HTTP handlers
	h := handlers.NewHandler(hub, store)

	// Register routes
	http.HandleFunc("/new", h.HandleNew)
	http.HandleFunc("/sse/", h.HandleSSE)
	http.HandleFunc("/move/", h.HandleMove)
	http.HandleFunc("/react/", h.HandleReact)
	http.HandleFunc("/release/", h.HandleRelease)
	http.HandleFunc("/forget/", h.HandleForget)
	http.HandleFunc("/api/stats", h.HandleStats)
	http.HandleFunc("/", h.HandlePage)

	log.Printf("Tiny Chess listening on http://localhost:8080 …")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
