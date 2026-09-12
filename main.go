package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"tinychess/internal/frontend"
	"tinychess/internal/game"
	"tinychess/internal/handlers"
	"tinychess/internal/logging"
	"tinychess/internal/storage"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logging")
	staticDir := flag.String("static-dir", "", "serve static frontend files for local development")
	flag.Parse()
	logging.Debug = *debug

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		if _, err := storage.New(dsn); err != nil {
			log.Fatalf("failed to initialize database: %v", err)
		}
	}

	// Initialize game hub
	hub := game.NewHub()

	mux := handlers.NewRouter(hub, commit)
	if *staticDir != "" {
		mux.Handle("/", frontend.Handler(*staticDir))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Tiny Chess listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
