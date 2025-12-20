package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"tinychess/internal/game"
	"tinychess/internal/logging"
	"tinychess/internal/templates"

	"github.com/corentings/chess/v2"
	"github.com/google/uuid"
)

// Handler contains dependencies for HTTP handlers
type Handler struct {
	Hub             *game.Hub
	HashbrownURL    string
	HashbrownAPIKey string
}

// NewHandler creates a new handler instance
func NewHandler(hub *game.Hub, hashbrownURL, hashbrownAPIKey string) *Handler {
	return &Handler{
		Hub:             hub,
		HashbrownURL:    hashbrownURL,
		HashbrownAPIKey: hashbrownAPIKey,
	}
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
	_, _ = h.Hub.Get(path, "")
	templates.WriteGameHTML(w, path)
}

// HandleSSE handles Server-Sent Events for real-time game updates
func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/sse/")
	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		clientID = uuid.NewString()
	}
	g, col := h.Hub.Get(id, clientID)

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
	id := strings.TrimPrefix(r.URL.Path, "/release/")
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

// HandleCoach proxies structured coach requests to Hashbrown.
func (h *Handler) HandleCoach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}

	if h.HashbrownURL == "" {
		WriteJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "coach unavailable"})
		return
	}

	var req CoachRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "bad json"})
		return
	}

	if strings.TrimSpace(req.FEN) == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing fen"})
		return
	}

	payload, err := json.Marshal(req)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "failed to encode request"})
		return
	}

	hbReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.HashbrownURL, bytes.NewReader(payload))
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "failed to build request"})
		return
	}
	hbReq.Header.Set("Content-Type", "application/json")
	if h.HashbrownAPIKey != "" {
		hbReq.Header.Set("Authorization", "Bearer "+h.HashbrownAPIKey)
	}

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(hbReq)
	if err != nil {
		logging.Debugf("hashbrown request failed: %v", err)
		WriteJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "coach request failed"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		WriteJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "failed to read coach response"})
		return
	}

	if resp.StatusCode >= http.StatusBadRequest {
		logging.Debugf("hashbrown error: status=%d body=%s", resp.StatusCode, string(body))
		WriteJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "coach error"})
		return
	}

	var hbRaw map[string]any
	if err := json.Unmarshal(body, &hbRaw); err != nil {
		WriteJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "invalid coach response"})
		return
	}

	suggestions := buildCoachSuggestions(req.FEN, hbRaw)
	explanation := extractCoachExplanation(hbRaw)

	WriteJSON(w, http.StatusOK, CoachResponse{
		OK:          true,
		Suggestions: suggestions,
		Explanation: explanation,
		Raw:         hbRaw,
	})
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

// CoachRequest is the structured payload sent to Hashbrown.
type CoachRequest struct {
	FEN         string   `json:"fen"`
	PGN         string   `json:"pgn"`
	MoveHistory []string `json:"moveHistory"`
	SideToMove  string   `json:"sideToMove"`
	Question    string   `json:"question"`
}

// CoachSuggestion is a validated coach move.
type CoachSuggestion struct {
	Move  string `json:"move"`
	Valid bool   `json:"valid"`
}

// CoachResponse is returned to the client.
type CoachResponse struct {
	OK          bool              `json:"ok"`
	Error       string            `json:"error,omitempty"`
	Suggestions []CoachSuggestion `json:"suggestions,omitempty"`
	Explanation string            `json:"explanation,omitempty"`
	Raw         map[string]any    `json:"raw,omitempty"`
}

func buildCoachSuggestions(fen string, hbRaw map[string]any) []CoachSuggestion {
	candidates := extractMoveCandidates(hbRaw)
	if len(candidates) == 0 {
		return nil
	}

	fenOpt, err := chess.FEN(fen)
	if err != nil {
		return nil
	}
	tmp := chess.NewGame(fenOpt)
	uci := chess.UCINotation{}
	san := chess.AlgebraicNotation{}

	out := make([]CoachSuggestion, 0, len(candidates))
	for _, move := range candidates {
		mv, err := uci.Decode(tmp.Position(), move)
		if err != nil {
			mv, err = san.Decode(tmp.Position(), move)
		}
		out = append(out, CoachSuggestion{
			Move:  move,
			Valid: err == nil && mv != nil,
		})
	}
	return out
}

func extractMoveCandidates(hbRaw map[string]any) []string {
	keys := []string{"moves", "suggestions", "candidateMoves", "candidates"}
	for _, key := range keys {
		if val, ok := hbRaw[key]; ok {
			return normalizeMoveList(val)
		}
	}
	return nil
}

func normalizeMoveList(val any) []string {
	switch v := val.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if strings.TrimSpace(item) != "" {
				out = append(out, item)
			}
		}
		return out
	default:
		return nil
	}
}

func extractCoachExplanation(hbRaw map[string]any) string {
	keys := []string{"explanation", "analysis", "reasoning", "summary"}
	for _, key := range keys {
		if val, ok := hbRaw[key]; ok {
			if s, ok := val.(string); ok {
				return s
			}
		}
	}
	return ""
}
