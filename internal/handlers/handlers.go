package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/corentings/chess/v2"
	"github.com/google/uuid"

	"tinychess/internal/game"
	"tinychess/internal/logging"
	"tinychess/internal/storage"
	"tinychess/internal/templates"
)

// Handler contains dependencies for HTTP handlers.
type Handler struct {
	Hub   *game.Hub
	Store *storage.Store
}

// NewHandler creates a new handler instance.
func NewHandler(hub *game.Hub, store *storage.Store) *Handler {
	return &Handler{Hub: hub, Store: store}
}

// HandleNew creates a new game. POST requests respond with JSON, while GET
// requests redirect to the new game URL.
func (h *Handler) HandleNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		var body struct {
			UserID string `json:"userId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
			return
		}
		userID := strings.TrimSpace(body.UserID)
		if userID == "" {
			WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing user id"})
			return
		}

		id, color, err := h.Hub.CreateGame(ctx, userID)
		if err != nil {
			logging.Debugf("create game failed: %v", err)
			WriteJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "could not create game"})
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id, "color": color.String()})
	default:
		userID := strings.TrimSpace(r.URL.Query().Get("userId"))
		if userID == "" {
			http.Error(w, "missing user id", http.StatusBadRequest)
			return
		}
		id, _, err := h.Hub.CreateGame(ctx, userID)
		if err != nil {
			logging.Debugf("create game failed: %v", err)
			http.Error(w, "failed to create game", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/"+id, http.StatusFound)
	}
}

// HandlePage serves the home page or game page.
func (h *Handler) HandlePage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" || path == "index.html" {
		templates.WriteHomeHTML(w)
		return
	}
	if _, _, err := h.Hub.Get(r.Context(), path, ""); err != nil && !errors.Is(err, storage.ErrNotFound) {
		logging.Debugf("ensure game %s failed: %v", path, err)
	}
	templates.WriteGameHTML(w, path)
}

// HandleSSE handles Server-Sent Events for real-time game updates.
func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/sse/")
	clientID := strings.TrimSpace(r.URL.Query().Get("clientId"))
	if clientID == "" {
		clientID = strings.TrimSpace(r.Header.Get("X-User-ID"))
	}
	if clientID == "" {
		clientID = uuid.NewString()
	}

	g, col, err := h.Hub.Get(r.Context(), id, clientID)
	if err != nil {
		http.Error(w, "game unavailable", http.StatusInternalServerError)
		return
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

	lastSeen := g.Touch()
	if err := h.persistLastSeen(r.Context(), id, lastSeen); err != nil {
		logging.Debugf("update last seen failed: %v", err)
	}

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
	id := strings.TrimPrefix(r.URL.Path, "/move/")
	g, _, err := h.Hub.Get(r.Context(), id, "")
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "game unavailable"})
		return
	}

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

	from := uci[:2]

	g.Mu.Lock()
	state := g.StateLocked()
	playerColor, ok := g.Clients[clientID]
	isOwner := g.OwnerID == clientID
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

	lastSeen := g.Touch()

	if err := g.MakeMove(uci); err != nil {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error(), "state": state})
		return
	}

	go g.Broadcast()

	g.Mu.Lock()
	state = g.StateLocked()
	moveNumber := len(state.UCI)
	g.Mu.Unlock()

	outcome := g.Outcome()

	if err := h.persistGameState(r.Context(), id, state, outcome, lastSeen); err != nil {
		logging.Debugf("persist game state failed: %v", err)
	}
	if err := h.recordMove(r.Context(), id, clientID, moveNumber, uci, playerColor, isOwner, lastSeen); err != nil {
		logging.Debugf("record move failed: %v", err)
	}

	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "state": state})
}

// HandleReact processes a reaction/emoji.
func (h *Handler) HandleReact(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/react/")
	g, _, err := h.Hub.Get(r.Context(), id, "")
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "game unavailable"})
		return
	}

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
	id := strings.TrimPrefix(r.URL.Path, "/release/")
	g, _, err := h.Hub.Get(r.Context(), id, "")
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "game unavailable"})
		return
	}

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
	if err := h.deactivateSession(r.Context(), id, body.TargetID); err != nil {
		logging.Debugf("deactivate session failed: %v", err)
	}
	go g.Broadcast()
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// HandleForget ends a game when the owner forgets it from the home page.
func (h *Handler) HandleForget(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/forget/")
	var body struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
		return
	}
	userID := strings.TrimSpace(body.UserID)
	if userID == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing user id"})
		return
	}

	g, _, err := h.Hub.Get(r.Context(), id, userID)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "game unavailable"})
		return
	}

	g.Mu.Lock()
	owner := g.OwnerID
	g.Mu.Unlock()
	if owner != userID {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "not owner"})
		return
	}

	if err := h.markGameForgotten(r.Context(), id); err != nil {
		logging.Debugf("mark forgotten failed: %v", err)
	}

	g.Mu.Lock()
	for cid := range g.Clients {
		delete(g.Clients, cid)
	}
	g.OwnerID = ""
	g.OwnerColor = chess.NoColor
	g.Mu.Unlock()

	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// HandleStats returns aggregate statistics for display on the home page.
func (h *Handler) HandleStats(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "stats": storage.Stats{}})
		return
	}

	stats, err := h.Store.FetchStats(r.Context())
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"ok": false})
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "stats": stats})
}

func (h *Handler) persistLastSeen(ctx context.Context, id string, ts time.Time) error {
	if h.Store == nil {
		return nil
	}
	gameID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return h.Store.UpdateLastSeen(ctx, gameID, ts)
}

func (h *Handler) persistGameState(ctx context.Context, id string, state game.GameState, outcome chess.Outcome, lastSeen time.Time) error {
	if h.Store == nil {
		return nil
	}
	gameID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	fen := state.FEN
	pgn := state.PGN
	status := state.Status
	active := outcome == chess.NoOutcome
	upd := storage.GameStateUpdate{
		FEN:      &fen,
		PGN:      &pgn,
		Status:   &status,
		Active:   &active,
		LastSeen: &lastSeen,
	}
	if !active {
		result := outcome.String()
		if result != "" {
			upd.Result = &result
		}
		completedAt := lastSeen
		upd.CompletedAt = &completedAt
	}
	return h.Store.SaveGameState(ctx, gameID, upd)
}

func (h *Handler) recordMove(ctx context.Context, gameID, clientID string, number int, uci string, color chess.Color, isOwner bool, lastSeen time.Time) error {
	if h.Store == nil {
		return nil
	}
	gid, err := uuid.Parse(gameID)
	if err != nil {
		return err
	}
	uid, err := uuid.Parse(clientID)
	if err != nil {
		return err
	}
	colorStr := "white"
	if color == chess.Black {
		colorStr = "black"
	}
	if err := h.Store.RecordMove(ctx, gid, uid, number, uci, colorStr); err != nil {
		return err
	}
	role := "player"
	if isOwner {
		role = "owner"
	}
	return h.Store.EnsureUserSession(ctx, gid, uid, colorStr, role, lastSeen)
}

func (h *Handler) deactivateSession(ctx context.Context, gameID, userID string) error {
	if h.Store == nil {
		return nil
	}
	gid, err := uuid.Parse(gameID)
	if err != nil {
		return err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return h.Store.DeactivateUserSession(ctx, gid, uid)
}

func (h *Handler) markGameForgotten(ctx context.Context, id string) error {
	if h.Store == nil {
		return nil
	}
	gameID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	now := time.Now()
	if err := h.Store.ForgetGame(ctx, gameID, now); err != nil {
		return err
	}
	return h.Store.DeactivateAllSessions(ctx, gameID)
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
