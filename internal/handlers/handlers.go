package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"tinychess/internal/coach"
	"tinychess/internal/game"
	"tinychess/internal/logging"
	"tinychess/internal/storage"

	"github.com/corentings/chess/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler contains dependencies for HTTP handlers. DB and Coach are
// optional; when nil the corresponding endpoints fail loud (404/503) but
// the rest of the API keeps working — this lets unit tests and the e2e
// harness run without a database or LLM credentials.
type Handler struct {
	Hub   *game.Hub
	DB    *gorm.DB
	Coach coach.Provider
}

// NewHandler creates a new handler instance with no persistence layer.
// Production callers set DB after construction (or via NewHandlerWithStore).
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
	WriteJSON(w, http.StatusOK, map[string]any{"id": id})
}

// HandleNewRedirect creates a new game and redirects to it (GET /new).
func (h *Handler) HandleNewRedirect(w http.ResponseWriter, r *http.Request) {
	id := uuid.NewString()
	http.Redirect(w, r, "/"+id, http.StatusFound)
}

// HandleSSE handles Server-Sent Events for real-time game updates.
func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("gameId")
	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		clientID = uuid.NewString()
	}
	g, col := h.Hub.Get(id, clientID)

	// Best-effort persistence: ensure a Game row exists. Skip silently on error
	// so DB outages don't break gameplay.
	if err := storage.EnsureGame(h.DB, id, time.Now()); err != nil {
		logging.Debugf("EnsureGame %s: %v", id, err)
	}

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
	state := g.StateLocked()
	g.Mu.Unlock()

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
	g, _ := h.Hub.Get(id, "")

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

	if len(uci) == 4 {
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

	// Best-effort persistence — log on failure, don't break gameplay.
	if h.DB != nil {
		ply := len(state.UCI)
		if err := storage.RecordMove(h.DB, id, ply, uci, state.FEN); err != nil {
			logging.Debugf("RecordMove %s/%d: %v", id, ply, err)
		}
		if state.Status != "" {
			result := resultFromPGN(state.PGN)
			if err := storage.FinishGame(h.DB, id, result, state.PGN, state.FEN, time.Now(), ply); err != nil {
				logging.Debugf("FinishGame %s: %v", id, err)
			}
		} else {
			if err := storage.UpdateGamePosition(h.DB, id, state.FEN, state.PGN, ply); err != nil {
				logging.Debugf("UpdateGamePosition %s: %v", id, err)
			}
		}
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
	g, _ := h.Hub.Get(id, "")

	var body game.ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
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
	g, _ := h.Hub.Get(id, "")

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

	g.Mu.Lock()
	owner := g.OwnerID
	g.Mu.Unlock()
	if body.ClientID != owner {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "not owner"})
		return
	}

	g.RemoveClient(body.TargetID)
	go g.Broadcast()
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
