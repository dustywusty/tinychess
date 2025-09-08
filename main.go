package main

import (
	"flag"
	"log"
	"net/http"

	"tinychess/internal/game"
	"tinychess/internal/handlers"
	"tinychess/internal/logging"
	"tinychess/internal/templates"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()
	logging.Debug = *debug

	templates.SetCommit(commit)

	// Initialize game hub
	hub := game.NewHub()

	// Initialize HTTP handlers
	h := handlers.NewHandler(hub)

	// Register routes
	http.HandleFunc("/new", h.HandleNew)
	http.HandleFunc("/sse/", h.HandleSSE)
	http.HandleFunc("/move/", h.HandleMove)
	http.HandleFunc("/react/", h.HandleReact)
	http.HandleFunc("/reset/", h.HandleReset)
	http.HandleFunc("/", h.HandlePage)

	log.Printf("Tiny Chess listening on http://localhost:8080 …")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
