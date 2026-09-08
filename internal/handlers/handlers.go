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

	"github.com/corentings/chess/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler contains dependencies for HTTP handlers. A nil DB selects volatile
// in-memory games. With a DB, storage errors must never fall back to memory.
type Handler struct {
	Hub *game.Hub
	DB  *gorm.DB
}

// NewHandler creates a new handler instance with no persistence layer.
// Production callers use NewHandlerWithStore for durable games.
func NewHandler(hub *game.Hub) *Handler {
	return &Handler{Hub: hub}
}

// NewHandlerWithStore creates a handler with persistence enabled.
func NewHandlerWithStore(hub *game.Hub, db *gorm.DB) *Handler {
	return &Handler{Hub: hub, DB: db}
}

// HandleCreateGame creates a new game and returns its id (POST /api/games).
func (h *Handler) HandleCreateGame(w http.ResponseWriter, r *http.Request) {
	id := uuid.NewString()
	if _, err := h.operate(r.Context(), id, true, nil); err != nil {
		gameStorageError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"id": id})
}

// HandleNewRedirect creates a new game and redirects to it (GET /new).
func (h *Handler) HandleNewRedirect(w http.ResponseWriter, r *http.Request) {
	id := uuid.NewString()
	if _, err := h.operate(r.Context(), id, true, nil); err != nil {
		gameStorageError(w, err)
		return
	}
	http.Redirect(w, r, "/g/"+id, http.StatusFound)
}

// HandleSnapshot returns the current authoritative state and assigns an
// anonymous seat when one is available. Mobile uses this during the SSE to
// WebSocket migration; clients should poll sparingly.
func (h *Handler) HandleSnapshot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("gameId")
	clientID := strings.TrimSpace(r.URL.Query().Get("clientId"))
	if clientID == "" {
		clientID = uuid.NewString()
	}
	var col *chess.Color
	g, err := h.operate(r.Context(), id, false, func(candidate *game.Game) bool {
		col = candidate.AssignClient(clientID)
		return false
	})
	if err != nil {
		gameStorageError(w, err)
		return
	}

	g.Mu.Lock()
	state := g.StateLocked()
	g.Mu.Unlock()

	result := game.ClientState{GameState: state, Role: "spectator", ClientID: clientID}
	if col != nil {
		color := col.String()
		result.Color = &color
		result.Role = "player"
	}
	WriteJSON(w, http.StatusOK, result)
}

// HandleSSE handles Server-Sent Events for real-time game updates.
func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("gameId")
	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		clientID = uuid.NewString()
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	var col *chess.Color
	g, err := h.operate(r.Context(), id, false, func(candidate *game.Game) bool {
		col = candidate.AssignClient(clientID)
		return false
	})
	if err != nil {
		gameStorageError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan []byte, 16)
	g.OpMu.Lock()
	g.AddWatcher(ch)

	g.Mu.Lock()
	state := g.StateLocked()
	g.Mu.Unlock()
	g.OpMu.Unlock()

	initial := game.ClientState{GameState: state, Role: "spectator", ClientID: clientID}
	if col != nil {
		c := col.String()
		initial.Color = &c
		initial.Role = "player"
	}
	initialJSON, _ := json.Marshal(initial)

	_, _ = fmt.Fprintf(w, "data: %s\n\n", initialJSON)
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

// HandleMove processes a chess move.
func (h *Handler) HandleMove(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("gameId")

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
	if len(uci) != 4 && len(uci) != 5 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid uci"})
		return
	}
	if parseSquare(uci[:2]) == chess.NoSquare || parseSquare(uci[2:4]) == chess.NoSquare {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid uci"})
		return
	}

	if len(uci) == 4 {
		if uci == "e1g1" || uci == "e1c1" || uci == "e8g8" || uci == "e8c8" {
			logging.Debugf("Castling move detected: %s", uci)
		}
	}

	var moveErr error
	var state game.GameState
	g, err := h.operate(r.Context(), id, false, func(g *game.Game) bool {
		_, moveErr = g.MakeMoveFor(clientID, uci)
		g.Mu.Lock()
		state = g.StateLocked()
		g.Mu.Unlock()
		return moveErr == nil
	})
	if err != nil {
		gameStorageError(w, err)
		return
	}
	// Watchers belong to the live process, not the private durable candidate.
	g.Mu.Lock()
	state.Watchers = len(g.Watchers)
	g.Mu.Unlock()
	if moveErr != nil {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": moveErr.Error(), "state": state})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "state": state})
}

// resultFromPGN extracts the result token ("1-0" / "0-1" / "1/2-1/2") from a
// PGN if present. Returns "" if the PGN is missing a terminator.
func resultFromPGN(pgn string) string {
	for _, token := range []string{"1-0", "0-1", "1/2-1/2"} {
		if strings.Contains(pgn, " "+token) || strings.HasSuffix(pgn, token) {
			return token
		}
	}
	return ""
}

// HandleReact processes a reaction/emoji.
func (h *Handler) HandleReact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("gameId")

	var body game.ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
		return
	}

	g, err := h.operate(r.Context(), id, false, nil)
	if err != nil {
		gameStorageError(w, err)
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

// HandleRelease removes a client from a game if requested by the owner.
func (h *Handler) HandleRelease(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("gameId")

	var body struct {
		ClientID string `json:"clientId"`
		TargetID string `json:"targetId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
		return
	}

	if body.ClientID == "" || body.TargetID == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing client id"})
		return
	}

	allowed := false
	_, err := h.operate(r.Context(), id, false, func(g *game.Game) bool {
		g.Mu.Lock()
		allowed = body.ClientID == g.OwnerID
		g.Mu.Unlock()
		if allowed {
			g.RemoveClient(body.TargetID)
		}
		return allowed
	})
	if err != nil {
		gameStorageError(w, err)
		return
	}
	if !allowed {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "not owner"})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ClientIP extracts the client IP from the request.
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
