package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"tinychess/internal/game"
	"tinychess/internal/logging"
	"tinychess/internal/templates"

	"github.com/google/uuid"
	"github.com/notnil/chess"
)

// Handler contains dependencies for HTTP handlers
type Handler struct {
	Hub *game.Hub
}

// NewHandler creates a new handler instance
func NewHandler(hub *game.Hub) *Handler {
	return &Handler{Hub: hub}
}

// HandleNew creates a new game and redirects to it
func (h *Handler) HandleNew(w http.ResponseWriter, r *http.Request) {
	id := uuid.NewString()
	http.Redirect(w, r, "/"+id, http.StatusFound)
}

// HandlePage serves the home page or game page
func (h *Handler) HandlePage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" || path == "index.html" {
		templates.WriteHomeHTML(w)
		return
	}
	_ = h.Hub.Get(path, "")
	templates.WriteGameHTML(w, path)
}

// HandleSSE handles Server-Sent Events for real-time game updates
func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/sse/")
	clientID := r.URL.Query().Get("clientId")
	g := h.Hub.Get(id, clientID)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan []byte, 16)
	g.AddWatcher(ch)

	g.Mu.Lock()
	initial, _ := json.Marshal(g.StateLocked())
	g.Mu.Unlock()

	_, _ = fmt.Fprintf(w, "data: %s\n\n", initial)
	flusher.Flush()

	g.Touch()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	defer g.RemoveWatcher(ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// heartbeat
			_, _ = w.Write([]byte("data: {}\n\n"))
			flusher.Flush()
		case msg := <-ch:
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(msg)
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

// HandleMove processes a chess move
func (h *Handler) HandleMove(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/move/")
	g := h.Hub.Get(id, "")

	var m game.MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
		return
	}

	clientID := strings.TrimSpace(m.ClientID)
	if clientID == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing client id"})
		return
	}

	uci := strings.ToLower(strings.TrimSpace(m.UCI))
	uci = appendPromotionIfPawn(g, uci)

	// Handle castling moves - ensure they're properly formatted
	if len(uci) == 4 {
		// Check for castling moves
		if uci == "e1g1" || uci == "e1c1" || uci == "e8g8" || uci == "e8c8" {
			logging.Debugf("Castling move detected: %s", uci)
		}
	}

	from := uci[:2]

	g.Mu.Lock()
	state := g.StateLocked()
	playerColor, ok := g.Clients[clientID]
	g.Mu.Unlock()

	fenOpt, err := chess.FEN(state.FEN)
	if err != nil {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "bad fen", "state": state})
		return
	}
	tmp := chess.NewGame(fenOpt)
	board := tmp.Position().Board()
	fsq := parseSquare(from)
	piece := board.Piece(fsq)
	turn := tmp.Position().Turn()

	if !ok {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "unknown client", "state": state})
		return
	}

	if piece == chess.NoPiece || piece.Color() != playerColor {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "wrong color", "state": state})
		return
	}

	if turn != playerColor {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "not your turn", "state": state})
		return
	}

	g.Touch()

	if err := g.MakeMove(uci); err != nil {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error(), "state": state})
		return
	}

	go g.Broadcast()

	g.Mu.Lock()
	state = g.StateLocked()
	g.Mu.Unlock()

	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "state": state})
}

// HandleReact processes a reaction/emoji
func (h *Handler) HandleReact(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/react/")
	g := h.Hub.Get(id, "")

	var body game.ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
		return
	}

	if !isAllowedEmoji(body.Emoji) {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "unsupported emoji"})
		return
	}

	canReact, wait := g.CanReact(body.Sender)
	if !canReact {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": fmt.Sprintf("cooldown %ds", wait)})
		return
	}

	payload := game.ReactionPayload{
		Kind:   "emoji",
		Emoji:  body.Emoji,
		At:     time.Now().UnixMilli(),
		Sender: body.Sender,
	}

	g.BroadcastReaction(payload)
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// HandleReset resets a game to the starting position
func (h *Handler) HandleReset(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/reset/")
	g := h.Hub.Get(id, "")

	g.Reset()

	g.Mu.Lock()
	state := g.StateLocked()
	g.Mu.Unlock()

	go g.Broadcast()
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "state": state})
}

// ClientIP extracts the client IP from the request
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
