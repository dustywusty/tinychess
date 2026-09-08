package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"tinychess/internal/game"
	"tinychess/internal/handlers"
	"tinychess/internal/logging"
	"tinychess/internal/storage"
)

//go:embed all:web/dist
var spaFS embed.FS

func main() {
	debug := flag.Bool("debug", false, "enable debug logging")
	healthcheck := flag.Bool("healthcheck", false, "check the local HTTP health endpoint")
	flag.Parse()
	logging.Debug = *debug
	port, err := serverPort(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}
	if *healthcheck {
		client := &http.Client{Timeout: 3 * time.Second}
		res, err := client.Get("http://127.0.0.1:" + port + "/healthz")
		if err != nil {
			log.Fatal("health check failed")
		}
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}

	hub := game.NewHub()
	h := handlers.NewHandler(hub)

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := storage.New(dsn)
		if err != nil {
			log.Fatalf("failed to initialize database: %v", err)
		}
		h.DB = db
		log.Printf("persistence: enabled")
	} else {
		log.Printf("persistence: disabled (set DATABASE_URL to enable game review)")
	}

	dist, err := fs.Sub(spaFS, "web/dist")
	if err != nil {
		log.Fatalf("embed: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)

	// API
	mux.HandleFunc("POST /api/games", h.HandleCreateGame)
	mux.HandleFunc("GET /api/games/{gameId}/snapshot", h.HandleSnapshot)
	mux.HandleFunc("GET /api/sse/{gameId}", h.HandleSSE)
	mux.HandleFunc("POST /api/games/{gameId}/move", h.HandleMove)
	mux.HandleFunc("POST /api/games/{gameId}/react", h.HandleReact)
	mux.HandleFunc("POST /api/games/{gameId}/release", h.HandleRelease)
	mux.HandleFunc("GET /api/games/{gameId}", h.HandleGetGame)
	mux.HandleFunc("GET /api/games/{gameId}/evals", h.HandleGetEvals)
	mux.HandleFunc("POST /api/games/{gameId}/evals", h.HandleAppendEvals)
	mux.HandleFunc("GET /api/version", versionHandler)

	// Legacy redirect for <a href="/new">
	mux.HandleFunc("GET /new", h.HandleNewRedirect)

	// SPA + assets fallback (must be last)
	mux.HandleFunc("GET /", handlers.SpaHandler(dist))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{
		Addr: ":" + port, Handler: mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		// SSE responses remain open. Do not set a global WriteTimeout.
	}
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			// Close long-lived SSE connections after the drain period.
			_ = server.Close()
		}
		close(done)
	}()
	log.Printf("Your Move listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	<-done
}

func serverPort(value string) (string, error) {
	if value == "" {
		return "8080", nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 65535 {
		return "", errors.New("PORT must be an integer from 1 to 65535")
	}
	return strconv.Itoa(n), nil
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(`{"commit":"` + commit + `"}`))
}
